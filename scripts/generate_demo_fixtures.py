#!/usr/bin/env python3
"""
Generate demo fixture files for atask screenshots.
Creates a realistic mix of projects and tasks with various states.
"""

import os
import random
from datetime import datetime, timedelta
from pathlib import Path

# Demo data
PROJECTS = [
    {
        "title": "Website Redesign",
        "tags": ["project", "work", "design"],
        "area": "work",
        "priority": "p1",
        "status": "active",
        "due_date": 15,  # days from now
        "description": "Complete overhaul of the company website with modern design"
    },
    {
        "title": "Home Renovation",
        "tags": ["project", "personal", "home"],
        "area": "personal",
        "priority": "p2",
        "status": "active",
        "due_date": 90,
        "description": "Kitchen and bathroom renovation project"
    },
    {
        "title": "Mobile App Development",
        "tags": ["project", "work", "development", "mobile"],
        "area": "work",
        "priority": "p1",
        "status": "active",
        "due_date": 60,
        "description": "Build native iOS and Android apps for our service"
    },
    {
        "title": "Learning Spanish",
        "tags": ["project", "personal", "education"],
        "area": "personal",
        "priority": "p3",
        "status": "active",
        "due_date": 180,
        "description": "Achieve conversational Spanish fluency"
    },
    {
        "title": "Q1 Marketing Campaign",
        "tags": ["project", "work", "marketing"],
        "area": "work",
        "priority": "p2",
        "status": "completed",
        "due_date": -10,  # completed 10 days ago
        "description": "Launch new product marketing campaign"
    },
    {
        "title": "API Migration",
        "tags": ["project", "work", "backend", "technical"],
        "area": "work",
        "priority": "p1",
        "status": "paused",
        "due_date": 30,
        "description": "Migrate legacy API to new microservices architecture"
    },
    {
        "title": "Fitness Goals 2025",
        "tags": ["project", "personal", "health", "fitness"],
        "area": "personal",
        "priority": "p2",
        "status": "active",
        "due_date": 365,
        "description": "Run a marathon and improve overall fitness"
    }
]

