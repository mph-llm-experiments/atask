package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mph-llm-experiments/acore"
)

var (
	// Legacy Denote filename pattern for backward compatibility during migration
	legacyDenotePattern = regexp.MustCompile(`^(\d{8}T\d{6})-{1,2}([^_]+)(?:__(.+))?\.md$`)
)

// ParseTaskFile reads and parses a task file using acore.
func ParseTaskFile(path string) (*Task, error) {
	var task Task
	content, err := acore.ReadFile(path, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}
	task.Content = content
	task.FilePath = path

	// Get file modification time
	if info, err := os.Stat(path); err == nil {
		task.ModTime = info.ModTime()
	}

	// If ID not in frontmatter, extract from filename (legacy Denote files)
	if task.ID == "" {
		base := filepath.Base(path)
		if m := legacyDenotePattern.FindStringSubmatch(base); len(m) > 1 {
			task.ID = m[1]
		}
	}

	// Set defaults per spec
	if task.Status == "" {
		task.Status = TaskStatusOpen
	}
	if task.Type == "" {
		task.Type = TypeTask
	}

	// Ensure relation slices for JSON output
	task.EnsureSlices()

	return &task, nil
}

// ParseProjectFile reads and parses a project file using acore.
func ParseProjectFile(path string) (*Project, error) {
	var project Project
	content, err := acore.ReadFile(path, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project file: %w", err)
	}
	project.Content = content
	project.FilePath = path

	// Get file modification time
	if info, err := os.Stat(path); err == nil {
		project.ModTime = info.ModTime()
	}

	// If ID not in frontmatter, extract from filename (legacy Denote files)
	if project.ID == "" {
		base := filepath.Base(path)
		if m := legacyDenotePattern.FindStringSubmatch(base); len(m) > 1 {
			project.ID = m[1]
		}
	}

	// Set defaults per spec
	if project.Status == "" {
		project.Status = ProjectStatusActive
	}
	if project.Type == "" {
		project.Type = TypeProject
	}

	// Ensure relation slices for JSON output
	project.EnsureSlices()

	return &project, nil
}

// ParseActionFile reads and parses an action file using acore.
func ParseActionFile(path string) (*Action, error) {
	var action Action
	content, err := acore.ReadFile(path, &action)
	if err != nil {
		return nil, fmt.Errorf("failed to parse action file: %w", err)
	}
	action.Content = content
	action.FilePath = path

	if info, err := os.Stat(path); err == nil {
		action.ModTime = info.ModTime()
	}

	if action.ID == "" {
		base := filepath.Base(path)
		if m := legacyDenotePattern.FindStringSubmatch(base); len(m) > 1 {
			action.ID = m[1]
		}
	}

	if action.Status == "" {
		action.Status = ActionPending
	}
	if action.Type == "" {
		action.Type = TypeAction
	}
	if action.Fields == nil {
		action.Fields = make(map[string]string)
	}

	action.EnsureSlices()

	return &action, nil
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
