package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/config"
	"github.com/mph-llm-experiments/atask/internal/denote"
	"github.com/mph-llm-experiments/atask/internal/query"
	"github.com/mph-llm-experiments/atask/internal/recurrence"
	"github.com/mph-llm-experiments/atask/internal/task"
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
		taskShowCommand(cfg),
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

// lookupTask finds a task by integer index_id or ULID string.
func lookupTask(dir string, identifier string) (*denote.Task, error) {
	// Try as integer index_id first
	if num, err := strconv.Atoi(identifier); err == nil {
		return task.FindTaskByID(dir, num)
	}
	// Otherwise treat as ULID / entity ID
	return task.FindTaskByEntityID(dir, identifier)
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
		recur    string
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
	cmd.Flags.StringVar(&recur, "recur", "", "Recurrence pattern (daily, weekly, monthly, yearly, every Nd/Nw/Nm/Ny, every mon,wed,fri)")

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

		// Validate recurrence pattern if provided
		var recurPattern string
		if recur != "" {
			if due == "" {
				return fmt.Errorf("--due is required when --recur is set")
			}
			var err error
			recurPattern, err = recurrence.ParsePattern(recur)
			if err != nil {
				return fmt.Errorf("invalid recurrence pattern: %v", err)
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
		if priority != "" || dueDate != "" || project != "" || estimate > 0 || recurPattern != "" {
			t, err := denote.ParseTaskFile(taskFile.FilePath)
			if err != nil {
				return fmt.Errorf("failed to read created task: %v", err)
			}

			if priority != "" {
				t.TaskMetadata.Priority = priority
			}
			if dueDate != "" {
				t.TaskMetadata.DueDate = dueDate
			}
			if project != "" {
				projectNum, err := strconv.Atoi(project)
				if err != nil {
					return fmt.Errorf("invalid project ID: %s (must be a numeric index_id)", project)
				}
				p, err := task.FindProjectByID(cfg.NotesDirectory, projectNum)
				if err != nil {
					return fmt.Errorf("project %d not found", projectNum)
				}
				t.TaskMetadata.ProjectID = strconv.Itoa(p.IndexID)
			}
			if estimate > 0 {
				t.TaskMetadata.Estimate = estimate
			}
			if recurPattern != "" {
				t.TaskMetadata.Recur = recurPattern
			}

			if err := task.UpdateTaskFile(t.FilePath, t); err != nil {
				return fmt.Errorf("failed to update task metadata: %v", err)
			}
		}

		// Reload to get final state (after any metadata updates)
		final, err := denote.ParseTaskFile(taskFile.FilePath)
		if err != nil {
			final = taskFile
		}

		if globalFlags.JSON {
			data, _ := json.MarshalIndent(final, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if !globalFlags.Quiet {
			fmt.Printf("Created task: %s\n", final.FilePath)
		}

		return nil
	}

	return cmd
}

// taskShowCommand shows details for a single task
func taskShowCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "show",
		Usage:       "atask show <id>",
		Description: "Show task details by index_id or ULID",
		Run: func(cmd *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask show <id>")
			}

			t, err := lookupTask(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			if globalFlags.JSON {
				data, err := json.MarshalIndent(t, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Text output
			fmt.Printf("# %s (#%d)\n\n", t.Title, t.IndexID)

			fmt.Printf("  ID:       %s\n", t.ID)
			fmt.Printf("  Status:   %s\n", t.TaskMetadata.Status)
			if t.TaskMetadata.Priority != "" {
				fmt.Printf("  Priority: %s\n", t.TaskMetadata.Priority)
			}
			if t.TaskMetadata.DueDate != "" {
				dueStr := t.TaskMetadata.DueDate
				if denote.IsOverdue(t.TaskMetadata.DueDate) && t.TaskMetadata.Status != denote.TaskStatusDone {
					dueStr += " (OVERDUE)"
				}
				fmt.Printf("  Due:      %s\n", dueStr)
			}
			if t.TaskMetadata.StartDate != "" {
				fmt.Printf("  Start:    %s\n", t.TaskMetadata.StartDate)
			}
			if t.TaskMetadata.Area != "" {
				fmt.Printf("  Area:     %s\n", t.TaskMetadata.Area)
			}
			if t.TaskMetadata.ProjectID != "" {
				projectName := t.TaskMetadata.ProjectID
				if p, err := task.FindProjectByID(cfg.NotesDirectory, func() int {
					n, _ := strconv.Atoi(t.TaskMetadata.ProjectID)
					return n
				}()); err == nil {
					projectName = fmt.Sprintf("%s (#%d)", p.Title, p.IndexID)
				}
				fmt.Printf("  Project:  %s\n", projectName)
			}
			if t.TaskMetadata.Estimate > 0 {
				fmt.Printf("  Estimate: %d\n", t.TaskMetadata.Estimate)
			}
			if t.TaskMetadata.Assignee != "" {
				fmt.Printf("  Assignee: %s\n", t.TaskMetadata.Assignee)
			}
			if t.TaskMetadata.Recur != "" {
				fmt.Printf("  Recur:    %s\n", t.TaskMetadata.Recur)
			}
			fmt.Println()

			if t.Created != "" {
				fmt.Printf("  Created:  %s\n", t.Created)
			}
			if t.Modified != "" {
				fmt.Printf("  Modified: %s\n", t.Modified)
			}

			var tagStrs []string
			for _, tag := range t.Tags {
				if tag != "task" {
					tagStrs = append(tagStrs, "#"+tag)
				}
			}
			if len(tagStrs) > 0 {
				fmt.Printf("\n  Tags: %s\n", strings.Join(tagStrs, " "))
			}

			if len(t.RelatedPeople) > 0 || len(t.RelatedTasks) > 0 || len(t.RelatedIdeas) > 0 {
				fmt.Println()
				if len(t.RelatedPeople) > 0 {
					fmt.Printf("  Related people: %s\n", strings.Join(t.RelatedPeople, ", "))
				}
				if len(t.RelatedTasks) > 0 {
					fmt.Printf("  Related tasks:  %s\n", strings.Join(t.RelatedTasks, ", "))
				}
				if len(t.RelatedIdeas) > 0 {
					fmt.Printf("  Related ideas:  %s\n", strings.Join(t.RelatedIdeas, ", "))
				}
			}

			if strings.TrimSpace(t.Content) != "" {
				fmt.Printf("\n---\n%s", t.Content)
			}

			return nil
		},
	}
}

// taskListCommand lists tasks
func taskListCommand(cfg *config.Config) *Command {
	var (
		all        bool
		area       string
		status     string
		priority   string
		project    string
		overdue    bool
		soon       bool
		sortBy     string
		reverse    bool
		search     string
		plannedFor string
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
	cmd.Flags.StringVar(&plannedFor, "planned-for", "", "Filter by planned_for date (today, YYYY-MM-DD, or any)")
	cmd.Flags.StringVar(&sortBy, "sort", "modified", "Sort by: modified, priority, due, created")
	cmd.Flags.BoolVar(&reverse, "reverse", false, "Reverse sort order")

	cmd.Flags.BoolVar(&all, "a", false, "Show all tasks (short)")
	cmd.Flags.StringVar(&sortBy, "s", "modified", "Sort by (short)")
	cmd.Flags.BoolVar(&reverse, "r", false, "Reverse sort (short)")

	cmd.Run = func(c *Command, args []string) error {
		if globalFlags.TUI {
			return fmt.Errorf("TUI integration not yet implemented")
		}

		scanner := denote.NewScanner(cfg.NotesDirectory)

		// Get all projects for name lookup and hidden status
		projects, _ := scanner.FindProjects()
		projectNames := make(map[string]string)
		hiddenProjectIDs := make(map[string]bool)
		for _, p := range projects {
			idStr := strconv.Itoa(p.IndexID)
			projectNames[idStr] = p.Title
			if p.ProjectMetadata.Status == denote.ProjectStatusPaused ||
				p.ProjectMetadata.Status == denote.ProjectStatusCancelled ||
				p.HasNotBegun() {
				hiddenProjectIDs[idStr] = true
			}
		}

		// Get all tasks
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		// Filter tasks
		var tasks []denote.Task
		for _, t := range allTasks {
			if !all && status == "" && t.TaskMetadata.Status != denote.TaskStatusOpen && t.TaskMetadata.Status != "" {
				continue
			}
			if status != "" && t.TaskMetadata.Status != status {
				continue
			}
			if !all && t.TaskMetadata.ProjectID != "" && hiddenProjectIDs[t.TaskMetadata.ProjectID] {
				continue
			}

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
			if plannedFor != "" {
				switch strings.ToLower(plannedFor) {
				case "any":
					if t.PlannedFor == "" {
						continue
					}
				case "today":
					if t.PlannedFor != time.Now().Format("2006-01-02") {
						continue
					}
				default:
					if t.PlannedFor != plannedFor {
						continue
					}
				}
			}
			tasks = append(tasks, *t)
		}

		sortTasks(tasks, sortBy, reverse)

		if globalFlags.JSON {
			type TaskJSON struct {
				denote.Task
				ProjectName string `json:"project_name,omitempty"`
			}
			type Output struct {
				Tasks []TaskJSON `json:"tasks"`
				Count int        `json:"count"`
			}

			jsonTasks := make([]TaskJSON, len(tasks))
			for i, t := range tasks {
				jsonTasks[i] = TaskJSON{
					Task:        t,
					ProjectName: projectNames[t.ProjectID],
				}
			}

			output := Output{Tasks: jsonTasks, Count: len(tasks)}
			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		if globalFlags.NoColor || color.NoColor {
			color.NoColor = true
		}

		doneColor := color.New(color.FgGreen)
		overdueColor := color.New(color.FgRed, color.Bold)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		if !globalFlags.Quiet {
			fmt.Printf("Tasks (%d):\n\n", len(tasks))
		}

		for _, t := range tasks {
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

			priorityStr := "    "
			if t.TaskMetadata.Priority != "" {
				pStr := fmt.Sprintf("[%s]", t.TaskMetadata.Priority)
				switch t.TaskMetadata.Priority {
				case "p1":
					priorityStr = priorityHighColor.Sprint(pStr)
				case "p2":
					priorityStr = priorityMedColor.Sprint(pStr)
				default:
					priorityStr = pStr
				}
			}

			dueStr := "            "
			if t.TaskMetadata.DueDate != "" {
				ds := fmt.Sprintf("[%s]", t.TaskMetadata.DueDate)
				if denote.IsOverdue(t.TaskMetadata.DueDate) {
					dueStr = overdueColor.Sprint(ds)
				} else {
					dueStr = ds
				}
			}

			title := t.Title
			if t.TaskMetadata.Recur != "" {
				title = "↻ " + title
			}
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			areaStr := ""
			if t.TaskMetadata.Area != "" {
				areaStr = t.TaskMetadata.Area
				if len(areaStr) > 10 {
					areaStr = areaStr[:7] + "..."
				}
			}

			projectName := ""
			if t.TaskMetadata.ProjectID != "" {
				if name, ok := projectNames[t.TaskMetadata.ProjectID]; ok && name != "" {
					projectName = "→ " + name
				} else {
					projectName = "→ " + t.TaskMetadata.ProjectID
				}
			}

			line := fmt.Sprintf("%3d %s %s %s  %-50s %-10s %s",
				t.IndexID,
				statusIcon,
				priorityStr,
				dueStr,
				title,
				areaStr,
				projectName,
			)

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
			pi := priorityValue(tasks[i].TaskMetadata.Priority)
			pj := priorityValue(tasks[j].TaskMetadata.Priority)
			less = pi < pj

		case "due":
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
		return 999
	}
}

// parseTaskIdentifiers parses task ID arguments, returning integer IDs and
// string entity IDs (ULIDs) separately. Supports ranges and comma-separated
// lists for integer IDs.
func parseTaskIdentifiers(args []string) (intIDs []int, entityIDs []string, err error) {
	seenInt := make(map[int]bool)
	seenStr := make(map[string]bool)

	for _, arg := range args {
		parts := strings.Split(arg, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)

			// Check if it looks like an integer range (e.g. "1-5")
			if strings.Contains(part, "-") && !strings.HasPrefix(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) == 2 {
					start, errS := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
					end, errE := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
					if errS == nil && errE == nil {
						if start > end {
							return nil, nil, fmt.Errorf("invalid range: %d > %d", start, end)
						}
						for i := start; i <= end; i++ {
							if !seenInt[i] {
								intIDs = append(intIDs, i)
								seenInt[i] = true
							}
						}
						continue
					}
				}
				// Not a valid integer range — treat as entity ID (ULIDs contain hyphens? No, but fall through)
			}

			// Try as integer
			if num, err := strconv.Atoi(part); err == nil {
				if !seenInt[num] {
					intIDs = append(intIDs, num)
					seenInt[num] = true
				}
			} else {
				// Treat as ULID / entity ID
				if !seenStr[part] {
					entityIDs = append(entityIDs, part)
					seenStr[part] = true
				}
			}
		}
	}

	sort.Ints(intIDs)
	return intIDs, entityIDs, nil
}

// parseTaskIDs parses task ID arguments (supports ranges and lists).
// Only returns integer IDs — use parseTaskIdentifiers for ULID support.
func parseTaskIDs(args []string) ([]int, error) {
	intIDs, entityIDs, err := parseTaskIdentifiers(args)
	if err != nil {
		return nil, err
	}
	if len(entityIDs) > 0 {
		return nil, fmt.Errorf("invalid task ID: %s", entityIDs[0])
	}
	return intIDs, nil
}

func taskUpdateCommand(cfg *config.Config) *Command {
	var (
		title        string
		priority     string
		due          string
		begin        string
		area         string
		project      string
		estimate     int
		status       string
		recur        string
		planFor      string
		addPerson    string
		removePerson string
		addTask      string
		removeTask   string
		addIdea      string
		removeIdea   string
	)

	cmd := &Command{
		Name:        "update",
		Usage:       "atask task update [options] <task-ids>",
		Description: "Update task metadata",
		Flags:       flag.NewFlagSet("task-update", flag.ExitOnError),
	}

	cmd.Flags.StringVar(&title, "title", "", "Set title")
	cmd.Flags.StringVar(&priority, "p", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&priority, "priority", "", "Set priority (p1, p2, p3)")
	cmd.Flags.StringVar(&due, "due", "", "Set due date")
	cmd.Flags.StringVar(&begin, "begin", "", "Set begin/start date")
	cmd.Flags.StringVar(&area, "area", "", "Set area")
	cmd.Flags.StringVar(&project, "project", "", "Set project")
	cmd.Flags.IntVar(&estimate, "estimate", -1, "Set time estimate")
	cmd.Flags.StringVar(&status, "status", "", "Set status (open, done, paused, delegated, dropped)")
	cmd.Flags.StringVar(&recur, "recur", "", "Set recurrence (use 'none' to clear)")
	cmd.Flags.StringVar(&planFor, "plan-for", "", "Set planned_for date (natural language, YYYY-MM-DD, or 'none' to clear)")

	cmd.Flags.StringVar(&addPerson, "add-person", "", "Add related contact (ULID)")
	cmd.Flags.StringVar(&removePerson, "remove-person", "", "Remove related contact (ULID)")
	cmd.Flags.StringVar(&addTask, "add-task", "", "Add related task (ULID)")
	cmd.Flags.StringVar(&removeTask, "remove-task", "", "Remove related task (ULID)")
	cmd.Flags.StringVar(&addIdea, "add-idea", "", "Add related idea (ULID)")
	cmd.Flags.StringVar(&removeIdea, "remove-idea", "", "Remove related idea (ULID)")

	cmd.Run = func(c *Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("task IDs required")
		}

		var recurPattern string
		var clearRecur bool
		if recur != "" {
			if strings.ToLower(recur) == "none" {
				clearRecur = true
			} else {
				var err error
				recurPattern, err = recurrence.ParsePattern(recur)
				if err != nil {
					return fmt.Errorf("invalid recurrence pattern: %v", err)
				}
			}
		}

		intIDs, entityIDs, err := parseTaskIdentifiers(args)
		if err != nil {
			return err
		}

		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		tasksByID := make(map[int]*denote.Task)
		tasksByEntityID := make(map[string]*denote.Task)
		for _, t := range allTasks {
			tasksByID[t.IndexID] = t
			tasksByEntityID[t.ID] = t
		}

		// Track updated tasks for JSON output
		var updatedTasks []*denote.Task

		// Collect tasks to update from both integer and entity IDs
		var tasksToUpdate []*denote.Task
		for _, id := range intIDs {
			t, ok := tasksByID[id]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %d not found\n", id)
				continue
			}
			tasksToUpdate = append(tasksToUpdate, t)
		}
		for _, eid := range entityIDs {
			t, ok := tasksByEntityID[eid]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %s not found\n", eid)
				continue
			}
			tasksToUpdate = append(tasksToUpdate, t)
		}

		updated := 0
		for _, t := range tasksToUpdate {

			changed := false
			if title != "" {
				t.Title = title
				changed = true
			}
			if priority != "" {
				t.TaskMetadata.Priority = priority
				changed = true
			}
			if due != "" {
				parsedDue, err := denote.ParseNaturalDate(due)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid due date for task ID %d: %v\n", t.IndexID, err)
					continue
				}
				t.TaskMetadata.DueDate = parsedDue
				changed = true
			}
			if begin != "" {
				parsedBegin, err := denote.ParseNaturalDate(begin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid begin date for task ID %d: %v\n", t.IndexID, err)
					continue
				}
				t.TaskMetadata.StartDate = parsedBegin
				changed = true
			}
			if area != "" {
				t.TaskMetadata.Area = area
				changed = true
			}
			if project != "" {
				projectNum, err := strconv.Atoi(project)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid project ID for task %d: %s (must be numeric)\n", t.IndexID, project)
					continue
				}
				p, err := task.FindProjectByID(cfg.NotesDirectory, projectNum)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Project %d not found for task %d\n", projectNum, t.IndexID)
					continue
				}
				t.TaskMetadata.ProjectID = strconv.Itoa(p.IndexID)
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
			if clearRecur {
				t.TaskMetadata.Recur = ""
				changed = true
			} else if recurPattern != "" {
				t.TaskMetadata.Recur = recurPattern
				changed = true
			}

			if planFor != "" {
				if strings.ToLower(planFor) == "none" {
					t.PlannedFor = ""
					changed = true
				} else {
					parsed, err := denote.ParseNaturalDate(planFor)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Invalid --plan-for date for task ID %d: %v\n", t.IndexID, err)
						continue
					}
					t.PlannedFor = parsed
					changed = true
				}
			}

			// Cross-app relationship updates
			if addPerson != "" {
				acore.AddRelation(&t.RelatedPeople, addPerson)
				acore.SyncRelation(t.Type, t.ID, addPerson)
				changed = true
			}
			if removePerson != "" {
				acore.RemoveRelation(&t.RelatedPeople, removePerson)
				acore.UnsyncRelation(t.Type, t.ID, removePerson)
				changed = true
			}
			if addTask != "" {
				acore.AddRelation(&t.RelatedTasks, addTask)
				acore.SyncRelation(t.Type, t.ID, addTask)
				changed = true
			}
			if removeTask != "" {
				acore.RemoveRelation(&t.RelatedTasks, removeTask)
				acore.UnsyncRelation(t.Type, t.ID, removeTask)
				changed = true
			}
			if addIdea != "" {
				acore.AddRelation(&t.RelatedIdeas, addIdea)
				acore.SyncRelation(t.Type, t.ID, addIdea)
				changed = true
			}
			if removeIdea != "" {
				acore.RemoveRelation(&t.RelatedIdeas, removeIdea)
				acore.UnsyncRelation(t.Type, t.ID, removeIdea)
				changed = true
			}

			if changed {
				if err := task.UpdateTaskFile(t.FilePath, t); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update task ID %d: %v\n", t.IndexID, err)
					continue
				}
				updated++
				updatedTasks = append(updatedTasks, t)
				if !globalFlags.JSON && !globalFlags.Quiet {
					fmt.Printf("Updated task ID %d: %s\n", t.IndexID, t.Title)
				}
			}
		}

		if globalFlags.JSON && len(updatedTasks) > 0 {
			// Reload from disk for accurate output
			var results []*denote.Task
			for _, t := range updatedTasks {
				if reloaded, err := denote.ParseTaskFile(t.FilePath); err == nil {
					results = append(results, reloaded)
				} else {
					results = append(results, t)
				}
			}
			if len(results) == 1 {
				data, _ := json.MarshalIndent(results[0], "", "  ")
				fmt.Println(string(data))
			} else {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(data))
			}
			return nil
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

		intIDs, entityIDs, err := parseTaskIdentifiers(args)
		if err != nil {
			return err
		}

		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to scan directory: %v", err)
		}

		tasksByID := make(map[int]*denote.Task)
		tasksByEntityID := make(map[string]*denote.Task)
		for _, t := range allTasks {
			tasksByID[t.IndexID] = t
			tasksByEntityID[t.ID] = t
		}

		var tasksToUpdate []*denote.Task
		for _, id := range intIDs {
			t, ok := tasksByID[id]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %d not found\n", id)
				continue
			}
			tasksToUpdate = append(tasksToUpdate, t)
		}
		for _, eid := range entityIDs {
			t, ok := tasksByEntityID[eid]
			if !ok {
				fmt.Fprintf(os.Stderr, "Task with ID %s not found\n", eid)
				continue
			}
			tasksToUpdate = append(tasksToUpdate, t)
		}

		updated := 0
		for _, t := range tasksToUpdate {
			t.TaskMetadata.Status = denote.TaskStatusDone
			if err := task.UpdateTaskFile(t.FilePath, t); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to mark task %d as done: %v\n", t.IndexID, err)
				continue
			}
			updated++
			if !globalFlags.Quiet {
				fmt.Printf("✓ Task ID %d marked as done: %s\n", t.IndexID, t.Title)
			}

			if err := handleRecurrence(cfg, t); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create recurring task for ID %d: %v\n", t.IndexID, err)
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

		t, err := lookupTask(cfg.NotesDirectory, args[0])
		if err != nil {
			return err
		}

		message := strings.Join(args[1:], " ")

		if err := denote.AddLogEntry(t.FilePath, message); err != nil {
			return fmt.Errorf("failed to add log entry: %v", err)
		}
		if !globalFlags.Quiet {
			fmt.Printf("Added log entry to task ID %d: %s\n", t.IndexID, t.Title)
		}
		return nil
	}

	return cmd
}

func taskEditCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "edit",
		Usage:       "atask task edit <task-id>",
		Description: "Open task file in $EDITOR",
		Run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask task edit <task-id>")
			}

			t, err := lookupTask(cfg.NotesDirectory, args[0])
			if err != nil {
				return err
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			cmd := exec.Command(editor, t.FilePath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
}

func taskDeleteCommand(cfg *config.Config) *Command {
	return &Command{
		Name:        "delete",
		Usage:       "atask task delete <task-id> [--confirm]",
		Description: "Delete a task file",
		Run: func(c *Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: atask task delete <task-id> [--confirm]")
			}

			confirm := false
			idRef := ""
			for _, arg := range args {
				if arg == "--confirm" {
					confirm = true
				} else if idRef == "" {
					idRef = arg
				}
			}
			if idRef == "" {
				return fmt.Errorf("usage: atask task delete <task-id> [--confirm]")
			}

			t, err := lookupTask(cfg.NotesDirectory, idRef)
			if err != nil {
				return err
			}

			if !confirm {
				return fmt.Errorf("use --confirm to delete task '%s' (%s)", t.Title, t.FilePath)
			}

			if err := os.Remove(t.FilePath); err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}

			if globalFlags.JSON {
				result := map[string]interface{}{
					"deleted":  true,
					"index_id": t.IndexID,
					"title":    t.Title,
					"file":     t.FilePath,
				}
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			if !globalFlags.Quiet {
				fmt.Printf("Deleted task #%d: %s\n", t.IndexID, t.Title)
			}
			return nil
		},
	}
}

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

		queryStr := args[0]

		ast, err := query.Parse(queryStr)
		if err != nil {
			return fmt.Errorf("query parse error: %v", err)
		}

		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

		projects, _ := scanner.FindProjects()
		projectNames := make(map[string]string)
		for _, p := range projects {
			projectNames[strconv.Itoa(p.IndexID)] = p.Title
		}

		var tasks []denote.Task
		for _, t := range allTasks {
			if ast.Evaluate(t, cfg) {
				tasks = append(tasks, *t)
			}
		}

		sortTasks(tasks, sortBy, reverse)

		if globalFlags.JSON {
			type TaskJSON struct {
				denote.Task
				ProjectName string `json:"project_name,omitempty"`
			}
			type Output struct {
				Tasks []TaskJSON `json:"tasks"`
				Count int        `json:"count"`
			}

			jsonTasks := make([]TaskJSON, len(tasks))
			for i, t := range tasks {
				jsonTasks[i] = TaskJSON{
					Task:        t,
					ProjectName: projectNames[t.ProjectID],
				}
			}

			output := Output{Tasks: jsonTasks, Count: len(tasks)}
			jsonBytes, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		if globalFlags.NoColor || color.NoColor {
			color.NoColor = true
		}

		doneColor := color.New(color.FgGreen)
		overdueColor := color.New(color.FgRed, color.Bold)
		priorityHighColor := color.New(color.FgRed, color.Bold)
		priorityMedColor := color.New(color.FgYellow)

		if !globalFlags.Quiet {
			fmt.Printf("Tasks (%d):\n\n", len(tasks))
		}

		for _, t := range tasks {
			statusIcon := "○"
			switch t.TaskMetadata.Status {
			case denote.TaskStatusDone:
				statusIcon = doneColor.Sprint("✓")
			case denote.TaskStatusPaused:
				statusIcon = "⏸"
			case denote.TaskStatusDropped:
				statusIcon = "⨯"
			case denote.TaskStatusDelegated:
				statusIcon = "→"
			}

			priorityStr := "   "
			if t.TaskMetadata.Priority != "" {
				switch t.TaskMetadata.Priority {
				case denote.PriorityP1:
					priorityStr = priorityHighColor.Sprintf("[%s]", t.TaskMetadata.Priority)
				case denote.PriorityP2:
					priorityStr = priorityMedColor.Sprintf("[%s]", t.TaskMetadata.Priority)
				default:
					priorityStr = fmt.Sprintf("[%s]", t.TaskMetadata.Priority)
				}
			}

			dueStr := "            "
			if t.TaskMetadata.DueDate != "" {
				if denote.IsOverdue(t.TaskMetadata.DueDate) && t.TaskMetadata.Status != denote.TaskStatusDone {
					dueStr = overdueColor.Sprintf("[%s]", t.TaskMetadata.DueDate)
				} else {
					dueStr = fmt.Sprintf("[%s]", t.TaskMetadata.DueDate)
				}
			}

			title := t.Title
			if t.TaskMetadata.Recur != "" {
				title = "↻ " + title
			}
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			areaStr := ""
			if t.TaskMetadata.Area != "" {
				areaStr = t.TaskMetadata.Area
			}

			projectName := ""
			if t.TaskMetadata.ProjectID != "" {
				if name, ok := projectNames[t.TaskMetadata.ProjectID]; ok && name != "" {
					projectName = "→ " + name
				} else {
					projectName = "→ " + t.TaskMetadata.ProjectID
				}
			}

			line := fmt.Sprintf("%3d %s %s %s  %-50s %-10s %s",
				t.IndexID,
				statusIcon,
				priorityStr,
				dueStr,
				title,
				areaStr,
				projectName,
			)

			fmt.Println(line)
		}

		return nil
	}

	return cmd
}

