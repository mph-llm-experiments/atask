# Changelog

All notable changes to denote-tasks will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.18.0] - 2026-02-15

### Added
- **Tag for Today feature** for morning time-blocking workflow
  - Tasks tagged for today appear at top of list with ★ indicator
  - Visual separator line between today tasks and regular tasks
  - Hotkey 'y' to toggle today tag on selected task
  - Hotkey 'Y' (shift) to clear all today tags with confirmation dialog
  - Auto-clears overnight when date changes (no manual cleanup needed)
  - Query language support: `today:tagged`, `today_date:YYYY-MM-DD`
  - Added `TodayDate` field to TaskMetadata
- SKILL.md documentation section for "Daily Planning with Tag for Today"
  - Proactive agent behavior patterns
  - CLI and TUI usage examples
  - Agent workflow examples for morning planning

### Fixed
- **TUI alignment issues** - tasks and projects now align perfectly
  - Fixed nested ANSI escape sequences causing column misalignment
  - Pad text BEFORE applying color instead of after
  - Avoid wrapping entire lines in additional .Render() calls
  - Projects and tasks use identical line format
- Stable sort now properly handles tasks vs projects (all tasks before all projects)

## [0.17.3] - 2026-02-15

### Fixed
- Made "due today" cutline responsive to terminal width - now adjusts dynamically when window is resized

## [0.17.2] - 2026-02-15

### Fixed
- Added footer wrapping to task and project detail views for better responsiveness

## [0.17.1] - 2026-02-15

### Added
- Added SKILL.md to repository root for AI agent integration
- Added skill installation instructions to README

## [0.17.0] - 2026-02-15

### Added
- **JSON output** for all list/query commands (`--json` flag) - enables programmatic parsing for AI agents
- **Query command** with full boolean expression support (AND/OR/NOT with parentheses)
- **Full-text search** in task/project content (`--search` flag and `content:` field in queries)
- **Batch-update command** with conditional filters (`--where` clause) for updating multiple tasks at once
- Support for complex queries with comparison operators (`:`, `=`, `!=`, `>`, `<`)
- Special date values in queries (`due:overdue`, `due:soon`, `due:today`, `due:week`)
- Enhanced query language with 15+ searchable fields (status, priority, area, content, tags, etc.)
- Comprehensive AI agent integration documentation and skill file

### Fixed
- Fixed project status menu in TUI (now shows active/completed/paused/cancelled instead of task statuses)
- Removed restriction preventing state changes on projects in TUI
- Added responsive footer wrapping based on terminal width

### Changed
- Updated README with query language reference and AI agent integration guide
- Created comprehensive SKILL.md for AI agents with decision trees and best practices

## [0.16.1] - 2025-01-14

### Fixed
- Various TUI improvements and bug fixes

## Earlier Versions

See git tags and commit history for changes prior to v0.17.0.

[0.17.3]: https://github.com/mph-llm-experiments/denote-tasks/compare/v0.17.2...v0.17.3
[0.17.2]: https://github.com/mph-llm-experiments/denote-tasks/compare/v0.17.1...v0.17.2
[0.17.1]: https://github.com/mph-llm-experiments/denote-tasks/compare/v0.17.0...v0.17.1
[0.17.0]: https://github.com/mph-llm-experiments/denote-tasks/compare/v0.16.1...v0.17.0
[0.16.1]: https://github.com/mph-llm-experiments/denote-tasks/releases/tag/v0.16.1
