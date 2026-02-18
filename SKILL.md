---
name: atask
description: Local task and project management using the atask CLI. Use when creating tasks, managing projects, tracking work, generating task reports, or organizing personal/work items.
---

# atask CLI Assistant

Expert in creating and managing tasks and projects using the atask CLI tool. This skill enables agents to help users track tasks, organize projects, and generate reports on their work.

## 🤖 AI Agent Features

**This tool is optimized for AI agent use with:**
- **JSON output** (`--json`) - Machine-readable structured data for all list/query commands
- **Query language** - Build complex filters programmatically without parsing CLI output
- **Full-text search** - Find tasks by content/context (`--search` or `content:` field)
- **Batch operations** - Update multiple tasks with conditional logic (`batch-update --where`)

**🎯 Agent Best Practice:** Prefer JSON output and query language over parsing text output!

---

## 🚀 Quick Decision Tree for Agents

**What do you need to do?**

```
├─ 📖 READ/FIND TASKS
│  ├─ Simple list? → atask list --json
│  ├─ Complex filter (AND/OR)? → atask query "expression" --json
│  ├─ Search by keyword/content? → atask list --search "term" --json
│  └─ Find by content + filters? → atask query "content:term AND field:value" --json
│
├─ ✏️ CREATE TASKS
│  ├─ Single task → atask new "title" [--options]
│  └─ Task in project → atask new "title" --project <project-id> [--options]
│
├─ 🔄 UPDATE TASKS
│  ├─ One/few tasks by ID → atask update --priority p1 28,35
│  ├─ Multiple by condition → atask batch-update --where "query" [--options]
│  └─ Mark done → atask done 28,35
│
├─ 📊 GENERATE REPORTS
│  ├─ Status report → atask query "(priority:p1 OR due:overdue) AND NOT status:done" --json
│  ├─ Project tasks → atask project tasks <id> --json
│  └─ Custom report → atask query "your-query" --json | jq 'your-filter'
│
└─ 💬 ADD NOTES
   └─ Log entry → atask log 28 "message"
```

**🎯 Golden Rule:** If the output will be processed → add `--json` to the command!

---

## Quick Reference

### 🤖 Agent-Optimized Commands (Use These First!)

```bash
# Query with JSON output (PREFERRED for agents)
atask query "status:open AND priority:p1" --json

# Full-text search in content
atask list --search "API integration" --json

# Batch update with conditions
atask batch-update --where "due:overdue" --priority p1 --dry-run

# List with JSON output
atask list --area work --json
atask project list --json
```

### Basic Task Operations

```bash
# Create tasks
atask new "Task title" [options]

# List and filter
atask list [filters]                    # Text output
atask list [filters] --json             # JSON output (preferred for agents)
atask query "filter expression" --json  # Complex filtering (preferred for agents)

# Update tasks
atask update [options] <task-ids>       # Specific IDs
atask batch-update --where "query" [options]  # Conditional (preferred for agents)
atask done <task-ids>

# Other operations
atask log <task-id> <message>
```

### Project Operations

```bash
atask project new "Project title" [options]
atask project list [filters] --json     # JSON output (preferred for agents)
atask project update [options] <project-ids>
atask project tasks <project-id> --json # JSON output (preferred for agents)
```

### Critical Global Flags

```bash
--json                  # 🤖 ALWAYS use for programmatic access
--area <area>           # Filter by area (work, personal, etc.)
--search <term>         # Full-text content search
--config <path>         # Use specific config file
--dir <path>            # Override task directory
```

---

## 🎯 Agent Workflow Guide

**When an agent needs to:**

1. **Find tasks** → Use `query` with `--json`:
   ```bash
   atask query "area:work AND (due:soon OR priority:p1)" --json
   ```

2. **Search by content/keywords** → Use `--search` or `content:`:
   ```bash
   atask list --search "API" --json
   atask query "content:blocker AND status:open" --json
   ```

3. **Get all tasks/projects** → Use `list` with `--json`:
   ```bash
   atask list --json
   atask project list --json
   ```

4. **Update multiple tasks** → Use `batch-update`:
   ```bash
   atask batch-update --where "tag:sprint-42 AND status:open" --status done
   ```

