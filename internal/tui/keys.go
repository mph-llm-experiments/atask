package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	
	"github.com/charmbracelet/bubbletea"
	"github.com/mph-llm-experiments/atask/internal/denote"
	"github.com/mph-llm-experiments/atask/internal/task"
)

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeHelp:
		return m.handleHelpKeys(msg)
	case ModeSearch:
		return m.handleSearchKeys(msg)
	case ModePreview:
		return m.handlePreviewKeys(msg)
	case ModeCreate:
		return m.handleCreateKeys(msg)
	case ModeCreateTags:
		return m.handleCreateTagsKeys(msg)
	case ModeTaskView:
		return m.handleTaskViewKeys(msg)
	case ModeProjectView:
		return m.handleProjectViewKeys(msg)
	case ModeAreaFilter:
		return m.handleAreaFilterKeys(msg)
	case ModeSort:
		return m.handleSortKeys(msg)
	case ModeStateMenu:
		return m.handleStateMenuKeys(msg)
	case ModeConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ModeConfirmClearToday:
		return m.handleConfirmClearTodayKeys(msg)
	case ModeFilterMenu:
		return m.handleFilterMenuKeys(msg)
	case ModePriorityFilter:
		return m.handlePriorityFilterKeys(msg)
	case ModeStateFilter:
		return m.handleStateFilterKeys(msg)
	case ModeLogEntry:
		return m.handleLogEntryKeys(msg)
	case ModeProjectSelect:
		return m.handleProjectSelectKeys(msg)
	case ModeCreateProject:
		return m.handleCreateProjectKeys(msg)
	case ModeCreateProjectTags:
		return m.handleCreateProjectTagsKeys(msg)
	case ModeDateEdit:
		return m.handleDateEditKeys(msg)
	case ModeTagsEdit:
		return m.handleTagsEditKeys(msg)
	case ModeEstimateEdit:
		return m.handleEstimateEditKeys(msg)
	default:
		return m.handleNormalKeys(msg)
	}
}

func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear status message on any key press
	m.statusMsg = ""
	
	// Always use task mode key handling now
	return m.handleTaskModeKeys(msg)
}

func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.mode = ModeNormal
	}
	return m, nil
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search and restore previous state
		m.mode = ModeNormal
		m.searchInput = ""
		m.searchQuery = ""
		m.cursor = 0
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "ctrl+c":
		// Exit search and keep current results
		m.mode = ModeNormal
		
	case "enter":
		// Confirm search and exit search mode
		m.mode = ModeNormal
		
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			m.searchQuery = m.searchInput
			m.cursor = 0
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
		
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
			m.searchQuery = m.searchInput
			m.cursor = 0
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		}
	}
	
	return m, nil
}

func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "enter":
		m.mode = ModeNormal
		m.previewFile = nil
		
	case "E":
		if m.config.Editor != "" && m.previewFile != nil {
			return m, m.editFile(m.previewFile.Path)
		}
		
	default:
		// Use navigation handler for scroll navigation
		nav := NewNavigationHandler(m.previewMaxScroll+1, false)
		nav.cursor = m.previewScroll
		newScroll := nav.HandleKey(msg.String())
		
		// For preview, we want different behavior for g/G
		switch msg.String() {
		case "g":
			m.previewScroll = 0
		case "G":
			m.previewScroll = m.previewMaxScroll
		default:
			m.previewScroll = newScroll
		}
	}
	
	return m, nil
}

func (m Model) handleCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Full task creation form
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.resetCreateFields()
		
	case "enter":
		// Special handling for project field
		if m.createField == 4 { // Project field
			// Load projects and switch to selection mode
			m.loadProjectsForSelection()
			if len(m.projectSelectList) > 0 {
				m.projectSelectFor = "create"
				m.mode = ModeProjectSelect
				return m, nil
			} else {
				m.statusMsg = "No projects found"
				return m, nil
			}
		}
		
		// Validate and save
		if m.createTitle == "" {
			m.statusMsg = "Title is required"
			return m, nil
		}
		// Create the task and exit to normal mode
		m.mode = ModeNormal
		return m, m.createTask()
		
	case "up", "shift+tab":
		// Move to previous field
		m.createField--
		if m.createField < 0 {
			m.createField = 6 // Wrap to last field (tags)
		}
		// Skip area field if filtered
		if m.createField == 3 && m.areaFilter != "" {
			m.createField--
		}
		
	case "down", "tab":
		// Move to next field
		m.createField++
		if m.createField > 6 {
			m.createField = 0 // Wrap to first field
		}
		// Skip area field if filtered
		if m.createField == 3 && m.areaFilter != "" {
			m.createField++
		}
		
	case "backspace":
		// Edit current field
		switch m.createField {
		case 0: // Title
			if len(m.createTitle) > 0 {
				m.createTitle = m.createTitle[:len(m.createTitle)-1]
			}
		case 1: // Priority
			if len(m.createPriority) > 0 {
				m.createPriority = m.createPriority[:len(m.createPriority)-1]
			}
		case 2: // Due Date
			if len(m.createDue) > 0 {
				m.createDue = m.createDue[:len(m.createDue)-1]
			}
		case 3: // Area
			if len(m.createArea) > 0 {
				m.createArea = m.createArea[:len(m.createArea)-1]
			}
		case 4: // Project
			// Skip - project is selected via project selection mode
		case 5: // Estimate
			if len(m.createEstimate) > 0 {
				m.createEstimate = m.createEstimate[:len(m.createEstimate)-1]
			}
		case 6: // Tags
			if len(m.createTags) > 0 {
				m.createTags = m.createTags[:len(m.createTags)-1]
			}
		}
		
	default:
		// Type into current field
		if len(msg.String()) == 1 {
			switch m.createField {
			case 0: // Title
				m.createTitle += msg.String()
			case 1: // Priority
				m.createPriority += msg.String()
			case 2: // Due Date
				m.createDue += msg.String()
			case 3: // Area
				m.createArea += msg.String()
			case 4: // Project
				// Skip - project is selected via project selection mode
			case 5: // Estimate
				m.createEstimate += msg.String()
			case 6: // Tags
				m.createTags += msg.String()
			}
		}
	}
	
	return m, nil
}

