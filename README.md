# atask

A focused task management tool using the Denote file naming convention for portability. This project has no association with the Denote project.

## Important consideration before using this code or interacting with this codebase

This application is an experiment in using Claude Code as the primary driver the development of a small, focused app that concerns itself with the owner's particular point of view on the task it is accomplishing.

As such, this is not meant to be what people think of as "an open source project," because I don't have a commitment to building a community around it and don't have the bandwidth to maintain it beyond "fix bugs I find in the process of pushing it in a direction that works for me."

It's important to understand this for a few reasons:

1. If you use this code, you'll be using something largely written by an LLM with all the things we know this entails in 2025: Potential inefficiency, security risks, and the risk of data loss.

2. If you use this code, you'll be using something that works for me the way I would like it to work. If it doesn't do what you want it to do, or if it fails in some way particular to your preferred environment, tools, or use cases, your best option is to take advantage of its very liberal license and fork it.

3. I'll make a best effort to only tag the codebase when it is in a working state with no bugs that functional testing has revealed.

While I appreciate and applaud assorted efforts to certify code and projects AI-free, I think it's also helpful to post commentary like this up front: Yes, this was largely written by an LLM so treat it accordingly. Don't think of it like code you can engage with, think of it like someone's take on how to do a task or solve a problem.

That said:

## What it is

`atask` is a specialized tool for managing tasks and projects using plain text files with the Denote naming convention. Each task is a Markdown file with YAML frontmatter containing task metadata (priority, due date, project assignment, etc.).

## What it isn't

- **Not a general note-taking app** - Use Denote in Emacs or other tools for that
- **Not a calendar** - Though it tracks due dates
- **Not a time tracker** - Though it supports time estimates
- **For most people?** - It's very aligned with one particular person's idea of just how they wanted CLI/TUI task management to work.

## Quick Start

```bash
# Install
go install github.com/mph-llm-experiments/atask@latest

# Create your first task
atask new -p p1 --due tomorrow "Review pull request"

# Launch the TUI
atask --tui

# Filter by area in TUI
atask --tui --area work
```

## Features

- **Task-focused** - Built specifically for task management, not general notes
- **Works with Denote** - Uses standard Denote file naming for compatibility
- **Project support** - Organize tasks by project with automatic linking
- **Dual interface** - Both CLI and TUI for different workflows
- **Advanced filtering** - Query language with boolean expressions (AND/OR/NOT)
- **Full-text search** - Search within task/project content, not just metadata
- **JSON output** - Machine-readable output for AI agents and automation
- **Batch operations** - Update multiple tasks with conditional filters

### AI Agent Integration

`atask` is designed to work seamlessly with AI agents like Claude, localgpt, and openclaw:

- **JSON output** (`--json`) for all list/query commands enables programmatic parsing
- **Query language** allows agents to construct complex filters without parsing CLI flags
- **Full-text search** (`--search` or `content:`) helps agents find tasks by context
- **Batch updates** enable agents to modify multiple tasks in one operation

Example agent workflow:
```bash
# Agent searches for relevant tasks
atask query "area:work AND content:API" --json | jq '.[] | .index_id'

# Agent updates multiple tasks
atask batch-update --where "tag:sprint-42 AND status:open" --status done

# Agent creates task from conversation context
atask new -p p1 --due "next monday" --area work "Implement OAuth flow"
```

## Installation

```bash
go install github.com/mph-llm-experiments/atask@latest
```

Or for a specific version:

```bash
go install github.com/mph-llm-experiments/atask@v0.17.0
```

## Usage

```bash
# Create a new task
atask new "Fix search bug"
atask new -p p1 --due tomorrow "Call client"

# List tasks
atask list
atask list -p p1 --area work
atask list --json  # Machine-readable output

# Search in content
atask list --search "API integration"
atask project list --search "Q1"

# Advanced queries with boolean logic
atask query "status:open AND priority:p1"
atask query "due:soon OR due:overdue"
atask query "area:work AND content:blocker"
atask query "(priority:p1 OR priority:p2) AND NOT status:done"

# Update tasks (uses index_id from list)
atask update -p p2 28
atask done 28,35

# Batch update with conditions
atask batch-update --where "area:work AND status:paused" --status open
atask batch-update --where "due:overdue" --priority p1 --dry-run

# Add log entries
atask log 28 "Found root cause"

# Interactive TUI
atask --tui
atask --tui --area work  # Start filtered by area

# Project management
atask project new "Q1 Planning"
atask project list
atask project list --json
atask project tasks 15  # Show tasks for project
```

### TUI Hotkeys

**Navigation:**

