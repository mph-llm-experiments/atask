package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mph-llm-experiments/acore"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

// CreateTask creates a new task file with YAML frontmatter using acore conventions.
func CreateTask(dir, title, content string, tags []string, area string) (*denote.Task, error) {
	// Get ID counter
	counter, err := acore.NewIndexCounter(dir, "atask")
	if err != nil {
		return nil, fmt.Errorf("failed to get ID counter: %w", err)
	}

	indexID, err := counter.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to get next index ID: %w", err)
	}

	id := acore.NewID()
	now := acore.Now()

	// Ensure "task" tag is included
	if !contains(tags, "task") {
		tags = append([]string{"task"}, tags...)
	}

	task := &denote.Task{}
	task.ID = id
	task.Title = title
	task.IndexID = indexID
	task.Type = denote.TypeTask
	task.Tags = tags
	task.Created = now
	task.Modified = now
	task.Status = denote.TaskStatusOpen
	task.Area = area

	// Build filename and path
	filename := acore.BuildFilename(id, title, "task")
	filepath := dir + "/" + filename
	task.FilePath = filepath

	if err := acore.WriteFile(filepath, task, content); err != nil {
		return nil, fmt.Errorf("failed to write task file: %w", err)
	}

	// Return the created task
	return denote.ParseTaskFile(filepath)
}

// CreateProject creates a new project file with YAML frontmatter using acore conventions.
func CreateProject(dir, title, content string, tags []string) (*denote.Project, error) {
	counter, err := acore.NewIndexCounter(dir, "atask")
	if err != nil {
		return nil, fmt.Errorf("failed to get ID counter: %w", err)
	}

	indexID, err := counter.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to get next index ID: %w", err)
	}

	id := acore.NewID()
	now := acore.Now()

	// Ensure "project" tag is included
	if !contains(tags, "project") {
		tags = append([]string{"project"}, tags...)
	}

	project := &denote.Project{}
	project.ID = id
	project.Title = title
	project.IndexID = indexID
	project.Type = denote.TypeProject
	project.Tags = tags
	project.Created = now
	project.Modified = now
	project.Status = denote.ProjectStatusActive

	filename := acore.BuildFilename(id, title, "project")
	filepath := dir + "/" + filename
	project.FilePath = filepath

	if err := acore.WriteFile(filepath, project, content); err != nil {
		return nil, fmt.Errorf("failed to write project file: %w", err)
	}

	return denote.ParseProjectFile(filepath)
}

