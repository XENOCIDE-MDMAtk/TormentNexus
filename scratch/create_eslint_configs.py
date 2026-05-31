import os

packages = [
    "apps/vscode",
    "packages/agents",
    "packages/ai",
    "packages/cli",
    "packages/core",
    "packages/search",
    "packages/tools",
    "packages/ui"
]

config_content = "export default [];\n"

for pkg in packages:
    try:
        file_path = os.path.join(pkg, "eslint.config.mjs")
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(config_content)
        print(f"Created {file_path}")
    except Exception as e:
        print(f"Failed to create {file_path}: {e}")