5. **Parse output** → ALWAYS use `--json` and parse with jq:
   ```bash
   atask query "status:open" --json | jq '.[] | .index_id'
   ```

6. **Preview changes** → ALWAYS use `--dry-run` first:
   ```bash
   atask batch-update --where "area:work" --priority p1 --dry-run
   ```

**❌ Avoid:** Parsing text output, using `list` without `--json`, updating tasks one-by-one when batch-update would work

---

## Configuration

**Config file:** `~/.config/atask/config.toml`

**Key settings:**
- `notes_directory` - Where task files are stored (default: `~/tasks`)
- `default_area` - Default area for new tasks
- `soon_horizon` - Days ahead for "due soon" filter (default: 3)
- `[tasks] default_state_filter` - State filter applied at TUI launch (default: `"incomplete"` — hides done tasks and completed projects)

**Check config location:**
```bash
cat ~/.config/atask/config.toml | grep notes_directory
```

## Creating Tasks

### Basic Task Creation

```bash
# Simple task
atask new "Review pull request"

# Task with priority and due date
atask new -p p1 --due tomorrow "Call client"

# Task with all metadata
atask new \
  -p p2 \
  --due "2026-02-20" \
  --area work \
  --estimate 5 \
  --tags "urgent,review" \
  "Fix authentication bug"
```

### Task Creation Options

- `-p, --priority` - Set priority (p1=high, p2=medium, p3=low)
- `--due` - Due date (YYYY-MM-DD or natural language: tomorrow, monday, next week)
- `--area` - Task area/context (work, personal, home, etc.)
- `--project` - Project ID to associate with
- `--estimate` - Time estimate (Fibonacci: 1,2,3,5,8,13)
- `--tags` - Comma-separated tags
- `--recur` - Recurrence pattern (requires `--due`)

### Linking Tasks to Projects

**Get project ID first:**
```bash
atask project list
# Note the Denote ID from output (e.g., 20260215T143010)
```

**Create linked task:**
```bash
atask new \
  --project 20260215T143010 \
  -p p1 \
  "Implement search feature"
```

## Listing and Filtering Tasks

### Basic Listing

```bash
# List open tasks (default)
atask list

# List all tasks including completed
atask list --all

# List with specific filters
atask list -p p1              # Only p1 priority
atask list --area work        # Only work tasks
atask list --status open      # Only open tasks
atask list --overdue          # Only overdue tasks
atask list --soon             # Tasks due soon
```

### Sorting

```bash
# Sort by different fields
atask list --sort priority    # By priority
atask list --sort due         # By due date
atask list --sort created     # By creation date
atask list --sort modified    # By modification date (default)
atask list --sort priority -r # Reverse order
```

### Combined Filters

```bash
# High priority work tasks due soon
atask list -p p1 --area work --soon

# Overdue personal tasks
atask list --area personal --overdue

# All done tasks in work area
atask list --all --status done --area work
```

### Full-Text Search

```bash
# Search in task content (not just title/metadata)
atask list --search "API integration"
atask project list --search "Q1"

# Combine with other filters
atask list --search "blocker" --area work --soon
```

## Advanced Filtering with Query Language

The `query` command provides powerful filtering with boolean expressions.

### Query Syntax

**Boolean Operators:**
- `AND` - Both conditions must be true
- `OR` - Either condition must be true
- `NOT` - Negate a condition
- `( )` - Group expressions for precedence

**Comparison Operators:**
- `:` or `=` - Equals (case-insensitive)
- `!=` - Not equals
- `>` - Greater than (numbers only)
- `<` - Less than (numbers only)

**Searchable Fields:**
- `status` - open, done, paused, delegated, dropped
- `priority` - p1, p2, p3
- `area` - Context/area
- `project_id` - Associated project (use "empty" or "set")
- `assignee` - Person responsible
- `due`, `due_date` - YYYY-MM-DD or special values (overdue, today, week, soon, empty, set)
- `start`, `start_date` - YYYY-MM-DD, empty, set
- `estimate` - Fibonacci numbers (1,2,3,5,8,13)
- `title` - Task title
- `tag`, `tags` - Tags (checks if any tag matches)
- `recur` - Recurrence pattern (use "empty" or "set", or match pattern like "weekly")
- `content`, `body`, `text` - Full-text search in file content
- `index_id` - Numeric ID

