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
	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/config"
	"github.com/mph-llm-experiments/atask/internal/denote"
	"github.com/mph-llm-experiments/atask/internal/task"
)

// ProjectCommand creates the project command with all subcommands
func ProjectCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "project",
		Usage:       "atask project <command> [options]",
		Description: "Manage projects",
	}

	cmd.Subcommands = []*Command{
		projectNewCommand(cfg),
		projectListCommand(cfg),
		projectShowCommand(cfg),
		projectTasksCommand(cfg),
		projectUpdateCommand(cfg),
		projectLogCommand(cfg),
	}

	return cmd
}

// lookupProject finds a project by integer index_id or ULID string.
func lookupProject(dir string, identifier string) (*denote.Project, error) {
	if num, err := strconv.Atoi(identifier); err == nil {
		return task.FindProjectByID(dir, num)
	}
	return task.FindProjectByEntityID(dir, identifier)
}

// projectShowCommand shows details for a single project
func projectShowCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "show",
		Usage:       "atask project show <id>",
		Description: "Show project details by index_id or ULID",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask project show <id>")
			}

			p, err := lookupProject(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if globalFlags.JSON {
				type jsonProject struct {
					*denote.Project
					Content string `json:"content,omitempty"`
				}
				jp := jsonProject{Project: p, Content: p.Content}
				data, err := json.MarshalIndent(jp, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Text output
			fmt.Printf("# %s (#%d)\n\n", p.Title, p.IndexID)

			fmt.Printf("  ID:       %s\n", p.ID)
			fmt.Printf("  Status:   %s\n", p.ProjectMetadata.Status)
			if p.ProjectMetadata.Priority != "" {
				fmt.Printf("  Priority: %s\n", p.ProjectMetadata.Priority)
			}
			if p.ProjectMetadata.DueDate != "" {
				dueStr := p.ProjectMetadata.DueDate
				if denote.IsOverdue(p.ProjectMetadata.DueDate) && p.ProjectMetadata.Status == denote.ProjectStatusActive {
					dueStr += " (OVERDUE)"
				}
				fmt.Printf("  Due:      %s\n", dueStr)
			}
			if p.ProjectMetadata.StartDate != "" {
				fmt.Printf("  Start:    %s\n", p.ProjectMetadata.StartDate)
			}
			if p.ProjectMetadata.Area != "" {
				fmt.Printf("  Area:     %s\n", p.ProjectMetadata.Area)
			}
			fmt.Println()

			if p.Created != "" {
				fmt.Printf("  Created:  %s\n", p.Created)
			}
			if p.Modified != "" {
				fmt.Printf("  Modified: %s\n", p.Modified)
			}

			var tagStrs []string
			for _, tag := range p.Tags {
				if tag != "project" {
					tagStrs = append(tagStrs, "#"+tag)
				}
			}
			if len(tagStrs) > 0 {
				fmt.Printf("\n  Tags: %s\n", strings.Join(tagStrs, " "))
			}

			if len(p.RelatedPeople) > 0 || len(p.RelatedTasks) > 0 || len(p.RelatedIdeas) > 0 {
				fmt.Println()
				if len(p.RelatedPeople) > 0 {
					fmt.Printf("  Related people: %s\n", strings.Join(p.RelatedPeople, ", "))
				}
				if len(p.RelatedTasks) > 0 {
					fmt.Printf("  Related tasks:  %s\n", strings.Join(p.RelatedTasks, ", "))
				}
				if len(p.RelatedIdeas) > 0 {
					fmt.Printf("  Related ideas:  %s\n", strings.Join(p.RelatedIdeas, ", "))
				}
			}

			if strings.TrimSpace(p.Content) != "" {
				fmt.Printf("\n---\n%s", p.Content)
			}

			return nil
		},
	}
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
		Usage:       "atask project new <title> [options]",
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
			if err := denote.UpdateProjectFile(projectFile.FilePath, projectFile); err != nil {
				return fmt.Errorf("failed to update project metadata: %v", err)
			}
		}

		if !globalFlags.Quiet {
			fmt.Printf("Created project: %s (ID: %d)\n", projectFile.FilePath, projectFile.IndexID)
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
		Usage:       "atask project list [options]",
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
					TaskCount: taskCounts[strconv.Itoa(p.IndexID)],
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
			title := p.Title
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
			taskCount := taskCounts[strconv.Itoa(p.IndexID)]
			taskStr := fmt.Sprintf("(%d tasks)", taskCount)

			// Build the line with fixed-width columns
			line := fmt.Sprintf("%3d %s %s %s  %-40s %-10s %s",
				p.IndexID,
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
		Usage:       "atask project tasks <project-id> [options]",
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

		// Parse project ID (can be numeric index or ULID)
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
			// Try as ULID
			targetProject, err = task.FindProjectByEntityID(cfg.NotesDirectory, projectIdentifier)
			if err != nil {
				return fmt.Errorf("project with ID %s not found", projectIdentifier)
			}
		}

		// Get all tasks for this project
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		// Filter tasks by project (using index_id)
		projectIDStr := strconv.Itoa(targetProject.IndexID)
		var projectTasks []*denote.Task
		for _, t := range allTasks {
			if t.TaskMetadata.ProjectID == projectIDStr {
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
		fmt.Printf("Project: %s\n", targetProject.Title)
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
			title := t.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}

			// Build line
			line := fmt.Sprintf("%3d %s %s %s  %s",
				t.IndexID,
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
		title        string
		priority     string
		due          string
		area         string
		status       string
		startDate    string
		addPerson    string
		removePerson string
		addTask      string
		removeTask   string
		addIdea      string
		removeIdea   string
	)

	cmd := &Command{
		Name:        "update",
		Usage:       "atask project update [options] <project-ids>",
		Description: "Update project metadata",
		Flags:       flag.NewFlagSet("project-update", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&title, "title", "", "Set title")
	cmd.Flags.StringVar(&priority, "p", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Set due date")
	cmd.Flags.StringVar(&startDate, "start", "", "Set start date")
	cmd.Flags.StringVar(&area, "area", "", "Set area")
	cmd.Flags.StringVar(&status, "status", "", "Set status (active, completed, paused, cancelled)")

	// Cross-app relationship flags
	cmd.Flags.StringVar(&addPerson, "add-person", "", "Add related contact (ULID)")
	cmd.Flags.StringVar(&removePerson, "remove-person", "", "Remove related contact (ULID)")
	cmd.Flags.StringVar(&addTask, "add-task", "", "Add related task (ULID)")
	cmd.Flags.StringVar(&removeTask, "remove-task", "", "Remove related task (ULID)")
	cmd.Flags.StringVar(&addIdea, "add-idea", "", "Add related idea (ULID)")
	cmd.Flags.StringVar(&removeIdea, "remove-idea", "", "Remove related idea (ULID)")

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
			projectsByID[p.IndexID] = p
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
			if title != "" {
				p.Title = title
				changed = true
			}
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

			// Apply cross-app relationship updates
			if addPerson != "" {
				acore.AddRelation(&p.RelatedPeople, addPerson)
				acore.SyncRelation(p.Type, p.ID, addPerson)
				changed = true
			}
			if removePerson != "" {
				acore.RemoveRelation(&p.RelatedPeople, removePerson)
				acore.UnsyncRelation(p.Type, p.ID, removePerson)
				changed = true
			}
			if addTask != "" {
				acore.AddRelation(&p.RelatedTasks, addTask)
				acore.SyncRelation(p.Type, p.ID, addTask)
				changed = true
			}
			if removeTask != "" {
				acore.RemoveRelation(&p.RelatedTasks, removeTask)
				acore.UnsyncRelation(p.Type, p.ID, removeTask)
				changed = true
			}
			if addIdea != "" {
				acore.AddRelation(&p.RelatedIdeas, addIdea)
				acore.SyncRelation(p.Type, p.ID, addIdea)
				changed = true
			}
			if removeIdea != "" {
				acore.RemoveRelation(&p.RelatedIdeas, removeIdea)
				acore.UnsyncRelation(p.Type, p.ID, removeIdea)
				changed = true
			}

			if changed {
				if err := denote.UpdateProjectFile(p.FilePath, p); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update project ID %d: %v\n", id, err)
					continue
				}
				updated++
				if !globalFlags.Quiet {
					fmt.Printf("Updated project ID %d: %s\n", id, p.Title)
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

// projectLogCommand adds or deletes a timestamped log entry on a project
func projectLogCommand(cfg *config.Config) *Command {
	var deleteLine string

	cmd := &Command{
		Name:        "log",
		Usage:       "atask project log <project-id> [message] [--delete <line>]",
		Description: "Add or delete a timestamped log entry on a project",
		Flags:       flag.NewFlagSet("project-log", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&deleteLine, "delete", "", "Delete a log entry matching this exact line")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("project ID required")
		}

		p, err := lookupProject(cfg.NotesDirectory, args[0])
		if err != nil {
			return err
		}

		if deleteLine != "" {
			if err := denote.DeleteLogEntry(p.FilePath, deleteLine); err != nil {
				return fmt.Errorf("failed to delete log entry: %v", err)
			}
			if !globalFlags.Quiet {
				fmt.Printf("Deleted log entry from project ID %d: %s\n", p.IndexID, p.Title)
			}
			return nil
		}

		if len(args) < 2 {
			return fmt.Errorf("message required (or use --delete)")
		}

		message := strings.Join(args[1:], " ")

		if err := denote.AddLogEntry(p.FilePath, message); err != nil {
			return fmt.Errorf("failed to add log entry: %v", err)
		}
		if !globalFlags.Quiet {
			fmt.Printf("Added log entry to project ID %d: %s\n", p.IndexID, p.Title)
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

		case "begin", "start":
			// Sort by start date (earliest first, empty last)
			si := projects[i].ProjectMetadata.StartDate
			sj := projects[j].ProjectMetadata.StartDate
			if si == "" && sj == "" {
				less = false
			} else if si == "" {
				less = false
			} else if sj == "" {
				less = true
			} else {
				less = si < sj
			}

		case "created":
			less = projects[i].ID < projects[j].ID

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
			less = tasks[i].ID < tasks[j].ID

		default:
			less = tasks[i].ModTime.After(tasks[j].ModTime)
		}

		if reverse {
			return !less
		}
		return less
	})
}
