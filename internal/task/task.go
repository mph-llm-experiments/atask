package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pdxmph/atask/internal/denote"
	"gopkg.in/yaml.v3"
)

// CreateTask creates a new task file with YAML frontmatter
func CreateTask(dir, title, content string, tags []string, area string) (*denote.Task, error) {
	// Get ID counter
	counter, err := denote.GetIDCounter(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get ID counter: %w", err)
	}

	// Get next index ID
	indexID, err := counter.NextIndexID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next index ID: %w", err)
	}

	// Generate Denote ID
	now := time.Now()
	denoteID := now.Format("20060102T150405")

	// Create slug from title
	slug := titleToSlug(title)

	// Ensure "task" tag is included
	if !contains(tags, "task") {
		tags = append([]string{"task"}, tags...)
	}

	// Build filename
	tagStr := ""
	if len(tags) > 0 {
		tagStr = "__" + strings.Join(tags, "_")
	}
	filename := fmt.Sprintf("%s--%s%s.md", denoteID, slug, tagStr)
	filepath := filepath.Join(dir, filename)

	// Create task metadata
	metadata := denote.TaskMetadata{
		Title:   title,
		IndexID: indexID,
		Type:    denote.TypeTask,
		Status:  denote.TaskStatusOpen,
		Area:    area,
	}

	// Build content with frontmatter
	var builder strings.Builder
	builder.WriteString("---\n")
	
	// Marshal metadata to YAML
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	builder.Write(yamlData)
	builder.WriteString("---\n\n")
	
	if content != "" {
		builder.WriteString(content)
	}

	// Write file
	if err := os.WriteFile(filepath, []byte(builder.String()), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Return the created task
	return denote.ParseTaskFile(filepath)
}

// CreateProject creates a new project file with YAML frontmatter
func CreateProject(dir, title, content string, tags []string) (*denote.Project, error) {
	// Get ID counter
	counter, err := denote.GetIDCounter(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get ID counter: %w", err)
	}

	// Get next index ID
	indexID, err := counter.NextIndexID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next index ID: %w", err)
	}

	// Generate Denote ID
	now := time.Now()
	denoteID := now.Format("20060102T150405")

	// Create slug from title
	slug := titleToSlug(title)

	// Ensure "project" tag is included
	if !contains(tags, "project") {
		tags = append([]string{"project"}, tags...)
	}

	// Build filename
	tagStr := ""
	if len(tags) > 0 {
		tagStr = "__" + strings.Join(tags, "_")
	}
	filename := fmt.Sprintf("%s--%s%s.md", denoteID, slug, tagStr)
	filepath := filepath.Join(dir, filename)

	// Create project metadata
	metadata := denote.ProjectMetadata{
		Title:   title,
		IndexID: indexID,
		Type:    denote.TypeProject,
		Status:  denote.ProjectStatusActive,
	}

	// Build content with frontmatter
	var builder strings.Builder
	builder.WriteString("---\n")
	
	// Marshal metadata to YAML
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	builder.Write(yamlData)
	builder.WriteString("---\n\n")
	
	if content != "" {
		builder.WriteString(content)
	}

	// Write file
	if err := os.WriteFile(filepath, []byte(builder.String()), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Return the created project
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

// FindProjectByDenoteID finds a project by its Denote ID
func FindProjectByDenoteID(dir string, denoteID string) (*denote.Project, error) {
	scanner := denote.NewScanner(dir)
	projects, err := scanner.FindProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.ID == denoteID {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project with Denote ID %s not found", denoteID)
}

// titleToSlug converts a title to a kebab-case slug
func titleToSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	
	// Replace non-alphanumeric with hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, slug)
	
	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	
	return slug
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