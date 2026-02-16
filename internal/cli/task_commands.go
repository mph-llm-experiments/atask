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
	"github.com/pdxmph/atask/internal/config"
	"github.com/pdxmph/atask/internal/denote"
	"github.com/pdxmph/atask/internal/query"
	"github.com/pdxmph/atask/internal/task"
)

// TaskCommand creates the task command with all subcommands
func TaskCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "task",
		Usage:       "atask task <command> [options]",
		Description: "Manage tasks",
	}

	cmd.Subcommands = []*Command{
		taskNewCommand(cfg),
		taskListCommand(cfg),
		taskQueryCommand(cfg),
		taskUpdateCommand(cfg),
		taskBatchUpdateCommand(cfg),
		taskDoneCommand(cfg),
		taskLogCommand(cfg),
		taskEditCommand(cfg),
		taskDeleteCommand(cfg),
	}

	return cmd
}

// taskNewCommand creates a new task
func taskNewCommand(cfg *config.Config) *Command {
	var (
		priority string
		due      string
		area     string
		project  string
		estimate int
		tags     string
	)

	cmd := &Command{
		Name:        "new",
		Usage:       "atask task new <title> [options]",
		Description: "Create a new task",
		Flags:       flag.NewFlagSet("task-new", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&priority, "p", "", "Priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Due date (YYYY-MM-DD or natural language)")
	cmd.Flags.StringVar(&area, "area", "", "Task area")
	cmd.Flags.StringVar(&project, "project", "", "Project name or ID")
	cmd.Flags.IntVar(&estimate, "estimate", 0, "Time estimate")
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

		// Parse due date if provided
		var dueDate string
		if due != "" {
			parsed, err := denote.ParseNaturalDate(due)
			if err != nil {
				return fmt.Errorf("invalid due date: %v", err)
			}
			dueDate = parsed
		}

		// Create the task (use global area flag)
		taskFile, err := task.CreateTask(cfg.NotesDirectory, title, "", tagList, globalFlags.Area)
		if err != nil {
			return fmt.Errorf("failed to create task: %v", err)
		}

		// Update metadata if provided
		if priority != "" || dueDate != "" || project != "" || estimate > 0 {
			// Read the task
			t, err := denote.ParseTaskFile(taskFile.Path)
			if err != nil {
				return fmt.Errorf("failed to read created task: %v", err)
			}

			// Update fields
			if priority != "" {
				t.TaskMetadata.Priority = priority
			}
			if dueDate != "" {
				t.TaskMetadata.DueDate = dueDate
			}
			if project != "" {
				// TODO: Look up project by name/ID
				t.TaskMetadata.ProjectID = project
			}
			if estimate > 0 {
				t.TaskMetadata.Estimate = estimate
			}

			// Write back
			if err := task.UpdateTaskFile(taskFile.Path, t.TaskMetadata); err != nil {
				return fmt.Errorf("failed to update task metadata: %v", err)
			}
		}

		if !globalFlags.Quiet {
			fmt.Printf("Created task: %s\n", taskFile.Path)
		}

		// Launch TUI if requested
		if globalFlags.TUI {
			// TODO: Launch TUI in task view for this task
			return fmt.Errorf("TUI integration not yet implemented")
		}

		return nil
	}

	return cmd
}

