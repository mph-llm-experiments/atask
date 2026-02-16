# Denote Tasks Demo

This directory contains demo fixtures for taking screenshots and demonstrating atask functionality.

## Quick Start

1. Generate demo fixtures (if not already present):
   ```bash
   python3 scripts/generate_demo_fixtures.py demo-notes
   ```

2. Run atask with the demo configuration:
   ```bash
   ./atask --config demo-config.toml --tui
   ```

## Demo Data Overview

The demo includes:

### 7 Projects
- **Website Redesign** (p1, work) - Active project with design focus
- **Home Renovation** (p2, personal) - Personal home improvement project  
- **Mobile App Development** (p1, work) - Technical development project
- **Learning Spanish** (p3, personal) - Educational/self-improvement
- **Q1 Marketing Campaign** (completed) - Example of finished project
- **API Migration** (paused) - Technical project on hold
- **Fitness Goals 2025** (p2, personal) - Health and fitness tracking

### 50+ Tasks
- Mix of priorities (p1, p2, p3)
- Various states (open, done, paused, delegated, dropped)
- Different areas (work, personal)
- Assigned to projects and standalone
- Range of due dates (overdue, due soon, future)
- Time estimates using Fibonacci sequence
- Rich tagging (#design, #backend, #urgent, etc.)

## Screenshot Tips

1. **Task List View**: Shows priority sorting, due dates, and status symbols
2. **Project Filter**: Press `p` to show only projects
3. **Area Filter**: Press `a` then type "work" or "personal"
4. **Search**: Press `/` and search for tags like "#design" or "#urgent"
5. **Task Details**: Press Enter on any task to see full metadata
6. **Project View**: Enter a project to see assigned tasks
7. **State Filter**: Press `f` to filter by task state

## Regenerating Demo Data

To get fresh demo data with current dates:
```bash
rm -rf demo-notes
python3 scripts/generate_demo_fixtures.py demo-notes
```

The script creates realistic-looking tasks and projects with:
- Proper Denote ID timestamps
- Meaningful titles and descriptions  
- Realistic tag combinations
- Varied priorities and due dates
- Different task states
- Project associations