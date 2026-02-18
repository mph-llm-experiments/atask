package cli

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"

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
		migrateProjectIDCommand(cfg),
	}

	return cmd
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
			denoteToIndex[p.File.ID] = strconv.Itoa(p.IndexID)
		}

		if !globalFlags.Quiet {
			fmt.Printf("Found %d projects\n", len(projects))
			for _, p := range projects {
				fmt.Printf("  %s -> %d (%s)\n", p.File.ID, p.IndexID, p.ProjectMetadata.Title)
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
						t.TaskMetadata.IndexID, t.TaskMetadata.Title, pid)
				}
				skipped++
				continue
			}

			if dryRun {
				fmt.Printf("  Would migrate task %d (%s): project_id %s -> %s\n",
					t.TaskMetadata.IndexID, t.TaskMetadata.Title, pid, indexIDStr)
				migrated++
				continue
			}

			// Update the task
			t.TaskMetadata.ProjectID = indexIDStr
			if err := task.UpdateTaskFile(t.File.Path, t.TaskMetadata); err != nil {
				fmt.Printf("  ERROR: Failed to update task %d: %v\n", t.TaskMetadata.IndexID, err)
				skipped++
				continue
			}

			if !globalFlags.Quiet {
				fmt.Printf("  Migrated task %d (%s): project_id %s -> %s\n",
					t.TaskMetadata.IndexID, t.TaskMetadata.Title, pid, indexIDStr)
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