// taskListCommand lists tasks
func taskListCommand(cfg *config.Config) *Command {
	var (
		all      bool
		area     string
		status   string
		priority string
		project  string
		overdue  bool
		soon     bool
		sortBy   string
		reverse  bool
		search   string
	)

	cmd := &Command{
		Name:        "list",
		Usage:       "atask task list [options]",
		Description: "List tasks",
		Flags:       flag.NewFlagSet("task-list", flag.ExitOnError),
	}

	cmd.Flags.BoolVar(&all, "all", false, "Show all tasks (default: open only)")
	cmd.Flags.StringVar(&area, "area", "", "Filter by area")
	cmd.Flags.StringVar(&status, "status", "", "Filter by status")
	cmd.Flags.StringVar(&priority, "p", "", "Filter by priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Filter by priority (p1, p2, p3)")
	cmd.Flags.StringVar(&project, "project", "", "Filter by project")
	cmd.Flags.BoolVar(&overdue, "overdue", false, "Show only overdue tasks")
	cmd.Flags.BoolVar(&soon, "soon", false, "Show tasks due soon")
	cmd.Flags.StringVar(&search, "search", "", "Search in task content (full-text)")
	cmd.Flags.StringVar(&sortBy, "sort", "modified", "Sort by: modified, priority, due, created")
	cmd.Flags.BoolVar(&reverse, "reverse", false, "Reverse sort order")
	
	// Convenience flags
	cmd.Flags.BoolVar(&all, "a", false, "Show all tasks (short)")
	cmd.Flags.StringVar(&sortBy, "s", "modified", "Sort by (short)")
	cmd.Flags.BoolVar(&reverse, "r", false, "Reverse sort (short)")

	cmd.Run = func(c *Command, args []string) error {
		// Launch TUI if requested
		if globalFlags.TUI {
			// TODO: Launch TUI with these filters applied
			return fmt.Errorf("TUI integration not yet implemented")
		}

		// Otherwise, list tasks in CLI
		scanner := denote.NewScanner(cfg.NotesDirectory)
		files, err := scanner.FindAllTaskAndProjectFiles()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// First pass: collect all projects for name lookup
		projectNames := make(map[string]string) // ID -> Title
		for _, file := range files {
			if file.IsProject() {
				p, err := denote.ParseProjectFile(file.Path)
				if err == nil {
					projectNames[file.ID] = p.ProjectMetadata.Title
				}
			}
		}

		// Second pass: filter to tasks only
		var tasks []denote.Task
		for _, file := range files {
			if !file.IsTask() {
				continue
			}

			// Parse task metadata
			t, err := denote.ParseTaskFile(file.Path)
			if err != nil {
				continue // Skip files we can't parse
			}

			// Apply filters
			if !all && status == "" && t.TaskMetadata.Status != denote.TaskStatusOpen && t.TaskMetadata.Status != "" {
				continue
			}

			if status != "" && t.TaskMetadata.Status != status {
				continue
			}

			// Use command-specific area filter or fall back to global
			filterArea := area
			if filterArea == "" {
				filterArea = globalFlags.Area
			}
			if filterArea != "" && t.TaskMetadata.Area != filterArea {
				continue
			}

			if priority != "" && t.TaskMetadata.Priority != priority {
				continue
			}

			if project != "" && t.TaskMetadata.ProjectID != project {
				continue
			}

			if overdue && !denote.IsOverdue(t.TaskMetadata.DueDate) {
				continue
			}

			if soon && !denote.IsDueSoon(t.TaskMetadata.DueDate, cfg.SoonHorizon) {
				continue
			}


		if search != "" {
			if !strings.Contains(strings.ToLower(t.Content), strings.ToLower(search)) {
				continue
			}
		}
			tasks = append(tasks, *t)
		}

		// Sort tasks
		sortTasks(tasks, sortBy, reverse)

		// Display tasks
		if globalFlags.JSON {
			// Create JSON output structure
			type TaskJSON struct {
				denote.Task
				ProjectName string `json:"project_name,omitempty"`
			}

			type Output struct {
				Tasks []TaskJSON `json:"tasks"`
				Count int        `json:"count"`
			}

			// Build JSON output with project names
			jsonTasks := make([]TaskJSON, len(tasks))
			for i, t := range tasks {
				jsonTasks[i] = TaskJSON{
					Task:        t,
					ProjectName: projectNames[t.ProjectID],
				}
			}

			output := Output{
				Tasks: jsonTasks,
				Count: len(tasks),
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
		doneColor := color.New(color.FgGreen)
		overdueColor := color.New(color.FgRed, color.Bold)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		// Display header
		if !globalFlags.Quiet {
			fmt.Printf("Tasks (%d):\n\n", len(tasks))
		}

		// Display tasks with clean, TUI-like formatting
		for _, t := range tasks {
			// Status icon
			status := "○"
			switch t.TaskMetadata.Status {
			case denote.TaskStatusDone:
				status = "✓"
			case denote.TaskStatusPaused:
				status = "⏸"
			case denote.TaskStatusDelegated:
				status = "→"
			case denote.TaskStatusDropped:
				status = "⨯"
			}

			// Priority with padding
			priority := "    " // 4 spaces for alignment
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

			// Due date with fixed width
			due := "            " // 12 spaces for alignment
			if t.TaskMetadata.DueDate != "" {
				dueStr := fmt.Sprintf("[%s]", t.TaskMetadata.DueDate)
				if denote.IsOverdue(t.TaskMetadata.DueDate) {
					due = overdueColor.Sprint(dueStr)
				} else {
					due = dueStr
				}
			}

			// Title - truncate to 50 chars
			title := t.TaskMetadata.Title
			if title == "" {
				title = t.File.Title
			}
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			// Area - truncate to 10 chars
			area := ""
			if t.TaskMetadata.Area != "" {
				area = t.TaskMetadata.Area
				if len(area) > 10 {
					area = area[:7] + "..."
				}
			}

			// Project name (look up actual name)
			projectName := ""
			if t.TaskMetadata.ProjectID != "" {
				if name, ok := projectNames[t.TaskMetadata.ProjectID]; ok && name != "" {
					projectName = "→ " + name
				} else {
					// Fallback to ID if name not found
					projectName = "→ " + t.TaskMetadata.ProjectID
				}
			}

			// Build the line with fixed-width columns
			// Format: ID Status Priority Due Title(50) Area(10) Project
			line := fmt.Sprintf("%3d %s %s %s  %-50s %-10s %s",
				t.TaskMetadata.IndexID,
				status,
				priority,
				due,
				title,
				area,
				projectName,
			)

			// Apply line coloring for done tasks
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

// sortTasks sorts tasks by the specified field
func sortTasks(tasks []denote.Task, sortBy string, reverse bool) {
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
			
		case "modified":
			fallthrough
		default:
			less = tasks[i].ModTime.After(tasks[j].ModTime)
		}
		
		if reverse {
			return !less
		}
		return less
	})
}

// priorityValue converts priority to numeric value for sorting
func priorityValue(p string) int {
	switch p {
	case "p1":
		return 1
	case "p2":
		return 2
	case "p3":
		return 3
	default:
		return 999 // No priority sorts last
	}
}

// parseTaskIDs parses task ID arguments (supports ranges and lists)
func parseTaskIDs(args []string) ([]int, error) {
	var numbers []int
	seen := make(map[int]bool)

	for _, arg := range args {
		// Handle comma-separated lists
		parts := strings.Split(arg, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			
			// Handle ranges (e.g., "3-5")
			if strings.Contains(part, "-") && !strings.HasPrefix(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) != 2 {
					return nil, fmt.Errorf("invalid range: %s", part)
				}
				
				start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
				}
				
				end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
				}
				
				if start > end {
					return nil, fmt.Errorf("invalid range: %d > %d", start, end)
				}
				
				for i := start; i <= end; i++ {
					if !seen[i] {
						numbers = append(numbers, i)
						seen[i] = true
					}
				}
			} else {
				// Single number
				num, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("invalid task ID: %s", part)
				}
				if !seen[num] {
					numbers = append(numbers, num)
					seen[num] = true
				}
			}
		}
	}

	sort.Ints(numbers)
	return numbers, nil
}

