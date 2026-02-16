# Task Mode Test Plan

## Test Environment Setup

```bash
# Build the application
go build ./cmd/atask

# Run with test config
./atask --config test-config.toml --tui
```

## Test Cases for Issue #4

### 1. Mode Toggle Infrastructure

**Test Steps:**
1. Launch TUI - should start in Notes Mode
2. Check header shows "Denote Notes"
3. Check status line shows "X notes"
4. Press 't' key
5. Verify header changes to "Denote Tasks"
6. Verify status line shows "X tasks"
7. Press 't' again to return to Notes Mode

**Expected Results:**
- Mode toggles correctly between Notes and Tasks
- Status message briefly shows "Task Mode" or "Notes Mode"
- File list updates to show appropriate files

### 2. File Filtering

**In Notes Mode:**
- Should see all files EXCEPT those with `__task` or `__project` tags
- Example files that should appear:
  - `20250101T120000--meeting-notes__work`
  - `20250102T090000--project-ideas`

**In Task Mode:**
- Should see ONLY files with `__task` or `__project` tags
- Example files that should appear:
  - `20250103T140000--implement-feature__task`
  - `20250104T100000--website-redesign__project`

### 3. Task-Specific Hotkeys

**Test in Task Mode:**
1. Press '1', '2', or '3' - should show "Priority setting not yet implemented"
2. Press 'e' - should show "Edit task not yet implemented"
3. Press 'u' - should show "Update metadata not yet implemented"
4. Press '?' - should show Task mode help with different hotkeys
5. Common keys should still work: j/k, /, enter, etc.

### 4. UI Updates

**Footer in Notes Mode:**
- Should show: `j/k:nav • /:search • enter:preview • n:new • t:task mode • s:sort • r:reverse • ?:help • q:quit`

**Footer in Task Mode:**
- Should show: `j/k:nav • /:search • enter:preview • 1/2/3:priority • e:edit • u:update • t:notes mode • ?:help • q:quit`

**Help Screen:**
- Press '?' in each mode
- Should show mode-specific help text
- Task mode should list task-specific actions

### 5. Search Functionality

**Test that search works in both modes:**
1. In Notes Mode, press '/' and search - should filter notes only
2. In Task Mode, press '/' and search - should filter tasks only
3. Tag search with '#' should work in both modes

### 6. Empty States

**If no tasks exist:**
- Task mode should show "No tasks found"

**If no notes exist (all files are tasks):**
- Notes mode should show "No notes found"

## Summary

This completes the basic infrastructure for Issue #4. The mode toggle, file filtering, and UI updates are all implemented. The task-specific functionality (priority setting, metadata updates) will be implemented when Issue #5 (task metadata handling) is completed.

**IMPORTANT:** All features are IMPLEMENTED BUT NOT TESTED. Human testing required before marking as complete.