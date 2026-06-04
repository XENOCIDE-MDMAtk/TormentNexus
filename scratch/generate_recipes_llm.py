import sqlite3
import urllib.request
import json
import uuid
import re
import argparse
import time
import sys

sys.stdout.reconfigure(encoding='utf-8')

PROXY_URL = "http://localhost:4000/v1/chat/completions"
LMSTUDIO_URL = "http://localhost:1234/v1/chat/completions"
DB_PATH = r"c:\Users\hyper\workspace\borg\catalog.db"

SYSTEM_PROMPT = """You are an expert developer building a database of Model Context Protocol (MCP) server launch commands.
Your task is to read the provided README.md of a GitHub repository and extract the EXACT commands required to launch the MCP server via stdio.
Many MCP servers are Node.js (requiring `npx` or `node build/index.js`), Python (requiring `uvx` or `python src/server.py`), or Go (requiring `go run`).

Output ONLY a JSON object with the following schema, and NO markdown wrapping or other text:
{
  "type": "stdio",
  "command": "the executable to run (e.g., 'npx', 'node', 'uvx', 'python', 'go')",
  "args": ["array", "of", "arguments"],
  "env": {"OPTIONAL_ENV_VAR": "value"}
}

If you cannot determine the launch command, output a best guess or fallback to a standard `npx -y @owner/repo`. Do NOT output any markdown, only raw JSON.
"""

def fetch_readme(owner, repo):
    # Try main first, then master
    branches = ['main', 'master']
    for branch in branches:
        url = f"https://raw.githubusercontent.com/{owner}/{repo}/{branch}/README.md"
        try:
            req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
            with urllib.request.urlopen(req, timeout=5) as response:
                if response.status == 200:
                    text = response.read().decode('utf-8', errors='ignore')
                    # truncate to avoid long processing time; 8k chars is plenty for launch commands
                    return text[:8000]
        except Exception as e:
            # print(f"    [Fetch Error] {branch}: {e}")
            continue
    return None

def call_api(url, model_name, readme_content, timeout):
    payload = {
        "model": model_name,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"Here is the README:\n\n{readme_content}\n\nExtract the stdio launch command JSON."}
        ],
        "temperature": 0.1,
        "stream": False
    }
    
    req = urllib.request.Request(url, data=json.dumps(payload).encode('utf-8'), headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=timeout) as response:
        res_data = json.loads(response.read().decode('utf-8'))
        msg = res_data['choices'][0]['message']['content']
        
        # strip markdown json blocks
        msg = msg.strip()
        if msg.startswith('```json'): msg = msg[7:]
        if msg.startswith('```'): msg = msg[3:]
        if msg.endswith('```'): msg = msg[:-3]
        
        return json.loads(msg.strip())

def query_llm(readme_content):
    try:
        # Try fast proxy first
        return call_api(PROXY_URL, "gpt-4o-mini", readme_content, 60)
    except Exception as e:
        print(f"  [Proxy Failed] {e}. Falling back to LMStudio...")
        try:
            return call_api(LMSTUDIO_URL, "gemma-4-e4b-uncensored-hauhaucs-aggressive", readme_content, 120)
        except Exception as e2:
            print(f"  [LMStudio Error] {e2}")
            return None

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--limit', type=int, default=50, help='Max servers to process')
    args = parser.parse_args()

    print(f"=== LLM RECIPE GENERATION (Limit: {args.limit}) ===")
    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    # Find discovered servers without a recipe
    c.execute("""
        SELECT s.uuid, s.canonical_id
        FROM published_mcp_servers s
        LEFT JOIN published_mcp_config_recipes r ON s.uuid = r.server_uuid
        WHERE s.status = 'discovered' AND r.uuid IS NULL
        LIMIT ?
    """, (args.limit,))
    
    candidates = c.fetchall()
    print(f"Found {len(candidates)} servers needing recipes.")

    success = 0
    
    for uuid_str, cid in candidates:
        print(f"\nProcessing: {cid}")
        parts = cid.split('/')
        if len(parts) != 3 or parts[0] != 'github':
            print("  Not a github repo, skipping.")
            continue
            
        owner, repo = parts[1], parts[2]
        
        readme = fetch_readme(owner, repo)
        if not readme:
            print("  Could not fetch README.md (tried main/master).")
            continue
            
        print(f"  Fetched README ({len(readme)} chars). Querying LMStudio...")
        
        start_t = time.time()
        recipe_json = query_llm(readme)
        elapsed = time.time() - start_t
        
        if recipe_json and 'command' in recipe_json:
            print(f"  [LLM {elapsed:.1f}s] Success: {recipe_json['command']} {' '.join(recipe_json.get('args', []))}")
            
            # Insert recipe
            c.execute("""
                INSERT INTO published_mcp_config_recipes (uuid, server_uuid, template, created_at, updated_at)
                VALUES (?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
            """, (str(uuid.uuid4()), uuid_str, json.dumps(recipe_json)))
            conn.commit()
            success += 1
        else:
            print(f"  [LLM {elapsed:.1f}s] Failed to extract valid JSON recipe.")

    print(f"\nDone! Generated {success} recipes out of {len(candidates)} attempts.")
    conn.close()

if __name__ == "__main__":
    main()
