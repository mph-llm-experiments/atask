package cli

import (
	"flag"
	"fmt"
	"sort"

	"github.com/pdxmph/atask/internal/config"
	"github.com/pdxmph/atask/internal/denote"
)

// CompletionCommand returns the completion command
func CompletionCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "completion",
		Usage:       "completion <type>",
		Description: "Output completion data for shell scripts",
		Flags:       flag.NewFlagSet("completion", flag.ContinueOnError),
		Run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("completion type required: task-ids, project-ids, areas, tags")
			}

			scanner := denote.NewScanner(cfg.NotesDirectory)
			files, err := scanner.FindAllTaskAndProjectFiles()
			if err != nil {
				return fmt.Errorf("failed to scan directory: %v", err)
			}

			switch args[0] {
			case "task-ids":
				return outputTaskIDs(files)
			case "project-ids":
				return outputProjectIDs(files)
			case "areas":
				return outputAreas(files)
			case "tags":
				return outputTags(files)
			default:
				return fmt.Errorf("unknown completion type: %s", args[0])
			}
		},
	}

	return cmd
}

func outputTaskIDs(files []denote.File) error {
	var ids []int
	seen := make(map[int]bool)

	for _, file := range files {
		if file.IsTask() {
			task, err := denote.ParseTaskFile(file.Path)
			if err == nil && task.TaskMetadata.IndexID > 0 {
				if !seen[task.TaskMetadata.IndexID] {
					ids = append(ids, task.TaskMetadata.IndexID)
					seen[task.TaskMetadata.IndexID] = true
				}
			}
		}
	}

	sort.Ints(ids)
	for _, id := range ids {
		fmt.Println(id)
	}
	return nil
}

func outputProjectIDs(files []denote.File) error {
	projects := make(map[string]string) // ID -> Title

	for _, file := range files {
		if file.IsProject() {
			project, err := denote.ParseProjectFile(file.Path)
			if err == nil {
				title := project.ProjectMetadata.Title
				if title == "" {
					title = file.Title
				}
				projects[file.ID] = title
			}
		}
	}

	// Output as "ID:Title" for richer completion
	var items []string
	for id, title := range projects {
		items = append(items, fmt.Sprintf("%s:%s", id, title))
	}
	sort.Strings(items)
	
	for _, item := range items {
		fmt.Println(item)
	}
	return nil
}

func outputAreas(files []denote.File) error {
	areas := make(map[string]bool)

	for _, file := range files {
		if file.IsTask() {
			task, err := denote.ParseTaskFile(file.Path)
			if err == nil && task.TaskMetadata.Area != "" {
				areas[task.TaskMetadata.Area] = true
			}
		} else if file.IsProject() {
			project, err := denote.ParseProjectFile(file.Path)
			if err == nil && project.ProjectMetadata.Area != "" {
				areas[project.ProjectMetadata.Area] = true
			}
		}
	}

	// Sort and output
	var areaList []string
	for area := range areas {
		areaList = append(areaList, area)
	}
	sort.Strings(areaList)
	
	for _, area := range areaList {
		fmt.Println(area)
	}
	return nil
}

func outputTags(files []denote.File) error {
	tags := make(map[string]bool)

	for _, file := range files {
		for _, tag := range file.Tags {
			// Skip special tags
			if tag != "task" && tag != "project" {
				tags[tag] = true
			}
		}
	}

	// Sort and output
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	sort.Strings(tagList)
	
	for _, tag := range tagList {
		fmt.Println(tag)
	}
	return nil
}