func taskBatchUpdateCommand(cfg *config.Config) *Command {
	var (
		whereClause string
		priority    string
		due         string
		area        string
		project     string
		estimate    int
		status      string
		recur       string
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
	cmd.Flags.StringVar(&recur, "recur", "", "Set recurrence (use 'none' to clear)")
	cmd.Flags.BoolVar(&preview, "preview", false, "Preview changes without applying them")

	cmd.Run = func(c *Command, args []string) error {
		if whereClause == "" {
			return fmt.Errorf("--where clause required\n\nExample:\n  atask batch-update --where \"status:open AND due:past\" --status paused")
		}

		if priority == "" && due == "" && area == "" && project == "" && estimate == -1 && status == "" && recur == "" {
			return fmt.Errorf("at least one field to update must be specified (--priority, --due, --area, --project, --estimate, --status, or --recur)")
		}

		ast, err := query.Parse(whereClause)
		if err != nil {
			return fmt.Errorf("failed to parse --where clause: %v", err)
		}

		scanner := denote.NewScanner(cfg.NotesDirectory)
		allTasks, err := scanner.FindTasks()
		if err != nil {
			return fmt.Errorf("failed to find tasks: %v", err)
		}

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

		fmt.Printf("Found %d matching task(s):\n\n", len(matchingTasks))
		for _, t := range matchingTasks {
			fmt.Printf("  %d: %s\n", t.IndexID, t.Title)
		}
		fmt.Println()

		var parsedDue string
		if due != "" {
			parsedDue, err = denote.ParseNaturalDate(due)
			if err != nil {
				return fmt.Errorf("invalid due date: %v", err)
			}
		}

		var recurPattern string
		var clearRecur bool
		if recur != "" {
			if strings.ToLower(recur) == "none" {
				clearRecur = true
			} else {
				var err error
				recurPattern, err = recurrence.ParsePattern(recur)
				if err != nil {
					return fmt.Errorf("invalid recurrence pattern: %v", err)
				}
			}
		}

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
		if clearRecur {
			changes = append(changes, "recur → (cleared)")
		} else if recurPattern != "" {
			changes = append(changes, fmt.Sprintf("recur → %s", recurPattern))
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
				projectNum, err := strconv.Atoi(project)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid project ID: %s (must be numeric)\n", project)
					continue
				}
				p, err := task.FindProjectByID(cfg.NotesDirectory, projectNum)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Project %d not found\n", projectNum)
					continue
				}
				t.TaskMetadata.ProjectID = strconv.Itoa(p.IndexID)
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
			if clearRecur {
				t.TaskMetadata.Recur = ""
				changed = true
			} else if recurPattern != "" {
				t.TaskMetadata.Recur = recurPattern
				changed = true
			}

			if changed {
				if err := task.UpdateTaskFile(t.FilePath, t); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to update task %d: %v\n", t.IndexID, err)
					continue
				}
				updated++

				if status == denote.TaskStatusDone {
					if err := handleRecurrence(cfg, t); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to create recurring task for ID %d: %v\n", t.IndexID, err)
					}
				}
			}
		}

		fmt.Printf("✓ Updated %d task(s)\n", updated)
		return nil
	}

	return cmd
}

// handleRecurrence checks if a completed task has a recurrence pattern and creates the next instance.
func handleRecurrence(cfg *config.Config, t *denote.Task) error {
	if t.TaskMetadata.Recur == "" || t.TaskMetadata.DueDate == "" {
		return nil
	}

	currentDue, err := time.ParseInLocation("2006-01-02", t.TaskMetadata.DueDate, time.Now().Location())
	if err != nil {
		return fmt.Errorf("failed to parse due date %q: %w", t.TaskMetadata.DueDate, err)
	}

	nextDue, err := recurrence.NextDueDate(t.TaskMetadata.Recur, currentDue)
	if err != nil {
		return fmt.Errorf("failed to compute next due date: %w", err)
	}

	newDueStr := nextDue.Format("2006-01-02")

	newTask, err := task.CloneTaskForRecurrence(cfg.NotesDirectory, t, newDueStr)
	if err != nil {
		return fmt.Errorf("failed to clone task: %w", err)
	}

	if !globalFlags.Quiet {
		fmt.Printf("↻ Created recurring task ID %d: %s (due %s)\n",
			newTask.IndexID, newTask.Title, newDueStr)
	}

	return nil
}
