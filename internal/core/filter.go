package core

import (
	"github.com/pdxmph/atask/internal/denote"
)

// FilterOptions represents options for filtering tasks
type FilterOptions struct {
	Status    string
	Area      string
	ProjectID string
	Priority  string
	Overdue   bool
	DueToday  bool
	DueWeek   bool
}

// ApplyFilters applies multiple filters to a task list
func ApplyFilters(tasks []*denote.Task, opts FilterOptions) []*denote.Task {
	filtered := tasks

	// Apply status filter
	if opts.Status != "" {
		filtered = denote.FilterTasks(filtered, "status", opts.Status)
	}

	// Apply area filter
	if opts.Area != "" {
		filtered = denote.FilterTasks(filtered, "area", opts.Area)
	}

	// Apply project filter
	if opts.ProjectID != "" {
		filtered = denote.FilterTasks(filtered, "project", opts.ProjectID)
	}

	// Apply priority filter
	if opts.Priority != "" {
		filtered = denote.FilterTasks(filtered, "priority", opts.Priority)
	}

	// Apply overdue filter
	if opts.Overdue {
		filtered = denote.FilterTasks(filtered, "overdue", "")
	}

	// Apply due today filter
	if opts.DueToday {
		filtered = denote.FilterTasks(filtered, "today", "")
	}

	// Apply due this week filter
	if opts.DueWeek {
		filtered = denote.FilterTasks(filtered, "week", "")
	}

	return filtered
}

// GetProjectTasks returns all tasks associated with a project
func GetProjectTasks(tasks []*denote.Task, projectID string) []*denote.Task {
	return denote.FilterTasks(tasks, "project", projectID)
}

// GetAreaTasks returns all tasks in a specific area
func GetAreaTasks(tasks []*denote.Task, area string) []*denote.Task {
	return denote.FilterTasks(tasks, "area", area)
}

// GetActiveTasks returns all active (not done/dropped) tasks
func GetActiveTasks(tasks []*denote.Task) []*denote.Task {
	return denote.FilterTasks(tasks, "active", "")
}

// GetOpenTasks returns all open tasks
func GetOpenTasks(tasks []*denote.Task) []*denote.Task {
	return denote.FilterTasks(tasks, "open", "")
}

// GetDoneTasks returns all done tasks
func GetDoneTasks(tasks []*denote.Task) []*denote.Task {
	return denote.FilterTasks(tasks, "done", "")
}