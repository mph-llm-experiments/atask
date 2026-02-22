package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

var (
	// Base style for normal text - brighter for better readability
	baseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))  // Light gray, almost white
		
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))
		
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")). // Orange for selected
		Bold(true)
		
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))  // Brighter for better readability in menus
		
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
		
	// Task-specific styles
	doneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("70"))  // Green for done tasks
		
	pausedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))  // Dim gray for paused tasks
		
	delegatedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))   // Blue for delegated tasks
		
	droppedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))  // Dark gray for dropped tasks
		
	overdueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red for overdue
		Bold(true)
		
	priorityHighStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Red for p1
		Bold(true)
		
	priorityMediumStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))  // Orange for p2
		
	priorityLowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))  // Lighter gray for p3, better contrast
		
	projectStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("135")). // Purple for projects
		Bold(true)
		
	cyanStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("51"))   // Cyan for active projects
)

func (m Model) renderNormal() string {
	var sections []string
	
	// Header
	header := m.renderHeader()
	sections = append(sections, header)
	
	// File list
	list := m.renderFileList()
	sections = append(sections, list)
	
	// Footer
	footer := m.renderFooter()
	sections = append(sections, footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	// Title shows current mode (always tasks now)
	titleText := "Denote Tasks"
	if m.projectFilter {
		titleText = "Denote Projects"
	}
	title := titleStyle.Render(titleText)
	
	// Filter info
	filterInfo := []string{}
	if m.searchQuery != "" {
		if strings.HasPrefix(m.searchQuery, "#") {
			filterInfo = append(filterInfo, fmt.Sprintf("Tag: %s", strings.TrimPrefix(m.searchQuery, "#")))
		} else {
			filterInfo = append(filterInfo, fmt.Sprintf("Search: %s", m.searchQuery))
		}
	}
	if m.areaFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("Area: %s", m.areaFilter))
	}
	if m.priorityFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("Priority: %s", m.priorityFilter))
	}
	if m.stateFilter != "" {
		filterInfo = append(filterInfo, fmt.Sprintf("State: %s", m.stateFilter))
	}
	if m.looseFilter {
		filterInfo = append(filterInfo, "Loose")
	}
	if m.soonFilter {
		filterInfo = append(filterInfo, fmt.Sprintf("Soon: %dd", m.config.SoonHorizon))
	}
	if m.todayFilter {
		filterInfo = append(filterInfo, "Due today")
	}
	
	// Sort info
	sortInfo := fmt.Sprintf(SortFormatString, m.sortBy)
	if m.reverseSort {
		sortInfo += " ↓"
	} else {
		sortInfo += " ↑"
	}
	
	// Status line - count tasks and projects separately
	var taskCount, projectCount int
	for _, f := range m.filtered {
		if f.IsProject() {
			projectCount++
		} else {
			taskCount++
		}
	}
	var status string
	if m.projectFilter {
		status = fmt.Sprintf("%d projects", projectCount)
	} else if projectCount > 0 {
		status = fmt.Sprintf("%d tasks / %d projects", taskCount, projectCount)
	} else {
		status = fmt.Sprintf("%d tasks", taskCount)
	}
	if len(filterInfo) > 0 {
		status += " | " + strings.Join(filterInfo, " | ")
	}
	status += " | " + sortInfo
	if m.statusMsg != "" {
		status += " | " + m.statusMsg
	}
	
	statusLine := statusStyle.Render(status)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, statusLine, "")
}

func (m Model) renderFileList() string {
	if len(m.filtered) == 0 {
		msg := "No tasks found"
		if m.projectFilter {
			msg = "No projects found"
		}
		return helpStyle.Render(msg)
	}
	
	// Calculate visible range
	visibleHeight := m.height - HeaderFooterHeight // Leave room for header and footer
	if visibleHeight < 1 {
		visibleHeight = DefaultVisibleHeight // Default
	}
	
	start := m.scrollOffset
	// Ensure cursor is visible within viewport
	if m.cursor < start {
		start = m.cursor
	}
	if m.cursor >= start+visibleHeight {
		start = m.cursor - visibleHeight + 1
	}
	end := start + visibleHeight
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	

	var lines []string
	for i := start; i < end; i++ {
		lines = append(lines, m.renderFileLine(i))
	}
	
	return strings.Join(lines, "\n")
}

