import os
import re
import hashlib
import json
import sqlite3
import yaml

ROOT_DIR = r"c:\Users\hyper\workspace"
TARGET_DIR = r"c:\Users\hyper\workspace\borg\.tormentnexus\skills"
MAIN_SKILL_MD = r"c:\Users\hyper\workspace\SKILL.md"

BOOKMARKS_DB = r"c:\Users\hyper\workspace\bobbybookmarks\bookmarks.db"
CATALOG_DB = r"c:\Users\hyper\workspace\borg\catalog.db"

IGNORED_DIRS = {
    "node_modules", ".venv", ".git", ".ruff_cache", ".pytest_cache", 
    "site-packages", "__pycache__", ".next", ".turbo"
}

def clean_content(text):
    """Normalize whitespace and lowercase to compare content core logic."""
    # Remove markdown titles and frontmatter
    text = re.sub(r"^---[\s\S]*?---", "", text)
    text = re.sub(r"^#\s+.*", "", text, flags=re.MULTILINE)
    # Remove spacing and normalize
    return "".join(text.split()).lower()

def extract_frontmatter(file_path):
    try:
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            content = f.read()
            
        m = re.match(r"^---[\s\S]*?---", content)
        if m:
            fm_text = m.group(0).strip("-\n ")
            fm = yaml.safe_load(fm_text) or {}
            body = content[m.end():].strip()
            return fm, body, content
        else:
            # Try to parse title as name
            title_match = re.search(r"^#\s+(.*)", content, re.MULTILINE)
            name = title_match.group(1).strip() if title_match else os.path.basename(os.path.dirname(file_path))
            return {"name": name, "description": ""}, content.strip(), content
    except Exception as e:
        return {}, "", ""

def parse_main_skill_md():
    """Parse c:\\Users\\hyper\\workspace\\SKILL.md which is a consolidated list of 41 skills."""
    skills = []
    if not os.path.exists(MAIN_SKILL_MD):
        return skills
        
    try:
        with open(MAIN_SKILL_MD, "r", encoding="utf-8", errors="ignore") as f:
            content = f.read()
            
        pattern = r'(^##\s+(\d+)\.\s+([^\n]+)$)'
        matches = list(re.finditer(pattern, content, re.MULTILINE))
        
        for idx, match in enumerate(matches):
            name = match.group(3).strip()
            start_pos = match.end()
            end_pos = matches[idx+1].start() if idx + 1 < len(matches) else len(content)
            section_content = content[start_pos:end_pos].strip()
            section_content = re.sub(r'\n---+\n*$', '', section_content).strip()
            
            clean_name = name.lower().replace(" ", "_").replace("-", "_")
            skills.append({
                "name": name,
                "id": clean_name,
                "description": f"Main consolidated skill: {name}",
                "body": section_content,
                "raw": f"# {name}\n\n{section_content}",
                "source": MAIN_SKILL_MD
            })
    except Exception as e:
        print(f"Error parsing main SKILL.md: {e}")
    return skills

def normalize_link(url):
    url = url.strip().lower()
    if url.startswith("git+"):
        url = url[4:]
    if url.endswith(".git"):
        url = url[:-4]
    if url.endswith("/"):
        url = url[:-1]
    return url

def scrape_external_links():
    links = {}
    
    # 1. Scrape from bobbybookmarks/bookmarks.db
    if os.path.exists(BOOKMARKS_DB):
        try:
            print(f"Connecting to bookmarks DB: {BOOKMARKS_DB}")
            conn = sqlite3.connect(BOOKMARKS_DB)
            c = conn.cursor()
            c.execute("SELECT url, category, short_description, long_description FROM bookmarks WHERE url LIKE '%skill%' OR category LIKE '%skill%'")
            for url, category, short_desc, long_desc in c.fetchall():
                normalized = normalize_link(url)
                links[normalized] = {
                    "url": url,
                    "source": "bobbybookmarks.db/bookmarks",
                    "category": category,
                    "description": short_desc or long_desc or ""
                }
            conn.close()
        except Exception as e:
            print(f"Error reading bookmarks.db: {e}")
            
    # 2. Scrape from internal resource registry (catalog.db)
    if os.path.exists(CATALOG_DB):
        try:
            print(f"Connecting to catalog DB: {CATALOG_DB}")
            conn = sqlite3.connect(CATALOG_DB)
            c = conn.cursor()
            c.execute("SELECT repository_url, display_name, description FROM published_mcp_servers WHERE repository_url LIKE '%skill%'")
            for repo_url, name, desc in c.fetchall():
                if repo_url:
                    normalized = normalize_link(repo_url)
                    # Deduplicate or enrich existing
                    if normalized not in links:
                        links[normalized] = {
                            "url": repo_url,
                            "source": "catalog.db/published_mcp_servers",
                            "name": name,
                            "description": desc or ""
                        }
                    else:
                        links[normalized]["source"] += ", catalog.db/published_mcp_servers"
            conn.close()
        except Exception as e:
            print(f"Error reading catalog.db: {e}")
            
    return list(links.values())

