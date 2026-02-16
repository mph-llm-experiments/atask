package task

import (
	"fmt"
	"os"

	"github.com/pdxmph/atask/internal/denote"
)

// UpdateTaskFile updates the task metadata in a file
func UpdateTaskFile(path string, metadata denote.TaskMetadata) error {
	// Read the current file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse existing frontmatter to get the content
	fm, err := denote.ParseFrontmatterFile(content)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Write updated content
	newContent, err := denote.WriteFrontmatterFile(metadata, fm.Content)
	if err != nil {
		return fmt.Errorf("failed to write frontmatter: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}