func (m Model) handleCreateTagsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.createTitle = ""
		m.createTags = ""
		
	case "enter":
		m.mode = ModeNormal
		return m, m.create()
		
	case "backspace":
		if len(m.createTags) > 0 {
			m.createTags = m.createTags[:len(m.createTags)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.createTags += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleCreateProjectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.createTitle = ""
		
	case "enter":
		if m.createTitle != "" {
			m.mode = ModeCreateProjectTags
		}
		
	case "backspace":
		if len(m.createTitle) > 0 {
			m.createTitle = m.createTitle[:len(m.createTitle)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.createTitle += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleCreateProjectTagsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.createTitle = ""
		m.createTags = ""
		
	case "enter":
		m.mode = ModeNormal
		return m, m.create()
		
	case "backspace":
		if len(m.createTags) > 0 {
			m.createTags = m.createTags[:len(m.createTags)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.createTags += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleTaskModeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// Common navigation keys work the same
	case "q", "ctrl+c":
		return m, tea.Quit
		
	case "j", "down", "k", "up", "ctrl+d", "ctrl+u":
		// Use navigation handler for main list
		if len(m.filtered) > 0 {
			nav := NewNavigationHandler(len(m.filtered), false)
			nav.cursor = m.cursor
			newCursor := nav.HandleKey(msg.String())
			if newCursor != m.cursor {
				m.cursor = newCursor
				m.adjustScrollOffset()
				m.loadVisibleMetadata()
			}
		}

	case "g":
		if m.lastKey == "g" {
			m.cursor = 0
			m.scrollOffset = 0
			m.lastKey = ""
			m.loadVisibleMetadata()
		} else {
			m.lastKey = "g"
		}

	case "G":
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
			m.adjustScrollOffset()
			m.loadVisibleMetadata()
		}
		
	case "/":
		m.mode = ModeSearch
		m.searchInput = m.searchQuery
		
	case "enter":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Always load fresh from disk
			if file.IsTask() {
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					m.mode = ModeTaskView
					m.viewingTask = task
					m.viewingProject = nil
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
				} else {
					m.statusMsg = fmt.Sprintf("Error loading task: %v", err)
				}
			} else if file.IsProject() {
				if project, err := denote.ParseProjectFile(file.Path); err == nil {
					m.mode = ModeProjectView
					m.viewingTask = nil
					m.viewingProject = project
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
					m.projectViewTab = 0 // Start on overview tab
					m.loadProjectTasks() // Load tasks assigned to this project
				} else {
					m.statusMsg = fmt.Sprintf("Error loading project: %v", err)
				}
			} else {
				// Fall back to preview for non-task files
				m.mode = ModePreview
				m.previewFile = &file
			}
		}
		
	case "?":
		m.mode = ModeHelp
		
	// Task-specific keys
	case "1", "2", "3":
		// Set priority
		priority := "p" + msg.String()
		if err := m.updateTaskPriority(priority); err != nil {
			m.statusMsg = fmt.Sprintf(ErrorFormat, err)
		}
		
	case "0":
		// Clear priority
		if err := m.updateTaskPriority(""); err != nil {
			m.statusMsg = fmt.Sprintf(ErrorFormat, err)
		}

	case "y":
		// Toggle "today" tag
		if err := m.toggleTodayTag(); err != nil {
			m.statusMsg = fmt.Sprintf(ErrorFormat, err)
		}

	case "Y":
		// Clear "today" tag from all tasks (with confirmation)
		m.mode = ModeConfirmClearToday
		m.statusMsg = "Clear 'today' tag from ALL tasks? (y/n)"

	case "E":
		// Edit task in external editor (uppercase for Edit action)
		if m.config.Editor != "" && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			return m, m.editFile(file.Path)
		} else {
			m.statusMsg = "No editor configured"
		}
		
	case "u":
		// Update task metadata - same as enter
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Always load fresh from disk
			if file.IsTask() {
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					m.mode = ModeTaskView
					m.viewingTask = task
					m.viewingProject = nil
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
				} else {
					m.statusMsg = fmt.Sprintf("Error loading task: %v", err)
				}
			} else if file.IsProject() {
				if project, err := denote.ParseProjectFile(file.Path); err == nil {
					m.mode = ModeProjectView
					m.viewingTask = nil
					m.viewingProject = project
					m.viewingFile = &file
					m.editingField = ""
					m.editBuffer = ""
					m.projectViewTab = 0 // Start on overview tab
					m.loadProjectTasks() // Load tasks assigned to this project
				} else {
					m.statusMsg = fmt.Sprintf("Error loading project: %v", err)
				}
			}
		}
		
	case "c":
		// Create new task or project depending on current view
		if m.projectFilter {
			// In project list, create a project
			m.mode = ModeCreateProject
			m.createTitle = ""
			m.createTags = ""
		} else {
			// In task list, create a task
			m.mode = ModeCreate
			m.resetCreateFields()
		}
		
	case "f":
		// Filter menu
		m.mode = ModeFilterMenu
		
	case "s":
		// State change menu - for tasks and projects
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.mode = ModeStateMenu
		}
		
	case "S":
		// Sort mode (uppercase S since lowercase is now for state)
		m.mode = ModeSort
		
	case "x", "delete":
		// Delete task confirmation
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() || file.IsProject() {
				m.mode = ModeConfirmDelete
			}
		}
		
	case "r":
		// Toggle reverse sort
		m.reverseSort = !m.reverseSort
		m.sortFiles()
		
	case "l":
		// Log entry - only for tasks
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() {
				m.mode = ModeLogEntry
				m.logInput = ""
				m.loggingFile = &file
			} else {
				m.statusMsg = "Log entries only available for tasks"
			}
		}
		
	case "d":
		// Edit due date
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() || file.IsProject() {
				m.mode = ModeDateEdit
				m.editingField = "d"
				// Load current due date from disk
				if file.IsTask() {
					if task, err := denote.ParseTaskFile(file.Path); err == nil {
						m.editBuffer = task.TaskMetadata.DueDate
					} else {
						m.editBuffer = ""
					}
				} else if file.IsProject() {
					if project, err := denote.ParseProjectFile(file.Path); err == nil {
						m.editBuffer = project.ProjectMetadata.DueDate
					} else {
						m.editBuffer = ""
					}
				}
				m.editCursor = len(m.editBuffer)
			}
		}
		
	case "e":
		// Edit estimate
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() {
				m.mode = ModeEstimateEdit
				m.editingField = "e"
				// Load current estimate from disk
				if task, err := denote.ParseTaskFile(file.Path); err == nil {
					if task.TaskMetadata.Estimate > 0 {
						m.editBuffer = fmt.Sprintf("%d", task.TaskMetadata.Estimate)
					} else {
						m.editBuffer = ""
					}
				} else {
					m.editBuffer = ""
				}
				m.editCursor = len(m.editBuffer)
			}
		}
		
	case "t":
		// Edit tags (lowercase for action)
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			if file.IsTask() || file.IsProject() {
				m.mode = ModeTagsEdit
				m.editingField = "t"
				// Load current tags from disk
				if file.IsTask() {
					if task, err := denote.ParseTaskFile(file.Path); err == nil {
						// Filter out system tags from metadata
						var userTags []string
						for _, tag := range task.TaskMetadata.Tags {
							if tag != "task" && tag != "project" {
								userTags = append(userTags, tag)
							}
						}
						// If no metadata tags, fall back to filename tags
						if len(userTags) == 0 && len(file.Tags) > 0 {
							for _, tag := range file.Tags {
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
				} else if file.IsProject() {
					if project, err := denote.ParseProjectFile(file.Path); err == nil {
						// Filter out system tags from metadata
						var userTags []string
						for _, tag := range project.ProjectMetadata.Tags {
							if tag != "task" && tag != "project" {
								userTags = append(userTags, tag)
							}
						}
						// If no metadata tags, fall back to filename tags
						if len(userTags) == 0 && len(file.Tags) > 0 {
							for _, tag := range file.Tags {
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
				}
				m.editCursor = len(m.editBuffer)
			}
		}
		
	case "P":
		// Toggle project filter (uppercase for filter)
		m.projectFilter = !m.projectFilter
		if m.projectFilter {
			m.statusMsg = "Showing projects only"
			// Clear state filter when switching to projects
			m.stateFilter = ""
		} else {
			m.statusMsg = "Showing all tasks"
			// Restore default state filter when going back to tasks
			m.stateFilter = m.config.Tasks.DefaultStateFilter
		}
		m.cursor = 0
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "T":
		// Go to task list (opposite of 'P' for projects, uppercase for filter)
		if m.projectFilter {
			// Currently in project list, switch to task list
			m.projectFilter = false
			m.statusMsg = "Showing tasks"
			// Restore default state filter for tasks
			if m.stateFilter == "" {
				m.stateFilter = m.config.Tasks.DefaultStateFilter
			}
			m.applyFilters()
			m.sortFiles()
			m.loadVisibleMetadata()
		} else {
			// Already in task list
			m.statusMsg = "Already showing tasks"
		}
		
	default:
		m.lastKey = ""
	}
	
	return m, nil
}

func (m Model) handleAreaFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.searchInput = ""
		
	case "enter":
		m.areaFilter = m.searchInput
		m.mode = ModeNormal
		m.searchInput = ""
		if m.areaFilter != "" {
			m.statusMsg = fmt.Sprintf("Filtering by area: %s", m.areaFilter)
		} else {
			m.statusMsg = "Area filter cleared"
		}
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}
	
	return m, nil
}

func (m Model) handleSortKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Store the previous mode to return to
	previousMode := ModeNormal
	if m.viewingProject != nil {
		previousMode = ModeProjectView
	}
	
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = previousMode
		
	case "d":
		// Sort by due date
		m.sortBy = "due"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by due date"
		
	case "p":
		// Sort by priority
		m.sortBy = "priority"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by priority"
		
	case "j":
		// Sort by project
		m.sortBy = "project"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by project"
		
	case "e":
		// Sort by estimate
		m.sortBy = "estimate"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by estimate"
		
	case "t":
		// Sort by title (both modes)
		m.sortBy = "title"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by title"
		
	case "c":
		// Sort by created date
		m.sortBy = "created"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by created date"
		
	case "m":
		// Sort by modified date
		m.sortBy = "modified"
		m.mode = previousMode
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.viewingProject != nil {
			m.loadProjectTasks()
		}
		m.statusMsg = "Sorted by modified date"
		
	case "r":
		// Toggle reverse sort
		m.reverseSort = !m.reverseSort
		m.mode = ModeNormal
		m.sortFiles()
		m.loadVisibleMetadata()
		if m.reverseSort {
			m.statusMsg = fmt.Sprintf("Sorted by %s (descending)", m.sortBy)
		} else {
			m.statusMsg = fmt.Sprintf("Sorted by %s (ascending)", m.sortBy)
		}
	}
	
	return m, nil
}