### Query Examples

```bash
# Basic queries
atask query "status:open AND priority:p1"
atask query "due:soon OR due:overdue"
atask query "area:work AND NOT status:done"

# Content search in queries
atask query "content:blocker AND area:work"
atask query "status:open AND content:API"

# Complex queries with grouping
atask query "(priority:p1 OR priority:p2) AND status:open"
atask query "area:work AND (due:overdue OR due:today)"

# Special values
atask query "project_id:empty"          # Tasks not in a project
atask query "due:set AND estimate>5"    # Tasks with due date and large estimates
atask query "tag:sprint-42"              # Tasks tagged with sprint-42

# Combine with output formats
atask query "status:open AND area:work" --json
atask query "due:soon" --sort due
```

### When to Use Query vs List

- **Use `list`** for simple filters (single field, common patterns)
  - `atask list -p p1 --area work`
  - `atask list --overdue --soon`

- **Use `query`** for complex logic (AND/OR/NOT, multiple conditions)
  - `atask query "area:work AND (due:overdue OR priority:p1)"`
  - `atask query "NOT project_id:empty AND content:blocker"`

## Updating Tasks

### Update Task Metadata

**Note:** Options must come BEFORE task IDs due to flag parsing.

```bash
# Update single task
atask update -p p2 28

# Update multiple tasks (comma-separated)
atask update --area work 28,35,61

# Update range of tasks
atask update --due "next week" 10-15

# Update mixed IDs
atask update --status paused 28,35-40,61
```

### Update Options

- `-p, --priority` - Change priority
- `--due` - Change due date
- `--area` - Change area
- `--project` - Change project association
- `--estimate` - Change time estimate
- `--status` - Change status (open, done, paused, delegated, dropped)
- `--recur` - Set recurrence pattern (use `none` to clear)

### Marking Tasks Complete

```bash
# Mark single task done
atask done 28

# Mark multiple tasks done
atask done 28,35,61

# Mark range done
atask done 10-15
```

### Recurring Tasks

Tasks can have a recurrence pattern. When a recurring task is marked done (via CLI or TUI), a new task is automatically created with the next due date. Recurring tasks show a `↻` indicator in list output.

**Recurrence requires `--due` to be set.**

```bash
# Create recurring tasks
atask new "Weekly review" --due monday --recur weekly
atask new "Daily standup" --due tomorrow --recur daily
atask new "Monthly report" --due "2026-03-01" --recur monthly
atask new "Biweekly 1:1" --due friday --recur "every 2w"
atask new "MWF workout" --due monday --recur "every mon,wed,fri"

# Supported patterns:
#   daily, weekly, monthly, yearly
#   every <N>d, every <N>w, every <N>m, every <N>y  (e.g. every 2w, every 14d)
#   every mon,wed,fri  (day-of-week, any combination)

# Add recurrence to existing task
atask update --recur weekly --due "2026-02-20" 28

# Remove recurrence
atask update --recur none 28

# When marked done, next instance is auto-created:
atask done 28
# Output: ✓ Task ID 28 marked as done: Weekly review
#         ↻ Created recurring task ID 131: Weekly review (due 2026-02-27)

# Query recurring tasks
atask query "recur:set" --json              # All recurring tasks
atask query "recur:weekly" --json           # Tasks with weekly recurrence
atask query "recur:empty AND due:set" --json  # Non-recurring tasks with due dates
```

