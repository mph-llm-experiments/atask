package denote

import (
	"strings"
	"time"
)

// File represents the basic Denote file structure
type File struct {
	ID      string    `json:"denote_id"`           // Denote timestamp ID (e.g., "20250704T151739")
	Title   string    `json:"-"`                   // Human-readable title from slug (don't serialize, use metadata title)
	Slug    string    `json:"slug,omitempty"`      // Kebab-case title from filename
	Tags    []string  `json:"filename_tags,omitempty"` // Tags from filename
	Path    string    `json:"path,omitempty"`      // Full file path
	ModTime time.Time `json:"-"`                   // File modification time (don't serialize, use Task/Project ModTime)
}

// IsTask checks if the file is a task based on tags
func (f *File) IsTask() bool {
	return f.HasTag("task")
}

// IsProject checks if the file is a project based on tags
func (f *File) IsProject() bool {
	return f.HasTag("project")
}

// HasTag checks if the file has a specific tag
func (f *File) HasTag(tag string) bool {
	for _, t := range f.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// MatchesSearch checks if the file matches a search query using fuzzy matching
func (f *File) MatchesSearch(query string) bool {
	query = strings.ToLower(query)
	
	// Fuzzy search in title
	if fuzzyMatch(strings.ToLower(f.Title), query) {
		return true
	}
	
	// Fuzzy search in slug
	if fuzzyMatch(strings.ToLower(f.Slug), query) {
		return true
	}
	
	// Fuzzy search in tags
	for _, tag := range f.Tags {
		if fuzzyMatch(strings.ToLower(tag), query) {
			return true
		}
	}
	
	return false
}

// MatchesTag checks if the file has a tag matching the query (fuzzy match)
func (f *File) MatchesTag(query string) bool {
	query = strings.ToLower(query)
	
	for _, tag := range f.Tags {
		if fuzzyMatch(strings.ToLower(tag), query) {
			return true
		}
	}
	
	return false
}

// fuzzyMatch performs true fuzzy matching - query letters must appear in order but can be non-consecutive
func fuzzyMatch(text, pattern string) bool {
	if pattern == "" {
		return true
	}
	
	patternIdx := 0
	for _, ch := range text {
		if patternIdx < len(pattern) && ch == rune(pattern[patternIdx]) {
			patternIdx++
		}
	}
	
	return patternIdx == len(pattern)
}

// NoteMetadata represents general note frontmatter
type NoteMetadata struct {
	Title   string   `yaml:"title"`               // Required: Human-readable title
	Type    string   `yaml:"type,omitempty"`      // Optional: "note" (determined by filename)
	Created string   `yaml:"created,omitempty"`   // YYYY-MM-DD format
	Tags    []string `yaml:"tags,omitempty"`      // Optional tags
}

// TaskMetadata represents task-specific frontmatter per spec v2.0.0
type TaskMetadata struct {
	Title     string   `yaml:"title" json:"title"`                      // Required: Human-readable title
	IndexID   int      `yaml:"index_id" json:"index_id"`                // Required: Sequential ID for CLI
	Type      string   `yaml:"type,omitempty" json:"type,omitempty"`    // Optional: "task" (determined by __task tag)
	Status    string   `yaml:"status,omitempty" json:"status,omitempty"` // Default: "open"
	Priority  string   `yaml:"priority,omitempty" json:"priority,omitempty"` // p1, p2, p3
	DueDate   string   `yaml:"due_date,omitempty" json:"due_date,omitempty"` // YYYY-MM-DD format
	StartDate string   `yaml:"start_date,omitempty" json:"start_date,omitempty"` // YYYY-MM-DD format
	TodayDate string   `yaml:"today_date,omitempty" json:"today_date,omitempty"` // YYYY-MM-DD format - tagged for today
	Estimate  int      `yaml:"estimate,omitempty" json:"estimate,omitempty"` // Fibonacci: 1,2,3,5,8,13
	ProjectID string   `yaml:"project_id,omitempty" json:"project_id,omitempty"` // Denote ID of project (v2.0.0)
	Area      string   `yaml:"area,omitempty" json:"area,omitempty"`    // Life context
	Assignee  string   `yaml:"assignee,omitempty" json:"assignee,omitempty"` // Person responsible
	Tags      []string `yaml:"tags,omitempty" json:"tags,omitempty"`    // Additional tags beyond filename
}

// ProjectMetadata represents project-specific frontmatter per spec v2.0.0
type ProjectMetadata struct {
	Title     string   `yaml:"title" json:"title"`                      // Required: Human-readable title
	IndexID   int      `yaml:"index_id" json:"index_id"`                // Required: Sequential ID for CLI
	Type      string   `yaml:"type,omitempty" json:"type,omitempty"`    // Optional: "project" (determined by __project tag)
	Status    string   `yaml:"status,omitempty" json:"status,omitempty"` // Default: "active"
	Priority  string   `yaml:"priority,omitempty" json:"priority,omitempty"` // p1, p2, p3
	DueDate   string   `yaml:"due_date,omitempty" json:"due_date,omitempty"` // YYYY-MM-DD format
	StartDate string   `yaml:"start_date,omitempty" json:"start_date,omitempty"` // YYYY-MM-DD format
	Area      string   `yaml:"area,omitempty" json:"area,omitempty"`    // Life context
	Tags      []string `yaml:"tags,omitempty" json:"tags,omitempty"`    // Additional tags beyond filename
}

// Task combines File info with TaskMetadata
type Task struct {
	File
	TaskMetadata
	ModTime time.Time `json:"modified_at"` // File modification time
	Content string    `json:"-"`           // Don't serialize full content in lists
}

// IsTaggedForToday checks if the task is tagged for today
func (t *Task) IsTaggedForToday() bool {
	if t.TaskMetadata.TodayDate == "" {
		return false
	}
	today := time.Now().Format("2006-01-02")
	return t.TaskMetadata.TodayDate == today
}

// Project combines File info with ProjectMetadata
type Project struct {
	File
	ProjectMetadata
	ModTime time.Time `json:"modified_at"` // File modification time
	Content string    `json:"-"`           // Don't serialize full content in lists
}

// Common status values
const (
	// Task statuses
	TaskStatusOpen      = "open"
	TaskStatusDone      = "done"
	TaskStatusPaused    = "paused"
	TaskStatusDelegated = "delegated"
	TaskStatusDropped   = "dropped"

	// Project statuses
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusPaused    = "paused"
	ProjectStatusCancelled = "cancelled"

	// Priority levels
	PriorityP1 = "p1"
	PriorityP2 = "p2"
	PriorityP3 = "p3"

	// File types
	TypeTask    = "task"
	TypeProject = "project"
)

// IsValidTaskStatus checks if a status is valid for tasks
func IsValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusOpen, TaskStatusDone, TaskStatusPaused, TaskStatusDelegated, TaskStatusDropped:
		return true
	}
	return false
}