func (m Model) renderFileLine(index int) string {
	file := m.filtered[index]
	
	// Always render task/project lines with fresh metadata
	if file.IsTask() {
		if task, err := denote.ParseTaskFile(file.Path); err == nil {
			return m.renderTaskLine(index, file, task)
		}
	} else if file.IsProject() {
		if project, err := denote.ParseProjectFile(file.Path); err == nil {
			return m.renderProjectLine(index, file, project)
		}
		// Project without metadata - show debug
		line := fmt.Sprintf("%s %s %-15s [NO METADATA] %-40s", 
			" ", ">", file.ID, truncate(file.Title, 40))
		return baseStyle.Render(line)
	}
	
	// Fallback for debugging - show basic file info
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	line := fmt.Sprintf("%s %-15s [UNKNOWN] %-40s", selector, file.ID, truncate(file.Title, 40))
	
	if index == m.cursor {
		return selectedStyle.Render(line)
	}
	return baseStyle.Render(line)
}

func (m Model) renderTaskLine(index int, file denote.File, task *denote.Task) string {
	// Selection indicator
	selector := " "
	if index == m.cursor {
		selector = ">"
	}

	// Today indicator
	todayIndicator := " "
	if task.IsTaggedForToday() {
		todayIndicator = "★"
	}

	// Format: Status Priority Title (Area) [Due Date]
	status := StatusSymbolOpen // open
	if task.TaskMetadata.Status == denote.TaskStatusDone {
		status = StatusSymbolDone
	} else if task.TaskMetadata.Status == denote.TaskStatusPaused {
		status = StatusSymbolPaused
	} else if task.TaskMetadata.Status == denote.TaskStatusDelegated {
		status = StatusSymbolDelegated
	} else if task.TaskMetadata.Status == denote.TaskStatusDropped {
		status = StatusSymbolDropped
	}
	
	// Priority with color - pad FIRST, then apply color
	priorityStr := "    " // Default: 4 spaces
	switch task.TaskMetadata.Priority {
	case PriorityLevels[0]:
		priorityStr = "[p1]"
	case PriorityLevels[1]:
		priorityStr = "[p2]"
	case PriorityLevels[2]:
		priorityStr = "[p3]"
	}

	// Ensure exactly 4 chars before applying color
	priorityStr = fmt.Sprintf("%-4s", priorityStr)

	// Now apply color to the padded string
	var priority string
	switch task.TaskMetadata.Priority {
	case PriorityLevels[0]:
		priority = priorityHighStyle.Render(priorityStr)
	case PriorityLevels[1]:
		priority = priorityMediumStyle.Render(priorityStr)
	case PriorityLevels[2]:
		priority = priorityLowStyle.Render(priorityStr)
	default:
		priority = priorityStr // Already 4 spaces
	}

	// Estimate
	estimate := "     " // Default empty space for alignment (5 spaces)
	if task.TaskMetadata.Estimate > 0 {
		estimate = fmt.Sprintf("[%3d]", task.TaskMetadata.Estimate)
	}
	
	title := task.Title
	if title == "" {
		title = file.Title
	}
	
	// Check if task has notes by parsing content
	hasNotes := false
	if task.Content != "" {
		if fm, err := denote.ParseFrontmatterFile([]byte(task.Content)); err == nil {
			hasNotes = strings.TrimSpace(fm.Content) != ""
		}
	}
	
	// Add notes indicator to title
	if hasNotes {
		title = "≡ " + title
	}

	// Add recurrence indicator
	if task.TaskMetadata.Recur != "" {
		title = "↻ " + title
	}

	area := ""
	// Only show area if we're not filtering by area
	if task.TaskMetadata.Area != "" && m.areaFilter == "" {
		area = fmt.Sprintf("(%s)", task.TaskMetadata.Area)
	}
	
	// Project name
	projectName := ""
	if task.TaskMetadata.ProjectID != "" {
		// Look up project name by index_id
		for _, f := range m.files {
			if f.IsProject() {
				if proj, err := denote.ParseProjectFile(f.Path); err == nil {
					if strconv.Itoa(proj.IndexID) == task.TaskMetadata.ProjectID {
						projTitle := truncate(proj.Title, 15)
						isActiveProject := (proj.ProjectMetadata.Status == denote.ProjectStatusActive || proj.ProjectMetadata.Status == "")

						if projTitle != "" {
							if isActiveProject {
								projectName = cyanStyle.Render("→ " + projTitle)
							} else {
								projectName = fmt.Sprintf("→ %s", projTitle)
							}
						}
						break
					}
				}
			}
		}
	}
	
	// Due date with consistent width - pad FIRST, then apply color
	var due string
	dateStr := ""
	if task.TaskMetadata.DueDate != "" {
		dateStr = fmt.Sprintf("[%s]", task.TaskMetadata.DueDate)
	}

	// Pad to consistent width BEFORE applying any color
	dateStr = fmt.Sprintf("%-*s", ColumnWidthDueSpaces, dateStr)

	// Now apply color to the padded string
	if task.TaskMetadata.DueDate != "" {
		if denote.IsOverdue(task.TaskMetadata.DueDate) {
			due = overdueStyle.Render(dateStr)
		} else if denote.IsDueSoon(task.TaskMetadata.DueDate, m.config.SoonHorizon) {
			due = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dateStr)
		} else {
			due = dateStr // No color, already padded
		}
	} else {
		due = dateStr // Empty, already padded with spaces
	}
	
	// Tags - use metadata tags (from YAML frontmatter), not filename tags
	var displayTags []string
	for _, tag := range task.Tags {
		if tag != "task" && tag != "project" {
			displayTags = append(displayTags, tag)
		}
	}
	tagStr := ""
	if len(displayTags) > 0 {
		tagStr = fmt.Sprintf("[%s]", strings.Join(displayTags, ", "))
	}
	
	// Build the line by padding each field to consistent width
	// This avoids mixing lipgloss Width() (which handles ANSI codes) with fmt.Sprintf (which doesn't)

	// Pad title, tags, and area to fixed widths using plain padding (no colors on these fields)
	titlePadded := fmt.Sprintf("%-*s", ColumnWidthTitle, truncate(title, ColumnWidthTitle))
	tagsPadded := fmt.Sprintf("%-*s", ColumnWidthTags, truncate(tagStr, ColumnWidthTags))
	areaPadded := fmt.Sprintf("%-*s", ColumnWidthArea, truncate(area, ColumnWidthArea))

	// Build line: selector today status priority estimate due title tags area project
	line := selector + todayIndicator + " " + status + " " + priority + " " + estimate + " " + due + "  " + titlePadded + " " + tagsPadded + " " + areaPadded + " " + projectName

	// Apply styling only to parts that don't already have color
	// Don't wrap the entire line in .Render() as it interferes with ANSI width calculations

	// Color the selector for selected items
	if index == m.cursor {
		selector = selectedStyle.Render(">")
		// Rebuild line with colored selector
		line = selector + todayIndicator + " " + status + " " + priority + " " + estimate + " " + due + "  " + titlePadded + " " + tagsPadded + " " + areaPadded + " " + projectName
	}

	// Return the line without additional .Render() wrapping
	// This preserves the width calculations for the colored priority and due fields
	return line
}