func (m Model) handleStateMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Determine which mode to return to
	returnMode := ModeNormal
	if m.projectViewTab == 1 && m.viewingProject != nil {
		returnMode = ModeProjectView
	}

	// Check if current item is a task or project
	if m.cursor >= len(m.filtered) {
		m.mode = returnMode
		return m, nil
	}

	file := m.filtered[m.cursor]
	isProject := file.IsProject()

	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = returnMode

	case "a":
		// Active (projects only)
		if isProject {
			err := m.updateCurrentProjectStatus(denote.ProjectStatusActive)
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Project status changed to active"
			}
			m.mode = returnMode
		}

	case "c":
		// Completed (projects only)
		if isProject {
			err := m.updateCurrentProjectStatus(denote.ProjectStatusCompleted)
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Project status changed to completed"
			}
			m.mode = returnMode
		}

	case "x":
		// Cancelled (projects only)
		if isProject {
			err := m.updateCurrentProjectStatus(denote.ProjectStatusCancelled)
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Project status changed to cancelled"
			}
			m.mode = returnMode
		}

	case "o":
		// Open (tasks only)
		if !isProject {
			var err error
			if returnMode == ModeProjectView {
				err = m.updateProjectTaskStatus(denote.TaskStatusOpen)
			} else {
				err = m.updateCurrentTaskStatus(denote.TaskStatusOpen)
			}
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task status changed to open"
			}
			m.mode = returnMode
		}

	case "p":
		// Paused (both tasks and projects)
		var err error
		if isProject {
			err = m.updateCurrentProjectStatus(denote.ProjectStatusPaused)
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Project status changed to paused"
			}
		} else {
			if returnMode == ModeProjectView {
				err = m.updateProjectTaskStatus(denote.TaskStatusPaused)
			} else {
				err = m.updateCurrentTaskStatus(denote.TaskStatusPaused)
			}
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task status changed to paused"
			}
		}
		m.mode = returnMode

	case "d":
		// Done (tasks only)
		if !isProject {
			var err error
			var taskPath string
			if returnMode == ModeProjectView {
				if m.projectTasksCursor < len(m.projectTasks) {
					taskPath = m.projectTasks[m.projectTasksCursor].File.Path
				}
				err = m.updateProjectTaskStatus(denote.TaskStatusDone)
			} else {
				if m.cursor < len(m.filtered) {
					taskPath = m.filtered[m.cursor].Path
				}
				err = m.updateCurrentTaskStatus(denote.TaskStatusDone)
			}
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				recurMsg := m.handleTaskRecurrence(taskPath)
				if recurMsg != "" {
					m.scanFiles()
				}
				m.statusMsg = "Task status changed to done" + recurMsg
			}
			m.mode = returnMode
		}

	case "e":
		// Delegated (tasks only)
		if !isProject {
			var err error
			if returnMode == ModeProjectView {
				err = m.updateProjectTaskStatus(denote.TaskStatusDelegated)
			} else {
				err = m.updateCurrentTaskStatus(denote.TaskStatusDelegated)
			}
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task status changed to delegated"
			}
			m.mode = returnMode
		}

	case "r":
		// Dropped (tasks only)
		if !isProject {
			var err error
			if returnMode == ModeProjectView {
				err = m.updateProjectTaskStatus(denote.TaskStatusDropped)
			} else {
				err = m.updateCurrentTaskStatus(denote.TaskStatusDropped)
			}
			if err != nil {
				m.statusMsg = fmt.Sprintf(ErrorFormat, err)
			} else {
				m.statusMsg = "Task status changed to dropped"
			}
			m.mode = returnMode
		}
	}

	return m, nil
}