// Stub implementations for other commands
func taskUpdateCommand(cfg *config.Config) *Command {
	var (
		priority string
		due      string
		area     string
		project  string
		estimate int
		status   string
	)

	cmd := &Command{
		Name:        "update",
		Usage:       "atask task update [options] <task-ids>",
		Description: "Update task metadata",
		Flags:       flag.NewFlagSet("task-update", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&priority, "p", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Set due date")
	cmd.Flags.StringVar(&area, "area", "", "Set area")
	cmd.Flags.StringVar(&project, "project", "", "Set project")
	cmd.Flags.IntVar(&estimate, "estimate", -1, "Set time estimate")
	cmd.Flags.StringVar(&status, "status", "", "Set status (open, done, paused, delegated, dropped)")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("task IDs required")
		}

		// Parse task IDs
		numbers, err := parseTaskIDs(args)
		if err != nil {
			return err
		}

		// Get all tasks
		scanner := denote.NewScanner(cfg.NotesDirectory)
		files, err := scanner.FindAllTaskAndProjectFiles()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Build index of tasks by index_id
		tasksByID := make(map[int]*denote.Task)
		for _, file := range files {
			if !file.IsTask() {
				continue
			}
			t, err := denote.ParseTaskFile(file.Path)
			if err != nil {
				continue
			}
			tasksByID[t.TaskMetadata.IndexID] = t
		}

		// Update each task
		updated := 0
		for _, id := range numbers {
			t, ok := tasksByID[id]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %d not found\n", id)
				continue
			}

			// Apply updates
			changed := false
			if priority != "" {
				t.TaskMetadata.Priority = priority
				changed = true
			}
			if due != "" {
				parsedDue, err := denote.ParseNaturalDate(due)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid due date for task ID %d: %v\n", id, err)
					continue
				}
				t.TaskMetadata.DueDate = parsedDue
				changed = true
			}
			if area != "" {
				t.TaskMetadata.Area = area
				changed = true
			}
			if project != "" {
				t.TaskMetadata.ProjectID = project
				changed = true
			}
			if estimate >= 0 {
				t.TaskMetadata.Estimate = estimate
				changed = true
			}
			if status != "" {
				t.TaskMetadata.Status = status
				changed = true
			}

			if changed {
				if err := task.UpdateTaskFile(t.File.Path, t.TaskMetadata); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update task ID %d: %v\n", id, err)
					continue
				}
				updated++
				if !globalFlags.Quiet {
					fmt.Printf("Updated task ID %d: %s\n", id, t.TaskMetadata.Title)
				}
			}
		}

		if updated == 0 && !globalFlags.Quiet {
			fmt.Println("No tasks updated")
		}

		return nil
	}

	return cmd
}

func taskDoneCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "done",
		Usage:       "atask task done <task-ids>",
		Description: "Mark tasks as done",
	}

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("task IDs required")
		}

		// Parse task IDs
		numbers, err := parseTaskIDs(args)
		if err != nil {
			return err
		}

		// Get all tasks
		scanner := denote.NewScanner(cfg.NotesDirectory)
		files, err := scanner.FindAllTaskAndProjectFiles()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Build index of tasks by index_id
		tasksByID := make(map[int]*denote.Task)
		for _, file := range files {
			if !file.IsTask() {
				continue
			}
			t, err := denote.ParseTaskFile(file.Path)
			if err != nil {
				continue
			}
			tasksByID[t.TaskMetadata.IndexID] = t
		}

		// Mark tasks as done
		updated := 0
		for _, id := range numbers {
			t, ok := tasksByID[id]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %d not found\n", id)
				continue
			}

			t.TaskMetadata.Status = denote.TaskStatusDone
			if err := task.UpdateTaskFile(t.File.Path, t.TaskMetadata); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to mark task ID %d as done: %v\n", id, err)
				continue
			}
			updated++
			if !globalFlags.Quiet {
				fmt.Printf("✓ Task ID %d marked as done: %s\n", id, t.TaskMetadata.Title)
			}
		}

		if updated == 0 && !globalFlags.Quiet {
			fmt.Println("No tasks marked as done")
		}

		return nil
	}

	return cmd
}

func taskLogCommand(cfg *config.Config) *Command {
	cmd := &Command{
		Name:        "log",
		Usage:       "atask task log <task-id> <message>",
		Description: "Add a timestamped log entry to a task",
	}

	cmd.Run = func(c *Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("task ID and message required")
		}

		// Parse task ID
		taskNum, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID: %s", args[0])
		}

		// Join remaining args as log message
		message := strings.Join(args[1:], " ")

		// Get all tasks
		scanner := denote.NewScanner(cfg.NotesDirectory)
		files, err := scanner.FindAllTaskAndProjectFiles()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Find the task by index_id
		for _, file := range files {
			if !file.IsTask() {
				continue
			}
			// Parse the task
			task, err := denote.ParseTaskFile(file.Path)
			if err != nil {
				continue
			}
			if task.TaskMetadata.IndexID == taskNum {
				// Add log entry
				if err := denote.AddLogEntry(file.Path, message); err != nil {
					return fmt.Errorf("failed to add log entry: %v", err)
				}

				if !globalFlags.Quiet {
					fmt.Printf("Added log entry to task ID %d: %s\n", taskNum, task.TaskMetadata.Title)
				}
				return nil
			}
		}

		return fmt.Errorf("task with ID %d not found", taskNum)
	}

	return cmd
}

func taskEditCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "edit",
		Usage:       "atask task edit <task-id>",
		Description: "Edit task in external editor or TUI",
		Run: func(c *Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
}

func taskDeleteCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "delete",
		Usage:       "atask task delete <task-ids>",
		Description: "Delete tasks (with confirmation)",
		Run: func(c *Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
}

