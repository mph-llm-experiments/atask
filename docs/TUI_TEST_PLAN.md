# TUI Test Plan for Issue #1

## Overview

This document provides a comprehensive test plan for the basic TUI implementation that browses Denote notes.

## Test Environment Setup

1. Build the application:
   ```bash
   go build ./cmd/atask
   ```

2. Use the provided test configuration:
   ```bash
   ./atask --config test-config.toml --tui
   ```

3. Or create your own test environment:
   ```bash
   mkdir ~/test-denote-notes
   # Copy various Denote-formatted .md files there
   # Edit test-config.toml to point to your directory
   ```

## Feature Tests

### 1. Initial Display

**Test**: Launch the TUI
**Expected**:
- Header shows "Denote Notes"
- Status line shows file count and sort info
- File list displays with columns: ID, Title, Tags
- Footer shows keyboard shortcuts
- No crash or panic

### 2. Navigation

**Test**: Basic movement
- Press `j` or `↓` - Cursor moves down
- Press `k` or `↑` - Cursor moves up
- Press `gg` - Jump to first file
- Press `G` - Jump to last file

**Expected**: Smooth navigation, selection highlight follows cursor

### 3. Search Function

**Test**: Search for files
- Press `/` - Search prompt appears
- Type search term (e.g., "task")
- Press `Enter` - List filters to matching files
- Press `/` again and `Esc` - Cancel search

**Expected**: 
- Search filters by title, slug, and tags
- Case-insensitive search
- File count updates in header

### 4. Sorting

**Test**: Sort options
- Press `s` - Toggle between date/title sort
- Press `r` - Reverse sort order

**Expected**:
- Header shows current sort mode
- Files reorder appropriately
- Selection stays on same file if possible

### 5. File Preview

**Test**: View file content
- Select a file and press `Enter`
- Press `q` or `Esc` to exit preview

**Expected**:
- Full file content displays
- Can exit back to list view
- No formatting errors

### 6. Help Screen

**Test**: Show help
- Press `?` - Help screen appears
- Press any key to exit

**Expected**: 
- Clear keyboard shortcut list
- Returns to previous view

### 7. Quit

**Test**: Exit application
- Press `q` from main view

**Expected**: Clean exit to terminal

## Edge Cases

### Empty Directory

**Test**: Point config to empty directory
**Expected**: Message "No notes found"

### Mixed Files

**Test**: Directory with both Denote and non-Denote files
**Expected**: Only Denote-formatted files appear

### Large Directory

**Test**: Directory with 100+ files
**Expected**: 
- Performance remains responsive
- Scrolling works smoothly
- No memory issues

### Long Filenames

**Test**: Files with very long titles
**Expected**: Titles truncate with "..." appropriately

### Missing Frontmatter

**Test**: Denote files without YAML frontmatter
**Expected**: Title derived from filename slug

## Regression Tests

After any changes, verify:
1. All keyboard shortcuts still work
2. No panic on unexpected input
3. Config file is never modified
4. Test files are never modified

## Performance Checklist

- [ ] Startup time < 1 second
- [ ] Navigation feels instant
- [ ] Search results appear quickly
- [ ] No lag when scrolling

## Known Limitations

These are expected behaviors in v1:
- No markdown rendering in preview
- No scrolling in preview for long files
- No file editing capabilities
- No task-specific features

## Bug Report Template

If issues found:
```
**Description**: [What happened]
**Steps to Reproduce**: [Exact steps]
**Expected**: [What should happen]
**Actual**: [What actually happened]
**Environment**: [OS, terminal, Go version]
```