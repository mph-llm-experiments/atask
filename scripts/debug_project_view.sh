#!/bin/bash

echo "Project View Debug Script"
echo "========================"
echo

# Check if atask binary exists
if [ ! -f "./atask" ]; then
    echo "Building atask..."
    go build ./cmd/atask
fi

# Create test project
echo "Creating test project..."
cat > test-notes/20250113T150000--debug-project__project.md << 'EOF'
---
title: Debug Project
type: project
status: active
priority: p2
area: testing
due_date: 2025-01-30
tags: [debug, test]
---

# Debug Project

This project is for debugging the hotkey issue.

## Overview

Testing project view hotkeys.
EOF

echo "Test project created."
echo
echo "To debug:"
echo "1. Run: ./atask --config test-config.toml --tui"
echo "2. Press 't' for Task Mode"
echo "3. Press 'p' for projects only"
echo "4. Open 'Debug Project'"
echo "5. Try hotkeys: p, s, d, a, g"
echo
echo "Expected: Each hotkey should show an edit prompt at the bottom"
echo "If not working, note:"
echo "- Which tab is active (Overview should be highlighted)"
echo "- Any error messages"
echo "- Whether Tab key switches between tabs"