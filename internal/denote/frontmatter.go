package denote

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterResult holds the result of parsing a file with YAML frontmatter.
type FrontmatterResult struct {
	Content string // Body content after frontmatter
}

// ParseFrontmatterFile splits file content into frontmatter and body.
// Returns the body content for display/editing.
func ParseFrontmatterFile(data []byte) (*FrontmatterResult, error) {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return &FrontmatterResult{Content: content}, nil
	}
	rest := content[3:]
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return nil, fmt.Errorf("unterminated frontmatter")
	}
	body := strings.TrimLeft(rest[idx+3:], "\n")
	return &FrontmatterResult{Content: body}, nil
}

// WriteFrontmatterFile creates file content with YAML frontmatter.
// Kept for backward compatibility with code that constructs file bytes directly.
func WriteFrontmatterFile(metadata interface{}, content string) ([]byte, error) {
	// Validate required fields based on type
	switch m := metadata.(type) {
	case NoteMetadata:
		if m.Title == "" {
			return nil, fmt.Errorf("note title is required")
		}
	case *Task:
		if m.Title == "" {
			return nil, fmt.Errorf("task title is required")
		}
		if m.IndexID <= 0 {
			return nil, fmt.Errorf("task index ID must be positive")
		}
	case *Project:
		if m.Title == "" {
			return nil, fmt.Errorf("project title is required")
		}
		if m.IndexID <= 0 {
			return nil, fmt.Errorf("project index ID must be positive")
		}
	}

	// Marshal to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(metadata); err != nil {
		return nil, fmt.Errorf("failed to encode metadata: %w", err)
	}

	fileContent := fmt.Sprintf("---\n%s---\n\n%s", buf.String(), content)

	return []byte(fileContent), nil
}
