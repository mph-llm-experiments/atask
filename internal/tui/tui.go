package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pdxmph/atask/internal/config"
)

// Run starts the TUI application with optional filters
func Run(cfg *config.Config, filters ...string) error {
	// Parse optional filters
	var areaFilter string
	var showProjects bool
	
	if len(filters) > 0 {
		areaFilter = filters[0]
	}
	if len(filters) > 1 && filters[1] == "true" {
		showProjects = true
	}
	model, err := NewModel(cfg)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	
	// Handle initial flags
	if showProjects {
		// Show projects list
		model.projectFilter = true      // Filter to show only projects
		// Use configured defaults for tasks
		model.sortBy = cfg.Tasks.SortBy
		if model.sortBy == "" {
			model.sortBy = "due"
		}
		model.reverseSort = cfg.Tasks.SortOrder == "reverse"
		
		// Apply area filter if also provided
		if areaFilter != "" {
			model.areaFilter = areaFilter
		}
		
		model.applyFilters()
		model.sortFiles()
		model.loadVisibleMetadata()
	} else if areaFilter != "" {
		// Just area filter (tasks)
		model.areaFilter = areaFilter
		model.stateFilter = "active"    // Apply active filter
		// Use configured defaults for tasks
		model.sortBy = cfg.Tasks.SortBy
		if model.sortBy == "" {
			model.sortBy = "due"
		}
		model.reverseSort = cfg.Tasks.SortOrder == "reverse"
		model.applyFilters()
		model.sortFiles()               // Apply the sort after filters
		model.loadVisibleMetadata()     // Load metadata for visible items
	}
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	
	return nil
}