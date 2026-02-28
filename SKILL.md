---
name: atask
description: Local task and project management using the atask CLI. Use when creating tasks, managing projects, tracking work, generating task reports, or organizing personal/work items.
---

# atask -- Task and Project Management

Manage tasks and projects using the atask CLI. Tasks are actionable work items with priorities, due dates, and areas. Projects group related tasks. This is a sibling tool to anote (ideas) and apeople (contacts).

All data is stored as plain markdown files with YAML frontmatter. Filename format: `{ulid}--{slug}__{type}.md`

## Commands

### new -- Create a task

```bash
atask new "Task title" [options]
```

Options:
- `-p, --priority` -- p1 (high), p2 (medium), p3 (low)
- `--due` -- Due date (YYYY-MM-DD or natural language: tomorrow, monday, next week)
- `--area` -- Context (work, personal, etc.)
- `--project` -- Project index_id to associate with (numeric, e.g. `195`)
- `--estimate` -- Time estimate (integer)
- `--tags` -- Comma-separated tags
- `--recur` -- Recurrence pattern (requires `--due`): daily, weekly, monthly, yearly, every Nd/Nw/Nm/Ny, every mon,wed,fri

### list -- List tasks

```bash
atask list [options] --json
```

Default: shows open tasks only, hides done/paused/delegated/dropped tasks and tasks belonging to inactive projects.

Options:
- `--all, -a` -- Show all tasks including completed
- `-p, --priority` -- Filter by priority
- `--area` -- Filter by area
- `--status` -- Filter by status
- `--project` -- Filter by project
- `--overdue` -- Show only overdue tasks
- `--soon` -- Show tasks due soon
- `--search` -- Full-text search in task content
- `--planned-for` -- Filter by planned_for date (today, YYYY-MM-DD, or any)
- `--sort, -s` -- Sort by: modified (default), priority, due, created
- `--reverse, -r` -- Reverse sort order

### show -- Show task details

```bash
atask show <index_id_or_ulid> --json
```

Accepts index_id (numeric) or ULID.

### query -- Complex filtering

```bash
atask query "<expression>" --json [--sort <field>] [--reverse]
```

Boolean operators: `AND`, `OR`, `NOT`, `( )`
Comparison operators: `:` or `=` (equals), `!=` (not equals), `>` `<` (numeric)

Fields:
- `status` -- open, done, paused, delegated, dropped
- `priority` -- p1, p2, p3
- `area` -- any area string
- `project_id` -- project index_id, or special values: `empty`, `set`
- `assignee` -- person responsible
- `due`, `due_date` -- YYYY-MM-DD or special: overdue, today, week, soon, empty, set
- `start`, `start_date` -- YYYY-MM-DD, empty, set
- `estimate` -- numeric comparison (e.g. `estimate>5`)
- `index_id` -- numeric comparison
- `title` -- substring match
- `tag`, `tags` -- matches any tag
- `recur` -- pattern string, or: empty, set
- `content`, `body`, `text` -- full-text search in file content

Examples:
```bash
atask query "status:open AND priority:p1" --json
atask query "area:work AND (due:overdue OR due:today)" --json
atask query "content:blocker AND NOT status:done" --json
atask query "project_id:empty AND due:soon" --json
atask query "tag:sprint-42 AND status:open" --json
```

### update -- Update task metadata

```bash
atask update [options] <task-ids>
```

Task IDs: single (`28`), comma-separated (`28,35,61`), or range (`10-15`). Options come BEFORE IDs.

Options:
- `-p, --priority` -- Set priority
- `--due` -- Set due date
- `--begin` -- Set begin/start date
- `--area` -- Set area
- `--project` -- Set project (index_id)
- `--estimate` -- Set time estimate
- `--status` -- Set status (open, done, paused, delegated, dropped)
- `--title` -- Set title
- `--tags` -- Set tags (comma-separated, use `none` to clear)
- `--recur` -- Set recurrence (use `none` to clear)
- `--plan-for` -- Set planned_for date (natural language, YYYY-MM-DD, or `none` to clear)

