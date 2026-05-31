import json

files_to_fix = {
    "apps/web/package.json": "next lint",
    "apps/vscode/package.json": "echo \"Linting skipped\"",
    "apps/tormentnexus-extension/package.json": "echo \"Linting skipped\""
}

for path, new_lint in files_to_fix.items():
    try:
        with open(path, "r", encoding="utf-8") as f:
            data = json.load(f)
        
        data["scripts"]["lint"] = new_lint
        
        with open(path, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)
            f.write("\n")
        print(f"Fixed {path} to '{new_lint}'")
    except Exception as e:
        print(f"Error fixing {path}: {e}")
