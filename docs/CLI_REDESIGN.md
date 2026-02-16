# CLI Redesign Plan

## Overview

Replace the current basic CLI with a comprehensive entity-first command structure that:
1. Provides full CLI functionality for scripting and quick operations
2. Can launch the TUI with pre-applied filters using `-tui` flag
3. Maintains consistency between CLI and TUI operations

## Command Structure

### Base Commands

```bash
# Launch TUI directly
atask --tui
atask -t

# Entity-based commands
atask task ...
atask project ...
atask note ...
```

### Task Commands

```bash
# Create new task
atask task new "Fix login bug" -p p1 -due tomorrow -area work

# List tasks (CLI output)
atask task list
atask task list -p1 -overdue -area work

# Launch TUI with filters
atask task list -tui                    # TUI in task mode
atask task list -tui -p1 -area work    # TUI with filters applied
atask task list -tui -project webapp    # TUI filtered to project

# Update tasks
atask task update 3 -status done
atask task update 3-5 -p p2
atask task update 3,5,7 -project newproject

# Quick actions
atask task done 3-5
atask task log 3 "Started working on this"
atask task edit 3        # Opens in external editor
atask task edit 3 -tui   # Opens in TUI task view

# Delete tasks
atask task delete 3,5
```

### Project Commands

```bash
# Create project
atask project new "Website Redesign" -area work -p p1

# List projects
atask project list
atask project list -tui              # TUI in project filter mode
atask project list -tui -area work   # TUI with area filter

# View project
atask project show "Website Redesign"
atask project tasks 1                 # List tasks for project
atask project view 1 -tui            # Open project in TUI

# Update project
atask project update 1 -status completed
```

### Note Commands

```bash
# Create note
atask note new "Meeting Notes" -tags "meeting,q1"

# List notes
atask note list
atask note list -tui                 # TUI in notes mode
atask note list -tag daily -tui      # TUI with tag filter

# Edit note
atask note edit 3
atask note edit 3 -tui              # Open in TUI preview
```

### Global Flags

```bash
# Configuration
--config PATH       # Use specific config file
--dir PATH         # Override notes directory

# Output control  
--json            # JSON output for scripting
--no-color        # Disable color output
--quiet           # Minimal output

# TUI launch
--tui, -t         # Launch TUI with context
```

## Implementation Strategy

### Phase 1: Core Structure
1. Implement command routing (task/project/note subcommands)
2. Add argument parsing for ranges and lists
3. Implement basic list/new/update/delete for tasks

### Phase 2: TUI Integration
1. Add -tui flag support to relevant commands
2. Pass filters and context to TUI initialization
3. Ensure TUI can start in specific modes/views

### Phase 3: Advanced Features
1. Project management commands
2. Note operations
3. Shell completions
4. JSON output mode

### Phase 4: Polish
1. Backward compatibility aliases
2. Color output with NO_COLOR support
3. Progress indicators for bulk operations
4. Error handling and confirmations

## Key Benefits

1. **Flexible Usage**: Same filters work for CLI output or TUI launch
2. **Discoverability**: Entity-first structure is intuitive
3. **Power User Features**: Ranges, bulk operations, scriptability
4. **Consistency**: CLI and TUI share the same concepts
5. **Context Preservation**: TUI launches with relevant filters/mode

## Examples of Dual-Purpose Commands

```bash
# Morning review - CLI
atask task list -due today -area work

# Morning review - TUI
atask task list -due today -area work -tui

# Project dashboard - CLI
atask project tasks "webapp" -status open

# Project dashboard - TUI  
atask project view "webapp" -tui

# Quick task entry then review
atask task new "Call client" -p p1 -due today
atask task list -due today -tui
```