# Project View Hotkey Debug Guide

## Test Scenario

1. Start the TUI: `./atask --config test-config.toml --tui`
2. Press `t` to switch to Task Mode
3. Press `p` to show projects only
4. Navigate to a project and press Enter to open it

## Expected Behavior

When you open a project, you should see:
- Two tabs at the top: "Overview" and "Tasks"
- The "Overview" tab should be highlighted (active) by default
- You should be on tab 0 (overview)

## Test the Hotkeys

While on the Overview tab (tab 0):
1. Press `p` - Should show "Enter priority (1/2/3):" at the bottom
2. Press `esc` to cancel
3. Press `s` - Should show "Enter status (active/completed/paused/cancelled):" at the bottom
4. Press `esc` to cancel
5. Press `d` - Should show "Enter due date..." at the bottom
6. Press `esc` to cancel
7. Press `a` - Should show "Enter area:" at the bottom
8. Press `esc` to cancel
9. Press `g` - Should show "Enter tags (space-separated):" at the bottom
10. Press `esc` to cancel

## Tab Switching Test

1. Press `Tab` key - Should switch to "Tasks" tab
2. Now try pressing `p`, `d`, `a`, `g` - These should NOT trigger editing
3. Press `s` - Should open state menu (different behavior than on Overview tab)
4. Press `Tab` again - Should switch back to "Overview" tab
5. Try the hotkeys again - They should work now

## Debugging Questions

1. Are you on the Overview tab when trying the hotkeys?
2. Does the Tab key successfully switch between tabs?
3. Do the hotkeys work when you first enter project view?
4. Do they stop working after switching tabs?
5. What exactly happens when you press a hotkey that doesn't work?

## Visual Indicators

The active tab should be highlighted in orange/yellow color (color 214).
The inactive tab should be gray (color 241).

## Alternative Test

If hotkeys don't work at all:
1. Try editing a task (not a project) to confirm task hotkeys work
2. Check if there are any error messages at the bottom of the screen
3. Try pressing `?` for help to see if key handling works at all