# Session Handoff - Task Mode Implementation

## Continue work on the atask project. Please read these files first:

1. `CLAUDE.md` - Project-specific AI assistant guidelines
2. `PROJECT_CHARTER.md` - Core principles and vision
3. `docs/DENOTE_TASK_SPEC.md` - File format specification (v2.0.0)
4. `PROGRESS.md` - Current implementation status (see Issue #4)

## Current Status (2025-01-13)

### Completed Issues:
- ✅ Issue #1: Basic TUI for browsing all Denote notes
- ✅ Issue #2: General note creation with 'n' key
- ✅ Issue #3: Fuzzy search with in-place filtering
  - True fuzzy matching
  - '#' prefix for tag-only search
  - Real-time results as you type

### Currently Working On: Issue #4 - Task Mode

**IMPORTANT:** This is NOT just a filter. It's a full modal interface change.

#### Requirements:
1. **Notes Mode** (default)
   - Shows all files EXCEPT task/project files
   - Current hotkeys remain unchanged

2. **Task Mode** (toggle with 't' from Notes mode)
   - Shows ONLY task/project files
   - Different hotkey set:
     - 1/2/3 - Set priority
     - e - Edit task
     - u - Update metadata
     - Different sorts/filters for tasks
   - Area filtering is "sticky" and affects new tasks

#### Implementation Progress:
- [ ] Basic mode toggle infrastructure
- [ ] File filtering based on mode
- [ ] Task-specific hotkey handlers
- [ ] Task metadata features

### Key Architecture Notes:
- Tasks are identified by `__task` tag in filename
- Projects are identified by `__project` tag in filename
- Many existing task files in `test-notes/` for testing
- Using Bubble Tea for TUI

### Testing:
```bash
go build ./cmd/atask
./atask --config test-config.toml --tui
```

### CRITICAL REMINDER:
**NEVER mark features as complete without human testing!**
Features must be marked as "IMPLEMENTED BUT NOT TESTED" until verified.