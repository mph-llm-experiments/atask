package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"regexp"
	"strconv"

	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/config"
	"github.com/mph-llm-experiments/atask/internal/denote"
	"github.com/mph-llm-experiments/atask/internal/task"
)

// MigrateCommand creates the migrate command
func MigrateCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "migrate",
		Usage:       "atask migrate <migration-name>",
		Description: "Run data migrations",
	}

	cmd.Subcommands = []*Command{
		migrateAcoreCommand(cfg),
		migrateProjectIDCommand(cfg),
	}

	return cmd
}

// migrateAcoreCommand migrates Denote-format files to acore ULID format
func migrateAcoreCommand(cfg *config.Config) *Command {
	fs := flag.NewFlagSet("migrate-acore", flag.ContinueOnError)
	applyMap := fs.String("apply-map", "", "Apply a migration map from another app")

	return &Command{
		Name:        "acore",
		Usage:       "atask migrate acore [--apply-map <path>]",
		Description: "Migrate files from Denote format to acore ULID format",
		Flags:       fs,
		Run: func(cmd *Command, args []string) error {
			if *applyMap != "" {
				migMap, err := acore.ReadMigrationMap(*applyMap)
				if err != nil {
					return fmt.Errorf("failed to read migration map: %w", err)
				}

				if err := acore.ApplyMappings(cfg.NotesDirectory, migMap.Mappings); err != nil {
					return fmt.Errorf("failed to apply mappings: %w", err)
				}

				if !globalFlags.Quiet {
					fmt.Printf("Applied %d mappings from %s\n", len(migMap.Mappings), migMap.App)
				}
				return nil
			}

			// Migrate tasks
			taskMap, err := acore.MigrateDirectory(cfg.NotesDirectory, "task", "atask")
			if err != nil {
				return fmt.Errorf("task migration failed: %w", err)
			}

			// Migrate projects
			projMap, err := acore.MigrateDirectory(cfg.NotesDirectory, "project", "atask")
			if err != nil {
				return fmt.Errorf("project migration failed: %w", err)
			}

			// Merge mappings
			combined := &acore.MigrationMap{
				App:        "atask",
				MigratedAt: taskMap.MigratedAt,
				Mappings:   make(map[string]string),
			}
			for k, v := range taskMap.Mappings {
				combined.Mappings[k] = v
			}
			for k, v := range projMap.Mappings {
				combined.Mappings[k] = v
			}

			if len(combined.Mappings) == 0 {
				if !globalFlags.Quiet {
					fmt.Println("No files to migrate.")
				}
				return nil
			}

			// Initialize the index counter from migrated files
			migrateStore := acore.NewLocalStore(cfg.NotesDirectory)
			counter, err := acore.NewIndexCounter(migrateStore, "atask")
			if err != nil {
				return fmt.Errorf("failed to create counter: %w", err)
			}
			readIndexID := func(name string) (int, error) {
				var entity struct {
					acore.Entity `yaml:",inline"`
				}
				if _, err := acore.ReadFile(migrateStore, name, &entity); err != nil {
					return 0, err
				}
				return entity.IndexID, nil
			}
			// Init from both task and project files
			if err := counter.InitFromFiles("task", readIndexID); err != nil {
				return fmt.Errorf("counter init from tasks: %w", err)
			}
			if err := counter.InitFromFiles("project", readIndexID); err != nil {
				return fmt.Errorf("counter init from projects: %w", err)
			}

			mapPath := cfg.NotesDirectory + "/migration-map.json"
			if err := acore.WriteMigrationMap(mapPath, combined); err != nil {
				return fmt.Errorf("failed to write migration map: %w", err)
			}

			if globalFlags.JSON {
				data, _ := json.MarshalIndent(combined, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Migrated %d files (%d tasks, %d projects). Mapping saved to %s\n",
					len(combined.Mappings), len(taskMap.Mappings), len(projMap.Mappings), mapPath)
				fmt.Println("Run 'apeople migrate --apply-map " + mapPath + "' and 'anote migrate --apply-map " + mapPath + "' to update cross-references.")
			}
			return nil
		},
	}
}

// migrateProjectIDCommand migrates project_id from Denote ID to index_id
func migrateProjectIDCommand(cfg *config.Config) *Command {
	var dryRun bool

	cmd := &Command{
		Name:        "project-id-to-index",
		Usage:       "atask migrate project-id-to-index [--dry-run]",
		Description: "Migrate task project_id from Denote timestamp ID to sequential index_id",
	}

	cmd.Flags = flag.NewFlagSet("migrate-project-id", flag.ExitOnError)
	cmd.Flags.BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")

	cmd.Run = func(c *Command, args []string) error {
		scanner := denote.NewScanner(cfg.NotesDirectory)

		// Build map of Denote ID -> index_id for all projects
		projects, err := scanner.FindProjects()
		if err != nil {
			return fmt.Errorf("failed to find projects: %v", err)
		}

		denoteToIndex := make(map[string]string) // Denote ID -> index_id string
		for _, p := range projects {
			denoteToIndex[p.ID] = strconv.Itoa(p.IndexID)
		}

		if !globalFlags.Quiet {
			fmt.Printf("Found %d projects\n", len(projects))
			for _, p := range projects {
				fmt.Printf("  %s -> %d (%s)\n", p.ID, p.IndexID, p.Title)
			}
			fmt.Println()
		}

		// Denote timestamp pattern: YYYYMMDDTHHMMSS (e.g., 20260217T181159)
		denotePattern := regexp.MustCompile(`^\d{8}T\d{6}$`)

		// Find all tasks and check their project_id
		tasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		migrated := 0
		skipped := 0
		alreadyDone := 0

		for _, t := range tasks {
			pid := t.TaskMetadata.ProjectID
			if pid == "" {
				continue
			}

			// Check if it looks like a Denote timestamp
			if !denotePattern.MatchString(pid) {
				// Already an index_id or something else
				alreadyDone++
				continue
			}

			// Look up the corresponding index_id
			indexIDStr, ok := denoteToIndex[pid]
			if !ok {
				if !globalFlags.Quiet {
					fmt.Printf("  WARN: Task %d (%s) has project_id %s which doesn't match any project\n",
						t.IndexID, t.Title, pid)
				}
				skipped++
				continue
			}

			if dryRun {
				fmt.Printf("  Would migrate task %d (%s): project_id %s -> %s\n",
					t.IndexID, t.Title, pid, indexIDStr)
				migrated++
				continue
			}

			// Update the task
			t.TaskMetadata.ProjectID = indexIDStr
			if err := task.UpdateTaskFile(t.FilePath, t); err != nil {
				fmt.Printf("  ERROR: Failed to update task %d: %v\n", t.IndexID, err)
				skipped++
				continue
			}

			if !globalFlags.Quiet {
				fmt.Printf("  Migrated task %d (%s): project_id %s -> %s\n",
					t.IndexID, t.Title, pid, indexIDStr)
			}
			migrated++
		}

		if dryRun {
			fmt.Printf("\nDry run: would migrate %d task(s), %d already done, %d skipped\n",
				migrated, alreadyDone, skipped)
		} else {
			fmt.Printf("\nMigrated %d task(s), %d already done, %d skipped\n",
				migrated, alreadyDone, skipped)
		}

		return nil
	}

	return cmd
}
