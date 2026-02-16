#!/usr/bin/env python3

import os
import glob
import re
import json

print("Fixing test data to match Denote Task Spec v3.0.0...")

# Counter for the unified index_id
current_index = 1

# Track project Denote IDs for task associations
project_denote_ids = {}

def parse_frontmatter(content):
    """Simple frontmatter parser"""
    if not content.startswith('---\n'):
        return None, content
    
    # Find the closing ---
    end_match = re.search(r'\n---\n', content[4:])
    if not end_match:
        return None, content
    
    fm_text = content[4:4+end_match.start()]
    body = content[4+end_match.end():]
    
    # Parse the frontmatter
    fm = {}
    for line in fm_text.split('\n'):
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        
        match = re.match(r'^(\w+):\s*(.*)$', line)
        if match:
            key = match.group(1)
            value = match.group(2).strip()
            
            # Remove quotes if present
            if value.startswith('"') and value.endswith('"'):
                value = value[1:-1]
            elif value.startswith("'") and value.endswith("'"):
                value = value[1:-1]
            
            # Try to parse as int
            try:
                value = int(value)
            except ValueError:
                pass
            
            fm[key] = value
    
    return fm, body

def write_frontmatter(fm, body):
    """Write frontmatter and body back"""
    content = "---\n"
    
    # Write fields in a sensible order
    field_order = ['title', 'index_id', 'type', 'status', 'priority', 
                  'due_date', 'start_date', 'estimate', 'project_id', 
                  'area', 'assignee', 'tags']
    
    for field in field_order:
        if field in fm:
            value = fm[field]
            if isinstance(value, str) and (' ' in value or ':' in value):
                content += f'{field}: "{value}"\n'
            else:
                content += f"{field}: {value}\n"
    
    # Add any remaining fields
    for field, value in fm.items():
        if field not in field_order:
            if isinstance(value, str) and (' ' in value or ':' in value):
                content += f'{field}: "{value}"\n'
            else:
                content += f"{field}: {value}\n"
    
    content += "---\n"
    content += body
    return content

def fix_file(filepath):
    """Fix a single file to match spec v3.0.0"""
    global current_index
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    fm, body = parse_frontmatter(content)
    if fm is None:
        print(f"  WARNING: Could not parse frontmatter in {filepath}")
        return False
    
    # Get filename info
    filename = os.path.basename(filepath)
    denote_id = filename.split('--')[0]
    is_task = '__task' in filename
    is_project = '__project' in filename
    
    # Remove incorrect fields
    if 'id' in fm:
        del fm['id']
    if 'identifier' in fm:
        del fm['identifier']
    if 'date' in fm:
        del fm['date']
    
    # Fix the index_id field
    if is_task:
        # Convert task_id to index_id
        if 'task_id' in fm:
            del fm['task_id']
        fm['index_id'] = current_index
        current_index += 1
    elif is_project:
        # Convert project_id to index_id
        if 'project_id' in fm and isinstance(fm['project_id'], int):
            del fm['project_id']
        fm['index_id'] = current_index
        current_index += 1
        
        # Track project Denote ID for later
        project_name = filename.split('--')[1].split('__')[0]
        project_denote_ids[project_name] = denote_id
        
        # Also track by title if available
        if 'title' in fm:
            title_key = fm['title'].lower().replace(' ', '-').replace('_', '-')
            project_denote_ids[title_key] = denote_id
    
    # Ensure required fields
    if 'title' not in fm and (is_task or is_project):
        # Generate title from filename
        title_slug = filename.split('--')[1].split('__')[0]
        title = title_slug.replace('-', ' ').title()
        fm['title'] = title
    
    # Write the file
    new_content = write_frontmatter(fm, body)
    with open(filepath, 'w') as f:
        f.write(new_content)
    
    return True

# First pass: Fix all files and collect project info
print("\nFirst pass: Fixing files and collecting project info...")
fixed_count = 0
for filepath in sorted(glob.glob("test-notes/*.md")):
    filename = os.path.basename(filepath)
    if '__task' in filename or '__project' in filename:
        if fix_file(filepath):
            fixed_count += 1
            print(f"  Fixed {filename}")

print(f"\nFixed {fixed_count} files")

# Create mapping of likely project associations
project_associations = {
    'opsgenie': '20250114T101200',
    'contacts': '20250114T100300', 
    'clockwise': '20250114T100800',
    'otter': '20250114T101400',
    'securityreviews': '20250114T101700',
    'security-reviews': '20250114T101700',
    'roadmaprethink': '20250114T101100',
    'roadmap-rethink': '20250114T101100',
    'oncall': '20250114T100900',
    'on-call': '20250114T100900',
    'auditoverhaul': '20250114T100600',
    'audit-overhaul': '20250114T100600',
    'lyon': '20250114T101000',
    'lyon-planning': '20250114T101000',
    'postfest': '20250114T101500',
    'ai': '20250114T100200',
    'crowdstrike': '20250114T100500',
    'anz': '20250114T100400',
    'roadmap': '20250114T101600',
    'teammtg': '20250114T101800',
    'team-mtg': '20250114T101800',
    'it': '20250114T100700',
    'apex': '20250114T100100',
    'sunsetting': '20250114T101300',
    'atask': '20250113T152000',
    'project-atask': '20250113T152000'
}

# Second pass: Fix project_id references in tasks
print("\nSecond pass: Fixing project_id references in tasks...")
for filepath in glob.glob("test-notes/*__task*.md"):
    with open(filepath, 'r') as f:
        content = f.read()
    
    fm, body = parse_frontmatter(content)
    if fm and 'project_id' in fm:
        old_id = fm['project_id']
        filename = os.path.basename(filepath).lower()
        content_lower = content.lower()
        
        # Try to determine correct project
        new_id = None
        
        # Check various patterns
        for key, denote_id in project_associations.items():
            if key in filename or key in content_lower:
                new_id = denote_id
                break
        
        # Special cases
        if not new_id:
            if 'ping' in filename or 'write-to' in filename:
                new_id = project_associations.get('contacts')
            elif 'pbp' in content_lower or 'deck' in content_lower:
                new_id = project_associations.get('oncall')
            elif 'maturity' in content_lower or 'challenges' in content_lower:
                new_id = project_associations.get('roadmaprethink')
        
        if new_id and new_id != old_id:
            print(f"  Updating {os.path.basename(filepath)}: {old_id} -> {new_id}")
            fm['project_id'] = new_id
            new_content = write_frontmatter(fm, body)
            with open(filepath, 'w') as f:
                f.write(new_content)

# Create/update the counter file
counter_data = {
    "next_index": current_index,
    "spec_version": "3.0.0"
}

counter_path = "test-notes/.denote-task-counter.json"
with open(counter_path, 'w') as f:
    json.dump(counter_data, f, indent=2)

print(f"\nCreated counter file with next_index: {current_index}")
print("\nDone! Test data should now conform to spec v3.0.0")