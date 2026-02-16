# DEPRECATED: Note Creation Test Plan

**Note: This feature has been removed. atask is now focused exclusively on task management.**

---

# Note Creation Test Plan (Legacy)

## Feature: Create New Note (Issue #2)

### Setup
```bash
./atask --config test-config.toml --tui
```

### Test Steps

#### 1. Basic Note Creation
1. **Press `n`** from the main list
2. **Verify**: "Create New Note" dialog appears with "Title:" prompt
3. **Type**: "Test Note Creation"
4. **Press Enter**
5. **Verify**: Dialog now shows tags prompt
6. **Type**: "test demo"  (space-separated tags)
7. **Press Enter**
8. **Expected**:
   - Note is created with proper Denote filename
   - Editor opens with the new file (if configured)
   - Returns to file list
   - New note appears in the list

#### 2. Cancel During Title
1. **Press `n`**
2. **Type**: "Some title"
3. **Press Esc**
4. **Expected**: Returns to file list without creating note

#### 3. Cancel During Tags
1. **Press `n`**
2. **Type title and press Enter**
3. **At tags prompt, press Esc**
4. **Expected**: Goes back to title prompt

#### 4. Note with No Tags
1. **Press `n`**
2. **Enter title**
3. **Press Enter** at tags prompt without typing
4. **Expected**: Creates note with no tags in filename

#### 5. Special Characters in Title
1. **Press `n`**
2. **Type**: "Test & Demo! With #Special"
3. **Continue creation**
4. **Expected**: Filename slug should be sanitized (test-demo-with-special)

### Verify Created Files

After creating notes, check the test-notes directory:
- Filenames follow pattern: `YYYYMMDDTHHMMSS--slug__tags.md`
- Files contain required YAML frontmatter:
  ```yaml
  ---
  title: Test Note Creation
  type: note
  created: 2025-01-13
  ---
  ```

### Edge Cases to Test
- Very long title (should work)
- Empty title (should not proceed to tags)
- Many tags (should all be included)
- Duplicate tag names (should be preserved)

### Integration Check
- After creating a note, it should immediately appear in the list
- Search should find the new note by title/tags
- Sort should include the new note properly