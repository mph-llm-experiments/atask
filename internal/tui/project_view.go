package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

var (
	tabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false)
		
	activeTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("214")).
		BorderForeground(lipgloss.Color("214")).
		Bold(true)
		
	inactiveTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("241"))
)

func (m Model) renderProjectView() string {
	if m.viewingProject == nil {
		return MsgNoProjectSelected
	}
	
	var sections []string
	
	// Title
	title := titleStyle.Render("Project: " + m.viewingProject.Title)
	sections = append(sections, title)
	
	// Tabs
	tabs := m.renderProjectTabs()
	sections = append(sections, tabs)
	
	// Tab content
	var content string
	switch m.projectViewTab {
	case 0:
		content = m.renderProjectMain() // Metadata + tasks
	case 1:
		content = m.renderProjectNotes() // Just the notes content
	}
	sections = append(sections, content)
	
	// Status message or edit prompt - JUST LIKE TASK VIEW
	if m.editingField != "" {
		var prompt string
		if m.editCursor < len(m.editBuffer) {
			prompt = fmt.Sprintf("\n%s %s█%s", m.statusMsg, m.editBuffer[:m.editCursor], m.editBuffer[m.editCursor:])
		} else {
			prompt = fmt.Sprintf("\n%s %s█", m.statusMsg, m.editBuffer)
		}
		sections = append(sections, editingStyle.Render(prompt))
	} else if m.statusMsg != "" {
		sections = append(sections, "\n"+statusStyle.Render(m.statusMsg))
	}
	
	// Footer with hints
	hints := m.getProjectViewHints()
	hintsText := strings.Join(hints, " • ")
	wrapped := hintStyle.
		Width(m.width).
		Render(hintsText)
	sections = append(sections, "\n"+wrapped)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderProjectTabs() string {
	tabs := []string{}
	
	// Tasks tab (main tab) with count
	taskCount := len(m.projectTasks)
	taskLabel := fmt.Sprintf("Tasks (%d)", taskCount)
	if m.projectViewTab == 0 {
		tabs = append(tabs, activeTabStyle.Render(taskLabel))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render(taskLabel))
	}
	
	// Notes tab
	if m.projectViewTab == 1 {
		tabs = append(tabs, activeTabStyle.Render("Notes"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("Notes"))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderProjectMain() string {
	project := m.viewingProject
	meta := project.ProjectMetadata

	var lines []string

	// Project metadata (similar to task view)
	lines = append(lines, m.renderFieldWithHotkey("Title", project.Title, "", "T"))
	
	// Status with color
	statusValue := meta.Status
	if statusValue == "" {
		statusValue = "active"
	}
	var statusColor string
	switch statusValue {
	case denote.ProjectStatusCompleted:
		statusColor = "70" // green
	case denote.ProjectStatusPaused:
		statusColor = "214" // orange
	case denote.ProjectStatusCancelled:
		statusColor = "241" // gray
	default:
		statusColor = "230" // default
	}
	statusStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render(statusValue)
	lines = append(lines, m.renderFieldWithHotkey("Status", statusStyled, "active", "s"))
	
	// Priority with color
	if meta.Priority != "" {
		var priorityColor string
		switch meta.Priority {
		case "p1":
			priorityColor = "196" // red
		case "p2":
			priorityColor = "214" // orange
		case "p3":
			priorityColor = "245" // gray
		default:
			priorityColor = "230"
		}
		priorityStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Bold(true).Render(meta.Priority)
		lines = append(lines, m.renderFieldWithHotkey("Priority", priorityStyled, "none", "p"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Priority", "", "not set", "p"))
	}
	
	// Due Date with overdue highlighting
	if meta.DueDate != "" {
		dueValue := meta.DueDate
		if denote.IsOverdue(meta.DueDate) {
			dueValue = overdueStyle.Render(dueValue + " (OVERDUE!)")
		} else if denote.IsDueThisWeek(meta.DueDate) {
			dueValue = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dueValue + " (this week)")
		}
		lines = append(lines, m.renderFieldWithHotkey("Due Date", dueValue, "not set", "d"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Due Date", "", "not set", "d"))
	}
	
	// Area
	if meta.Area != "" {
		lines = append(lines, m.renderFieldWithHotkey("Area", meta.Area, "not set", "a"))
	} else {
		lines = append(lines, m.renderFieldWithHotkey("Area", "", "not set", "a"))
	}
	
	// Tags
	tagsDisplay := ""
	if len(project.Tags) > 0 {
		tagsDisplay = strings.Join(project.Tags, " ")
	} else {
		tagsDisplay = "(not set)"
	}
	lines = append(lines, m.renderFieldWithHotkey("Tags", tagsDisplay, "not set", "t"))
	
	// Other metadata
	lines = append(lines, m.renderFieldWithHotkey("Begin Date", meta.StartDate, "not set", "B"))
	
	// File info
	lines = append(lines, "")
	lines = append(lines, m.renderFieldWithHotkey("File", m.viewingFile.Path, "", ""))
	lines = append(lines, m.renderFieldWithHotkey("ID", project.ID, "", ""))
	
	// Add horizontal rule
	lines = append(lines, "\n" + strings.Repeat("─", 60))
	
	// Add tasks section
	tasksContent := m.renderProjectTasks()
	lines = append(lines, tasksContent)
	
	return strings.Join(lines, "\n")
}

func (m Model) renderProjectTasks() string {
	if len(m.projectTasks) == 0 {
		return "\n" + helpStyle.Render("No tasks assigned to this project")
	}
	
	var lines []string
	
	// Calculate visible range
	visibleHeight := m.height - ProjectViewHeaderHeight // Leave room for header, tabs, and footer
	if visibleHeight < 1 {
		visibleHeight = ProjectViewHeaderHeight // Default
	}
	
	start := 0
	if m.projectTasksCursor >= visibleHeight {
		start = m.projectTasksCursor - visibleHeight + 1
	}
	
	end := start + visibleHeight
	if end > len(m.projectTasks) {
		end = len(m.projectTasks)
	}
	
	// Render each task
	for i := start; i < end; i++ {
		task := m.projectTasks[i]
		line := m.renderProjectTaskLine(i, task)
		lines = append(lines, line)
	}
	
	return "\n" + strings.Join(lines, "\n")
}

func (m Model) renderProjectTaskLine(index int, task denote.Task) string {
	// Selection indicator
	selector := " "
	if m.projectViewTab == 0 && index == m.projectTasksCursor {
		selector = ">"
	}
	
	// Status - use same symbols as main list
	status := "○" // open
	isDone := false
	isDelegated := false
	isDropped := false
	switch task.TaskMetadata.Status {
	case denote.TaskStatusDone:
		status = "✓"
		isDone = true
	case denote.TaskStatusPaused:
		status = "⏸"
	case denote.TaskStatusDelegated:
		status = "→"
		isDelegated = true
	case denote.TaskStatusDropped:
		status = "⨯"
		isDropped = true
	}
	
	// Priority with color
	priority := "    " // Default empty space for alignment
	switch task.TaskMetadata.Priority {
	case "p1":
		priority = priorityHighStyle.Render("[p1]")
	case "p2":
		priority = priorityMediumStyle.Render("[p2]")
	case "p3":
		priority = priorityLowStyle.Render("[p3]")
	}
	
	// Due date formatting with padding
	due := "            " // 12 spaces for alignment when no date
	isDueSoon := false
	isOverdue := false
	if task.TaskMetadata.DueDate != "" {
		dateStr := fmt.Sprintf("[%s]", task.TaskMetadata.DueDate)
		if denote.IsOverdue(task.TaskMetadata.DueDate) {
			due = overdueStyle.Render(dateStr)
			isOverdue = true
		} else if denote.IsDueSoon(task.TaskMetadata.DueDate, m.config.SoonHorizon) {
			due = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dateStr)
			isDueSoon = true
		} else {
			due = dateStr
		}
		// Pad to consistent width
		due = fmt.Sprintf("%-12s", due)
	}
	
	title := task.Title
	
	// Tags (excluding 'task' and 'project' tags)
	tags := ""
	filteredTags := []string{}
	for _, tag := range task.Tags {
		if tag != "task" && tag != "project" {
			filteredTags = append(filteredTags, tag)
		}
	}
	if len(filteredTags) > 0 {
		tags = strings.Join(filteredTags, " ")
	}
	
	// Area (if not filtering by area)
	area := ""
	if task.TaskMetadata.Area != "" && m.areaFilter == "" {
		area = task.TaskMetadata.Area
	}
	
	// Build the line with proper formatting
	// Format: selector status priority due title tags area
	line := fmt.Sprintf("%s %s %s %s  %-50s %-25s %-10s", 
		selector,
		status,
		priority,
		due,
		truncate(title, ColumnWidthTitle),
		truncate(tags, ColumnWidthTags),
		truncate(area, ColumnWidthArea))
	
	// Apply styling based on state
	if m.projectViewTab == 0 && index == m.projectTasksCursor {
		return selectedStyle.Render(line)
	} else if isDone || isDelegated || isDropped {
		return doneStyle.Render(line)
	} else if isOverdue && !isDueSoon {
		// Only apply full line overdue style if not also "due soon"
		return line
	}
	
	return line
}

func (m Model) renderProjectNotes() string {
	var lines []string
	
	// Add body content
	bodyContent := m.getProjectBodyContent()
	if bodyContent != "" {
		// Wrap long lines for readability
		maxWidth := 80
		if m.width > 0 && m.width < maxWidth {
			maxWidth = m.width - 4
		}
		wrapped := wrapText(bodyContent, maxWidth)
		lines = append(lines, "\n"+wrapped)
	} else {
		lines = append(lines, "\n"+helpStyle.Render("(no notes)"))
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) getProjectBodyContent() string {
	if m.viewingProject != nil {
		// Extract body from content (content after frontmatter)
		content := m.viewingProject.Content
		
		// Find the end of frontmatter
		if strings.HasPrefix(content, "---\n") {
			// Find the closing ---
			parts := strings.SplitN(content, "\n---\n", 3)
			if len(parts) >= 2 {
				body := strings.TrimSpace(parts[1])
				if len(parts) == 3 {
					body = strings.TrimSpace(parts[2])
				}
				return body
			}
		}
		
		// If no frontmatter, return the whole content
		return strings.TrimSpace(content)
	}
	
	return ""
}

func (m Model) getProjectViewHints() []string {
	hints := []string{
		"tab:switch tabs",
		"q/esc:back",
		"E:edit file",
		"n:new task",
	}
	
	if m.projectViewTab == 0 {
		// Main tab (metadata + tasks) hints
		// Metadata editing hints
		hints = append(hints,
			"T:title",
			"p:priority",
			"s:status",
			"d:due date",
			"a:area",
			"t:tags",
		)
		// Task list hints (if there are tasks)
		if len(m.projectTasks) > 0 {
			hints = append(hints,
				"j/k:nav",
				"enter:view task",
				"1/2/3:task priority",
				"S:sort",
				"x:delete task",
			)
		}
		// Project deletion
		hints = append(hints, "X:delete project")
	}
	// Tab 1 (Notes) has no special operations
	
	return hints
}