Cross-app relationship flags (values are ULIDs):
- `--add-person <ulid>` / `--remove-person <ulid>`
- `--add-task <ulid>` / `--remove-task <ulid>`
- `--add-idea <ulid>` / `--remove-idea <ulid>`

### batch-update -- Conditional bulk update

```bash
atask batch-update --where "<query>" [options]
```

Uses the same query language as `query`. Options: `--priority`, `--status`, `--area`, `--due`, `--project`, `--recur`, `--estimate`.

- `--preview` -- Preview changes without applying them. Always use this first.

```bash
atask batch-update --where "status:paused AND area:work" --status open --preview
atask batch-update --where "due:overdue" --priority p1
atask batch-update --where "tag:sprint-42 AND status:open" --status done
```

### done -- Mark tasks complete

```bash
atask done <task-ids>
```

Accepts same ID formats as update. Recurring tasks automatically create a new instance with the next due date.

### log -- Add timestamped log entry

```bash
atask log <task-id> "message"
```

### project -- Manage projects

```bash
atask project new "Title" [-p priority] [--due date] [--start date] [--area area] [--tags tags]
atask project list [--all] [--area area] [-p priority] [--status status] [--sort field] [--search term] --json
atask project update [options] <project-ids>    # --title, --priority, --due, --start, --area, --status
atask project tasks <project-id> [--all] [--sort field] [--status status] --json
```

Project statuses: active, completed, paused, cancelled.

Project update also supports cross-app relationship flags (`--add-person`, etc.).

## JSON Structure

### Task

```json
{
  "id": "01KJ1KJ3VFJFNDH5K6VEDS2G6G",
  "title": "Fix authentication bug",
  "index_id": 28,
  "type": "task",
  "tags": ["security"],
  "created": "2026-02-15T14:30:00Z",
  "modified": "2026-02-20T10:00:00Z",
  "related_people": ["01KJ1KHY4NFGESK9DDS4YEGH2J"],
  "related_tasks": [],
  "related_ideas": [],
  "file_path": "/path/to/01KJ1KJ3VF...--fix-authentication-bug__task.md",
  "status": "open",
  "priority": "p1",
  "due_date": "2026-02-20",
  "estimate": 5,
  "recur": "weekly",
  "project_id": "195",
  "area": "work",
  "project_name": "Website Redesign"
}
```

`atask list --json` returns `{"tasks": [...]}`. `atask show <id> --json` returns a single task object.

Notes:
- `project_name` appears in `list` output only, not in `show`
- `estimate`, `recur`, `project_id`, `project_name`, `due_date` are omitted from JSON when not set
- `atask show` does not include a `content` field (unlike anote/apeople show)

### Project

```json
{
  "id": "01KJ1TP6WH5VW27KFY98RWEHJB",
  "title": "JIRA Dev",
  "index_id": 195,
  "type": "project",
  "tags": ["project"],
  "created": "2026-02-22T04:45:54Z",
  "modified": "2026-02-22T04:46:03Z",
  "related_people": ["01KJ1KHY4NFGESK9DDS4YEGH2J"],
  "related_tasks": [],
  "related_ideas": [],
  "file_path": "/path/to/01KJ1TP6WH...--jira-dev__project.md",
  "status": "active",
  "priority": "p1",
  "task_count": 1
}
```

`atask project list --json` returns `{"projects": [...]}`.

Key fields:
- `id` -- ULID, the canonical identifier
- `index_id` -- stable numeric ID for CLI commands
- `project_id` -- string of the project's index_id (e.g. `"195"`), not a ULID
- `related_people`, `related_tasks`, `related_ideas` -- arrays of ULIDs (always `[]`, never null)

## Recurring Tasks

When a recurring task is marked done, a new task is automatically created with the next due date.

```bash
atask new "Weekly review" --due monday --recur weekly
atask new "Biweekly 1:1" --due friday --recur "every 2w"
atask new "MWF workout" --due monday --recur "every mon,wed,fri"
```

