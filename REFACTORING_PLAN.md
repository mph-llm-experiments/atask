# Refactoring Plan: atask Focus on Task Management

## Executive Summary

This document outlines the plan to refactor atask from a dual notes/tasks system to a focused task management tool. The project will maintain Denote format compatibility for backward compatibility and consistent file identification, but will remove all general notes functionality.

## Current State Analysis

### What We Have Now
1. **Dual Mode System**: Both notes and tasks functionality
2. **Complex CLI**: Commands for both notes and tasks
3. **Complex TUI**: View modes for notes and tasks
4. **Mixed Concerns**: Code that handles both notes and tasks interchangeably

### What We Want
1. **Task-Only System**: Focus exclusively on task management
2. **Simplified CLI**: Only task and project commands
3. **Simplified TUI**: Only task and project views
4. **Clear Purpose**: A dedicated task management tool using Denote format

## Architecture Changes

### Simplified Architecture
```
atask/
├── cmd/atask/        # Entry point
├── internal/
│   ├── cli/                 # CLI implementation (task-focused)
│   ├── config/              # Configuration
│   ├── core/                # Business logic (filters)
│   ├── denote/              # Denote file operations (keep)
│   ├── task/                # Task-specific logic (expand)
│   └── tui/                 # TUI implementation (task-only)
└── docs/                    # Specifications
```

### Components to Remove
1. **Notes-specific views** in TUI
2. **Notes commands** in CLI
3. **ViewMode switching** between notes and tasks
4. **General file browsing** capabilities

### Components to Keep
1. **Denote package** - For file naming and ID generation
2. **Task metadata handling** - Core functionality
3. **Project management** - Essential for task organization
4. **Filter system** - Adapted for tasks only
5. **Both CLI and TUI** - But simplified

## Detailed Refactoring Steps

### Phase 1: Define New Data Model

#### Task-Only File Types
- **Task Files**: `*__task*.md` with task metadata
- **Project Files**: `*__project*.md` with project metadata
- **No General Notes**: Remove support for files without task/project tags

#### Simplified View Model
```go
// Remove ViewMode enum, always in "task mode"
type Model struct {
    // Core data
    tasks         []*denote.Task
    projects      []*denote.Project
    filtered      []*denote.Task
    
    // UI state
    cursor        int
    mode          UIMode  // Normal, Search, Create, etc.
    
    // Filters (task-specific)
    areaFilter    string
    priorityFilter string
    statusFilter  string
    projectFilter string
    
    // Remove notes-specific fields
}
```

### Phase 2: CLI Simplification

#### Commands to Keep (Modified)
```bash
# Task commands
atask task new "Task title" [options]
atask task list [filters]
atask task update <id> [options]
atask task done <id>
atask task delete <id>
atask task log <id> "message"

# Project commands  
atask project new "Project title" [options]
atask project list [filters]
atask project tasks <project-id>
atask project update <id> [options]

# Global options
atask --tui  # Launch TUI in task mode
```

#### Commands to Remove
- Any generic "note" commands
- File browsing commands
- Mode switching options

### Phase 3: TUI Simplification

#### Views to Keep (Modified)
1. **Task List View** - Main view showing tasks
2. **Project List View** - Toggled with 'p' key
3. **Task Detail View** - For viewing/editing tasks
4. **Project Detail View** - Shows project and its tasks
5. **Create Forms** - For new tasks/projects

#### Views to Remove
1. **Notes browsing view**
2. **Mode switching UI**
3. **Generic file operations**

#### Simplified Key Bindings
```
Task List:
  j/k     - Navigate
  Enter   - View task details
  c       - Create new task
  x       - Delete task
  d       - Mark done
  p       - Toggle project view
  /       - Search tasks
  f       - Filter menu
  S       - Sort menu
  q       - Quit

Project List:
  j/k     - Navigate  
  Enter   - View project tasks
  c       - Create new project
  x       - Delete project
  t       - Back to task list
  /       - Search projects
  q       - Quit
```

### Phase 4: Code Changes

#### Files to Modify Heavily

1. **internal/tui/model.go**
   - Remove ViewMode enum and switching logic
   - Remove notes-specific fields
   - Simplify to task-only model

2. **internal/tui/views.go**
   - Remove renderFileLine for generic files
   - Keep only renderTaskLine and renderProjectLine
   - Simplify header to show "Denote Tasks" only

3. **internal/cli/cli.go**
   - Remove notes commands
   - Simplify command structure
   - Update help text

4. **internal/tui/keys.go**
   - Remove 't' key for mode switching
   - Remove notes-specific hotkeys
   - Simplify key handling

#### Files to Keep Mostly Unchanged

1. **internal/denote/** - Core Denote functionality still needed
2. **internal/task/** - Core task logic
3. **internal/config/** - Configuration system
4. **internal/core/filter.go** - Filtering logic (adapt for tasks)

### Phase 5: Documentation Updates

1. **Update PROJECT_CHARTER.md**
   - Remove references to notes functionality
   - Clarify task-only focus
   - Update examples

2. **Update README.md**
   - Clear description as task management tool
   - Remove notes examples
   - Focus on task workflows

3. **Update CLI_REFERENCE.md**
   - Remove notes commands
   - Update examples
   - Clarify task focus

## Migration Considerations

### For Existing Users
1. **Data Compatibility**: Task files remain unchanged
2. **Command Changes**: Document removed commands
3. **Workflow Changes**: Provide migration guide

### Backward Compatibility
1. **File Format**: Keep Denote format for all files
2. **Metadata**: Maintain same YAML structure
3. **IDs**: Continue using Denote IDs for references

## Testing Plan

### Unit Tests
1. Remove notes-specific tests
2. Enhance task operation tests
3. Add project management tests

### Integration Tests
1. CLI task workflows
2. TUI task operations
3. Project-task associations

### Manual Testing
1. Create tasks via CLI and TUI
2. Manage projects and associations
3. Filter and sort operations
4. Data integrity checks

## Benefits of Refactoring

1. **Clarity of Purpose**: Clear focus on task management
2. **Reduced Complexity**: Simpler codebase to maintain
3. **Better UX**: No mode confusion, clearer workflows
4. **Easier Onboarding**: Single purpose tool
5. **Maintainability**: Less code, fewer edge cases

## Risks and Mitigation

### Risk: User Confusion
**Mitigation**: Clear documentation about the change, migration guide

### Risk: Feature Requests for Notes
**Mitigation**: Clear project charter, recommend separate tools

### Risk: Breaking Changes
**Mitigation**: Maintain file format compatibility, careful testing

## Implementation Timeline

### Week 1: Foundation
- Define new data models
- Remove ViewMode infrastructure
- Simplify core types

### Week 2: CLI Refactoring  
- Remove notes commands
- Simplify command structure
- Update help and documentation

### Week 3: TUI Refactoring
- Remove notes views
- Simplify key bindings
- Focus on task workflows

### Week 4: Polish and Testing
- Update all documentation
- Comprehensive testing
- Performance optimization

## Success Criteria

1. **Focused Tool**: Only task management functionality
2. **Simpler Codebase**: Reduced by ~30-40%
3. **Clearer UX**: No mode switching or confusion
4. **Maintained Compatibility**: Denote format preserved
5. **Better Performance**: Faster operations due to focus

## Next Steps

1. Review and approve this plan
2. Create feature branch for refactoring
3. Implement changes incrementally
4. Test thoroughly at each phase
5. Update documentation continuously
6. Plan migration communication

---

*This refactoring will transform atask from a dual-purpose tool into a focused, powerful task management system while maintaining the benefits of the Denote format.*