TASK_TEMPLATES = [
    # Website Redesign tasks
    {"title": "Create wireframes for homepage", "project": 0, "tags": ["design", "ux"], "priority": "p1", "estimate": 5, "due_offset": 3},
    {"title": "Design mobile responsive layouts", "project": 0, "tags": ["design", "mobile"], "priority": "p1", "estimate": 8, "due_offset": 7},
    {"title": "Review competitor websites", "project": 0, "tags": ["research"], "priority": "p2", "estimate": 3, "due_offset": 1},
    {"title": "Set up staging environment", "project": 0, "tags": ["devops"], "priority": "p2", "estimate": 3, "due_offset": 5},
    {"title": "Implement new navigation menu", "project": 0, "tags": ["frontend"], "priority": "p1", "estimate": 5, "due_offset": 10},
    {"title": "Write content for About page", "project": 0, "tags": ["content"], "priority": "p3", "estimate": 2, "due_offset": 8},
    
    # Home Renovation tasks
    {"title": "Get quotes from contractors", "project": 1, "tags": ["planning"], "priority": "p1", "estimate": 3, "due_offset": 7},
    {"title": "Choose kitchen cabinet style", "project": 1, "tags": ["decisions"], "priority": "p2", "estimate": 2, "due_offset": 14},
    {"title": "Order appliances", "project": 1, "tags": ["shopping"], "priority": "p1", "estimate": 3, "due_offset": 30},
    {"title": "Schedule building permits", "project": 1, "tags": ["admin"], "priority": "p1", "estimate": 2, "due_offset": 5},
    {"title": "Pack kitchen items", "project": 1, "tags": ["preparation"], "priority": "p2", "estimate": 5, "due_offset": 20},
    
    # Mobile App tasks
    {"title": "Set up React Native environment", "project": 2, "tags": ["setup", "development"], "priority": "p1", "estimate": 3, "due_offset": 2},
    {"title": "Design app icon and splash screen", "project": 2, "tags": ["design"], "priority": "p2", "estimate": 3, "due_offset": 10},
    {"title": "Implement user authentication", "project": 2, "tags": ["backend", "security"], "priority": "p1", "estimate": 8, "due_offset": 15},
    {"title": "Create onboarding flow", "project": 2, "tags": ["ux", "frontend"], "priority": "p2", "estimate": 5, "due_offset": 20},
    {"title": "Set up push notifications", "project": 2, "tags": ["backend", "mobile"], "priority": "p3", "estimate": 5, "due_offset": 25},
    {"title": "Beta testing with users", "project": 2, "tags": ["testing", "qa"], "priority": "p1", "estimate": 8, "due_offset": 45},
    
    # Learning Spanish tasks
    {"title": "Complete Duolingo daily practice", "project": 3, "tags": ["daily", "practice"], "priority": "p2", "estimate": 1, "due_offset": 0},
    {"title": "Watch Spanish movie with subtitles", "project": 3, "tags": ["immersion"], "priority": "p3", "estimate": 2, "due_offset": 3},
    {"title": "Schedule conversation with tutor", "project": 3, "tags": ["speaking"], "priority": "p1", "estimate": 1, "due_offset": 7},
    {"title": "Review irregular verbs", "project": 3, "tags": ["grammar"], "priority": "p2", "estimate": 2, "due_offset": 2},
    
    # Marketing Campaign tasks (completed project)
    {"title": "Design campaign graphics", "project": 4, "tags": ["design", "marketing"], "priority": "p1", "estimate": 5, "status": "done"},
    {"title": "Write press release", "project": 4, "tags": ["content", "pr"], "priority": "p1", "estimate": 3, "status": "done"},
    {"title": "Set up Google Ads", "project": 4, "tags": ["advertising", "digital"], "priority": "p2", "estimate": 3, "status": "done"},
    {"title": "Create social media content", "project": 4, "tags": ["social", "content"], "priority": "p2", "estimate": 5, "status": "done"},
    
    # API Migration tasks (paused project)
    {"title": "Document current API endpoints", "project": 5, "tags": ["documentation"], "priority": "p1", "estimate": 5, "status": "paused"},
    {"title": "Design microservice architecture", "project": 5, "tags": ["architecture", "planning"], "priority": "p1", "estimate": 8, "status": "paused"},
    {"title": "Set up Docker containers", "project": 5, "tags": ["devops", "infrastructure"], "priority": "p2", "estimate": 5, "status": "paused"},
    
    # Fitness Goals tasks
    {"title": "Join running club", "project": 6, "tags": ["social", "running"], "priority": "p2", "estimate": 1, "due_offset": 7},
    {"title": "Buy new running shoes", "project": 6, "tags": ["equipment"], "priority": "p1", "estimate": 2, "due_offset": 3},
    {"title": "Schedule health check-up", "project": 6, "tags": ["health"], "priority": "p1", "estimate": 1, "due_offset": 14},
    {"title": "Plan weekly meal prep", "project": 6, "tags": ["nutrition", "planning"], "priority": "p2", "estimate": 3, "due_offset": 2},
    {"title": "Research marathon training plans", "project": 6, "tags": ["research", "running"], "priority": "p2", "estimate": 2, "due_offset": 10},
    
    # Standalone tasks (no project)
    {"title": "Renew driver's license", "tags": ["admin", "personal"], "priority": "p1", "estimate": 2, "due_offset": 30, "area": "personal"},
    {"title": "File quarterly taxes", "tags": ["finance", "business"], "priority": "p1", "estimate": 3, "due_offset": 20, "area": "work"},
    {"title": "Buy birthday gift for Mom", "tags": ["family", "shopping"], "priority": "p2", "estimate": 2, "due_offset": 5, "area": "personal"},
    {"title": "Update resume", "tags": ["career"], "priority": "p3", "estimate": 3, "due_offset": 60, "area": "personal"},
    {"title": "Fix bug in login system", "tags": ["bugfix", "urgent"], "priority": "p1", "estimate": 3, "due_offset": 1, "area": "work"},
    {"title": "Backup important documents", "tags": ["admin", "important"], "priority": "p2", "estimate": 2, "due_offset": 7, "area": "personal"},
    {"title": "Schedule dentist appointment", "tags": ["health"], "priority": "p2", "estimate": 1, "due_offset": 14, "area": "personal"},
    {"title": "Review team performance", "tags": ["management"], "priority": "p2", "estimate": 5, "due_offset": 10, "area": "work"},
    {"title": "Plan vacation itinerary", "tags": ["travel", "planning"], "priority": "p3", "estimate": 3, "due_offset": 45, "area": "personal"},
    {"title": "Update software licenses", "tags": ["admin", "software"], "priority": "p2", "estimate": 2, "due_offset": 15, "area": "work"},
    {"title": "Organize garage", "tags": ["home", "organizing"], "priority": "p3", "estimate": 8, "due_offset": 30, "area": "personal"},
    {"title": "Prepare presentation for board", "tags": ["presentation", "important"], "priority": "p1", "estimate": 5, "due_offset": 3, "area": "work"},
    {"title": "Call insurance company", "tags": ["admin", "urgent"], "priority": "p1", "estimate": 1, "due_offset": 2, "area": "personal"},
]

def generate_denote_id(offset_minutes=0):
    """Generate a Denote ID with optional time offset."""
    dt = datetime.now() - timedelta(minutes=offset_minutes)
    return dt.strftime("%Y%m%dT%H%M%S")

def title_to_slug(title):
    """Convert title to Denote slug format."""
    slug = title.lower()
    # Keep only alphanumeric and convert spaces to hyphens
    slug = ''.join(c if c.isalnum() or c == ' ' else '' for c in slug)
    slug = slug.replace(' ', '-')
    # Remove multiple hyphens
    while '--' in slug:
        slug = slug.replace('--', '-')
    return slug.strip('-')

def format_date(days_offset):
    """Format date for frontmatter."""
    date = datetime.now() + timedelta(days=days_offset)
    return date.strftime("%Y-%m-%d")