def main():
    print("=== DEDUPLICATING AND CATALOGING SKILLS ===")
    
    all_skills = []
    
    # 1. Parse main SKILL.md
    print("Parsing main consolidated SKILL.md...")
    main_skills = parse_main_skill_md()
    all_skills.extend(main_skills)
    print(f"Added {len(main_skills)} skills from main SKILL.md")
    
    # 2. Walk specific directories for individual SKILL.md files
    search_dirs = [
        r"c:\Users\hyper\workspace\.agent\skills",
        r"c:\Users\hyper\workspace\.agent\plugins",
        r"c:\Users\hyper\workspace\borg\.borg\skills"
    ]
    print(f"Scanning target directories: {search_dirs}...")
    for s_dir in search_dirs:
        if not os.path.exists(s_dir):
            continue
        for root, dirs, files in os.walk(s_dir):
            dirs[:] = [d for d in dirs if d not in IGNORED_DIRS]
            for file in files:
                if file.lower() in ("skill.md", "skill.markdown"):
                    full_path = os.path.join(root, file)
                    
                    if os.path.abspath(full_path) == os.path.abspath(MAIN_SKILL_MD):
                        continue
                        
                    fm, body, raw = extract_frontmatter(full_path)
                    if not fm:
                        continue
                        
                    name = fm.get("name", "") or os.path.basename(root)
                    desc = fm.get("description", "")
                    
                    clean_id = name.lower().replace(" ", "_").replace("-", "_")
                    
                    all_skills.append({
                        "name": name,
                        "id": clean_id,
                        "description": desc,
                        "body": body,
                        "raw": raw,
                        "source": full_path
                    })

    print(f"Found total {len(all_skills)} raw skill definitions across all sources.")
    
    # 3. Deduplicate by content hash
    deduped = {}
    duplicates_log = []
    
    for s in all_skills:
        body_cleaned = clean_content(s["body"])
        if not body_cleaned:
            continue
            
        content_hash = hashlib.sha256(body_cleaned.encode('utf-8')).hexdigest()
        
        if content_hash in deduped:
            existing = deduped[content_hash]
            duplicates_log.append({
                "name": s["name"],
                "source": s["source"],
                "duplicates_with": existing["source"]
            })
            # Optionally merge/append sources to keep track of all locations
            if s["source"] not in existing["sources"]:
                existing["sources"].append(s["source"])
        else:
            s["sources"] = [s["source"]]
            s["content_hash"] = content_hash
            deduped[content_hash] = s

    print(f"\nDeduplication complete:")
    print(f"  Unique Skills: {len(deduped)}")
    print(f"  Duplicate Skills Found: {len(duplicates_log)}")
    
    # 4. Scrape external links from bookmarks & resource registry
    print("\nScraping external links for skill repositories...")
    external_links = scrape_external_links()
    print(f"Found {len(external_links)} unique skill directory/repo links.")

    # 5. Write unique skills to target directory
    print(f"\nWriting unique skills to target directory: {TARGET_DIR}")
    os.makedirs(TARGET_DIR, exist_ok=True)
    
    catalog_index = []
    
    for s in deduped.values():
        skill_id = s["id"]
        # Make a safe folder name
        safe_id = re.sub(r'[^a-zA-Z0-9_-]', '', skill_id).strip().lower()
        if not safe_id:
            safe_id = "unknown_skill_" + hashlib.md5(s["name"].encode()).hexdigest()[:6]
            
        skill_folder = os.path.join(TARGET_DIR, safe_id)
        os.makedirs(skill_folder, exist_ok=True)
        
        # Build YAML frontmatter
        fm_data = {
            "name": s["name"],
            "description": s["description"] or f"Structured runbook for {s['name']}",
            "sources": s["sources"]
        }
        
        yaml_str = yaml.dump(fm_data, sort_keys=False, default_flow_style=False).strip()
        full_markdown = f"---\n{yaml_str}\n---\n\n{s['body']}"
        
        with open(os.path.join(skill_folder, "SKILL.md"), "w", encoding="utf-8") as f:
            f.write(full_markdown)
            
        catalog_index.append({
            "id": safe_id,
            "name": s["name"],
            "description": s["description"],
            "sources": s["sources"],
            "content_hash": s["content_hash"]
        })
        
    # Write catalog report index including the external links
    report_path = os.path.join(TARGET_DIR, "catalog_index.json")
    with open(report_path, "w", encoding="utf-8") as f:
        json.dump({
            "total_raw_found": len(all_skills),
            "total_unique_local": len(deduped),
            "total_external_links": len(external_links),
            "skills": catalog_index,
            "external_links": external_links,
            "duplicates": duplicates_log
        }, f, indent=2)
        
    print(f"Catalog index written to: {report_path}")
    print("Catalog and deduplication complete!")

if __name__ == "__main__":
    main()
