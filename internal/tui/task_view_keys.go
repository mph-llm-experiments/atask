package tui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

func (m Model) handleTaskViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If we're editing a field, handle input
	if m.editingField != "" {
		switch msg.String() {
		case "esc":
			m.editingField = ""
			m.editBuffer = ""
			m.editCursor = 0
			m.editCursor = 0
			
		case "enter":
			// Save the field
			if m.viewingTask != nil {
				switch m.editingField {
				case "title":
					if err := m.updateTaskField("title", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						m.statusMsg = "Title updated"
					}
				case "priority":
					if m.editBuffer == "" || m.editBuffer == "0" {
						// Clear priority
						if err := m.updateTaskField("priority", ""); err != nil {
							m.statusMsg = fmt.Sprintf(ErrorFormat, err)
						} else {
							m.statusMsg = "Priority removed"
						}
					} else if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						if err := m.updateTaskField("priority", "p"+m.editBuffer); err != nil {
							m.statusMsg = fmt.Sprintf(ErrorFormat, err)
						} else {
							m.statusMsg = fmt.Sprintf("Priority set to p%s", m.editBuffer)
						}
					} else {
						m.statusMsg = "Priority must be 0 (clear), 1, 2, or 3"
					}
				case "status":
					if err := m.updateTaskField("status", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "due":
					if err := m.updateTaskField("due_date", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "area":
					if err := m.updateTaskField("area", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "estimate":
					if err := m.updateTaskField("estimate", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "tags":
					if err := m.updateTaskField("tags", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				}
			} else if m.viewingProject != nil {
				// Handle project updates
				switch m.editingField {
				case "title":
					if err := m.updateProjectField("title", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						m.statusMsg = "Title updated"
					}
				case "priority":
					if m.editBuffer == "" || m.editBuffer == "0" {
						// Clear priority
						if err := m.updateProjectField("priority", ""); err != nil {
							m.statusMsg = fmt.Sprintf(ErrorFormat, err)
						} else {
							m.statusMsg = "Priority removed"
						}
					} else if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						if err := m.updateProjectField("priority", "p"+m.editBuffer); err != nil {
							m.statusMsg = fmt.Sprintf(ErrorFormat, err)
						} else {
							m.statusMsg = fmt.Sprintf("Priority set to p%s", m.editBuffer)
						}
					} else {
						m.statusMsg = "Priority must be 0 (clear), 1, 2, or 3"
					}
				case "status":
					if err := m.updateProjectField("status", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "due":
					if err := m.updateProjectField("due_date", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "area":
					if err := m.updateProjectField("area", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				case "tags":
					if err := m.updateProjectField("tags", m.editBuffer); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					}
				}
			}
			m.editingField = ""
			m.editBuffer = ""
			m.editCursor = 0
			
		case "backspace", "ctrl+h":
			if m.editCursor > 0 && len(m.editBuffer) > 0 {
				m.editBuffer = m.editBuffer[:m.editCursor-1] + m.editBuffer[m.editCursor:]
				m.editCursor--
			}

		case "delete", "ctrl+d":
			if m.editCursor < len(m.editBuffer) {
				m.editBuffer = m.editBuffer[:m.editCursor] + m.editBuffer[m.editCursor+1:]
			}

		case "left", "ctrl+b":
			if m.editCursor > 0 {
				m.editCursor--
			}

		case "right", "ctrl+f":
			if m.editCursor < len(m.editBuffer) {
				m.editCursor++
			}

		case "home", "ctrl+a":
			m.editCursor = 0

		case "end", "ctrl+e":
			m.editCursor = len(m.editBuffer)

		case "ctrl+k":
			// Kill to end of line
			m.editBuffer = m.editBuffer[:m.editCursor]

		case "ctrl+u":
			// Kill to beginning of line
			m.editBuffer = m.editBuffer[m.editCursor:]
			m.editCursor = 0

		case "ctrl+w":
			// Delete word backward
			if m.editCursor > 0 {
				i := m.editCursor - 1
				for i > 0 && m.editBuffer[i-1] == ' ' {
					i--
				}
				for i > 0 && m.editBuffer[i-1] != ' ' {
					i--
				}
				m.editBuffer = m.editBuffer[:i] + m.editBuffer[m.editCursor:]
				m.editCursor = i
			}

		default:
			if len(msg.String()) == 1 {
				m.editBuffer = m.editBuffer[:m.editCursor] + msg.String() + m.editBuffer[m.editCursor:]
				m.editCursor++
			}
		}
		
		return m, nil
	}
	
	// Normal task view navigation
	switch msg.String() {
	case "q", "esc":
		if m.returnToProject && m.viewingProject != nil {
			// Return to project view
			m.mode = ModeProjectView
			m.viewingTask = nil
			file := denote.FileFromProject(m.viewingProject)
			m.viewingFile = &file
			m.returnToProject = false
			// Re-sort project tasks in case metadata changed
			m.loadProjectTasks()
		} else {
			// Return to normal task list
			m.mode = ModeNormal
			m.viewingTask = nil
			m.viewingProject = nil
			m.viewingFile = nil
			// Re-sort the list in case metadata changed
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
		
	case "D":
		// Mark task as done (quick action)
		if m.viewingTask != nil && m.viewingFile != nil {
			if err := m.updateCurrentTaskStatus(denote.TaskStatusDone); err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				recurMsg := m.handleTaskRecurrence(m.viewingFile.Path)
				if recurMsg != "" {
					m.scanFiles()
				}
				m.statusMsg = "Task marked as done" + recurMsg
			}
		}

	case "E":
		// Edit in external editor (uppercase for Edit action)
		if m.config.Editor != "" && m.viewingFile != nil {
			return m, m.editFile(m.viewingFile.Path)
		}
		
	case "l":
		// Log entry - only for tasks
		if m.viewingTask != nil && m.viewingFile != nil {
			m.mode = ModeLogEntry
			m.logInput = ""
			m.loggingFile = m.viewingFile
		}
		
	// Field edit hotkeys
	case "T":
		// Title field (uppercase - different from tags)
		m.editingField = "title"
		if m.viewingTask != nil {
			m.editBuffer = m.viewingTask.Title
		} else if m.viewingProject != nil {
			m.editBuffer = m.viewingProject.Title
		}
		m.editCursor = len(m.editBuffer)
		m.statusMsg = "Enter title:"
		
	case "p":
		m.editingField = "priority"
		m.editBuffer = ""
		m.editCursor = 0
		m.statusMsg = "Enter priority (1/2/3):"
		
	case "s":
		m.editingField = "status"
		m.editBuffer = ""
		m.editCursor = 0
		if m.viewingTask != nil {
			m.statusMsg = "Enter status (open/done/paused/delegated/dropped):"
		} else {
			m.statusMsg = "Enter status (active/completed/paused/cancelled):"
		}
		
	case "d":
		m.editingField = "due"
		m.editBuffer = ""
		m.editCursor = 0
		m.statusMsg = "Enter due date (e.g. 2d, 1w, friday, jan 15, 2024-01-15):"
		
	case "a":
		m.editingField = "area"
		m.editBuffer = ""
		m.editCursor = 0
		m.statusMsg = "Enter area (work/personal/etc):"
		
	case "e":
		// Estimate field (lowercase for action)
		if m.viewingTask != nil {
			m.editingField = "estimate"
			m.editBuffer = ""
			m.editCursor = 0
			m.statusMsg = "Enter time estimate (1/2/3/5/8/13):"
		}
		
	case "j":
		// Project selection - only for tasks
		if m.viewingTask != nil {
			// Load projects and switch to selection mode
			m.loadProjectsForSelection()
			if len(m.projectSelectList) > 0 {
				m.projectSelectFor = "update"
				m.projectSelectTask = m.viewingTask
				m.mode = ModeProjectSelect
			} else {
				m.statusMsg = "No projects found"
			}
		}
		
	case "t":
		// Tags field (lowercase for action)
		m.editingField = "tags"
		// Pre-fill with current tags, filtering out system tags
		if m.viewingTask != nil {
			var userTags []string
			for _, tag := range m.viewingTask.Tags {
				if tag != "task" && tag != "project" {
					userTags = append(userTags, tag)
				}
			}
			// If no metadata tags, fall back to filename tags
			if len(userTags) == 0 && m.viewingFile != nil && len(m.viewingFile.Tags) > 0 {
				for _, tag := range m.viewingFile.Tags {
					if tag != "task" && tag != "project" {
						userTags = append(userTags, tag)
					}
				}
			}
			if len(userTags) > 0 {
				m.editBuffer = strings.Join(userTags, " ")
			} else {
				m.editBuffer = ""
			}
		} else if m.viewingProject != nil {
			var userTags []string
			for _, tag := range m.viewingProject.Tags {
				if tag != "task" && tag != "project" {
					userTags = append(userTags, tag)
				}
			}
			// If no metadata tags, fall back to filename tags
			if len(userTags) == 0 && m.viewingFile != nil && len(m.viewingFile.Tags) > 0 {
				for _, tag := range m.viewingFile.Tags {
					if tag != "task" && tag != "project" {
						userTags = append(userTags, tag)
					}
				}
			}
			if len(userTags) > 0 {
				m.editBuffer = strings.Join(userTags, " ")
			} else {
				m.editBuffer = ""
			}
		} else {
			m.editBuffer = ""
		}
		m.editCursor = len(m.editBuffer)
		m.statusMsg = "Enter tags (" + MsgSpaceSeparated + "):"
		
	case "r":
		// In acore format, tags are not part of the filename (only type is).
		// Tags changes are handled by updating the frontmatter only.
		// No filename rename is needed.
		m.statusMsg = "Tags are updated in frontmatter only (no filename change)"
		
	case "?":
		m.mode = ModeHelp
	}
	
	return m, nil
}