def create_project_file(project, index, base_dir):
    """Create a project file."""
    # Generate ID (older projects have older IDs)
    id_offset = (len(PROJECTS) - index) * 24 * 60  # Days in minutes
    denote_id = generate_denote_id(id_offset)
    
    # Create filename
    slug = title_to_slug(project["title"])
    # Ensure 'project' tag is in the filename
    file_tags = project["tags"] if "project" in project["tags"] else ["project"] + project["tags"]
    tags_part = "_".join(file_tags)
    filename = f"{denote_id}--{slug}__{tags_part}.md"
    
    # Create frontmatter
    due_date = format_date(project["due_date"]) if project.get("due_date") else ""
    
    content = f"""---
title: {project["title"]}
index_id: {100 + index}
type: project
status: {project["status"]}
priority: {project.get("priority", "")}
area: {project.get("area", "")}
due_date: "{due_date}"
tags:
{chr(10).join(f"  - {tag}" for tag in project["tags"])}
---

## Overview

{project["description"]}

## Objectives

- Define clear goals and milestones
- Establish timeline and deadlines
- Assign resources and responsibilities
- Track progress and deliverables

## Progress Log

[{format_date(-5)}]: Project kickoff meeting held
[{format_date(-3)}]: Initial planning phase completed
[{format_date(-1)}]: First milestone review
"""
    
    filepath = os.path.join(base_dir, filename)
    with open(filepath, 'w') as f:
        f.write(content)
    
    return denote_id, filename  # Return denote_id instead of index

def create_task_file(task_data, task_index, base_dir, project_map):
    """Create a task file."""
    # Generate ID (spread throughout the day)
    id_offset = task_index * 30  # 30 minutes apart
    denote_id = generate_denote_id(id_offset)
    
    # Build tags list - 'task' must be included for both filename and metadata
    user_tags = task_data.get("tags", [])
    file_tags = ["task"] + user_tags
    
    # Create filename with double underscore before tags
    slug = title_to_slug(task_data["title"])
    tags_part = "_".join(file_tags)
    filename = f"{denote_id}--{slug}__{tags_part}.md"
    
    # Determine status
    status = task_data.get("status", "open")
    
    # Calculate due date
    due_date = ""
    if "due_offset" in task_data and status != "done":
        due_date = format_date(task_data["due_offset"])
    elif status == "done":
        due_date = format_date(-random.randint(1, 20))  # Completed in the past
    
    # Get project ID if assigned
    project_id = ""
    if "project" in task_data and task_data["project"] in project_map:
        project_id = project_map[task_data["project"]]
    
    # Random area if not specified
    area = task_data.get("area", random.choice(["work", "personal"]))
    
    content = f"""---
title: {task_data["title"]}
index_id: {1000 + task_index}
type: task
status: {status}
priority: {task_data.get("priority", "")}
area: {area}
project_id: {project_id}
due_date: "{due_date}"
estimate: {task_data.get("estimate", 0)}
tags:
{chr(10).join(f"  - {tag}" for tag in file_tags)}
---

## Task Description

{task_data["title"]} - This task is part of the demo data set.

## Notes

- Created for demonstration purposes
- Shows various task states and metadata
"""
    
    # Add completion log for done tasks
    if status == "done":
        content += f"\n## Log\n\n[{format_date(-random.randint(1, 10))}]: Task completed successfully\n"
    
    filepath = os.path.join(base_dir, filename)
    with open(filepath, 'w') as f:
        f.write(content)
    
    return filename

def main():
    """Generate demo fixtures."""
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python generate_demo_fixtures.py <output-directory>")
        sys.exit(1)
    
    output_dir = sys.argv[1]
    
    # Create output directory
    Path(output_dir).mkdir(parents=True, exist_ok=True)
    
    print(f"Generating demo fixtures in {output_dir}")
    
    # Create projects
    project_map = {}
    for i, project in enumerate(PROJECTS):
        denote_id, filename = create_project_file(project, i, output_dir)
        project_map[i] = denote_id  # Map index to denote ID
        print(f"  Created project: {filename}")
    
    # Create tasks
    task_count = 0
    for task_data in TASK_TEMPLATES:
        filename = create_task_file(task_data, task_count, output_dir, project_map)
        print(f"  Created task: {filename}")
        task_count += 1
    
    # Add some random task variations
    statuses = ["open", "done", "paused", "delegated", "dropped"]
    for i in range(7):  # Add 7 more random tasks
        task = random.choice(TASK_TEMPLATES)
        task_copy = task.copy()
        task_copy["status"] = random.choice(statuses)
        if task_copy["status"] in ["done", "dropped"]:
            task_copy.pop("due_offset", None)
        
        filename = create_task_file(task_copy, task_count, output_dir, project_map)
        print(f"  Created task: {filename}")
        task_count += 1
    
    print(f"\nGenerated {len(PROJECTS)} projects and {task_count} tasks")
    print(f"Total: {len(PROJECTS) + task_count} files")

if __name__ == "__main__":
    main()