func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Determine which mode to return to
	returnMode := ModeNormal
	if m.viewingProject != nil {
		returnMode = ModeProjectView
	}
	
	switch msg.String() {
	case "y", "Y":
		// Handle project deletion specially
		if m.projectViewTab == 0 && m.viewingProject != nil {
			// Delete the project and clear project_id from affected tasks
			projectPath := m.viewingFile.Path
			projectTitle := m.viewingProject.ProjectMetadata.Title
			
			// First, clear project_id from all affected tasks
			for _, task := range m.affectedTasks {
				if err := m.clearProjectFromTask(task.File.Path); err != nil {
					m.statusMsg = fmt.Sprintf("Error updating task: %v", err)
					m.mode = returnMode
					return m, nil
				}
			}
			
			// Then delete the project file
			if err := m.deleteFile(projectPath); err != nil {
				m.statusMsg = fmt.Sprintf("Error deleting project: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Deleted project: %s", projectTitle)
				if len(m.affectedTasks) > 0 {
					m.statusMsg += fmt.Sprintf(" (cleared from %d tasks)", len(m.affectedTasks))
				}
				// Go back to main task list
				m.mode = ModeNormal
				m.viewingProject = nil
				m.viewingFile = nil
				m.affectedTasks = nil
				// Rescan files
				m.scanFiles()
				return m, nil
			}
		} else if m.projectViewTab == 1 && m.viewingProject != nil {
			// Delete task from project view
			if m.projectTasksCursor < len(m.projectTasks) {
				task := m.projectTasks[m.projectTasksCursor]
				filePath := task.File.Path
				fileTitle := task.TaskMetadata.Title
				
				if err := m.deleteFile(filePath); err != nil {
					m.statusMsg = fmt.Sprintf("Error deleting: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Deleted: %s", fileTitle)
					// Rescan files after deletion
					m.scanFiles()
					// Reload project tasks
					m.loadProjectTasks()
				}
			}
		} else {
			// Delete from normal view
			if m.cursor < len(m.filtered) {
				file := m.filtered[m.cursor]
				filePath := file.Path
				fileTitle := file.Title
				
				if err := m.deleteFile(filePath); err != nil {
					m.statusMsg = fmt.Sprintf("Error deleting: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Deleted: %s", fileTitle)
					// Rescan files after deletion
					m.scanFiles()
				}
			}
		}
		m.mode = returnMode
		
	case "n", "N", "esc", "ctrl+c":
		// Cancel delete
		m.mode = returnMode
		m.statusMsg = "Delete cancelled"
		// Clear affected tasks list
		m.affectedTasks = nil
	}
	
	return m, nil
}

