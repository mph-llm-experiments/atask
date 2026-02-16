# Unified Architecture: atask

Version: 2.0.0  
Date: 2025-01-14

## Vision

A focused task management tool with CLI and TUI interfaces, sharing the same underlying data layer and adhering to the Denote Task Specification v2.0.0. This tool is specifically designed for task and project management, not general note-taking.

## Core Principle

"Focus on tasks, not notes" - The CLI and TUI provide different interfaces to a task-focused workflow using Denote's file naming convention for consistency and compatibility.

## Architecture Overview

```
atask
├── cmd/
│   ├── notes/
│   │   ├── main.go           # Entry point, mode detection
│   │   ├── cli.go            # CLI command routing
│   │   └── tui.go            # TUI initialization
│   ├── migrate-projects/     # Migration tools
│   └── validate/             # Validation tools
│
├── internal/
│   ├── denote/               # Core Denote operations
│   │   ├── types.go          # Data structures (Task, Project)
│   │   ├── parser.go         # File parsing
│   │   ├── scanner.go        # Directory scanning
│   │   ├── id_counter.go     # Sequential ID management
│   │   └── spec.go           # Spec version handling
│   │
│   ├── core/                 # Business logic (spec-compliant)
│   │   ├── task.go           # Task operations
│   │   ├── project.go        # Project operations
│   │   ├── filter.go         # Shared filtering logic
│   │   ├── sort.go           # Shared sorting logic
│   │   └── cache.go          # Project metadata cache
│   │
│   ├── cli/                  # CLI interface layer
│   │   ├── commands.go       # Command definitions
│   │   ├── output.go         # Formatting (table, json, etc.)
│   │   └── interactive.go    # CLI prompts
│   │
│   ├── tui/                  # TUI interface layer
│   │   ├── model.go          # Bubble Tea model
│   │   ├── views/            # UI components
│   │   ├── keys.go           # Keyboard handling
│   │   └── theme.go          # Visual themes
│   │
│   └── config/               # Configuration
│       ├── config.go         # Unified config structure
│       └── paths.go          # Standard paths
│
├── pkg/                      # Public API (if needed)
│   └── denote/               # Exportable Denote operations
│
└── docs/
    ├── DENOTE_TASK_SPEC.md   # Specification (source of truth)
    ├── ARCHITECTURE.md       # This document
    └── MIGRATION.md          # Migration guide
```

## Shared Components

### 1. Filter System

Both interfaces use the same filter structure:

```go
type FilterSet struct {
    Area        string
    ProjectID   string      // Denote ID
    Status      []string    
    Priority    []string
    Assignee    string
    Tags        []string
    Overdue     bool
    DueWithin   time.Duration
    TextSearch  string
}
```

### 2. Command-Line Arguments

Shared flags that work in both modes:

```bash
# Filtering (applies to both CLI output and TUI initial state)
--area, -a VALUE          # Filter by area
--project, -p ID          # Filter by project (Denote ID or sequential)
--status, -s VALUE        # Filter by status (open,done,paused,etc.)
--priority VALUE          # Filter by priority (p1,p2,p3)
--assignee VALUE          # Filter by assignee
--tag, -t VALUE           # Filter by tag (repeatable)
--overdue                 # Show only overdue items
--due-within PERIOD       # Due within period (e.g., 1w, 3d)
--search, -/ TEXT         # Full-text search

# Mode selection
--tui                     # Force TUI mode
--cli                     # Force CLI mode (default with args)

# Output (CLI only)
--format, -f FORMAT       # table (default), json, csv, plain
--no-color                # Disable color output
--quiet, -q               # Minimal output

# Common operations
--version                 # Show version and spec compliance
--config PATH             # Use alternate config file
--tasks-dir PATH          # Override tasks directory
--no-cache                # Disable caching
```

### 3. Unified Configuration

Single config file at `~/.config/notes/config.toml`:

```toml
# Core settings (both interfaces)
notes_directory = "~/notes"
tasks_directory = "~/tasks"
default_area = "work"
spec_version = "2.0.0"

# CLI-specific settings
[cli]
default_format = "table"
color_output = true
page_size = 20
date_format = "2006-01-02"

# TUI-specific settings  
[tui]
theme = "default"
editor = "nvim"
preview_command = "glow"
show_titles = true
confirm_delete = true

# Caching (both interfaces)
[cache]
enabled = true
ttl_seconds = 300
max_entries = 1000
```

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)

1. **Create unified repository structure**
   - Set up new `notes` repository
   - Import shared code from both projects
   - Set up build system

2. **Extract and merge denote package**
   - Merge types, parser, scanner from both
   - Ensure Spec v2.0.0 compliance
   - Add spec version checking

3. **Build core business logic layer**
   - Task CRUD operations
   - Project CRUD operations
   - Filter/sort implementations
   - Cache system for project lookups

### Phase 2: Interface Layers (Weeks 3-4)

4. **Implement CLI interface**
   - Port notes-cli commands
   - Use cobra or similar for command structure
   - Add filter flag parsing
   - Implement output formatters

5. **Implement TUI interface**
   - Port notes-tui UI components
   - Integrate with shared filter system
   - Add project selection UI (per issue #49)
   - Implement context-aware features (per issue #53)

### Phase 3: Integration (Weeks 5-6)

6. **Unified entry point**
   - Mode detection logic
   - Shared flag parsing
   - Config loading
   - Initial filter application

7. **Testing and validation**
   - Unit tests for core logic
   - Integration tests for both interfaces
   - Spec compliance validation
   - Performance testing with large datasets

### Phase 4: Migration (Week 7)

8. **Migration tools**
   - Config migration from both tools
   - Update path for existing users
   - Symlink strategy for compatibility
   - Documentation updates

9. **Release preparation**
   - Update all documentation
   - Create migration guide
   - Beta testing period
   - Coordinated deprecation plan

## Spec Compliance

### Denote Task Spec v2.0.0 Requirements

1. **File Format**: ✓ Already compliant
2. **Project References**: Must migrate to Denote IDs
3. **Required Fields**: Add validation for title, type
4. **Counter File**: Rename to `.denote-task-counter.json`
5. **Spec Version**: Track in counter file and config

### Migration for Spec v2.0.0

```bash
# Built-in migration command
notes migrate --to-spec 2.0.0

# What it does:
# 1. Updates project references from strings to Denote IDs
# 2. Adds missing title fields
# 3. Adds type field to all tasks/projects
# 4. Renames counter file
# 5. Updates config with spec version
```

## Benefits of Unification

1. **Consistency**: Same behavior regardless of interface
2. **Efficiency**: Single codebase to maintain
3. **Features**: All improvements benefit both interfaces
4. **Learning**: One set of concepts and filters
5. **Flexibility**: Easy to switch between interfaces
6. **Future-proof**: Easy to add new interfaces (web, API)

## Example Usage Patterns

```bash
# Quick CLI operations
notes task add "Fix bug" --project 20250627T191225
notes task list --area work --overdue
notes task done 45

# Launch TUI with filters
notes --tui --area work
notes --tui --project 20250627T191225

# Seamless switching
alias n='notes'
alias nt='notes --tui'
n task list --area work     # Quick check in CLI
nt --area work              # Deep dive in TUI

# Advanced combinations
notes task list --area work --format json | jq '.[] | select(.priority=="p1")'
notes --tui --search "database" --status open
```

## Success Criteria

1. **Feature parity**: Both interfaces can do everything the original tools could
2. **Performance**: No regression in speed or responsiveness  
3. **Compatibility**: Existing workflows continue to work
4. **Spec compliance**: Full adherence to Denote Task Spec v2.0.0
5. **User satisfaction**: Smooth migration path, clear benefits

## Open Questions

1. **Binary name**: `notes`, `atask`, something else?
2. **Repository**: New unified repo or evolve one of the existing?
3. **Versioning**: Start at 1.0.0 or continue from one tool's version?
4. **Package structure**: Export public API for third-party tools?
5. **Plugin system**: Allow custom commands/views?

## Next Steps

1. Create GitHub issues for roadmap phases
2. Set up new repository structure
3. Begin extracting shared code
4. Create detailed design docs for each component
5. Establish testing strategy