**Behavior notes:**
- Late completions advance to the next **future** date (won't create past-due tasks)
- The new task copies priority, area, project, estimate, tags, and body content
- Status resets to `open`; start_date and today_date are cleared
- Works in both CLI (`atask done`) and TUI (press `d` in state menu)

### Batch Update with Conditional Filters

Update multiple tasks at once based on query conditions.

```bash
# Preview changes before applying (dry-run)
atask batch-update --where "area:work AND status:paused" --status open --dry-run

# Update all overdue tasks to high priority
atask batch-update --where "due:overdue" --priority p1

# Move all tasks from one project to another
atask batch-update --where "project_id:20260201T120000" --project 20260215T143010

# Update tasks by tag
atask batch-update --where "tag:sprint-42 AND status:open" --status done

# Clear due dates for paused tasks
atask batch-update --where "status:paused" --due ""
```

**Supported Update Options:**
- `--priority` - Change priority (p1, p2, p3, or "" to clear)
- `--status` - Change status
- `--area` - Change area
- `--due` - Change due date
- `--project` - Change project association
- `--recur` - Set recurrence pattern (use `none` to clear)
- `--dry-run` - Preview changes without applying them

**Important:** Always use `--dry-run` first to preview changes!

## JSON Output for Programmatic Access

All list and query commands support `--json` flag for machine-readable output.

### JSON Output Examples

```bash
# Get tasks as JSON
atask list --json
atask query "status:open" --json

# Projects as JSON
atask project list --json
atask project tasks 15 --json

# Parse with jq
atask query "area:work AND priority:p1" --json | jq '.[] | {id: .index_id, title: .title}'

# Extract specific fields
atask list --area work --json | jq '.[] | .index_id'

# Count tasks by status
atask list --all --json | jq 'group_by(.status) | map({status: .[0].status, count: length})'
```

### JSON Structure

**Task JSON:**
```json
{
  "denote_id": "20260215T143010",
  "slug": "fix-auth-bug",
  "filename_tags": ["task", "urgent", "work"],
  "path": "/path/to/file.md",
  "title": "Fix authentication bug",
  "index_id": 28,
  "type": "task",
  "status": "open",
  "priority": "p1",
  "due_date": "2026-02-20",
  "area": "work",
  "project_id": "20260201T120000",
  "estimate": 5,
  "recur": "weekly",
  "tags": ["urgent", "security"],
  "modified_at": "2026-02-15T14:30:00Z",
  "project_name": "Website Redesign"
}
```

**Project JSON:**
```json
{
  "denote_id": "20260201T120000",
  "slug": "website-redesign",
  "title": "Website Redesign",
  "index_id": 15,
  "status": "active",
  "priority": "p1",
  "due_date": "2026-03-31",
  "start_date": "2026-02-01",
  "area": "work",
  "tags": ["q1-goals"],
  "modified_at": "2026-02-15T14:30:00Z",
  "task_count": 12
}
```

## Task Logging

Add timestamped log entries to tasks:

```bash
# Add log entry
atask log 28 "Discussed with team, waiting for feedback"

# Add progress note
atask log 35 "Completed first draft, ready for review"
```

## Creating Projects

### Basic Project Creation

```bash
# Simple project
atask project new "Q1 Planning"

# Project with metadata
atask project new \
  -p p1 \
  --due "2026-03-31" \
  --start "2026-02-01" \
  --area work \
  --tags "quarterly,planning" \
  "Q1 2026 Goals"
```

### Project Creation Options

- `-p, --priority` - Set priority (p1, p2, p3)
- `--due` - Project due date
- `--start` - Project start date
- `--area` - Project area
- `--tags` - Comma-separated tags

## Managing Projects

### List Projects

```bash
# List active projects (default)
atask project list

# List all projects
atask project list --all

# Filter projects
atask project list --area work
atask project list -p p1
atask project list --status completed
```

### View Project Tasks

```bash
# Show all tasks for a project (by index_id)
atask project tasks 15

# Show only open tasks for project
atask project tasks 15

# Show all tasks including done
atask project tasks 15 --all
```

### Update Projects

```bash
# Update project metadata
atask project update -p p2 15
atask project update --due "2026-04-30" 15
atask project update --status completed 15
```

### Project Status Values

- `active` - Project in progress (default)
- `completed` - Project finished
- `paused` - Project on hold
- `cancelled` - Project abandoned

## Generating Reports

### Daily Task Report

```bash
# Check what's due today/overdue
atask list --overdue
atask list --soon

# Priority tasks for today
atask list -p p1 --soon
```

### Weekly Report

```bash
# Tasks completed this week (manual date filter needed)
atask list --all --status done

# Upcoming tasks
atask list --soon --sort due
```

### Project Status Report

```bash
# All active projects
atask project list

# Projects by area
atask project list --area work
atask project list --area personal

# View tasks for each project
atask project tasks <project-id>
```

### Area-Based Reports

```bash
# Work tasks summary
atask --area work list -p p1

# Personal tasks
atask --area personal list --soon

# All tasks in area
atask --area work list --all
```

## Common Workflows

### Morning Review Workflow

```bash
# 1. Check overdue tasks
atask list --overdue

# 2. Check what's due soon
atask list --soon --sort due

# 3. Review high priority tasks
atask list -p p1 --area work
```

### Creating a New Project with Tasks

```bash
# 1. Create project
atask project new -p p1 --area work "Website Redesign"

# 2. Get project ID from output or list
atask project list

# 3. Create tasks linked to project
atask new --project <project-id> -p p1 "Design mockups"
atask new --project <project-id> -p p2 "Frontend implementation"
atask new --project <project-id> -p p2 "Backend integration"

# 4. View project tasks
atask project tasks <project-id>
```

### Weekly Planning Workflow

```bash
# 1. Review completed work
atask list --all --status done --area work

# 2. Check upcoming deadlines
atask list --soon --sort due

# 3. Review project status
atask project list --area work

# 4. Create tasks for the week
atask new -p p1 --due monday "Plan sprint"
atask new -p p2 --due wednesday "Team sync"
```

### Bulk Task Management

```bash
# Traditional approach: Update specific tasks by ID
atask update --project <project-id> 28,35,61
atask done 10-15

# Modern approach: Conditional batch updates
# 1. Preview what will change
atask batch-update --where "area:work AND status:paused" --status open --dry-run

# 2. Apply the update
atask batch-update --where "area:work AND status:paused" --status open

# 3. Update all overdue tasks to high priority
atask batch-update --where "due:overdue" --priority p1

# 4. Associate all sprint tasks with a project
atask batch-update --where "tag:sprint-42" --project <project-id>
```

### Context-Based Task Management

```bash
# Find tasks by content and context
atask query "content:API AND area:work AND NOT status:done"

# Find tasks mentioning a specific topic
atask list --search "database migration" --area work

# Complex queries for specific situations
atask query "(due:overdue OR due:today) AND priority:p1"

# Export results for external processing
atask query "area:work AND status:open" --json | jq '.[] | {id, title, due_date}'
```

## Understanding IDs

### Denote ID (Canonical)
- Format: `YYYYMMDDTHHMMSS` (e.g., `20260215T143010`)
- Source: Timestamp in filename
- Use: Project associations, unique file identifier
- Example: `20260215T143010--website-redesign__project.md`

### Index ID (CLI Convenience)
- Format: Sequential integer (e.g., `28`)
- Source: `index_id` field in frontmatter
- Use: Quick reference in CLI commands
- Example: `atask done 28`

### How to Reference

```bash
# List shows both IDs
atask list
# Output: "28 ○ [p1] [2026-02-20] Fix authentication bug"
#          ^^ This is the index_id

# Use index_id for CLI commands
atask done 28
atask log 28 "Fixed the issue"

# Use Denote ID for project associations
atask new --project 20260215T143010 "New task"
```

## File Format Reference

### Task File Example
```
Filename: 20260215T143010--fix-auth-bug__task_urgent_work.md
```

```yaml
---
title: Fix authentication bug
index_id: 28
type: task
status: open
priority: p1
due_date: 2026-02-20
area: work
project_id: 20260201T120000
estimate: 5
recur: weekly
tags: [urgent, security]
---

Task description goes here.

## Notes
Additional context.

[2026-02-15] Task created
[2026-02-16] Investigated root cause
```

### Project File Example
```
Filename: 20260201T120000--website-redesign__project_work.md
```

```yaml
---
title: Website Redesign
index_id: 15
type: project
status: active
priority: p1
start_date: 2026-02-01
due_date: 2026-03-31
area: work
tags: [q1-goals, customer-facing]
---

Complete redesign of company website.

## Objectives
- Modernize design
- Improve mobile UX
- Increase conversions
```

## 🤖 Best Practices for Agents

### Critical Rules (ALWAYS Follow)

1. **🎯 ALWAYS use `--json` when reading output:**
   - ✅ `atask list --json`
   - ✅ `atask query "status:open" --json`
   - ❌ NEVER parse text output - use JSON instead

2. **🎯 PREFER `query` over `list` for complex filters:**
   - ✅ `atask query "area:work AND (due:soon OR priority:p1)" --json`
   - ❌ Avoid multiple `list` commands with different flags

3. **🎯 ALWAYS use `--dry-run` before `batch-update`:**
   - ✅ `atask batch-update --where "query" --status done --dry-run`
   - ✅ Then remove `--dry-run` to apply
   - ❌ NEVER batch-update without preview

4. **🎯 Use full-text search for context-based queries:**
   - ✅ `atask list --search "API integration" --json`
   - ✅ `atask query "content:blocker AND area:work" --json`
   - User mentions keywords → search in content, not just titles

5. **🎯 Use `batch-update` for multiple task changes:**
   - ✅ `atask batch-update --where "tag:sprint AND status:open" --status done`
   - ❌ Avoid running `update` or `done` multiple times in a loop

### Task Creation Best Practices

6. **Always use natural language dates** when creating tasks:
   - ✓ `--due tomorrow`
   - ✓ `--due "next friday"`
   - ✓ `--due "2026-02-20"`

7. **Create projects before tasks** when organizing work:
   - Create project first
   - Note the Denote ID from output
   - Link tasks using `--project <denote-id>`

8. **Add log entries** for important updates:
   - Track progress: `atask log 28 "50% complete"`
   - Document blockers: `atask log 28 "Waiting for API access"`

### Query Language Best Practices

9. **Build queries programmatically:**
   - Use parentheses for complex logic: `(A OR B) AND C`
   - Special values for dates: `due:overdue`, `due:soon`, `due:today`
   - Check for empty fields: `project_id:empty`, `due:empty`
   - Combine with `content:` for keyword searches

10. **Parse JSON output correctly:**
    - Extract IDs: `jq '.[] | .index_id'`
    - Filter fields: `jq '.[] | {id: .index_id, title: .title, due: .due_date}'`
    - Count results: `jq 'length'`
    - Check for empty: `jq 'if length == 0 then "No tasks found" else . end'`

## Troubleshooting

### Can't find tasks directory
```bash
# Check config
cat ~/.config/atask/config.toml | grep notes_directory

# Override with flag
atask --dir ~/custom/path list
```

### Task not showing in list
```bash
# Check if filtering is hiding it
atask list --all

# Check specific area
atask list --area work --all
```

### Getting "task not found" error
- Ensure you're using the `index_id` from the frontmatter
- List tasks first to verify ID: `atask list`
- Task IDs are stable - position in list doesn't matter

## Common Patterns

### Sprint Planning
```bash
# Create sprint project
atask project new -p p1 --start monday --due "in 2 weeks" "Sprint 42"

# Add sprint tasks
atask new --project <sprint-id> -p p1 --due "next week" "Feature A"
atask new --project <sprint-id> -p p2 --due "next week" "Feature B"

# Daily standup report
atask project tasks <sprint-id> --status open
```

### Weekly Review
```bash
# Completed work
atask list --all --status done --area work

# Carry-over tasks
atask list --status open --area work

# Projects status
atask project list --area work
```

### Area-Based Workflows
```bash
# Switch contexts with global area filter
atask --area work list --soon      # Work context
atask --area personal list         # Personal context

# Create area-specific tasks
atask new --area work -p p1 "Work task"
atask new --area personal "Personal task"
```

## Integration Tips

When helping users:

1. **Ask about context** - Work, personal, or specific project?
2. **Suggest priorities** - p1 for urgent, p2 for normal, p3 for low priority
3. **Recommend due dates** - Use natural language for user convenience
4. **Link to projects** - Keep related tasks organized
5. **Generate useful reports** - Daily, weekly, or project-based views
6. **Use batch operations** - When updating multiple related tasks
7. **Add log entries** - For important milestones or blockers

## Example Agent Interactions

**User:** "Show me what I need to do today"
```bash
atask list --soon --sort due
atask list --overdue
```

**User:** "Create a project for the website redesign"
```bash
atask project new -p p1 --area work --due "in 6 weeks" "Website Redesign"
# Note the Denote ID from output for future task creation
```

**User:** "Add tasks for the website project"
```bash
# Assuming project ID is 20260215T120000
atask new --project 20260215T120000 -p p1 --due "next week" "Create design mockups"
atask new --project 20260215T120000 -p p2 --due "in 2 weeks" "Implement frontend"
atask new --project 20260215T120000 -p p2 --due "in 3 weeks" "Backend integration"
```

**User:** "What's the status of my work projects?"
```bash
atask project list --area work
atask list --area work -p p1
```

**User:** "Mark these tasks done: 28, 35, 40"
```bash
atask done 28,35,40
```

**User:** "Find all work tasks that mention API and are overdue"
```bash
atask query "area:work AND content:API AND due:overdue"
```

**User:** "Show me high priority tasks as JSON so I can process them"
```bash
atask query "priority:p1 AND status:open" --json
```

**User:** "Move all paused tasks back to open status"
```bash
# Preview first
atask batch-update --where "status:paused" --status open --dry-run

# Then apply
atask batch-update --where "status:paused" --status open
```

**User:** "Find tasks that mention 'blocker' and make them high priority"
```bash
# Find them first
atask list --search "blocker"

# Update them
atask batch-update --where "content:blocker" --priority p1
```

**User:** "Get all tasks for the current sprint (tagged sprint-42) that aren't done yet"
```bash
atask query "tag:sprint-42 AND NOT status:done"
```

**User:** "Show me tasks without a project that are due soon"
```bash
atask query "project_id:empty AND due:soon"
```

---

## AI Agent Integration

**Key Features for Agents:**

1. **JSON Output** - All list/query commands support `--json` for structured data
2. **Query Language** - Build complex filters programmatically without parsing CLI flags
3. **Full-Text Search** - Find tasks by content, not just metadata
4. **Batch Operations** - Update multiple tasks efficiently with conditional logic

**Example Agent Workflow:**

```bash
# 1. Agent searches for relevant tasks based on conversation context
atask query "area:work AND content:authentication" --json | jq '.[] | .index_id'

# 2. Agent creates new task from user request
atask new -p p1 --due "next monday" --area work "Implement OAuth flow"

# 3. Agent updates related tasks
atask batch-update --where "content:auth AND status:open" --project <new-project-id>

# 4. Agent generates status report
atask query "(priority:p1 OR due:overdue) AND NOT status:done" --json
```

---

## 📋 Agent Command Cheat Sheet

**Copy-paste ready templates for common agent tasks:**

```bash
# Find tasks by criteria
atask query "status:open AND priority:p1" --json

# Search by content/keywords
atask list --search "authentication" --json

# Get all open work tasks
atask query "area:work AND status:open" --json

# Find overdue or high-priority tasks
atask query "(due:overdue OR priority:p1) AND NOT status:done" --json

# Update multiple tasks at once
atask batch-update --where "tag:sprint-42" --status done --dry-run

# Get task count by status
atask list --all --json | jq 'group_by(.status) | map({status: .[0].status, count: length})'

# Extract just task IDs
atask query "status:open" --json | jq '.[] | .index_id'

# Find tasks without a project
atask query "project_id:empty" --json

# Full-text search in specific area
atask query "area:work AND content:blocker" --json
```

---

## 🎯 Summary: Agent-First Workflow

1. **Always output JSON** when reading data (`--json`)
2. **Use `query`** for complex filters, not multiple `list` calls
3. **Search content** with `--search` or `content:` when user mentions keywords
4. **Batch update** instead of looping individual updates
5. **Preview first** with `--dry-run` before batch changes
6. **Parse with jq** - never parse text output

**The tool is designed for AI agents** - use the machine-friendly features!

---

**Note:** This CLI tool manages local markdown files in Denote format. All data is stored in plain text files in the configured `notes_directory`. The tool is designed for personal task management with optional project organization.