func (m Model) handleConfirmClearTodayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Clear all today tags
		if err := m.clearAllTodayTags(); err != nil {
			m.statusMsg = fmt.Sprintf("Error clearing today tags: %v", err)
		}
		m.mode = ModeNormal

	case "n", "N", "esc", "ctrl+c":
		// Cancel
		m.mode = ModeNormal
		m.statusMsg = "Cancelled"
	}

	return m, nil
}

func (m Model) handleFilterMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.mode = ModeNormal
		
	case "a":
		// Area filter
		m.mode = ModeAreaFilter
		m.searchInput = m.areaFilter
		
	case "p":
		// Priority filter
		m.mode = ModePriorityFilter
		
	case "s":
		// State filter  
		m.mode = ModeStateFilter
		
	case "l":
		// Loose filter toggle (tasks with no project)
		m.looseFilter = !m.looseFilter
		m.mode = ModeNormal
		if m.looseFilter {
			m.statusMsg = "Showing loose tasks (no project)"
		} else {
			m.statusMsg = "Loose filter disabled"
		}
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()

	case "d":
		// Soon filter toggle
		m.soonFilter = !m.soonFilter
		m.mode = ModeNormal
		if m.soonFilter {
			m.statusMsg = fmt.Sprintf("Showing tasks due within %d days", m.config.SoonHorizon)
		} else {
			m.statusMsg = "Soon filter disabled"
		}
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c":
		// Clear all filters
		m.areaFilter = ""
		m.priorityFilter = ""
		m.stateFilter = ""
		m.soonFilter = false
		m.looseFilter = false
		m.mode = ModeNormal
		m.statusMsg = "All filters cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}

