package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/pdxmph/denote-tasks/internal/config"
	"github.com/pdxmph/denote-tasks/internal/denote"
	"github.com/pdxmph/denote-tasks/internal/task"
)

// ProjectCommand creates the project command with all subcommands
func ProjectCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "project",
		Usage:       "denote-tasks project <command> [options]",
		Description: "Manage projects",
	}

	cmd.Subcommands = []*Command{
		projectNewCommand(cfg),
		projectListCommand(cfg),
		projectTasksCommand(cfg),
		projectUpdateCommand(cfg),
	}

	return cmd
}

// projectNewCommand creates a new project
func projectNewCommand(cfg *config.Config) *Command {
	var (
		priority  string
		due       string
		area      string
		startDate string
		tags      string
	)

	cmd := &Command{
		Name:        "new",
		Usage:       "denote-tasks project new <title> [options]",
		Description: "Create a new project",
		Flags:       flag.NewFlagSet("project-new", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&priority, "p", "", "Priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Due date (YYYY-MM-DD or natural language)")
	cmd.Flags.StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD or natural language)")
	cmd.Flags.StringVar(&area, "area", "", "Project area")
	cmd.Flags.StringVar(&tags, "tags", "", "Comma-separated tags")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("title required")
		}

		title := strings.Join(args, " ")

		// Parse tags
		var tagList []string
		if tags != "" {
			tagList = strings.Split(tags, ",")
			for i := range tagList {
				tagList[i] = strings.TrimSpace(tagList[i])
			}
		}

		// Create the project
		projectFile, err := task.CreateProject(cfg.NotesDirectory, title, "", tagList)
		if err != nil {
			return fmt.Errorf("failed to create project: %v", err)
		}

		// Update metadata if provided
		needsUpdate := false
		if priority != "" {
			projectFile.ProjectMetadata.Priority = priority
			needsUpdate = true
		}
		if due != "" {
			parsed, err := denote.ParseNaturalDate(due)
			if err != nil {
				return fmt.Errorf("invalid due date: %v", err)
			}
			projectFile.ProjectMetadata.DueDate = parsed
			needsUpdate = true
		}
		if startDate != "" {
			parsed, err := denote.ParseNaturalDate(startDate)
			if err != nil {
				return fmt.Errorf("invalid start date: %v", err)
			}
			projectFile.ProjectMetadata.StartDate = parsed
			needsUpdate = true
		}
		if area != "" {
			projectFile.ProjectMetadata.Area = area
			needsUpdate = true
		}

		// Write back if we have updates
		if needsUpdate {
			if err := updateProjectFile(projectFile.Path, projectFile.ProjectMetadata); err != nil {
				return fmt.Errorf("failed to update project metadata: %v", err)
			}
		}

		if !globalFlags.Quiet {
			fmt.Printf("Created project: %s (ID: %s)\n", projectFile.Path, projectFile.ID)
		}

		// Launch TUI if requested
		if globalFlags.TUI {
			// TODO: Launch TUI in project view for this project
			return fmt.Errorf("TUI integration not yet implemented")
		}

		return nil
	}

	return cmd
}

