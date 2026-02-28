package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mph-llm-experiments/acore"
)

// CreateNote creates a new note file using acore conventions.
func CreateNote(directory, title string, tags []string) (string, error) {
	if title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}

	id := acore.NewID()
	now := acore.Now()

	metadata := NoteMetadata{
		Title:   title,
		Type:    "note",
		Created: now,
		Tags:    tags,
	}

	filename := acore.BuildFilename(id, title, "note")
	path := filepath.Join(directory, filename)

	store, name := storeAndName(path)
	if err := acore.WriteFile(store, name, &metadata, ""); err != nil {
		return "", fmt.Errorf("failed to create note: %w", err)
	}

	return path, nil
}

// BuildFilename builds an acore filename from components.
func BuildFilename(id, title, entityType string) string {
	return acore.BuildFilename(id, title, entityType)
}

// RenameFileForType renames a file to reflect a new type tag.
func RenameFileForType(oldPath string, newType string) (string, error) {
	dir := filepath.Dir(oldPath)
	base := filepath.Base(oldPath)

	// Try acore format first: {ulid}--{slug}__{type}.md
	id, slug, _, err := acore.ParseFilename(base)
	if err != nil {
		// Try legacy Denote format
		parts := strings.SplitN(base, "--", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid filename format")
		}
		id = parts[0]
		remainingPart := strings.TrimSuffix(parts[1], ".md")
		slugParts := strings.SplitN(remainingPart, "__", 2)
		slug = slugParts[0]
	}

	newFilename := fmt.Sprintf("%s--%s__%s.md", id, slug, newType)
	newPath := filepath.Join(dir, newFilename)

	if newPath == oldPath {
		return oldPath, nil
	}

	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("target file already exists: %s", newPath)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	return newPath, nil
}
