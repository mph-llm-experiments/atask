package denote

import (
	"strings"
	"time"

	"github.com/mph-llm-experiments/acore"
)

// File represents a lightweight view of a task/project file for list display.
// Used by the TUI and scanner for browsing without loading full metadata.
type File struct {
	ID      string    `json:"id"`
	Title   string    `json:"-"`
	Tags    []string  `json:"tags,omitempty"`
	Path    string    `json:"path,omitempty"`
	ModTime time.Time `json:"-"`
	Type    string    `json:"type,omitempty"`
}

// IsTask checks if the file is a task
func (f *File) IsTask() bool {
	return f.Type == TypeTask
}

// IsProject checks if the file is a project
func (f *File) IsProject() bool {
	return f.Type == TypeProject
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

// TaskMetadata holds domain-specific task fields.
// Common fields (ID, Title, IndexID, Type, Tags, Created, Modified,
// RelatedPeople, RelatedTasks, RelatedIdeas) come from embedded acore.Entity.
type TaskMetadata struct {
	Status    string `yaml:"status,omitempty" json:"status,omitempty"`
	Priority  string `yaml:"priority,omitempty" json:"priority,omitempty"`
	DueDate   string `yaml:"due_date,omitempty" json:"due_date,omitempty"`
	StartDate string `yaml:"start_date,omitempty" json:"start_date,omitempty"`
	TodayDate string `yaml:"today_date,omitempty" json:"today_date,omitempty"`
	Estimate  int    `yaml:"estimate,omitempty" json:"estimate,omitempty"`
	ProjectID string `yaml:"project_id,omitempty" json:"project_id,omitempty"`
	Area      string `yaml:"area,omitempty" json:"area,omitempty"`
	Assignee  string `yaml:"assignee,omitempty" json:"assignee,omitempty"`
	Recur     string `yaml:"recur,omitempty" json:"recur,omitempty"`
}

// ProjectMetadata holds domain-specific project fields.
// Common fields come from embedded acore.Entity.
type ProjectMetadata struct {
	Status    string `yaml:"status,omitempty" json:"status,omitempty"`
	Priority  string `yaml:"priority,omitempty" json:"priority,omitempty"`
	DueDate   string `yaml:"due_date,omitempty" json:"due_date,omitempty"`
	StartDate string `yaml:"start_date,omitempty" json:"start_date,omitempty"`
	Area      string `yaml:"area,omitempty" json:"area,omitempty"`
}

// Task combines acore.Entity with task-specific metadata.
type Task struct {
	acore.Entity `yaml:",inline"`
	TaskMetadata `yaml:",inline"`
	ModTime      time.Time `yaml:"-" json:"-"`
	Content      string    `yaml:"-" json:"-"`
}

// Project combines acore.Entity with project-specific metadata.
type Project struct {
	acore.Entity    `yaml:",inline"`
	ProjectMetadata `yaml:",inline"`
	ModTime         time.Time `yaml:"-" json:"-"`
	Content         string    `yaml:"-" json:"-"`
}

// FileFromTask constructs a File view from a Task.
func FileFromTask(t *Task) File {
	return File{
		ID:      t.ID,
		Title:   t.Title,
		Tags:    t.Tags,
		Path:    t.FilePath,
		ModTime: t.ModTime,
		Type:    TypeTask,
	}
}

// FileFromProject constructs a File view from a Project.
func FileFromProject(p *Project) File {
	return File{
		ID:      p.ID,
		Title:   p.Title,
		Tags:    p.Tags,
		Path:    p.FilePath,
		ModTime: p.ModTime,
		Type:    TypeProject,
	}
}

// IsTaggedForToday checks if the task is tagged for today
func (t *Task) IsTaggedForToday() bool {
	if t.TaskMetadata.TodayDate == "" {
		return false
	}
	today := time.Now().Format("2006-01-02")
	return t.TaskMetadata.TodayDate == today
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
	TypeAction  = "action"

	// Action statuses
	ActionPending  = "pending"
	ActionApproved = "approved"
	ActionExecuted = "executed"
	ActionFailed   = "failed"
	ActionRejected = "rejected"

	// Valid action types
	ActionTypeTaskCreate   = "task_create"
	ActionTypeTaskUpdate   = "task_update"
	ActionTypeIdeaCreate   = "idea_create"
	ActionTypeIdeaUpdate   = "idea_update"
	ActionTypePeopleUpdate = "people_update"
	ActionTypePeopleLog    = "people_log"
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
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
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
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)

	daysUntil := int(dueStart.Sub(nowStart).Hours() / 24)

	return daysUntil >= 0 && daysUntil <= horizonDays
}

// DaysUntilDue returns the number of days until the due date
func DaysUntilDue(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return 0
	}
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

// HasNotBegun returns true if the project has a begin date in the future
func (p *Project) HasNotBegun() bool {
	if p.ProjectMetadata.StartDate == "" {
		return false
	}
	loc := time.Now().Location()
	start, err := time.ParseInLocation("2006-01-02", p.ProjectMetadata.StartDate, loc)
	if err != nil {
		return false
	}
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return start.After(today)
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

// IsValidEstimate checks if an estimate value is valid (Fibonacci)
func IsValidEstimate(estimate int) bool {
	validEstimates := []int{1, 2, 3, 5, 8, 13}
	for _, v := range validEstimates {
		if estimate == v {
			return true
		}
	}
	return false
}

// ActionMetadata holds domain-specific action queue fields.
type ActionMetadata struct {
	ActionType string            `yaml:"action_type" json:"action_type"`
	Status     string            `yaml:"status" json:"status"`
	ProposedAt string            `yaml:"proposed_at" json:"proposed_at"`
	ProposedBy string            `yaml:"proposed_by" json:"proposed_by"`
	Fields     map[string]string `yaml:"fields" json:"fields"`
}

// Action combines acore.Entity with action-specific metadata.
type Action struct {
	acore.Entity   `yaml:",inline"`
	ActionMetadata `yaml:",inline"`
	ModTime        time.Time `yaml:"-" json:"-"`
	Content        string    `yaml:"-" json:"-"`
}

// IsValidActionType checks if an action type is valid
func IsValidActionType(actionType string) bool {
	switch actionType {
	case ActionTypeTaskCreate, ActionTypeTaskUpdate,
		ActionTypeIdeaCreate, ActionTypeIdeaUpdate,
		ActionTypePeopleUpdate, ActionTypePeopleLog:
		return true
	}
	return false
}

// IsValidActionStatus checks if an action status is valid
func IsValidActionStatus(status string) bool {
	switch status {
	case ActionPending, ActionApproved, ActionExecuted, ActionFailed, ActionRejected:
		return true
	}
	return false
}

// NoteMetadata represents general note frontmatter (for legacy compatibility)
type NoteMetadata struct {
	Title   string   `yaml:"title"`
	Type    string   `yaml:"type,omitempty"`
	Created string   `yaml:"created,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}