// projectListCommand lists projects
func projectListCommand(cfg *config.Config) *Command {
	var (
		all      bool
		area     string
		status   string
		priority string
		sortBy   string
		reverse  bool
		search   string
	)

	cmd := &Command{
		Name:        "list",
		Usage:       "denote-tasks project list [options]",
		Description: "List projects",
		Flags:       flag.NewFlagSet("project-list", flag.ExitOnError),
	}

	cmd.Flags.BoolVar(&all, "all", false, "Show all projects (default: active only)")
	cmd.Flags.StringVar(&area, "area", "", "Filter by area")
	cmd.Flags.StringVar(&status, "status", "", "Filter by status")
	cmd.Flags.StringVar(&priority, "p", "", "Filter by priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Filter by priority (p1, p2, p3)")
	cmd.Flags.StringVar(&sortBy, "sort", "modified", "Sort by: modified, priority, due, created")
	cmd.Flags.BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags.StringVar(&search, "search", "", "Search in project content (full-text)")

	// Convenience flags
	cmd.Flags.BoolVar(&all, "a", false, "Show all projects (short)")
	cmd.Flags.StringVar(&sortBy, "s", "modified", "Sort by (short)")
	cmd.Flags.BoolVar(&reverse, "r", false, "Reverse sort (short)")

	cmd.Run = func(c *Command, args []string) error {
		// Launch TUI if requested
		if globalFlags.TUI {
			// TODO: Launch TUI with these filters applied
			return fmt.Errorf("TUI integration not yet implemented")
		}

		// Get all projects
		scanner := denote.NewScanner(cfg.NotesDirectory)
		projects, err := scanner.FindProjects()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Apply filters
		var filtered []*denote.Project
		for _, p := range projects {
			// Status filter
			if !all && status == "" && p.ProjectMetadata.Status != denote.ProjectStatusActive {
				continue
			}
			if status != "" && p.ProjectMetadata.Status != status {
				continue
			}

			// Area filter
			filterArea := area
			if filterArea == "" {
				filterArea = globalFlags.Area
			}
			if filterArea != "" && p.ProjectMetadata.Area != filterArea {
				continue
			}

			// Priority filter
			if priority != "" && p.ProjectMetadata.Priority != priority {
				continue
			}


		// Content search
		if search != "" {
			if !strings.Contains(strings.ToLower(p.Content), strings.ToLower(search)) {
				continue
			}
		}
			filtered = append(filtered, p)
		}

		// Sort projects
		sortProjects(filtered, sortBy, reverse)

		// Count tasks per project (needed for both JSON and text output)
		tasks, _ := scanner.FindTasks()
		taskCounts := make(map[string]int)
		for _, t := range tasks {
			if t.TaskMetadata.ProjectID != "" {
				taskCounts[t.TaskMetadata.ProjectID]++
			}
		}

		// Display projects
		if globalFlags.JSON {
			// Create JSON output structure
			type ProjectJSON struct {
				denote.Project
				TaskCount int `json:"task_count"`
			}

			type Output struct {
				Projects []ProjectJSON `json:"projects"`
				Count    int           `json:"count"`
			}

			// Build JSON output with task counts
			jsonProjects := make([]ProjectJSON, len(filtered))
			for i, p := range filtered {
				jsonProjects[i] = ProjectJSON{
					Project:   *p,
					TaskCount: taskCounts[p.File.ID],
				}
			}

			output := Output{
				Projects: jsonProjects,
				Count:    len(filtered),
			}

			// Marshal and print
			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		// Color setup
		if globalFlags.NoColor || color.NoColor {
			color.NoColor = true
		}

		// Status colors
		completedColor := color.New(color.FgGreen)
		pausedColor := color.New(color.FgYellow)
		cancelledColor := color.New(color.FgRed, color.Faint)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		// Display header
		if !globalFlags.Quiet {
			fmt.Printf("Projects (%d):\n\n", len(filtered))
		}

		// Display projects
		for _, p := range filtered {
			// Status icon
			status := "◆"
			switch p.ProjectMetadata.Status {
			case denote.ProjectStatusCompleted:
				status = "✓"
			case denote.ProjectStatusPaused:
				status = "⏸"
			case denote.ProjectStatusCancelled:
				status = "⨯"
			}

			// Priority with padding
			priority := "    " // 4 spaces for alignment
			if p.ProjectMetadata.Priority != "" {
				pStr := fmt.Sprintf("[%s]", p.ProjectMetadata.Priority)
				switch p.ProjectMetadata.Priority {
				case "p1":
					priority = priorityHighColor.Sprint(pStr)
				case "p2":
					priority = priorityMedColor.Sprint(pStr)
				default:
					priority = pStr
				}
			}

			// Due date with fixed width
			due := "            " // 12 spaces for alignment
			if p.ProjectMetadata.DueDate != "" {
				dueStr := fmt.Sprintf("[%s]", p.ProjectMetadata.DueDate)
				if denote.IsOverdue(p.ProjectMetadata.DueDate) && p.ProjectMetadata.Status == denote.ProjectStatusActive {
					due = color.New(color.FgRed, color.Bold).Sprint(dueStr)
				} else {
					due = dueStr
				}
			}

			// Title - truncate to 40 chars
			title := p.ProjectMetadata.Title
			if title == "" {
				title = p.File.Title
			}
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			// Area - truncate to 10 chars
			area := ""
			if p.ProjectMetadata.Area != "" {
				area = p.ProjectMetadata.Area
				if len(area) > 10 {
					area = area[:7] + "..."
				}
			}

			// Task count
			taskCount := taskCounts[p.File.ID]
			taskStr := fmt.Sprintf("(%d tasks)", taskCount)

			// Build the line with fixed-width columns
			line := fmt.Sprintf("%3d %s %s %s  %-40s %-10s %s",
				p.ProjectMetadata.IndexID,
				status,
				priority,
				due,
				title,
				area,
				taskStr,
			)

			// Apply line coloring for different statuses
			switch p.ProjectMetadata.Status {
			case denote.ProjectStatusCompleted:
				fmt.Println(completedColor.Sprint(line))
			case denote.ProjectStatusPaused:
				fmt.Println(pausedColor.Sprint(line))
			case denote.ProjectStatusCancelled:
				fmt.Println(cancelledColor.Sprint(line))
			default:
				fmt.Println(line)
			}
		}

		return nil
	}

	return cmd
}

// projectTasksCommand shows tasks for a specific project
func projectTasksCommand(cfg *config.Config) *Command {
	var (
		all    bool
		status string
		sortBy string
	)

	cmd := &Command{
		Name:        "tasks",
		Usage:       "denote-tasks project tasks <project-id> [options]",
		Description: "Show tasks for a specific project",
		Flags:       flag.NewFlagSet("project-tasks", flag.ExitOnError),
	}

	cmd.Flags.BoolVar(&all, "all", false, "Show all tasks (default: open only)")
	cmd.Flags.StringVar(&status, "status", "", "Filter by task status")
	cmd.Flags.StringVar(&sortBy, "sort", "priority", "Sort by: priority, due, created")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("project ID required")
		}

		// Parse project ID (can be numeric index or Denote ID)
		projectIdentifier := args[0]
		
		// Find the project
		scanner := denote.NewScanner(cfg.NotesDirectory)
		var targetProject *denote.Project
		
		// Try to parse as numeric ID first
		if projectNum, err := strconv.Atoi(projectIdentifier); err == nil {
			targetProject, err = task.FindProjectByID(cfg.NotesDirectory, projectNum)
			if err != nil {
				return fmt.Errorf("project with ID %d not found", projectNum)
			}
		} else {
			// Try as Denote ID
			targetProject, err = task.FindProjectByDenoteID(cfg.NotesDirectory, projectIdentifier)
			if err != nil {
				return fmt.Errorf("project with Denote ID %s not found", projectIdentifier)
			}
		}

		// Get all tasks for this project
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		// Filter tasks by project
		var projectTasks []*denote.Task
		for _, t := range allTasks {
			if t.TaskMetadata.ProjectID == targetProject.File.ID {
				// Apply status filter
				if !all && status == "" && t.TaskMetadata.Status != denote.TaskStatusOpen {
					continue
				}
				if status != "" && t.TaskMetadata.Status != status {
					continue
				}
				projectTasks = append(projectTasks, t)
			}
		}

		// Sort tasks
		sortProjectTasks(projectTasks, sortBy, false)

		// JSON output
		if globalFlags.JSON {
			type Output struct {
				Project *denote.Project `json:"project"`
				Tasks   []*denote.Task  `json:"tasks"`
				Count   int             `json:"task_count"`
			}

			output := Output{
				Project: targetProject,
				Tasks:   projectTasks,
				Count:   len(projectTasks),
			}

			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		// Display project header
		fmt.Printf("Project: %s\n", targetProject.ProjectMetadata.Title)
		if targetProject.ProjectMetadata.Status != denote.ProjectStatusActive {
			fmt.Printf("Status: %s\n", targetProject.ProjectMetadata.Status)
		}
		if targetProject.ProjectMetadata.DueDate != "" {
			fmt.Printf("Due: %s", targetProject.ProjectMetadata.DueDate)
			if denote.IsOverdue(targetProject.ProjectMetadata.DueDate) {
				fmt.Printf(" (OVERDUE)")
			}
			fmt.Println()
		}
		fmt.Printf("\nTasks (%d):\n\n", len(projectTasks))

		// Display tasks
		if len(projectTasks) == 0 {
			fmt.Println("No tasks found for this project")
			return nil
		}

		// Color setup
		if globalFlags.NoColor || color.NoColor {
			color.NoColor = true
		}

		doneColor := color.New(color.FgGreen)
		overdueColor := color.New(color.FgRed, color.Bold)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		for _, t := range projectTasks {
			// Status icon
			statusIcon := "○"
			switch t.TaskMetadata.Status {
			case denote.TaskStatusDone:
				statusIcon = "✓"
			case denote.TaskStatusPaused:
				statusIcon = "⏸"
			case denote.TaskStatusDelegated:
				statusIcon = "→"
			case denote.TaskStatusDropped:
				statusIcon = "⨯"
			}

			// Priority
			priority := "    "
			if t.TaskMetadata.Priority != "" {
				pStr := fmt.Sprintf("[%s]", t.TaskMetadata.Priority)
				switch t.TaskMetadata.Priority {
				case "p1":
					priority = priorityHighColor.Sprint(pStr)
				case "p2":
					priority = priorityMedColor.Sprint(pStr)
				default:
					priority = pStr
				}
			}

			// Due date
			due := "            "
			if t.TaskMetadata.DueDate != "" {
				dueStr := fmt.Sprintf("[%s]", t.TaskMetadata.DueDate)
				if denote.IsOverdue(t.TaskMetadata.DueDate) {
					due = overdueColor.Sprint(dueStr)
				} else {
					due = dueStr
				}
			}

			// Title
			title := t.TaskMetadata.Title
			if title == "" {
				title = t.File.Title
			}
			if len(title) > 60 {
				title = title[:57] + "..."
			}

			// Build line
			line := fmt.Sprintf("%3d %s %s %s  %s",
				t.TaskMetadata.IndexID,
				statusIcon,
				priority,
				due,
				title,
			)

			// Apply coloring for done tasks
			if t.TaskMetadata.Status == denote.TaskStatusDone {
				fmt.Println(doneColor.Sprint(line))
			} else {
				fmt.Println(line)
			}
		}

		return nil
	}

	return cmd
}

