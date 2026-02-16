# Project Charter: atask

Version: 2.0.0  
Date: 2025-01-14

## Project Vision

**atask** is a focused task management tool that uses the Denote file naming convention for consistent, unique identifiers. It provides powerful task and project management features through a clean, simple interface.

### Core Philosophy

"Focus on what matters: getting tasks done."

## Design Principles

1. **Task Focus**
   - Only task and project files are managed
   - No general notes or document management
   - Clear, single purpose

2. **Denote Format**
   - Use Denote naming for unique, stable IDs
   - Compatible file format with Emacs Denote
   - No proprietary extensions

3. **Simplicity**
   - Minimal configuration
   - Intuitive operations
   - No feature creep

4. **Two Interfaces, One Purpose**
   - CLI for quick task operations
   - TUI for interactive task management
   - Consistent task-focused functionality

## What We Are Building

### Included Features

- **Task Management**
  - Task files with `__task` tag
  - YAML frontmatter for metadata
  - Status tracking (open, done, paused, delegated, dropped)
  - Priority levels (p1, p2, p3)
  - Due dates with natural language parsing
  - Time estimates
  - Areas for contexts (work, personal, etc.)
  - Task logging with timestamps

- **Project Management**
  - Project files with `__project` tag
  - Task-project associations via Denote IDs
  - Project status tracking
  - Project timeline management
  - View all tasks in a project

- **Interfaces**
  - CLI: `atask task new "Fix bug"`
  - TUI: `atask --tui`
  - Consistent filtering and sorting

- **Organization**
  - Filter by area, priority, status, project
  - Sort by due date, priority, modification time
  - Search by title or tags
  - "Due soon" and overdue tracking

### Explicitly Excluded

- ❌ General note management
- ❌ Document browsing
- ❌ Non-task/project files
- ❌ TaskWarrior integration
- ❌ Multiple file formats
- ❌ Import/export systems
- ❌ File management operations beyond tasks

## Technical Architecture

```
atask/
├── cmd/
│   └── atask/
│       └── main.go          # Entry point
├── internal/
│   ├── denote/              # Denote format operations
│   │   ├── parser.go        # File parsing
│   │   ├── scanner.go       # Task/project scanning
│   │   └── types.go         # Core types
│   ├── task/                # Task logic
│   │   ├── task.go          # Task operations
│   │   └── update.go        # Task updates
│   ├── core/                # Business logic
│   │   └── filter.go        # Task filtering
│   ├── cli/                 # CLI interface
│   │   ├── task_commands.go # Task commands
│   │   └── project_commands.go # Project commands
│   └── tui/                 # TUI interface
│       ├── task_view.go     # Task list view
│       └── project_view.go  # Project view
└── docs/
    ├── DENOTE_TASK_SPEC.md  # File specification
    └── REFACTORING_PLAN.md  # Focus refactoring
```

## Data Format (Denote Task Spec v2.0.0)

### Task File Example

```
Filename: 20250113T143022--fix-search-bug__task_bug_urgent.md
```

```yaml
---
title: Fix search bug in project view
task_id: 234
type: task
status: open
priority: p1
due_date: 2025-01-15
project_id: 20250110T091533  # Project's Denote ID
area: work
---

The search function returns duplicate results when filtering by status.

## Steps to Reproduce
1. Open project view
2. Search for "open" tasks
3. See duplicates

[2025-01-13] Reported by user
```

### Project File Example

```
Filename: 20250110T091533--search-improvements__project.md
```

```yaml
---
title: Search Improvements Q1
project_id: 45
type: project  
status: active
priority: p2
due_date: 2025-03-31
area: work
---

Improve search functionality across the application.
```

## Key Differences from Legacy Tools

| Feature | notes-cli/tui | atask |
|---------|---------------|--------------|
| Purpose | Notes + Tasks | Tasks only |
| File types | All markdown | Task/Project files only |
| Directory structure | Separate tasks/notes | Task directory |
| TaskWarrior | Supported | Not needed |
| Note management | Yes | No |
| Project references | String matching | Denote IDs |
| Complexity | High | Minimal |

## Development Approach

### Phase 1: Core (Week 1-2)
- Denote file operations
- Task/project data structures
- YAML frontmatter parsing
- Spec v2.0.0 compliance

### Phase 2: Business Logic (Week 3-4)
- Filter system
- Sort operations  
- Cache layer
- Project associations

### Phase 3: Interfaces (Week 5-6)
- CLI commands
- TUI views
- Shared argument parsing
- Configuration

### Phase 4: Polish (Week 7-8)
- Testing
- Documentation
- Performance optimization
- Release preparation

## Configuration

Minimal configuration in `~/.config/atask/config.toml`:

```toml
# Required
notes_directory = "~/tasks"  # Where task files live

# Optional
editor = "nvim"
default_area = "work"
soon_days = 7  # Days ahead for "due soon"

[tui]
theme = "default"
```

## Success Criteria

1. **Focus**: Task management only, no scope creep
2. **Simplicity**: <3000 LOC total
3. **Performance**: Instant operations on thousands of tasks
4. **Reliability**: Zero data corruption
5. **Usability**: Intuitive task workflows

## Migration Guide

For users expecting note management:

1. **Purpose Change**: This tool manages tasks only
2. **Note Files**: Use a different tool for general notes
3. **Task Files**: Must have `__task` tag
4. **Project Files**: Must have `__project` tag

For users of notes-cli/tui:

1. **Task Format**: Compatible if using Denote format
2. **Notes**: Extract to separate tool/directory
3. **Projects**: Ensure they have `__project` tag
4. **Focus**: Embrace the simplified workflow

## Design Decisions

1. **Why Task-Only?**
   - Clear purpose and focus
   - Simpler codebase
   - Better user experience
   - No mode confusion

2. **Why Keep Denote Format?**
   - Excellent ID system
   - Guaranteed uniqueness
   - Stable references
   - Good filename convention

3. **Why No General Notes?**
   - Avoid feature creep
   - Maintain focus
   - Reduce complexity
   - Clear boundaries

4. **Why YAML Frontmatter?**
   - Human readable
   - Well-supported
   - Structured data
   - Easy to parse

## Future Possibilities

Once core is solid:

- Task templates
- Recurring tasks
- Time tracking
- Team features
- API access

But always maintaining task focus.

## Getting Started

```bash
# Create new task
atask task new "Write documentation" --priority p1

# List all open tasks
atask task list

# List tasks for a project
atask project tasks <project-id>

# Interactive mode
atask --tui

# Mark task as done
atask task done 123

# Create a project
atask project new "Q1 Goals"
```

## License

MIT License - Keep it simple and open.

## Summary

atask is a focused task management tool that leverages Denote's excellent file naming convention for stable, unique identifiers. By constraining ourselves to task management only, we achieve clarity of purpose, a simpler codebase, and an intuitive user experience. This is not a general-purpose note manager—it's a tool designed to do one thing exceptionally well: help you manage your tasks and projects.