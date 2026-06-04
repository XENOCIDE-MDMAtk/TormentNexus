import os
import sqlite3
import json

BOOKMARKS_DB = r"c:\Users\hyper\workspace\bobbybookmarks\bookmarks.db"
CATALOG_DB = r"c:\Users\hyper\workspace\borg\catalog.db"

def inspect_db(db_path, query, params=()):
    if not os.path.exists(db_path):
        print(f"Database not found: {db_path}")
        return []
    try:
        conn = sqlite3.connect(db_path)
        c = conn.cursor()
        c.execute(query, params)
        rows = c.fetchall()
        colnames = [desc[0] for desc in c.description]
        results = [dict(zip(colnames, row)) for row in rows]
        conn.close()
        return results
    except Exception as e:
        print(f"Error reading {db_path}: {e}")
        return []

def main():
    print("--- INSPECTING BOOKMARKS.DB ---")
    if os.path.exists(BOOKMARKS_DB):
        conn = sqlite3.connect(BOOKMARKS_DB)
        c = conn.cursor()
        c.execute("SELECT name FROM sqlite_master WHERE type='table'")
        print("Tables in bookmarks.db:", c.fetchall())
        c.execute("PRAGMA table_info(bookmarks)")
        print("Columns in bookmarks table:")
        for col in c.fetchall():
            print(f"  {col[1]} ({col[2]})")
        conn.close()
    
    # Query all bookmark entries to find potential skill directories or lists
    # We will fetch everything to analyze matches manually and print them.
    bookmarks = inspect_db(BOOKMARKS_DB, "SELECT * FROM bookmarks")
    print(f"\nTotal bookmarks found: {len(bookmarks)}")
    
    print("\n--- INSPECTING CATALOG.DB ---")
    if os.path.exists(CATALOG_DB):
        conn = sqlite3.connect(CATALOG_DB)
        c = conn.cursor()
        c.execute("SELECT name FROM sqlite_master WHERE type='table'")
        print("Tables in catalog.db:", c.fetchall())
        c.execute("PRAGMA table_info(published_mcp_servers)")
        print("Columns in published_mcp_servers table:")
        for col in c.fetchall():
            print(f"  {col[1]} ({col[2]})")
        conn.close()
        
    mcp_servers = inspect_db(CATALOG_DB, "SELECT * FROM published_mcp_servers")
    print(f"\nTotal published MCP servers found: {len(mcp_servers)}")

    # Look for anything with 'skill', 'mcp', 'plugin', 'tool', 'registry', 'directory'
    print("\n--- SKILL DIRECTORIES / LINKS REPORT ---")
    
    skill_keywords = ["skill", "plugin", "tool", "directory", "awesome-mcp", "registry", "mcp-server", "catalog"]
    
    matched_bookmarks = []
    for b in bookmarks:
        text_to_search = f"{b.get('url', '')} {b.get('category', '')} {b.get('short_description', '')} {b.get('long_description', '')} {b.get('title', '')}".lower()
        if any(kw in text_to_search for kw in skill_keywords):
            matched_bookmarks.append(b)
            
    matched_servers = []
    for s in mcp_servers:
        # Check repo urls, display name, etc.
        text_to_search = f"{s.get('repository_url', '')} {s.get('display_name', '')} {s.get('description', '')}".lower()
        matched_servers.append(s) # all MCP servers are technically tools/plugins/skill-like resources
        
    print(f"\nMatches in bookmarks ({len(matched_bookmarks)}):")
    for b in matched_bookmarks[:50]:
        print(f"- URL: {b.get('url')}\n  Title: {b.get('title')}\n  Category: {b.get('category')}\n  Desc: {b.get('short_description')}")
        
    print(f"\nMatches in catalog ({len(matched_servers)}):")
    for s in matched_servers[:50]:
        print(f"- URL: {s.get('repository_url')}\n  Name: {s.get('display_name')}\n  Desc: {s.get('description')}")

    # Write them out cleanly to a JSON file
    output_data = {
        "bookmarks_skill_links": matched_bookmarks,
        "catalog_mcp_links": matched_servers
    }
    with open("scratch/extracted_links.json", "w") as f:
        json.dump(output_data, f, indent=2)
    print(f"\nAll links written to scratch/extracted_links.json")

if __name__ == "__main__":
    main()