// FindTaskByID finds a task by its sequential ID
func FindTaskByID(dir string, id int) (*denote.Task, error) {
	scanner := denote.NewScanner(dir)
	tasks, err := scanner.FindTasks()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.IndexID == id {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task %d not found", id)
}

// FindProjectByID finds a project by its sequential ID
func FindProjectByID(dir string, id int) (*denote.Project, error) {
	scanner := denote.NewScanner(dir)
	projects, err := scanner.FindProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.IndexID == id {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project %d not found", id)
}

// FindTaskByEntityID finds a task by its ULID (or legacy Denote ID)
func FindTaskByEntityID(dir string, entityID string) (*denote.Task, error) {
	scanner := denote.NewScanner(dir)
	tasks, err := scanner.FindTasks()
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == entityID {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task with ID %s not found", entityID)
}

// FindProjectByEntityID finds a project by its ULID (or legacy Denote ID)
func FindProjectByEntityID(dir string, entityID string) (*denote.Project, error) {
	scanner := denote.NewScanner(dir)
	projects, err := scanner.FindProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.ID == entityID {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project with ID %s not found", entityID)
}

// CloneTaskForRecurrence creates a new task based on an existing recurring task
// with a new due date.
func CloneTaskForRecurrence(dir string, original *denote.Task, newDueDate string) (*denote.Task, error) {
	counter, err := acore.NewIndexCounter(dir, "atask")
	if err != nil {
		return nil, fmt.Errorf("failed to get ID counter: %w", err)
	}

	indexID, err := counter.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to get next index ID: %w", err)
	}

	id := acore.NewID()
	now := acore.Now()

	task := &denote.Task{}
	task.ID = id
	task.Title = original.Title
	task.IndexID = indexID
	task.Type = denote.TypeTask
	task.Tags = make([]string, len(original.Tags))
	copy(task.Tags, original.Tags)
	task.Created = now
	task.Modified = now
	task.Status = denote.TaskStatusOpen
	task.Priority = original.TaskMetadata.Priority
	task.DueDate = newDueDate
	task.Estimate = original.TaskMetadata.Estimate
	task.ProjectID = original.TaskMetadata.ProjectID
	task.Area = original.TaskMetadata.Area
	task.Assignee = original.TaskMetadata.Assignee
	task.Recur = original.TaskMetadata.Recur
	// StartDate and TodayDate intentionally left empty

	filename := acore.BuildFilename(id, original.Title, "task")
	filepath := dir + "/" + filename
	task.FilePath = filepath

	// Extract body content
	body := extractBody(original.Content)

	if err := acore.WriteFile(filepath, task, body); err != nil {
		return nil, fmt.Errorf("failed to write cloned task: %w", err)
	}

	return denote.ParseTaskFile(filepath)
}

// extractBody returns the content after the YAML frontmatter
func extractBody(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	rest := content[3:]
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return ""
	}
	return rest[idx+3:]
}

// CreateAction creates a new action file in the queue/ subdirectory.
func CreateAction(dir, title, actionType, proposedBy, body string, fields map[string]string) (*denote.Action, error) {
	queueDir := filepath.Join(dir, "queue")
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}

	counter, err := acore.NewIndexCounter(queueDir, "atask-action")
	if err != nil {
		return nil, fmt.Errorf("failed to get action ID counter: %w", err)
	}

	indexID, err := counter.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to get next action index ID: %w", err)
	}

	id := acore.NewID()
	now := acore.Now()

	action := &denote.Action{}
	action.ID = id
	action.Title = title
	action.IndexID = indexID
	action.Type = denote.TypeAction
	action.Tags = []string{"action"}
	action.Created = now
	action.Modified = now
	action.ActionType = actionType
	action.Status = denote.ActionPending
	action.ProposedAt = now
	action.ProposedBy = proposedBy
	action.Fields = fields

	filename := acore.BuildFilename(id, title, "action")
	fp := filepath.Join(queueDir, filename)
	action.FilePath = fp

	if err := acore.WriteFile(fp, action, body); err != nil {
		return nil, fmt.Errorf("failed to write action file: %w", err)
	}

	return denote.ParseActionFile(fp)
}

// FindActionByID finds an action by its index_id in the queue/ subdirectory.
func FindActionByID(dir string, id int) (*denote.Action, error) {
	scanner := denote.NewScanner(dir)
	actions, err := scanner.FindActions()
	if err != nil {
		return nil, err
	}

	for _, action := range actions {
		if action.IndexID == id {
			return action, nil
		}
	}

	return nil, fmt.Errorf("action %d not found", id)
}

// FindActionByEntityID finds an action by its ULID.
func FindActionByEntityID(dir string, entityID string) (*denote.Action, error) {
	scanner := denote.NewScanner(dir)
	actions, err := scanner.FindActions()
	if err != nil {
		return nil, err
	}

	for _, action := range actions {
		if action.ID == entityID {
			return action, nil
		}
	}

	return nil, fmt.Errorf("action with ID %s not found", entityID)
}

// ArchiveAction moves an action file to the queue/archive/ subdirectory.
func ArchiveAction(dir string, action *denote.Action) error {
	archiveDir := filepath.Join(dir, "queue", "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	newPath := filepath.Join(archiveDir, filepath.Base(action.FilePath))
	if err := os.Rename(action.FilePath, newPath); err != nil {
		return fmt.Errorf("failed to archive action: %w", err)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
