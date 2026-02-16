package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/atask/internal/denote"
)

// FieldRenderer handles consistent field display
type FieldRenderer struct {
	labelStyle lipgloss.Style
	valueStyle lipgloss.Style
	emptyStyle lipgloss.Style
}

// NewFieldRenderer creates a field renderer with default styles
func NewFieldRenderer() *FieldRenderer {
	return &FieldRenderer{
		labelStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("246")),
		valueStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		emptyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
	}
}

// RenderField renders a single field with consistent formatting
func (fr *FieldRenderer) RenderField(label, value, emptyText string, isEditing bool, editBuffer string) string {
	return fr.RenderFieldWithCursor(label, value, emptyText, isEditing, editBuffer, len(editBuffer))
}

// RenderFieldWithCursor renders a single field with cursor at specified position
func (fr *FieldRenderer) RenderFieldWithCursor(label, value, emptyText string, isEditing bool, editBuffer string, cursor int) string {
	displayValue := value
	if value == "" {
		displayValue = fr.emptyStyle.Render(emptyText)
	} else {
		displayValue = fr.valueStyle.Render(value)
	}

	if isEditing {
		if cursor < len(editBuffer) {
			displayValue = fmt.Sprintf("%s█%s", editBuffer[:cursor], editBuffer[cursor:])
		} else {
			displayValue = fmt.Sprintf("%s█", editBuffer)
		}
	}

	return fmt.Sprintf("%s %s", 
		fr.labelStyle.Render(fmt.Sprintf("%-12s:", label)),
		displayValue,
	)
}

// RenderPriority renders priority with appropriate styling
func (fr *FieldRenderer) RenderPriority(priority string, isEditing bool, editBuffer string) string {
	return fr.RenderPriorityWithCursor(priority, isEditing, editBuffer, len(editBuffer))
}

// RenderPriorityWithCursor renders priority with cursor at specified position
func (fr *FieldRenderer) RenderPriorityWithCursor(priority string, isEditing bool, editBuffer string, cursor int) string {
	if isEditing {
		return fr.RenderFieldWithCursor("Priority", "", "none", true, editBuffer, cursor)
	}

	var style lipgloss.Style
	switch priority {
	case "p1":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	case "p2":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case "p3":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	default:
		return fr.RenderField("Priority", "", "none", false, "")
	}

	return fmt.Sprintf("%s %s",
		fr.labelStyle.Render("Priority    :"),
		style.Render(priority),
	)
}

// RenderStatus renders status with appropriate symbol
func (fr *FieldRenderer) RenderStatus(status string, isEditing bool, editBuffer string) string {
	return fr.RenderStatusWithCursor(status, isEditing, editBuffer, len(editBuffer))
}

// RenderStatusWithCursor renders status with cursor at specified position
func (fr *FieldRenderer) RenderStatusWithCursor(status string, isEditing bool, editBuffer string, cursor int) string {
	if isEditing {
		return fr.RenderFieldWithCursor("Status", "", status, true, editBuffer, cursor)
	}

	symbol := StatusSymbolOpen
	var style lipgloss.Style

	switch status {
	case denote.TaskStatusDone:
		symbol = StatusSymbolDone
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("70"))
	case denote.TaskStatusPaused:
		symbol = StatusSymbolPaused
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	case denote.TaskStatusDelegated:
		symbol = StatusSymbolDelegated
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	case denote.TaskStatusDropped:
		symbol = StatusSymbolDropped
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	}

	return fmt.Sprintf("%s %s %s",
		fr.labelStyle.Render("Status      :"),
		style.Render(symbol),
		style.Render(status),
	)
}

// RenderTags renders tags as a space-separated list
func (fr *FieldRenderer) RenderTags(tags []string, isEditing bool, editBuffer string) string {
	return fr.RenderTagsWithCursor(tags, isEditing, editBuffer, len(editBuffer))
}

// RenderTagsWithCursor renders tags with cursor at specified position
func (fr *FieldRenderer) RenderTagsWithCursor(tags []string, isEditing bool, editBuffer string, cursor int) string {
	if isEditing {
		return fr.RenderFieldWithCursor("Tags", "", "none", true, editBuffer, cursor)
	}

	if len(tags) == 0 {
		return fr.RenderField("Tags", "", "none", false, "")
	}

	return fr.RenderField("Tags", strings.Join(tags, " "), "", false, "")
}

// RenderDueDate renders due date with overdue highlighting
func (fr *FieldRenderer) RenderDueDate(dueDate string, isEditing bool, editBuffer string) string {
	return fr.RenderDueDateWithCursor(dueDate, isEditing, editBuffer, len(editBuffer))
}

// RenderDueDateWithCursor renders due date with cursor at specified position
func (fr *FieldRenderer) RenderDueDateWithCursor(dueDate string, isEditing bool, editBuffer string, cursor int) string {
	if isEditing {
		return fr.RenderFieldWithCursor("Due Date", "", "not set", true, editBuffer, cursor)
	}

	if dueDate == "" {
		return fr.RenderField("Due Date", "", "not set", false, "")
	}

	// Check if overdue
	if denote.IsOverdue(dueDate) {
		return fmt.Sprintf("%s %s",
			fr.labelStyle.Render("Due Date    :"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(dueDate + " (overdue)"),
		)
	}

	return fr.RenderField("Due Date", dueDate, "", false, "")
}