Patterns: daily, weekly, monthly, yearly, every Nd/Nw/Nm/Ny, every mon,wed,fri.

Late completions advance to the next future date. The new task copies priority, area, project, estimate, tags, and body content. Status resets to open.

## Task States

open, done, paused, delegated, dropped

## Agent Workflows

### Morning review

```bash
atask list --overdue --json
atask list --soon --sort due --json
atask list -p p1 --json
```

### Create project with tasks

```bash
atask project new -p p1 --area work "Website Redesign"
# Parse index_id from output
atask new --project <index_id> -p p1 "Design mockups"
atask new --project <index_id> -p p2 "Frontend implementation"
```

### Cross-app: link task to contact

```bash
atask new "Follow up on proposal" --due "next friday" --json
# Parse id (ULID) from output
apeople update 5 --add-task <task-ulid>
atask update <task-index-id> --add-person <contact-ulid>
```

## Action Queue

The action queue lets agents propose actions for human review instead of executing them directly. Humans can review, modify fields, approve, or reject proposed actions via the web UI or CLI.

### action new -- Propose an action

```bash
atask action new "Title" --action-type <type> [--proposed-by agent-name] [--field key=value ...] [--body "reasoning"] --json
```

Action types: `task_create`, `task_update`, `idea_create`, `idea_update`, `people_update`, `people_log`

Fields vary by action type:
- `task_create`: `title`, `priority`, `due`, `area`, `project` (index_id), `tags` (comma-separated), `estimate`, `add_person` (ULID)
- `task_update`: `target_id` (required), `title`, `status`, `priority`, `due`, `area`, `project`, `plan_for`, `add_person` (ULID)
- `idea_create`: `title`, `kind`, `tags`
- `idea_update`: `target_id` (required), `title`, `state`, `kind`, `maturity`
- `people_update`: `target_id` (required), `state`, `plan_for`
- `people_log`: `target_id` (required), `note` (required), `interaction`

When proposing task actions, include as much context as you can determine:
- **project**: If the task clearly belongs to an existing project, look it up with `atask project list --json` and include the `index_id`
- **add_person**: If the task is related to a specific contact, look them up with `apeople list --json` and include their ULID (`id` field)
- **area**: Set to `work` or `personal` based on context
- **priority**: Set based on urgency/importance signals

Example:
```bash
atask action new "Create follow-up task for Sarah" \
  --action-type task_create \
  --proposed-by claude-code \
  --field title="Follow up on Sarah's proposal" \
  --field priority=p2 \
  --field due=2026-03-05 \
  --field area=work \
  --field project=195 \
  --field add_person=01KJ1KHY4NFGESK9DDS4YEGH2J \
  --body "Noticed task #42 paused 5 days, deadline moved up." \
  --json
```

### action list -- List pending actions

```bash
atask action list [--all] --json
```

### action show -- Show action details

```bash
atask action show <id> --json
```

### action update -- Modify before approval

```bash
atask action update <id> [--field key=value ...] [--title "new title"] [--action-type type] --json
```

### action approve -- Approve and execute

```bash
atask action approve <id> --json
```

Executes the proposed action (e.g., creates the task), archives the action file.

### action reject -- Reject and archive

```bash
atask action reject <id> --json
```

### When to use the action queue

Use `atask action new` instead of direct commands when:
- The user hasn't explicitly asked for the action
- The action involves changes the user should review first
- You're suggesting multiple actions as part of a workflow review
- The context is ambiguous and human judgment would help

For routine operations the user explicitly requests, use direct commands (`atask new`, `atask update`, etc.).

## Configuration

Config: `~/.config/acore/config.toml`

```toml
[directories]
atask = "/path/to/notes"
```

Override with `--dir` flag. Also supports `--config` for alternate config file.

## Global Options

```
--json         JSON output (always use for programmatic access)
--dir PATH     Override task directory
--config PATH  Use specific config file
--quiet, -q    Minimal output
--no-color     Disable color output
--area AREA    Filter by area (global, works with TUI too)
--tui, -t      Launch TUI interface
```
