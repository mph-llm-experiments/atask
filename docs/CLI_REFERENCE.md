# atask CLI Reference

## Overview

atask is a focused task management tool using the Denote file naming convention. It provides both CLI and TUI interfaces for managing tasks and projects.

The CLI uses an implicit task-first command structure:

```
atask <action> [options] [arguments]      # For tasks (implicit)
atask project <action> [options] [arguments]  # For projects (explicit)
```

**Important**: Task operations use `index_id` values from the files, not display positions. When you see task "28" in the list, you use 28 in commands.

## Global Options

These options can be used with any command:

- `--config PATH` - Use specific config file
- `--dir PATH` - Override task directory  
- `--area AREA` - Filter by area (for TUI or commands)
- `--tui, -t` - Launch TUI interface
- `--json` - Output in JSON format (not yet implemented)
- `--no-color` - Disable color output
- `--quiet, -q` - Minimal output

## Task Commands (Implicit)

Task commands don't require the "task" prefix - they're the default.

### new

Create a new task.

```bash
atask new [options] <title>
```

Options:
- `-p, --priority` - Set priority (p1, p2, p3)
- `--due` - Set due date (YYYY-MM-DD or natural language)
- `--area` - Set task area
- `--project` - Set project ID
- `--estimate` - Set time estimate
- `--tags` - Comma-separated tags

Examples:
```bash
atask new "Review budget proposal"
atask new -p p1 --due tomorrow "Call client"
atask new --area work --project 20240315T093000 "Update docs"
```

### task list

List tasks with filtering and sorting.

```bash
atask list [options]
```

Options:
- `-a, --all` - Show all tasks (default: open only)
- `--area` - Filter by area
- `--status` - Filter by status
- `-p, --priority` - Filter by priority (p1, p2, p3)
- `--project` - Filter by project ID
- `--overdue` - Show only overdue tasks
- `--soon` - Show tasks due soon
- `-s, --sort` - Sort by: modified (default), priority, due, created
- `-r, --reverse` - Reverse sort order

Examples:
```bash
atask list                    # List open tasks
atask list --all              # List all tasks
atask list -p p1              # List only p1 tasks
atask list --area work        # List work tasks
atask list --overdue          # List overdue tasks
atask list --sort priority    # Sort by priority
```

### task update

Update task metadata. **Note**: Options must come before task IDs.

```bash
atask update [options] <task-ids>
```

Options:
- `-p, --priority` - Set priority (p1, p2, p3)
- `--due` - Set due date
- `--area` - Set area
- `--project` - Set project ID
- `--estimate` - Set time estimate
- `--status` - Set status (open, done, paused, delegated, dropped)

Task IDs support:
- Single: `28`
- List: `28,35,61`
- Range: `28-35`
- Mixed: `28,35-40,61`

Examples:
```bash
atask update -p p2 28                # Change priority
atask update --due "next week" 35    # Set due date
atask update --status paused 28,35   # Pause multiple tasks
atask update --area personal 10-15   # Update area for range
```

### task done

Mark tasks as done.

```bash
atask done <task-ids>
```

Examples:
```bash
atask done 28           # Mark single task as done
atask done 28,35,61     # Mark multiple tasks as done
atask done 10-15        # Mark range as done
```

### task log

Add a timestamped log entry to a task.

```bash
atask log <task-id> <message>
```

Examples:
```bash
atask log 28 "Discussed with team, waiting for feedback"
atask log 35 "Completed first draft"
```

### task edit (not implemented)

Edit task in external editor or TUI.

```bash
atask edit <task-id>
```

### task delete (not implemented)

Delete tasks with confirmation.

```bash
atask delete <task-ids>
```

## Project Commands (not implemented)

```bash
atask project new <title>
atask project list [options]
atask project update [options] <project-ids>
atask project archive <project-ids>
```


## TUI Launch Examples

```bash
# Launch TUI
atask --tui

# Launch TUI filtered to work area
atask --tui --area work

# Launch TUI filtered to personal area
atask --tui --area personal
```

## Examples

### Using global area filter

```bash
# List only work tasks
atask --area work task list

# List only personal tasks with p1 priority
atask --area personal task list -p p1

# The global --area flag works with any command
atask --area work done 28,35
```

### Daily workflow

```bash
# Check what's due today
atask list --overdue
atask list --soon

# Add a new urgent task
atask new -p p1 --due today "Fix critical bug"

# Update task priority after meeting
atask update -p p2 28

# Log progress
atask log 28 "Found root cause, working on fix"

# Mark completed
atask done 28
```

### Bulk operations

```bash
# Update multiple tasks to a project
atask update --project 20240315T093000 28,35,61

# Mark a range as done
atask done 10-15

# Change area for all personal tasks
atask list --area personal  # See the IDs
atask update --area home 4,7,12-15,23
```

## Tips

1. **Index IDs are stable**: The number shown (e.g., 28) is the task's permanent ID, not its position in the list.

2. **Flags before IDs**: Due to Go's flag parsing, options must come before task IDs:
   - ✓ `atask update -p p1 28`
   - ✗ `atask update 28 -p p1`

3. **Natural date parsing**: The `--due` flag accepts natural language:
   - `today`, `tomorrow`, `next week`
   - `monday`, `next friday`
   - `2025-01-15` (ISO format)

4. **Filtering is additive**: Multiple filters work together:
   ```bash
   atask list -p p1 --area work --soon
   ```
   Shows only p1 work tasks due soon.