func (m Model) renderProjectLine(index int, file denote.File, project *denote.Project) string {
	// Selection indicator
	selector := " "
	if index == m.cursor {
		selector = ">"
	}
	
	// Use same status indicator style as tasks
	status := "▶" // Project indicator
	isActive := false
	hasNotBegun := project.HasNotBegun()

	switch project.ProjectMetadata.Status {
	case denote.ProjectStatusCompleted:
		status = "●"
	case denote.ProjectStatusPaused:
		status = "◐"
	case denote.ProjectStatusCancelled:
		status = "⨯"
	case denote.ProjectStatusActive, "":
		if hasNotBegun {
			status = "▷" // Not yet begun indicator (outline triangle)
		} else {
			isActive = true
			status = "▶"
		}
	default:
		status = "?"
	}
	
	// Priority - we'll color it later based on active status
	priority := "    " // Default empty space for alignment
	priorityRaw := ""
	switch project.ProjectMetadata.Priority {
	case "p1":
		priorityRaw = "[p1]"
	case "p2":
		priorityRaw = "[p2]"
	case "p3":
		priorityRaw = "[p3]"
	}
	
	title := project.Title
	if title == "" {
		title = file.Title
	}
	
	// Check if project has notes by parsing content
	hasNotes := false
	if project.Content != "" {
		if fm, err := denote.ParseFrontmatterFile([]byte(project.Content)); err == nil {
			hasNotes = strings.TrimSpace(fm.Content) != ""
		}
	}
	
	// Add notes indicator to title
	if hasNotes {
		title = "≡ " + title
	}
	
	// Truncate title first
	titleTruncated := truncate(title, ColumnWidthTitle)
	
	// No special styling for status - will be handled at line level
	
	area := ""
	// Only show area if we're not filtering by area
	if project.ProjectMetadata.Area != "" && m.areaFilter == "" {
		area = fmt.Sprintf("(%s)", project.ProjectMetadata.Area)
	}
	
	// Tags - use metadata tags (from YAML frontmatter), not filename tags
	var displayTags []string
	for _, tag := range project.Tags {
		if tag != "task" && tag != "project" {
			displayTags = append(displayTags, tag)
		}
	}
	tagStr := ""
	if len(displayTags) > 0 {
		tagStr = fmt.Sprintf("[%s]", strings.Join(displayTags, ", "))
	}


	// Build the line - exactly matching task format
	// Format: selector status priority due title tags area project

	// For priority, apply the color now
	if priorityRaw != "" {
		if hasNotBegun {
			priority = pausedStyle.Render(priorityRaw)
		} else {
			switch project.ProjectMetadata.Priority {
			case "p1":
				priority = priorityHighStyle.Render(priorityRaw)
			case "p2":
				priority = priorityMediumStyle.Render(priorityRaw)
			case "p3":
				priority = priorityLowStyle.Render(priorityRaw)
			}
		}
	}
	
	// For due date, we need to pad BEFORE coloring
	dueDisplay := ""
	if project.ProjectMetadata.DueDate != "" {
		dateStr := fmt.Sprintf("[%s]", project.ProjectMetadata.DueDate)
		// Pad to 12 chars
		dateStr = fmt.Sprintf("%-12s", dateStr)
		
		if hasNotBegun {
			dueDisplay = pausedStyle.Render(dateStr)
		} else if denote.IsOverdue(project.ProjectMetadata.DueDate) {
			dueDisplay = overdueStyle.Render(dateStr)
		} else if denote.IsDueSoon(project.ProjectMetadata.DueDate, m.config.SoonHorizon) {
			dueDisplay = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(dateStr)
		} else if isActive {
			dueDisplay = cyanStyle.Render(dateStr)
		} else {
			dueDisplay = dateStr
		}
	} else {
		dueDisplay = strings.Repeat(" ", ColumnWidthDueSpaces)
		if isActive {
			dueDisplay = cyanStyle.Render(dueDisplay)
		}
	}
	
	// Pad FIRST (plain text), then apply color
	// fmt.Sprintf counts ANSI escape code bytes as characters, so padding
	// must happen before color is applied
	titlePadded := fmt.Sprintf("%-*s", ColumnWidthTitle, titleTruncated)
	tagsPadded := fmt.Sprintf("%-*s", ColumnWidthTags, truncate(tagStr, ColumnWidthTags))
	areaPadded := fmt.Sprintf("%-*s", ColumnWidthArea, truncate(area, ColumnWidthArea))

	var statusDisplay string
	if isActive {
		statusDisplay = cyanStyle.Render(status)
		titlePadded = cyanStyle.Render(titlePadded)
		tagsPadded = cyanStyle.Render(tagsPadded)
		areaPadded = cyanStyle.Render(areaPadded)
	} else if hasNotBegun {
		statusDisplay = pausedStyle.Render(status)
		titlePadded = pausedStyle.Render(titlePadded)
		tagsPadded = pausedStyle.Render(tagsPadded)
		areaPadded = pausedStyle.Render(areaPadded)
	} else {
		statusDisplay = status
	}

	// Build line matching task format exactly:
	// selector + todayIndicator + " " + status + " " + priority + " " + estimate + " " + due + "  " + titlePadded + " " + tagsPadded + " " + areaPadded + " " + projectName
	todayIndicator := " " // Projects don't have today indicator
	estimate := "     "   // Projects don't have estimates, 5 spaces to match task format

	line := selector + todayIndicator + " " + statusDisplay + " " + priority + " " + estimate + " " + dueDisplay + "  " + titlePadded + " " + tagsPadded + " " + areaPadded + " "

	// Color the selector for selected items (like we do for tasks)
	if index == m.cursor {
		selector = selectedStyle.Render(">")
		// Rebuild line with colored selector
		line = selector + todayIndicator + " " + statusDisplay + " " + priority + " " + estimate + " " + dueDisplay + "  " + titlePadded + " " + tagsPadded + " " + areaPadded + " "
	}

	// Return the line without additional .Render() wrapping
	// This preserves the width calculations for the colored fields
	return line
}