// IsValidProjectStatus checks if a status is valid for projects
func IsValidProjectStatus(status string) bool {
	switch status {
	case ProjectStatusActive, ProjectStatusCompleted, ProjectStatusPaused, ProjectStatusCancelled:
		return true
	}
	return false
}

// IsValidPriority checks if a priority is valid
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityP1, PriorityP2, PriorityP3:
		return true
	}
	return false
}

// IsOverdue checks if a task/project is overdue
func IsOverdue(dueDateStr string) bool {
	if dueDateStr == "" {
		return false
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return dueStart.Before(nowStart)
}

// IsDueSoon checks if a task/project is due within the specified number of days
func IsDueSoon(dueDateStr string, horizonDays int) bool {
	if dueDateStr == "" {
		return false
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	// Calculate days until due
	daysUntil := int(dueStart.Sub(nowStart).Hours() / 24)
	
	// Due soon if due today or within horizon days (and not overdue)
	return daysUntil >= 0 && daysUntil <= horizonDays
}

// DaysUntilDue returns the number of days until the due date
func DaysUntilDue(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return 0
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return int(dueStart.Sub(nowStart).Hours() / 24)
}

// IsDueThisWeek checks if a task is due within the next 7 days
func IsDueThisWeek(dueDateStr string) bool {
	days := DaysUntilDue(dueDateStr)
	return days >= 0 && days <= 7
}

// GetParsedStartDate returns the parsed start date
func (t *Task) GetParsedStartDate() *time.Time {
	if t.StartDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", t.StartDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedDueDate returns the parsed due date
func (t *Task) GetParsedDueDate() *time.Time {
	if t.DueDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", t.DueDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedStartDate returns the parsed start date for a project
func (p *Project) GetParsedStartDate() *time.Time {
	if p.StartDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil
	}
	return &parsed
}

// GetParsedDueDate returns the parsed due date for a project
func (p *Project) GetParsedDueDate() *time.Time {
	if p.DueDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", p.DueDate)
	if err != nil {
		return nil
	}
	return &parsed
}