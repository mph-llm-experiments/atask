package denote

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mph-llm-experiments/acore"
)

// UpdateTaskStatus updates the status field in a task file.
func UpdateTaskStatus(filepath string, newStatus string) error {
	if !IsValidTaskStatus(newStatus) {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.Status = newStatus
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskPriority updates the priority field in a task file.
func UpdateTaskPriority(filepath string, newPriority string) error {
	if newPriority != "" && !IsValidPriority(newPriority) {
		return fmt.Errorf("invalid priority: %s", newPriority)
	}

	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.Priority = newPriority
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskProjectID updates the project_id field in a task file.
func UpdateTaskProjectID(filepath string, projectID string) error {
	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.ProjectID = projectID
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskDueDate updates the due_date field in a task file.
func UpdateTaskDueDate(filepath string, dueDate string) error {
	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.DueDate = dueDate
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskStartDate updates the start_date field in a task file.
func UpdateTaskStartDate(filepath string, startDate string) error {
	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.StartDate = startDate
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskEstimate updates the estimate field in a task file.
func UpdateTaskEstimate(filepath string, estimate int) error {
	if estimate != 0 && !IsValidEstimate(estimate) {
		return fmt.Errorf("invalid estimate: %d (must be 0, 1, 2, 3, 5, 8, or 13)", estimate)
	}

	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.Estimate = estimate
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskArea updates the area field in a task file.
func UpdateTaskArea(filepath string, area string) error {
	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.Area = area
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// UpdateTaskTags updates the tags field in a task file.
func UpdateTaskTags(filepath string, tags []string) error {
	task, err := ParseTaskFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse task: %w", err)
	}

	task.Tags = tags
	task.Modified = acore.Now()

	return acore.UpdateFrontmatter(filepath, task)
}

// BulkUpdateTaskStatus updates status for multiple tasks.
func BulkUpdateTaskStatus(filepaths []string, newStatus string) error {
	for _, filepath := range filepaths {
		if err := UpdateTaskStatus(filepath, newStatus); err != nil {
			return fmt.Errorf("failed to update %s: %w", filepath, err)
		}
	}
	return nil
}

// UpdateProjectFile updates a project file with new metadata.
func UpdateProjectFile(path string, project *Project) error {
	project.Modified = acore.Now()
	return acore.UpdateFrontmatter(path, project)
}

// AddLogEntry adds a timestamped log entry to a task file.
func AddLogEntry(filepath string, message string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	frontmatterEnd := -1
	inFrontmatter := false

	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && line == "---" {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == -1 {
		return fmt.Errorf("no frontmatter found in file")
	}

	now := time.Now()
	timestamp := now.Format("[2006-01-02 Mon]")
	logEntry := fmt.Sprintf("%s: %s", timestamp, message)

	var newLines []string
	newLines = append(newLines, lines[:frontmatterEnd+1]...)

	insertPos := frontmatterEnd + 1
	for insertPos < len(lines) && lines[insertPos] == "" {
		insertPos++
	}

	if insertPos < len(lines) {
		newLines = append(newLines, "")
	}

	newLines = append(newLines, logEntry)

	if insertPos < len(lines) {
		newLines = append(newLines, "")
		newLines = append(newLines, lines[insertPos:]...)
	}

	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(filepath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
