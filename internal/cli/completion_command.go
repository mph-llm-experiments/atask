package cli

import (
	"flag"
	"fmt"
	"sort"
	"strconv"

	"github.com/mph-llm-experiments/atask/internal/config"
	"github.com/mph-llm-experiments/atask/internal/denote"
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
			if err == nil && task.IndexID > 0 {
				if !seen[task.IndexID] {
					ids = append(ids, task.IndexID)
					seen[task.IndexID] = true
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
	type projectInfo struct {
		indexID int
		title   string
	}
	var projects []projectInfo

	for _, file := range files {
		if file.IsProject() {
			project, err := denote.ParseProjectFile(file.Path)
			if err == nil {
				title := project.Title
				if title == "" {
					title = file.Title
				}
				projects = append(projects, projectInfo{indexID: project.IndexID, title: title})
			}
		}
	}

	// Sort by index_id
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].indexID < projects[j].indexID
	})

	// Output as "index_id:Title" for richer completion
	for _, p := range projects {
		fmt.Printf("%s:%s\n", strconv.Itoa(p.indexID), p.title)
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