func (m Model) renderFooter() string {
	if m.mode == ModeSearch {
		// Show search input at bottom when in search mode
		prompt := "Search: " + m.searchInput + "█"
		help := MsgFuzzyMatch
		return "\n" + prompt + helpStyle.Render(help)
	}
	
	// Show appropriate hotkeys based on current view
	var help []string
	if m.projectFilter {
		// Project mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:view",
			"c:create project",
			"0-3:priority",
			"d:due date",
			"t:tags",
			"x:delete",
			"E:edit",
			"f:filter",
			"T:tasks",
			"S:sort",
			"?:help",
			"q:quit",
		}
	} else {
		// Task mode hotkeys
		help = []string{
			"j/k:nav",
			"/:search",
			"enter:preview",
			"c:create task",
			"0-3:priority",
			"s:state",
			"y:today",
			"d:due date",
			"t:tags",
			"x:delete",
			"E:edit",
			"l:log",
			"f:filter",
			"P:projects",
			"S:sort",
			"?:help",
			"q:quit",
		}
	}
	
	// Join help items and wrap based on terminal width
	helpText := strings.Join(help, " • ")

	// Use lipgloss to set max width and enable wrapping
	wrapped := helpStyle.
		Width(m.width).
		Render(helpText)

	return "\n" + wrapped
}