func (m Model) handlePriorityFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		
	case "1":
		m.priorityFilter = "p1"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p1"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "2":
		m.priorityFilter = "p2"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p2"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "3":
		m.priorityFilter = "p3"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by priority: p3"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c", "x":
		// Clear priority filter
		m.priorityFilter = ""
		m.mode = ModeNormal
		m.statusMsg = "Priority filter cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}

func (m Model) handleStateFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		
	case "i":
		// Incomplete (everything except done)
		m.stateFilter = "incomplete"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: incomplete (everything except done)"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()

	case "a":
		// Active (open + delegated)
		m.stateFilter = "active"
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: active (open + delegated)"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "o":
		// Open
		m.stateFilter = denote.TaskStatusOpen
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: open"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "p":
		// Paused
		m.stateFilter = denote.TaskStatusPaused
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: paused"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "d":
		// Done
		m.stateFilter = denote.TaskStatusDone
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: done"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "e":
		// Delegated
		m.stateFilter = denote.TaskStatusDelegated
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: delegated"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "r":
		// Dropped
		m.stateFilter = denote.TaskStatusDropped
		m.mode = ModeNormal
		m.statusMsg = "Filtering by state: dropped"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
		
	case "c", "x":
		// Clear state filter
		m.stateFilter = ""
		m.mode = ModeNormal
		m.statusMsg = "State filter cleared"
		m.applyFilters()
		m.sortFiles()
		m.loadVisibleMetadata()
	}
	
	return m, nil
}

func (m Model) handleLogEntryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		// Cancel log entry
		// Return to where we came from
		if m.viewingFile != nil {
			m.mode = ModeTaskView
		} else {
			m.mode = ModeNormal
		}
		m.logInput = ""
		m.loggingFile = nil
		m.statusMsg = "Log entry cancelled"
		
	case "enter":
		// Save log entry
		if m.logInput != "" && m.loggingFile != nil {
			if err := m.addLogEntry(); err != nil {
				m.statusMsg = fmt.Sprintf("Error adding log: %v", err)
			} else {
				m.statusMsg = "Log entry added"
				
				// If we're viewing this task, reload it completely to show the new log entry
				if m.viewingFile != nil && m.viewingFile.Path == m.loggingFile.Path {
					// Reload the task with fresh content
					if task, err := denote.ParseTaskFile(m.loggingFile.Path); err == nil {
						m.viewingTask = task
						// Return to task view
						m.mode = ModeTaskView
						m.logInput = ""
						m.loggingFile = nil
						return m, nil
					}
				}
				
				// If we're previewing this file, update the preview
				if m.previewFile != nil && m.previewFile.Path == m.loggingFile.Path {
					// Force refresh by clearing and resetting
					m.previewFile = m.loggingFile
				}
				
				// Reload file metadata after modification
				m.loadVisibleMetadata()
			}
		} else {
			m.statusMsg = "No log text entered"
		}
		
		// If we were in task view, return there, otherwise go to normal mode
		if m.viewingFile != nil {
			m.mode = ModeTaskView
		} else {
			m.mode = ModeNormal
		}
		m.logInput = ""
		m.loggingFile = nil
		
	case "backspace":
		if len(m.logInput) > 0 {
			m.logInput = m.logInput[:len(m.logInput)-1]
		}
		
	default:
		if len(msg.String()) == 1 {
			m.logInput += msg.String()
		}
	}
	
	return m, nil
}
func (m Model) handleProjectSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		// Return to previous mode
		if m.projectSelectFor == "create" {
			m.mode = ModeCreate
		} else if m.projectSelectFor == "update" {
			m.mode = ModeTaskView
		} else {
			m.mode = ModeNormal
		}
		
	case "enter":
		// Select project or unassign
		if m.projectSelectCursor == 0 {
			// "None" option selected - unassign from project
			if m.projectSelectFor == "create" {
				// Clear project ID in create form
				m.createProject = ""
				m.mode = ModeCreate
			} else if m.projectSelectFor == "update" && m.projectSelectTask != nil {
				// Clear project assignment
				m.projectSelectTask.TaskMetadata.ProjectID = ""
				if err := task.UpdateTaskFile(m.projectSelectTask.File.Path, m.projectSelectTask.TaskMetadata); err != nil {
					m.statusMsg = fmt.Sprintf("Error updating task: %v", err)
				} else {
					m.statusMsg = "Removed from project"
					// Reload task if we are viewing it
					if m.viewingTask != nil && m.viewingTask.File.Path == m.projectSelectTask.File.Path {
						m.viewingTask = m.projectSelectTask
					}
				}
				m.mode = ModeTaskView
			}
		} else if m.projectSelectCursor-1 < len(m.projectSelectList) {
			// Project selected (adjust for None option at index 0)
			selected := m.projectSelectList[m.projectSelectCursor-1]
			
			if m.projectSelectFor == "create" {
				// Set project ID in create form (using index_id)
				m.createProject = strconv.Itoa(selected.IndexID)
				m.mode = ModeCreate
			} else if m.projectSelectFor == "update" && m.projectSelectTask != nil {
				// Update task with selected project (using index_id)
				m.projectSelectTask.TaskMetadata.ProjectID = strconv.Itoa(selected.IndexID)
				if err := task.UpdateTaskFile(m.projectSelectTask.File.Path, m.projectSelectTask.TaskMetadata); err != nil {
					m.statusMsg = fmt.Sprintf("Error updating task: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Added to project: %s", selected.ProjectMetadata.Title)
					// Reload task if we are viewing it
					if m.viewingTask != nil && m.viewingTask.File.Path == m.projectSelectTask.File.Path {
						m.viewingTask = m.projectSelectTask
					}
				}
				m.mode = ModeTaskView
			}
		}
		
	case "j", "down", "k", "up", "g", "G", "ctrl+d", "ctrl+u":
		// Use navigation handler (account for None option at index 0)
		nav := NewNavigationHandler(len(m.projectSelectList)+1, false)
		nav.cursor = m.projectSelectCursor
		m.projectSelectCursor = nav.HandleKey(msg.String())
		
	// Allow number selection (0-9)
	case "0":
		// Select "None" option
		m.projectSelectCursor = 0
		// Auto-select
		return m.handleProjectSelectKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tea.KeyEnter)}})
		
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		num, _ := strconv.Atoi(msg.String())
		if num <= len(m.projectSelectList) {
			m.projectSelectCursor = num // 1-9 maps to cursor positions 1-9 (after None at 0)
			// Auto-select
			return m.handleProjectSelectKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(tea.KeyEnter)}})
		}
	}
	
	return m, nil
}

