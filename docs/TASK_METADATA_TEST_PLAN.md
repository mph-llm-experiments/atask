# Task Metadata Test Plan

## Test Environment Setup

```bash
# Build the application
go build ./cmd/atask

# Run with test config
./atask --config test-config.toml --tui
```

## Test Cases for Task Metadata Features

### 1. Task Metadata Display

**Test Steps:**
1. Launch TUI and press 't' to enter Task Mode
2. Observe the task list display

**Expected Results:**
- Tasks should show in format: `Status [Priority] Title (Area) [Due Date]`
- Status symbols:
  - ○ = open
  - ● = done
  - ◐ = paused
- Priority shows as [p1], [p2], [p3]
- Area shows in parentheses: (development), (work)
- Due dates show special formatting:
  - Overdue: `[DUE: 2025-01-01!]` (with exclamation)
  - Due this week: `[Due: 2025-01-15]`
  - Future: `[2025-02-01]`

### 2. Priority Update with 1/2/3 Keys

**Test Steps:**
1. In Task Mode, navigate to a task
2. Press '1' - should set priority to p1
3. Press '2' - should set priority to p2
4. Press '3' - should set priority to p3
5. Observe status message and list update

**Expected Results:**
- Priority updates immediately in the display
- Status message shows "Priority updated to p1/p2/p3"
- File is actually updated (can verify by pressing 'e' to edit)

### 3. Edit Task with 'e' Key

**Test Steps:**
1. Configure editor in config file (e.g., `editor = "vim"`)
2. In Task Mode, navigate to a task
3. Press 'e'

**Expected Results:**
- External editor opens with the task file
- After saving and exiting, TUI refreshes
- Any metadata changes are reflected in the display

**Also test in Notes Mode:**
- Press 'e' on any note should open it in editor

### 4. Error Handling

**Test Cases:**
1. **No editor configured**: Press 'e' without editor in config
   - Should show "No editor configured"
2. **Invalid task file**: Try updating priority on corrupted task
   - Should show error message
3. **File write permissions**: Test on read-only file
   - Should show appropriate error

### 5. Metadata Persistence

**Test Steps:**
1. Update a task's priority with '1'
2. Exit TUI (press 'q')
3. Check the file directly - frontmatter should show `priority: p1`
4. Relaunch TUI - priority should still be p1

### 6. Task List Features

**Visual Check:**
- Long titles are truncated properly
- Alignment is consistent
- Colors/highlighting work on selected task
- Empty priority/area/due date fields don't break layout

## Sample Task Files

The test-notes directory contains various task files:
- Tasks with all metadata fields
- Tasks with missing optional fields
- Tasks with different statuses
- Overdue tasks for testing date display

## Known Limitations

- 'u' key for full metadata update not yet implemented
- Task-specific sorting (by priority, due date) not yet implemented
- Area filtering not yet implemented

## Summary

**IMPLEMENTED BUT NOT TESTED:**
- Task metadata loading and display
- Priority updates with 1/2/3 keys
- External editor integration with 'e' key
- Proper formatting of tasks in list view