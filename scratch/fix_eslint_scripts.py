import json
import glob

files = [
    "apps/vscode/package.json",
    "packages/agents/package.json",
    "packages/ai/package.json",
    "packages/cli/package.json",
    "packages/core/package.json",
    "packages/search/package.json",
    "packages/tools/package.json",
    "packages/ui/package.json"
]

for file_path in files:
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
        
        # Replace the problematic lint scripts
        content = content.replace(
            '"lint": "eslint --no-eslintrc --parser @typescript-eslint/parser \\"src/**/*.{ts,tsx,js,mjs,cjs}\\""',
            '"lint": "eslint \\"src/**/*.{ts,tsx,js,mjs,cjs}\\""'
        )
        content = content.replace(
            '"lint": "eslint --no-eslintrc --no-inline-config --parser @typescript-eslint/parser \\"src/**/*.{ts,tsx,js,mjs,cjs}\\""',
            '"lint": "eslint \\"src/**/*.{ts,tsx,js,mjs,cjs}\\""'
        )
        
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Fixed {file_path}")
    except Exception as e:
        print(f"Failed to fix {file_path}: {e}")

