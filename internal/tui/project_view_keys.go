package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/atask/internal/denote"
)

func (m Model) handleProjectViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle editing mode first - EXACTLY like task view
	if m.editingField != "" {
		switch msg.String() {
		case "esc":
			m.editingField = ""
			m.editBuffer = ""
			m.editCursor = 0
			m.editCursor = 0

		case "enter":
			// Handle project field updates - map single letters to field names
			fieldMap := map[string]string{
				"T": "title",
				"p": "priority",
				"s": "status",
				"d": "due_date",
				"B": "start_date",
				"a": "area",
				"t": "tags",
			}

			if fieldName, ok := fieldMap[m.editingField]; ok {
				var updateErr error
				var updateValue string

				switch m.editingField {
				case "p":
					if m.editBuffer == "" || m.editBuffer == "0" {
						// Clear priority
						updateValue = ""
					} else if m.editBuffer == "1" || m.editBuffer == "2" || m.editBuffer == "3" {
						updateValue = "p" + m.editBuffer
					} else {
						m.statusMsg = "Priority must be 0 (clear), 1, 2, or 3"
						m.editingField = ""
						m.editBuffer = ""
						m.editCursor = 0
						return m, nil
					}
				default:
					updateValue = m.editBuffer
				}

				updateErr = m.updateProjectField(fieldName, updateValue)

				if updateErr != nil {
					m.statusMsg = fmt.Sprintf(ErrorFormat, updateErr)
				} else {
					// Special message for priority clearing
					if m.editingField == "p" && updateValue == "" {
						m.statusMsg = "Priority removed"
					} else {
						m.statusMsg = fmt.Sprintf("Updated %s", fieldName)
					}

					// Force reload the project from disk to ensure we show the updated data
					if project, err := denote.ParseProjectFile(m.viewingFile.Path); err == nil {
						m.viewingProject = project
						// Update the file title if it changed
						if project.Title != "" {
							m.viewingFile.Title = project.Title
						}
					}

					// Re-sort if we updated a field that could affect order or visibility
					if m.editingField == "d" || m.editingField == "p" || m.editingField == "B" {
						m.applyFilters()
						m.sortFiles()
						m.loadVisibleMetadata()
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

	// Normal navigation when not editing
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeNormal
		m.viewingProject = nil
		m.viewingFile = nil
		m.projectTasks = nil
		m.projectTasksCursor = 0
		m.projectViewTab = 0
		m.affectedTasks = nil
		// Re-sort the list in case metadata changed
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()

	case "tab":
		// Switch between tabs
		if m.projectViewTab == 0 {
			m.projectViewTab = 1
		} else {
			m.projectViewTab = 0
		}

	case "E":
		// Edit project file in external editor (uppercase for Edit action)
		if m.config.Editor != "" && m.viewingFile != nil {
			return m, m.editFile(m.viewingFile.Path)
		} else {
			m.statusMsg = "No editor configured"
		}

	// Field edit hotkeys - work on overview tab
	case "T":
		// Title field (uppercase - different from tags)
		if m.projectViewTab == 0 {
			m.editingField = "T" // Title
			m.editBuffer = m.viewingProject.Title
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter title:"
		}

	case "p":
		if m.projectViewTab == 0 {
			m.editingField = "p" // Use single letter like renderField expects
			m.editBuffer = strings.TrimPrefix(m.viewingProject.ProjectMetadata.Priority, "p")
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter priority (0 to clear, 1/2/3):"
		}

	case "s":
		if m.projectViewTab == 0 {
			m.editingField = "s" // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.Status
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter status (active/completed/paused/cancelled):"
		} else if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// In tasks tab, 's' opens state menu
			m.mode = ModeStateMenu
			return m, nil
		}

	case "d":
		if m.projectViewTab == 0 {
			m.editingField = "d" // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.DueDate
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter due date (YYYY-MM-DD or relative: 1d, 1w, tomorrow):"
		}

	case "B":
		if m.projectViewTab == 0 {
			m.editingField = "B"
			m.editBuffer = m.viewingProject.ProjectMetadata.StartDate
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter begin date (YYYY-MM-DD or relative: 1d, 1w, tomorrow):"
		}

	case "a":
		if m.projectViewTab == 0 {
			m.editingField = "a" // Use single letter
			m.editBuffer = m.viewingProject.ProjectMetadata.Area
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter area:"
		}

	case "t":
		// Tags field (lowercase for action)
		if m.projectViewTab == 0 {
			m.editingField = "t" // Tags
			// Filter out system tags
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
			m.editCursor = len(m.editBuffer)
			m.statusMsg = "Enter tags (space-separated):"
		}

	case "n":
		// Create new task with this project pre-selected
		m.mode = ModeCreate
		m.resetCreateFields()
		// Pre-fill the project ID from the current project
		if m.viewingProject != nil {
			m.createProject = strconv.Itoa(m.viewingProject.IndexID)
		}
		// Pre-fill area if the project has one
		if m.viewingProject != nil && m.viewingProject.ProjectMetadata.Area != "" {
			m.areaFilter = m.viewingProject.ProjectMetadata.Area
		}
		m.creatingFromProject = true
		m.statusMsg = "Creating new task for project"
		return m, nil

	// Keys for task navigation (on main tab)
	case "j", "down", "k", "up", "G", "ctrl+d", "ctrl+u":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Use navigation handler for task list in project view
			nav := NewNavigationHandler(len(m.projectTasks), false)
			nav.cursor = m.projectTasksCursor
			m.projectTasksCursor = nav.HandleKey(msg.String())
		}

	case "enter":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 && m.projectTasksCursor < len(m.projectTasks) {
			// Open the selected task
			task := m.projectTasks[m.projectTasksCursor]
			m.mode = ModeTaskView
			m.viewingTask = &task
			file := denote.FileFromTask(&task)
			m.viewingFile = &file
			m.editingField = ""
			m.editBuffer = ""
			m.editCursor = 0
			m.returnToProject = true // Remember to return to project view
			// Keep the project reference!
		}

	case "0":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Clear priority on selected task
			task := &m.projectTasks[m.projectTasksCursor]
			if err := m.updateTaskPriorityFromProject(task, ""); err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task priority cleared"
			}
		}

	case "1", "2", "3":
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 {
			// Set priority on selected task
			task := &m.projectTasks[m.projectTasksCursor]
			priority := "p" + msg.String()
			if err := m.updateTaskPriorityFromProject(task, priority); err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = fmt.Sprintf("Task priority updated to %s", priority)
			}
		}

	case "x":
		// On main tab (0): delete selected task if cursor is on a task
		if m.projectViewTab == 0 && len(m.projectTasks) > 0 && m.projectTasksCursor < len(m.projectTasks) {
			// Deleting a task
			m.mode = ModeConfirmDelete
			return m, nil
		}

	case "X":
		// Capital X to delete the project itself (only on main tab)
		if m.projectViewTab == 0 {
			// Deleting the project itself - find affected tasks first
			m.findTasksAffectedByProjectDeletion()
			m.mode = ModeConfirmDelete
			return m, nil
		}

	// Sorting keys (same as task mode - uppercase S since lowercase s is for status)
	case "S":
		// Open sort menu
		m.mode = ModeSort
		return m, nil
	}

	return m, nil
}

// updateTaskPriorityFromProject updates a task priority from the project view
func (m *Model) updateTaskPriorityFromProject(task *denote.Task, priority string) error {
	task.TaskMetadata.Priority = priority
	if err := denote.UpdateTaskPriority(task.FilePath, priority); err != nil {
		return err
	}
	return nil
}