// projectUpdateCommand updates project metadata
func projectUpdateCommand(cfg *config.Config) *Command {
	var (
		priority  string
		due       string
		area      string
		status    string
		startDate string
	)

	cmd := &Command{
		Name:        "update",
		Usage:       "denote-tasks project update [options] <project-ids>",
		Description: "Update project metadata",
		Flags:       flag.NewFlagSet("project-update", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&priority, "p", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Set due date")
	cmd.Flags.StringVar(&startDate, "start", "", "Set start date")
	cmd.Flags.StringVar(&area, "area", "", "Set area")
	cmd.Flags.StringVar(&status, "status", "", "Set status (active, completed, paused, cancelled)")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("project IDs required")
		}

		// Parse project IDs (support same format as tasks)
		numbers, err := parseTaskIDs(args) // Reuse the same ID parsing logic
		if err != nil {
			return err
		}

		// Get all projects
		scanner := denote.NewScanner(cfg.NotesDirectory)
		projects, err := scanner.FindProjects()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Build index of projects by index_id
		projectsByID := make(map[int]*denote.Project)
		for _, p := range projects {
			projectsByID[p.ProjectMetadata.IndexID] = p
		}

		// Update each project
		updated := 0
		for _, id := range numbers {
			p, ok := projectsByID[id]
			if !ok {
				fmt.Fprintf(os.Stderr, "Project with ID %d not found\n", id)
				continue
			}

			// Apply updates
			changed := false
			if priority != "" {
				p.ProjectMetadata.Priority = priority
				changed = true
			}
			if due != "" {
				parsedDue, err := denote.ParseNaturalDate(due)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid due date for project ID %d: %v\n", id, err)
					continue
				}
				p.ProjectMetadata.DueDate = parsedDue
				changed = true
			}
			if startDate != "" {
				parsedStart, err := denote.ParseNaturalDate(startDate)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid start date for project ID %d: %v\n", id, err)
					continue
				}
				p.ProjectMetadata.StartDate = parsedStart
				changed = true
			}
			if area != "" {
				p.ProjectMetadata.Area = area
				changed = true
			}
			if status != "" {
				if !denote.IsValidProjectStatus(status) {
					fmt.Fprintf(os.Stderr, "Invalid status for project ID %d: %s\n", id, status)
					continue
				}
				p.ProjectMetadata.Status = status
				changed = true
			}

			if changed {
				if err := updateProjectFile(p.File.Path, p.ProjectMetadata); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update project ID %d: %v\n", id, err)
					continue
				}
				updated++
				if !globalFlags.Quiet {
					fmt.Printf("Updated project ID %d: %s\n", id, p.ProjectMetadata.Title)
				}
			}
		}

		if updated == 0 && !globalFlags.Quiet {
			fmt.Println("No projects updated")
		}

		return nil
	}

	return cmd
}