func (m Model) renderHelp() string {
	help := `
Denote Tasks - Keyboard Shortcuts

Navigation:
  j/↓     Move down
  k/↑     Move up
  gg      Go to top
  G       Go to bottom

Actions (lowercase):
  Enter   Open task/project details
  c       Create new task/project
  d       Edit due date
  e       Edit estimate (tasks only)
  l       Add log entry (tasks only)
  r       Toggle sort order
  s       Change task state (open/done/etc)
  t       Edit tags
  u       Update task metadata
  x       Delete task/project
  /       Fuzzy search (use #tag for tag search)

Priority:
  0       Clear priority
  1/2/3   Set priority (p1/p2/p3)

Filters & Views (uppercase):
  E       Edit in external editor
  P       Toggle projects view
  T       Toggle tasks view
  S       Sort options menu
  f       Filter menu (area/priority/state/soon)
  
Other:
  ?       Toggle this help
  q       Quit

Press any key to continue...`

	return titleStyle.Render("Help") + help
}



func (m Model) renderPreview() string {
	if m.previewFile == nil {
		return "No file to preview"
	}
	
	// Read file content
	content, err := os.ReadFile(m.previewFile.Path)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	
	// Simple preview - just show the content
	// TODO: Add markdown rendering, scrolling, etc.
	preview := string(content)
	
	header := titleStyle.Render(m.previewFile.Title)
	footer := helpStyle.Render("\nPress q or Esc to close")
	
	return header + "\n\n" + preview + footer
}

func (m Model) renderCreate() string {
	// Always in task mode now
	// Full task creation form
	prompt := titleStyle.Render("Create New Task")
	
	// Build form with all fields
	var form strings.Builder
	form.WriteString("\n")
	
	// Look up project name if we have a project ID
	projectDisplay := m.createProject
	projectHint := MsgPressEnterSelect
	if m.createProject != "" {
		// Find project name by index_id
		for _, f := range m.files {
			if f.IsProject() {
				if proj, err := denote.ParseProjectFile(f.Path); err == nil {
					if strconv.Itoa(proj.IndexID) == m.createProject {
						if proj.Title != "" {
							projectDisplay = proj.Title
						} else if f.Title != "" {
							projectDisplay = f.Title
						}
						break
					}
				}
			}
		}
		projectHint = MsgPressEnterChange
	}
	
	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Title", m.createTitle, "required"},
		{"Priority", m.createPriority, "p1, p2, p3"},
		{"Due Date", m.createDue, "YYYY-MM-DD or natural language"},
		{"Area", m.createArea, "life context"},
		{"Project", projectDisplay, projectHint},
		{"Estimate", m.createEstimate, "numeric value"},
		{"Tags", m.createTags, "space-separated"},
	}
	
	for i, field := range fields {
		if i == m.createField {
			// Active field with cursor
			if field.label == "Area" && m.areaFilter != "" {
				// Area is read-only when filtered
				form.WriteString(fmt.Sprintf("  %s: %s (inherited)\n", field.label, field.value))
			} else if field.label == "Project" {
				// Project is read-only - selection only
				form.WriteString(fmt.Sprintf("→ %s: %s", field.label, field.value))
				if field.value == "" {
					form.WriteString(fmt.Sprintf(" (%s)", field.hint))
				} else {
					form.WriteString(fmt.Sprintf(" [%s]", field.hint))
				}
				form.WriteString("\n")
			} else {
				form.WriteString(fmt.Sprintf("→ %s: %s█", field.label, field.value))
				if field.hint != "" {
					form.WriteString(fmt.Sprintf(" (%s)", field.hint))
				}
				form.WriteString("\n")
			}
		} else {
			// Inactive field
			if field.label == "Area" && m.areaFilter != "" {
				form.WriteString(fmt.Sprintf("  %s: %s (inherited)\n", field.label, field.value))
			} else if field.value != "" {
				form.WriteString(fmt.Sprintf("  %s: %s\n", field.label, field.value))
			} else {
				form.WriteString(fmt.Sprintf("  %s: (%s)\n", field.label, field.hint))
			}
		}
	}
	
	help := helpStyle.Render("\n↑/↓ to navigate, Enter to save, Esc to cancel")
	
	return prompt + baseStyle.Render(form.String()) + help
}

