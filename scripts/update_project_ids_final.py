#!/usr/bin/env python3

import os
import re
import glob

print("Updating all project_id fields with new unique IDs...")

# Map old IDs to new IDs based on project names
old_to_new = {
    "20250704T124524": "20250114T100100",  # apex
    "20250704T124525": None,  # Multiple projects had this - need to determine by content
    "20250704T124526": None,  # Multiple projects had this
    "20250704T124527": None,  # Multiple projects had this
    "20250704T124528": None,  # Multiple projects had this
    "20250704T124529": None,  # Multiple projects had this
    "20250704T124530": None,  # Multiple projects had this
    "20250704T124531": "20250114T101400",  # otter
    "20250704T124532": "20250114T101500",  # postfest
    "20250704T124533": "20250114T101600",  # roadmap
    "20250704T124534": "20250114T101700",  # securityreviews
    "20250704T124535": "20250114T101800",  # teammtg
}

# Build project name to new ID mapping
project_name_to_id = {
    "apex": "20250114T100100",
    "ai": "20250114T100200",
    "contacts": "20250114T100300",
    "anz": "20250114T100400",
    "crowdstrike": "20250114T100500",
    "auditoverhaul": "20250114T100600",
    "it": "20250114T100700",
    "clockwise": "20250114T100800",
    "oncall": "20250114T100900",
    "lyon-planning": "20250114T101000",
    "lyon_planning": "20250114T101000",
    "roadmaprethink": "20250114T101100",
    "opsgenie": "20250114T101200",
    "sunsetting": "20250114T101300",
    "otter": "20250114T101400",
    "postfest": "20250114T101500",
    "roadmap": "20250114T101600",
    "securityreviews": "20250114T101700",
    "teammtg": "20250114T101800",
}

updated_count = 0

for task_file in glob.glob("test-notes/*__task*.md"):
    with open(task_file, 'r') as f:
        content = f.read()
    
    # Check if it has project_id
    id_match = re.search(r'^project_id:\s*([0-9T]+)\s*$', content, re.MULTILINE)
    if id_match:
        old_id = id_match.group(1)
        new_id = None
        
        # Skip if it's already the atask project
        if old_id == "20250113T152000":
            continue
        
        # Try direct mapping first
        if old_id in old_to_new and old_to_new[old_id]:
            new_id = old_to_new[old_id]
        else:
            # Determine by content/filename
            filename = os.path.basename(task_file).lower()
            content_lower = content.lower()
            
            if "opsgenie" in filename or "opsgenie" in content_lower:
                new_id = project_name_to_id["opsgenie"]
            elif "contact" in filename or "ping" in filename or "write-to" in filename:
                new_id = project_name_to_id["contacts"]
            elif "clockwise" in filename or "clockwise" in content_lower:
                new_id = project_name_to_id["clockwise"]
            elif "otter" in filename or "otter" in content_lower:
                new_id = project_name_to_id["otter"]
            elif "security" in filename and "review" in filename:
                new_id = project_name_to_id["securityreviews"]
            elif "roadmap" in filename and "rethink" in content_lower:
                new_id = project_name_to_id["roadmaprethink"]
            elif "oncall" in filename or "on-call" in content_lower or "tabletop" in content_lower:
                new_id = project_name_to_id["oncall"]
            elif "audit" in filename:
                new_id = project_name_to_id["auditoverhaul"]
            elif "lyon" in filename or "lyon" in content_lower:
                new_id = project_name_to_id["lyon-planning"]
            elif "postfest" in filename or ("post" in filename and "fest" in content_lower):
                new_id = project_name_to_id["postfest"]
            elif "ai" in filename and ("zoom" in content_lower or "slack" in content_lower):
                new_id = project_name_to_id["ai"]
            elif "crowdstrike" in filename:
                new_id = project_name_to_id["crowdstrike"]
            elif "anz" in filename or "anz" in content_lower:
                new_id = project_name_to_id["anz"]
            elif "roadmap" in filename and "os" in content_lower:
                new_id = project_name_to_id["roadmap"]
            elif "teammtg" in content_lower or ("monthly" in content_lower and "sync" in content_lower):
                new_id = project_name_to_id["teammtg"]
            elif "topical" in filename and "discussion" in filename:
                new_id = project_name_to_id["teammtg"]
            elif "maturity" in content_lower or "challenges" in content_lower or "mission" in content_lower:
                new_id = project_name_to_id["roadmaprethink"]
            elif "decision" in content_lower and "tree" in content_lower:
                new_id = project_name_to_id["oncall"]
            elif "pbp" in content_lower and "deck" in content_lower:
                new_id = project_name_to_id["oncall"]
        
        if new_id and new_id != old_id:
            print(f"  {os.path.basename(task_file)}: {old_id} -> {new_id}")
            content = re.sub(
                r'^project_id:\s*' + re.escape(old_id) + r'\s*$',
                f'project_id: {new_id}',
                content,
                flags=re.MULTILINE
            )
            
            with open(task_file, 'w') as f:
                f.write(content)
            
            updated_count += 1

print(f"\nUpdated {updated_count} task files with correct project IDs.")