// sortProjects sorts projects by the specified field
func sortProjects(projects []*denote.Project, sortBy string, reverse bool) {
	sort.Slice(projects, func(i, j int) bool {
		var less bool
		
		switch sortBy {
		case "priority":
			// Sort by priority (p1 < p2 < p3 < "")
			pi := priorityValue(projects[i].ProjectMetadata.Priority)
			pj := priorityValue(projects[j].ProjectMetadata.Priority)
			less = pi < pj
			
		case "due":
			// Sort by due date (earliest first, empty last)
			di := projects[i].ProjectMetadata.DueDate
			dj := projects[j].ProjectMetadata.DueDate
			if di == "" && dj == "" {
				less = false
			} else if di == "" {
				less = false
			} else if dj == "" {
				less = true
			} else {
				less = di < dj
			}
			
		case "created":
			less = projects[i].File.ID < projects[j].File.ID
			
		case "modified":
			fallthrough
		default:
			less = projects[i].ModTime.After(projects[j].ModTime)
		}
		
		if reverse {
			return !less
		}
		return less
	})
}

// sortProjectTasks sorts tasks by the specified field
func sortProjectTasks(tasks []*denote.Task, sortBy string, reverse bool) {
	sort.Slice(tasks, func(i, j int) bool {
		var less bool
		
		switch sortBy {
		case "priority":
			// Sort by priority (p1 < p2 < p3 < "")
			pi := priorityValue(tasks[i].TaskMetadata.Priority)
			pj := priorityValue(tasks[j].TaskMetadata.Priority)
			less = pi < pj
			
		case "due":
			// Sort by due date (earliest first, empty last)
			di := tasks[i].TaskMetadata.DueDate
			dj := tasks[j].TaskMetadata.DueDate
			if di == "" && dj == "" {
				less = false
			} else if di == "" {
				less = false
			} else if dj == "" {
				less = true
			} else {
				less = di < dj
			}
			
		case "created":
			less = tasks[i].File.ID < tasks[j].File.ID
			
		default:
			less = tasks[i].ModTime.After(tasks[j].ModTime)
		}
		
		if reverse {
			return !less
		}
		return less
	})
}

// updateProjectFile updates the project metadata in a file
func updateProjectFile(path string, metadata denote.ProjectMetadata) error {
	// Read the current file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse existing frontmatter to get the content
	fm, err := denote.ParseFrontmatterFile(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Write updated content
	newContent, err := denote.WriteFrontmatterFile(metadata, fm.Content)
	if err != nil {
		return fmt.Errorf("failed to write frontmatter: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}