// taskQueryCommand allows complex filtering with boolean expressions
func taskQueryCommand(cfg *config.Config) *Command {
	var sortBy string
	var reverse bool

	cmd := &Command{
		Name:        "query",
		Usage:       "atask query <expression> [options]",
		Description: "Query tasks with complex filter expressions",
		Flags:       flag.NewFlagSet("task-query", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&sortBy, "sort", "modified", "Sort by: priority, due, created, modified")
	cmd.Flags.BoolVar(&reverse, "r", false, "Reverse sort order")
	cmd.Flags.BoolVar(&reverse, "reverse", false, "Reverse sort order")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("query expression required\n\nExamples:\n  atask query \"status:open AND priority:p1\"\n  atask query \"area:work AND (priority:p1 OR priority:p2)\"\n  atask query \"due:soon AND NOT status:done\"")
		}

		// Query is the first argument (should be quoted if it contains spaces)
		queryStr := args[0]

		// Parse the query
		ast, err := query.Parse(queryStr)
		if err != nil {
			return fmt.Errorf("query parse error: %v", err)
		}

		// Get all tasks
		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		// Get project names for display
		projects, _ := scanner.FindProjects()
		projectNames := make(map[string]string)
		for _, p := range projects {
			projectNames[p.File.ID] = p.ProjectMetadata.Title
		}

		// Evaluate query against all tasks
		var tasks []denote.Task
		for _, t := range allTasks {
			if ast.Evaluate(t, cfg) {
				tasks = append(tasks, *t)
			}
		}

		// Sort tasks
		sortTasks(tasks, sortBy, reverse)

		// Display tasks
		if globalFlags.JSON {
			// Create JSON output structure
			type TaskJSON struct {
				denote.Task
				ProjectName string `json:"project_name,omitempty"`
			}

			type Output struct {
				Tasks []TaskJSON `json:"tasks"`
				Count int        `json:"count"`
			}

			// Build JSON output with project names
			jsonTasks := make([]TaskJSON, len(tasks))
			for i, t := range tasks {
				jsonTasks[i] = TaskJSON{
					Task:        t,
					ProjectName: projectNames[t.ProjectID],
				}
			}

			output := Output{
				Tasks: jsonTasks,
				Count: len(tasks),
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
		doneColor := color.New(color.FgGreen)
		overdueColor := color.New(color.FgRed, color.Bold)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		// Display header
		if !globalFlags.Quiet {
			fmt.Printf("Tasks (%d):\n\n", len(tasks))
		}

		// Display tasks with clean, TUI-like formatting
		for _, t := range tasks {
			// Status icon
			status := "○"
			switch t.TaskMetadata.Status {
			case denote.TaskStatusDone:
				status = doneColor.Sprint("✓")
			case denote.TaskStatusPaused:
				status = "⏸"
			case denote.TaskStatusDropped:
				status = "⨯"
			case denote.TaskStatusDelegated:
				status = "→"
			}

			// Priority with color
			priority := "   "
			if t.TaskMetadata.Priority != "" {
				switch t.TaskMetadata.Priority {
				case denote.PriorityP1:
					priority = priorityHighColor.Sprintf("[%s]", t.TaskMetadata.Priority)
				case denote.PriorityP2:
					priority = priorityMedColor.Sprintf("[%s]", t.TaskMetadata.Priority)
				default:
					priority = fmt.Sprintf("[%s]", t.TaskMetadata.Priority)
				}
			}

			// Due date with color if overdue
			due := "            "
			if t.TaskMetadata.DueDate != "" {
				if denote.IsOverdue(t.TaskMetadata.DueDate) && t.TaskMetadata.Status != denote.TaskStatusDone {
					due = overdueColor.Sprintf("[%s]", t.TaskMetadata.DueDate)
				} else {
					due = fmt.Sprintf("[%s]", t.TaskMetadata.DueDate)
				}
			}

			// Title (truncate if too long)
			title := t.TaskMetadata.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			// Area
			area := ""
			if t.TaskMetadata.Area != "" {
				area = t.TaskMetadata.Area
			}

			// Project name
			projectName := ""
			if t.TaskMetadata.ProjectID != "" {
				if name, ok := projectNames[t.TaskMetadata.ProjectID]; ok && name != "" {
					projectName = "→ " + name
				} else {
					// Fallback to ID if name not found
					projectName = "→ " + t.TaskMetadata.ProjectID
				}
			}

			// Build the line with fixed-width columns
			line := fmt.Sprintf("%3d %s %s %s  %-50s %-10s %s",
				t.TaskMetadata.IndexID,
				status,
				priority,
				due,
				title,
				area,
				projectName,
			)

			fmt.Println(line)
		}

		return nil
	}

	return cmd
}

// taskBatchUpdateCommand updates multiple tasks based on query conditions
func taskBatchUpdateCommand(cfg *config.Config) *Command {
	var (
		whereClause string
		priority    string
		due         string
		area        string
		project     string
		estimate    int
		status      string
		preview     bool
	)

	cmd := &Command{
		Name:        "batch-update",
		Usage:       "atask batch-update --where <query> --set <field=value> [options]",
		Description: "Update multiple tasks based on query conditions",
		Flags:       flag.NewFlagSet("task-batch-update", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&whereClause, "where", "", "Query expression to filter tasks")
	cmd.Flags.StringVar(&priority, "priority", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Set due date")
	cmd.Flags.StringVar(&area, "area", "", "Set area")
	cmd.Flags.StringVar(&project, "project", "", "Set project")
	cmd.Flags.IntVar(&estimate, "estimate", -1, "Set time estimate")
	cmd.Flags.StringVar(&status, "status", "", "Set status (open, done, paused, delegated, dropped)")
	cmd.Flags.BoolVar(&preview, "preview", false, "Preview changes without applying them")

	cmd.Run = func(c *Command, args []string) error {
		if whereClause == "" {
			return fmt.Errorf("--where clause required\n\nExample:\n  atask batch-update --where \"status:open AND due:past\" --status paused")
		}

		// Check that at least one field to update is specified
		if priority == "" && due == "" && area == "" && project == "" && estimate == -1 && status == "" {
			return fmt.Errorf("at least one field to update must be specified (--priority, --due, --area, --project, --estimate, or --status)")
		}

		// Parse the where clause
		ast, err := query.Parse(whereClause)
		if err != nil {
			return fmt.Errorf("failed to parse --where clause: %v", err)
		}

		// Get all tasks
		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		// Filter tasks using the query
		var matchingTasks []*denote.Task
		for _, t := range allTasks {
			if ast.Evaluate(t, cfg) {
				matchingTasks = append(matchingTasks, t)
			}
		}

		if len(matchingTasks) == 0 {
			fmt.Println("No tasks match the query")
			return nil
		}

		// Show what will be updated
		fmt.Printf("Found %d matching task(s):\n\n", len(matchingTasks))
		for _, t := range matchingTasks {
			fmt.Printf("  %d: %s\n", t.TaskMetadata.IndexID, t.TaskMetadata.Title)
		}
		fmt.Println()

		// Parse due date if provided
		var parsedDue string
		if due != "" {
			parsedDue, err = denote.ParseNaturalDate(due)
			if err != nil {
				return fmt.Errorf("invalid due date: %v", err)
			}
		}

		// Show what changes will be made
		changes := []string{}
		if priority != "" {
			changes = append(changes, fmt.Sprintf("priority → %s", priority))
		}
		if due != "" {
			changes = append(changes, fmt.Sprintf("due_date → %s", parsedDue))
		}
		if area != "" {
			changes = append(changes, fmt.Sprintf("area → %s", area))
		}
		if project != "" {
			changes = append(changes, fmt.Sprintf("project_id → %s", project))
		}
		if estimate >= 0 {
			changes = append(changes, fmt.Sprintf("estimate → %d", estimate))
		}
		if status != "" {
			changes = append(changes, fmt.Sprintf("status → %s", status))
		}

		fmt.Printf("Changes to apply:\n")
		for _, change := range changes {
			fmt.Printf("  • %s\n", change)
		}
		fmt.Println()

		if preview {
			fmt.Println("Preview mode: no changes applied")
			return nil
		}

		// Apply updates
		updated := 0
		for _, t := range matchingTasks {
			changed := false

			if priority != "" {
				t.TaskMetadata.Priority = priority
				changed = true
			}
			if due != "" {
				t.TaskMetadata.DueDate = parsedDue
				changed = true
			}
			if area != "" {
				t.TaskMetadata.Area = area
				changed = true
			}
			if project != "" {
				t.TaskMetadata.ProjectID = project
				changed = true
			}
			if estimate >= 0 {
				t.TaskMetadata.Estimate = estimate
				changed = true
			}
			if status != "" {
				t.TaskMetadata.Status = status
				changed = true
			}

			if changed {
				if err := task.UpdateTaskFile(t.File.Path, t.TaskMetadata); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update task %d: %v\n", t.TaskMetadata.IndexID, err)
					continue
				}
				updated++
			}
		}

		fmt.Printf("✓ Updated %d task(s)\n", updated)
		return nil
	}

	return cmd
}