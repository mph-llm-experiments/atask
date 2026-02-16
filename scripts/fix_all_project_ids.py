#!/usr/bin/env python3

import os
import glob
import re
from datetime import datetime, timedelta

print("Giving all projects unique IDs...")

# Get all project files
project_files = sorted(glob.glob("test-notes/*__project*.md"))

# Start from a base timestamp and increment by 1 minute for each project
base_time = datetime(2025, 1, 14, 10, 0, 0)  # Today at 10:00 AM

for i, project_file in enumerate(project_files):
    old_filename = os.path.basename(project_file)
    old_timestamp = old_filename.split('--')[0]
    
    # Skip if it's already unique (like atask project)
    if old_timestamp == "20250113T152000":
        continue
    
    # Generate new timestamp - increment by 1 minute for each project
    new_dt = base_time + timedelta(minutes=i)
    new_timestamp = new_dt.strftime("%Y%m%dT%H%M%S")
    
    # Build new filename
    new_filename = old_filename.replace(old_timestamp, new_timestamp)
    new_path = os.path.join(os.path.dirname(project_file), new_filename)
    
    print(f"Renaming: {old_filename}")
    print(f"      to: {new_filename}")
    
    # Rename the file
    os.rename(project_file, new_path)
    
    # Update the ID inside the file
    with open(new_path, 'r') as f:
        content = f.read()
    
    # Update the id field in frontmatter
    content = re.sub(
        r'^id:\s*"?' + re.escape(old_timestamp) + r'"?\s*$',
        f'id: "{new_timestamp}"',
        content,
        flags=re.MULTILINE
    )
    
    with open(new_path, 'w') as f:
        f.write(content)

print("\nAll projects now have unique IDs!")