func (m Model) handleDateEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "enter":
		// Parse and validate the date
		parsedDate, err := denote.ParseNaturalDate(m.editBuffer)
		if err != nil && m.editBuffer != "" {
			m.statusMsg = fmt.Sprintf("Invalid date: %s", err)
			return m, nil
		}
		
		// Update the due date
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			if file.IsTask() {
				// Load fresh metadata from disk
				if t, err := denote.ParseTaskFile(file.Path); err == nil {
					t.TaskMetadata.DueDate = parsedDate
					if err := task.UpdateTaskFile(file.Path, t.TaskMetadata); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						if parsedDate == "" {
							m.statusMsg = "Due date removed"
						} else {
							m.statusMsg = fmt.Sprintf("Due date set to %s", parsedDate)
						}
						// Force UI to refresh
						m.loadVisibleMetadata()
					}
				}
			} else if file.IsProject() {
				// Load fresh metadata from disk
				if project, err := denote.ParseProjectFile(file.Path); err == nil {
					project.ProjectMetadata.DueDate = parsedDate
					if err := denote.UpdateProjectFile(file.Path, project.ProjectMetadata); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						if parsedDate == "" {
							m.statusMsg = "Due date removed"
						} else {
							m.statusMsg = fmt.Sprintf("Due date set to %s", parsedDate)
						}
						// Force UI to refresh
						m.loadVisibleMetadata()
					}
				}
			}
		}
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "backspace":
		if m.editCursor > 0 && len(m.editBuffer) > 0 {
			// Remove character before cursor
			m.editBuffer = m.editBuffer[:m.editCursor-1] + m.editBuffer[m.editCursor:]
			m.editCursor--
		}
		
	case "delete":
		if m.editCursor < len(m.editBuffer) {
			// Remove character at cursor
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
		
	default:
		if len(msg.String()) == 1 {
			// Insert character at cursor position
			m.editBuffer = m.editBuffer[:m.editCursor] + msg.String() + m.editBuffer[m.editCursor:]
			m.editCursor++
		}
	}
	
	return m, nil
}