func (m Model) renderCreateTags() string {
	var itemType string
	if m.projectFilter {
		itemType = "Project"
	} else {
		itemType = "Task"
		if m.areaFilter != "" {
			itemType = fmt.Sprintf("Task (Area: %s)", m.areaFilter)
		}
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	titleLine := baseStyle.Render(fmt.Sprintf("\nTitle: %s", m.createTitle))
	input := baseStyle.Render(fmt.Sprintf("\nTags (space-separated): %s█", m.createTags))
	help := helpStyle.Render("\nEnter to create, Esc to go back")
	
	return prompt + titleLine + input + help
}

func (m Model) renderAreaFilter() string {
	prompt := titleStyle.Render("Filter by Area")
	input := baseStyle.Render(fmt.Sprintf("\nArea: %s█", m.searchInput))
	help := helpStyle.Render("\nEnter to apply filter, Esc to cancel, Empty to clear filter")
	
	return prompt + input + help
}

func (m Model) renderSort() string {
	prompt := titleStyle.Render("Sort Options")
	
	// Show current sort
	current := baseStyle.Render(fmt.Sprintf("\nCurrent: %s", m.sortBy))
	if m.reverseSort {
		current = baseStyle.Render(fmt.Sprintf("\nCurrent: %s (descending)", m.sortBy))
	} else {
		current = baseStyle.Render(fmt.Sprintf("\nCurrent: %s (ascending)", m.sortBy))
	}
	
	options := "\n\nSort by:"
	// Always in task mode now
	options += `
  (d) Due date
  (p) Priority
  (j) Project
  (e) Estimate
  (t) Title
  (c) Created date
  (m) Modified date`
	
	options += `
  
  (r) Toggle reverse
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderStateMenu() string {
	if m.cursor >= len(m.filtered) {
		return MsgNoTaskSelected
	}

	file := m.filtered[m.cursor]

	// Check if it's a task or project
	if file.IsTask() {
		// Always read fresh from disk
		task, err := denote.ParseTaskFile(file.Path)
		if err != nil {
			return "Failed to read task"
		}

		prompt := titleStyle.Render("Change Task Status")
		taskInfo := baseStyle.Render(fmt.Sprintf("\nTask: %s", task.Title))
		currentStatus := baseStyle.Render(fmt.Sprintf("\nCurrent status: %s", task.TaskMetadata.Status))

		options := `

Change to:
  (o) Open
  (p) Paused
  (d) Done
  (e) Delegated
  (r) Dropped

  Esc to cancel`

		return prompt + taskInfo + currentStatus + helpStyle.Render(options)
	} else if file.IsProject() {
		// Project status menu
		project, err := denote.ParseProjectFile(file.Path)
		if err != nil {
			return "Failed to read project"
		}

		prompt := titleStyle.Render("Change Project Status")
		projectInfo := baseStyle.Render(fmt.Sprintf("\nProject: %s", project.Title))
		currentStatus := baseStyle.Render(fmt.Sprintf("\nCurrent status: %s", project.ProjectMetadata.Status))

		options := `

Change to:
  (a) Active
  (c) Completed
  (p) Paused
  (x) Cancelled

  Esc to cancel`

		return prompt + projectInfo + currentStatus + helpStyle.Render(options)
	}

	return "Unknown item type"
}

func (m Model) renderConfirmDelete() string {
	// Handle project deletion from project view
	if m.viewingProject != nil && m.projectViewTab == 0 && m.mode == ModeConfirmDelete {
		prompt := titleStyle.Render("Confirm Project Deletion")
		
		warning := baseStyle.Render(fmt.Sprintf("\nAre you sure you want to delete project: %s?", m.viewingProject.Title))
		
		// Show affected tasks if any
		affectedInfo := ""
		if len(m.affectedTasks) > 0 {
			affectedInfo = fmt.Sprintf("\n\n⚠️  This will affect %d task(s):", len(m.affectedTasks))
			for i, task := range m.affectedTasks {
				if i < 10 { // Show first 10 tasks
					affectedInfo += fmt.Sprintf("\n  • %s", task.Title)
				} else if i == 10 {
					affectedInfo += fmt.Sprintf("\n  ... and %d more", len(m.affectedTasks)-10)
					break
				}
			}
			affectedInfo += "\n\nThe project_id will be removed from these tasks."
		}
		
		options := `

  (y) Yes, delete project and clear task associations
  (n) No, cancel
  
  This action cannot be undone!`
		
		dangerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		
		return prompt + warning + affectedInfo + "\n" + dangerStyle.Render(options)
	}
	
	// Handle task deletion from project view
	if m.viewingProject != nil && m.projectViewTab == 1 && len(m.projectTasks) > 0 && m.mode == ModeConfirmDelete {
		task := m.projectTasks[m.projectTasksCursor]
		prompt := titleStyle.Render("Confirm Delete")
		warning := baseStyle.Render("\nAre you sure you want to delete this task?")
		fileName := baseStyle.Render(fmt.Sprintf("\n\nTask: %s", task.Title))
		
		options := `

  (y) Yes, delete
  (n) No, cancel
  
  This action cannot be undone!`
		
		dangerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		
		return prompt + warning + fileName + "\n" + dangerStyle.Render(options)
	}
	
	// Handle normal deletion
	if m.cursor >= len(m.filtered) {
		return "No item selected"
	}
	
	file := m.filtered[m.cursor]
	
	prompt := titleStyle.Render("Confirm Delete")
	
	itemType := "file"
	if file.IsTask() {
		itemType = "task"
	} else if file.IsProject() {
		itemType = "project"
	}
	
	warning := baseStyle.Render(fmt.Sprintf("\nAre you sure you want to delete this %s?", itemType))
	fileName := baseStyle.Render(fmt.Sprintf("\n\nFile: %s", file.Title))
	
	options := `

  (y) Yes, delete
  (n) No, cancel
  
  This action cannot be undone!`
	
	dangerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	
	return prompt + warning + fileName + "\n" + dangerStyle.Render(options)
}

func (m Model) renderConfirmClearToday() string {
	prompt := titleStyle.Render("Clear All 'Today' Tags")

	// Count how many tasks are tagged for today
	count := 0
	for _, file := range m.files {
		if file.IsTask() {
			if task, err := denote.ParseTaskFile(file.Path); err == nil {
				if task.IsTaggedForToday() {
					count++
				}
			}
		}
	}

	warning := baseStyle.Render(fmt.Sprintf("\nClear 'today' tag from all %d task(s)?", count))

	options := `

  (y) Yes, clear all today tags
  (n) No, cancel`

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	return prompt + warning + "\n" + warningStyle.Render(options)
}

func (m Model) renderFilterMenu() string {
	prompt := titleStyle.Render("Filter Options")
	
	// Show current filters
	activeFilters := []string{}
	if m.areaFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("Area: %s", m.areaFilter))
	}
	if m.priorityFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("Priority: %s", m.priorityFilter))
	}
	if m.stateFilter != "" {
		activeFilters = append(activeFilters, fmt.Sprintf("State: %s", m.stateFilter))
	}
	if m.looseFilter {
		activeFilters = append(activeFilters, "Loose (no project)")
	}
	if m.soonFilter {
		activeFilters = append(activeFilters, fmt.Sprintf("Soon: %d days", m.config.SoonHorizon))
	}
	if m.todayFilter {
		activeFilters = append(activeFilters, "Due today")
	}
	
	current := "\n\nActive filters:"
	if len(activeFilters) == 0 {
		current += " None"
	} else {
		for _, f := range activeFilters {
			current += "\n  • " + f
		}
	}
	
	// Apply base style to current filters section
	current = baseStyle.Render(current)
	
	options := `

Filter by:
  (a) Area
  (p) Priority
  (s) State
  (l) Loose tasks (toggle) - no project
  (d) Due soon (toggle)
  (t) Due today (toggle)

  (c) Clear all filters

  Esc to close`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderPriorityFilter() string {
	prompt := titleStyle.Render("Filter by Priority")
	
	current := ""
	if m.priorityFilter != "" {
		current = baseStyle.Render(fmt.Sprintf("\n\nCurrent: %s", m.priorityFilter))
	}
	
	options := `

Select priority:
  (1) P1 - High priority
  (2) P2 - Medium priority
  (3) P3 - Low priority
  
  (c) Clear priority filter
  
  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func (m Model) renderStateFilter() string {
	prompt := titleStyle.Render("Filter by State")
	
	current := ""
	if m.stateFilter != "" {
		current = baseStyle.Render(fmt.Sprintf("\n\nCurrent: %s", m.stateFilter))
	}
	
	options := `

Select state:
  (i) Incomplete (everything except done)
  (a) Active (open + delegated)
  (o) Open only
  (p) Paused
  (d) Done
  (e) Delegated only
  (r) Dropped

  (c) Clear state filter

  Esc to cancel`
	
	return prompt + current + helpStyle.Render(options)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func (m Model) renderLogEntry() string {
	if m.loggingFile == nil {
		return MsgNoTaskSelected
	}
	
	title := m.loggingFile.Title
	// Always read fresh from disk
	if task, err := denote.ParseTaskFile(m.loggingFile.Path); err == nil && task.Title != "" {
		title = task.Title
	}
	
	prompt := titleStyle.Render("Add Log Entry")
	taskInfo := baseStyle.Render(fmt.Sprintf("\nTask: %s", title))
	input := baseStyle.Render(fmt.Sprintf("\n\nLog entry: %s█", m.logInput))
	help := helpStyle.Render("\n\nEnter to save, Esc to cancel")
	
	return prompt + taskInfo + input + help
}

func (m Model) renderProjectSelect() string {
	prompt := titleStyle.Render("Select Project")
	
	if len(m.projectSelectList) == 0 {
		return prompt + "\n\n" + helpStyle.Render("No projects found.\n\nPress Esc to go back")
	}
	
	// Build project list
	var lines []string
	
	// Add "None" option at the top to unassign
	selector := " "
	if m.projectSelectCursor == 0 {
		selector = ">"
	}
	noneLine := fmt.Sprintf("%s 0. ✗ (None - unassign from project)", selector)
	if m.projectSelectCursor == 0 {
		lines = append(lines, selectedStyle.Render(noneLine))
	} else {
		lines = append(lines, helpStyle.Render(noneLine))
	}
	
	// Add projects starting from index 1
	for i, project := range m.projectSelectList {
		// Selection indicator (adjust for None option at position 0)
		selector := " "
		if i+1 == m.projectSelectCursor {
			selector = ">"
		}
		
		// Number for quick selection (1-9)
		number := ""
		if i < 9 {
			number = fmt.Sprintf("%d. ", i+1)
		} else {
			number = "   "
		}
		
		// Project status indicator
		status := ""
		switch project.ProjectMetadata.Status {
		case denote.ProjectStatusActive, "":
			status = "●" // Active
		case denote.ProjectStatusCompleted:
			status = "✓" // Completed
		case denote.ProjectStatusPaused:
			status = "⏸" // Paused
		case denote.ProjectStatusCancelled:
			status = "⨯" // Cancelled
		}
		
		// Project title and area
		title := project.Title
		
		area := ""
		if project.ProjectMetadata.Area != "" {
			area = fmt.Sprintf(" (%s)", project.ProjectMetadata.Area)
		}
		
		// Due date
		due := ""
		if project.ProjectMetadata.DueDate != "" {
			due = fmt.Sprintf(" [%s]", project.ProjectMetadata.DueDate)
		}
		
		// Format line
		line := fmt.Sprintf("%s %s%s %s%s%s", selector, number, status, title, area, due)
		
		if i+1 == m.projectSelectCursor {
			lines = append(lines, selectedStyle.Render(line))
		} else if project.ProjectMetadata.Status == denote.ProjectStatusActive || project.ProjectMetadata.Status == "" {
			lines = append(lines, cyanStyle.Render(line))
		} else {
			lines = append(lines, baseStyle.Render(line))
		}
	}
	
	list := strings.Join(lines, "\n")
	
	help := helpStyle.Render("\n\nj/k or ↑/↓: navigate • 0: unassign • 1-9: quick select • Enter: select • Esc: cancel")
	
	return prompt + "\n\n" + list + help
}

func (m Model) renderCreateProject() string {
	itemType := "Project"
	if m.areaFilter != "" {
		itemType = fmt.Sprintf("Project (Area: %s)", m.areaFilter)
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	input := baseStyle.Render(fmt.Sprintf("\nTitle: %s█", m.createTitle))
	help := helpStyle.Render("\nEnter to continue, Esc to cancel")
	
	return prompt + input + help
}

func (m Model) renderCreateProjectTags() string {
	itemType := "Project"
	if m.areaFilter != "" {
		itemType = fmt.Sprintf("Project (Area: %s)", m.areaFilter)
	}
	prompt := titleStyle.Render(fmt.Sprintf("Create New %s", itemType))
	titleLine := baseStyle.Render(fmt.Sprintf("\nTitle: %s", m.createTitle))
	input := baseStyle.Render(fmt.Sprintf("\nTags (space-separated): %s█", m.createTags))
	if m.areaFilter != "" {
		help := helpStyle.Render(fmt.Sprintf("\nArea '%s' will be added automatically\nEnter to create, Esc to go back", m.areaFilter))
		return prompt + titleLine + input + help
	}
	help := helpStyle.Render("\nEnter to create, Esc to go back")
	
	return prompt + titleLine + input + help
}