- `j/k` or `↓/↑` - Move down/up
- `g g` - Go to top
- `G` - Go to bottom
- `Enter` - Open task/project details

**Actions (lowercase):**

- `c` - Create new task or project
- `d` - Edit due date
- `l` - Add log entry (tasks only)
- `r` - Toggle sort order
- `s` - Change state (task: open/done/paused/delegated/dropped; project: active/completed/paused/cancelled)
- `t` - Edit tags
- `u` - Update task metadata
- `x` - Delete task/project
- `/` - Search (use `#tag` for tag search)

**Priority:**

- `0` - Clear priority
- `1/2/3` - Set priority (p1/p2/p3)

**Filters & Views (uppercase):**

- `E` - Edit in external editor
- `P` - Toggle projects view
- `T` - Toggle tasks view
- `S` - Sort options menu
- `f` - Filter menu (area/priority/state/loose/soon)

**General:**

- `?` - Help screen
- `q` - Quit

See [CLI Reference](docs/CLI_REFERENCE.md) for full command documentation.

## Query Language

The `query` command supports complex filtering with boolean expressions:

**Boolean Operators:**
- `AND` - Both conditions must be true
- `OR` - Either condition must be true
- `NOT` - Negate a condition
- `( )` - Group expressions

**Comparison Operators:**
- `:` or `=` - Equals (case-insensitive)
- `!=` - Not equals
- `>` - Greater than (numbers only)
- `<` - Less than (numbers only)

**Searchable Fields:**
- `status` - Task status (open, done, paused, delegated, dropped)
- `priority` - Priority level (p1, p2, p3)
- `area` - Context/area
- `project_id` - Associated project (use "empty" or "set")
- `assignee` - Person responsible
- `due`, `due_date` - Due date or special values (overdue, today, week, soon, empty, set)
- `start`, `start_date` - Start date (YYYY-MM-DD, empty, set)
- `estimate` - Time estimate (Fibonacci numbers)
- `title` - Task title
- `tag`, `tags` - Tags (checks if any tag matches)
- `content`, `body`, `text` - Full-text search in file content
- `index_id` - Numeric ID

**Examples:**

```bash
# High priority open tasks
atask query "status:open AND priority:p1"

# Tasks due soon or overdue
atask query "due:soon OR due:overdue"

# Work tasks with specific content
atask query "area:work AND content:blocker"

# Complex queries with grouping
atask query "(priority:p1 OR priority:p2) AND NOT status:done"

# Tasks without a project
atask query "project_id:empty"

# Tasks with estimates over 5
atask query "estimate>5"

# Combine with output formats
atask query "status:open AND tag:v2mom" --json
```

## Configuration

Create `~/.config/atask/config.toml`:

```toml
notes_directory = "~/tasks"  # Where task files live (kept for backward compatibility)
editor = "vim"              # External editor for 'E' command
default_area = "work"       # Default area for new tasks
soon_horizon = 3            # Days ahead for "soon" filter

[tui]
theme = "default"           # UI theme

[tasks]
sort_by = "due"                        # Default sort: due, priority, project, title, created
sort_order = "normal"                  # normal or reverse
default_state_filter = "incomplete"    # Hide completed tasks at launch (incomplete, active, or "" for none)
```

## AI Agent Skill Installation

For AI agents (Claude Code, etc.), install the skill file for enhanced integration:

```bash
# Create skill directory if it doesn't exist
mkdir -p ~/.claude/skills/atask

# Copy the skill file
cp SKILL.md ~/.claude/skills/atask/

# The skill will be automatically available to Claude-based agents
```

The skill file provides comprehensive guidance for AI agents, emphasizing machine-friendly features like JSON output, query language, and batch operations.

## Documentation

- [Project Charter](PROJECT_CHARTER.md) - Vision and goals
- [Denote Task Specification](docs/DENOTE_TASK_SPEC.md) - File format (v2.0.0)
- [Architecture](docs/UNIFIED_ARCHITECTURE.md) - Technical design
- [AI Agent Skill](SKILL.md) - Comprehensive guide for AI agents

## Task File Format

Tasks are stored as markdown files with Denote naming:

```
20240315T093000--fix-search-bug__task_p1_work.md
┗──────────────┘ ┗─────────────┘┗───────────┘
   Denote ID       Title slug        Tags
```

With YAML frontmatter:

```yaml
---
title: Fix search bug in task list
index_id: 28
type: task
status: open
priority: p1
due_date: 2024-03-16
project_id: 20240301T100000
area: work
---
## Description
Search is not filtering tasks correctly when...
```

## License

MIT License