func (m Model) handleTagsEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "enter":
		// Update the tags
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Parse tags from input
			var newTags []string
			if m.editBuffer != "" {
				for _, tag := range strings.Fields(m.editBuffer) {
					tag = strings.TrimSpace(tag)
					if tag != "" && tag != "task" && tag != "project" {
						newTags = append(newTags, tag)
					}
				}
			}
			
			if file.IsTask() {
				// Load fresh metadata from disk
				if t, err := denote.ParseTaskFile(file.Path); err == nil {
					oldPath := file.Path
					// Always include "task" in metadata tags
					t.TaskMetadata.Tags = []string{"task"}
					t.TaskMetadata.Tags = append(t.TaskMetadata.Tags, newTags...)
					
					// First update the metadata
					if err := task.UpdateTaskFile(file.Path, t.TaskMetadata); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						// Now rename the file to reflect new tags
						allTags := []string{"task"} // Always include task tag
						for _, tag := range newTags {
							if tag != "task" {
								allTags = append(allTags, tag)
							}
						}
						
						// Rename file
						newPath, err := denote.RenameFileForTags(oldPath, allTags)
						if err != nil {
							m.statusMsg = fmt.Sprintf("Tags updated but rename failed: %v", err)
						} else {
							if len(newTags) == 0 {
								m.statusMsg = "Tags cleared"
							} else {
								m.statusMsg = fmt.Sprintf("Tags updated: %s", strings.Join(newTags, " "))
							}
							
							// Trigger a rescan if the file was renamed
							if newPath != oldPath {
								m.scanFiles()
							}
						}
					}
				}
			} else if file.IsProject() {
				// Load fresh metadata from disk
				if project, err := denote.ParseProjectFile(file.Path); err == nil {
					oldPath := file.Path
					// Always include "project" in metadata tags
					project.ProjectMetadata.Tags = []string{"project"}
					project.ProjectMetadata.Tags = append(project.ProjectMetadata.Tags, newTags...)
					
					// First update the metadata
					if err := denote.UpdateProjectFile(file.Path, project.ProjectMetadata); err != nil {
						m.statusMsg = fmt.Sprintf(ErrorFormat, err)
					} else {
						// Now rename the file to reflect new tags
						allTags := []string{"project"} // Always include project tag
						for _, tag := range newTags {
							if tag != "project" {
								allTags = append(allTags, tag)
							}
						}
						
						// Rename file
						newPath, err := denote.RenameFileForTags(oldPath, allTags)
						if err != nil {
							m.statusMsg = fmt.Sprintf("Tags updated but rename failed: %v", err)
						} else {
							if len(newTags) == 0 {
								m.statusMsg = "Tags cleared"
							} else {
								m.statusMsg = fmt.Sprintf("Tags updated: %s", strings.Join(newTags, " "))
							}
							
							// Trigger a rescan if the file was renamed
							if newPath != oldPath {
								m.scanFiles()
							}
						}
					}
				}
			}
		}
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "backspace":
		if m.editCursor > 0 && len(m.editBuffer) > 0 {
			// Remove character before cursor
			m.editBuffer = m.editBuffer[:m.editCursor-1] + m.editBuffer[m.editCursor:]
			m.editCursor--
		}
		
	case "delete":
		if m.editCursor < len(m.editBuffer) {
			// Remove character at cursor
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
		
	default:
		if len(msg.String()) == 1 {
			// Insert character at cursor position
			m.editBuffer = m.editBuffer[:m.editCursor] + msg.String() + m.editBuffer[m.editCursor:]
			m.editCursor++
		}
	}
	
	return m, nil
}

func (m Model) handleEstimateEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "enter":
		// Save the estimate
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			file := m.filtered[m.cursor]
			
			// Parse the estimate value
			var estimate int
			if m.editBuffer != "" {
				if val, err := strconv.Atoi(m.editBuffer); err == nil && val >= 0 {
					estimate = val
				} else {
					m.statusMsg = "Invalid estimate - must be a positive number"
					m.mode = ModeNormal
					m.editingField = ""
					m.editBuffer = ""
					m.editCursor = 0
					return m, nil
				}
			}
			
			// Update the task
			if file.IsTask() {
				// Read the file content
				content, err := os.ReadFile(file.Path)
				if err != nil {
					m.statusMsg = fmt.Sprintf("Failed to read file: %v", err)
					m.mode = ModeNormal
					m.editingField = ""
					m.editBuffer = ""
					m.editCursor = 0
					return m, nil
				}
				
				// Parse existing frontmatter
				fm, err := denote.ParseFrontmatterFile(content)
				if err != nil {
					m.statusMsg = fmt.Sprintf("Failed to parse frontmatter: %v", err)
					m.mode = ModeNormal
					m.editingField = ""
					m.editBuffer = ""
					m.editCursor = 0
					return m, nil
				}
				
				// Update the metadata
				if taskMeta, ok := fm.Metadata.(denote.TaskMetadata); ok {
					taskMeta.Estimate = estimate
					
					// Write updated content
					newContent, err := denote.WriteFrontmatterFile(taskMeta, fm.Content)
					if err != nil {
						m.statusMsg = fmt.Sprintf("Failed to write frontmatter: %v", err)
					} else {
						// Write back to file
						if err := os.WriteFile(file.Path, newContent, 0644); err != nil {
							m.statusMsg = fmt.Sprintf("Failed to update estimate: %v", err)
						} else {
							if estimate == 0 {
								m.statusMsg = "Estimate cleared"
							} else {
								m.statusMsg = fmt.Sprintf("Estimate set to %d", estimate)
							}
						}
					}
				}
			}
		}
		m.mode = ModeNormal
		m.editingField = ""
		m.editBuffer = ""
		m.editCursor = 0
		
	case "backspace":
		if m.editCursor > 0 {
			m.editBuffer = m.editBuffer[:m.editCursor-1] + m.editBuffer[m.editCursor:]
			m.editCursor--
		}
		
	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}
		
	case "right":
		if m.editCursor < len(m.editBuffer) {
			m.editCursor++
		}
		
	case "home", "ctrl+a":
		m.editCursor = 0
		
	case "end", "ctrl+e":
		m.editCursor = len(m.editBuffer)
		
	default:
		// Only allow numeric input
		if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
			// Insert character at cursor position
			m.editBuffer = m.editBuffer[:m.editCursor] + msg.String() + m.editBuffer[m.editCursor:]
			m.editCursor++
		}
	}
	
	return m, nil
}