import os
import json
import glob

workspace_dirs = [
    "apps/*",
    "packages/*"
]

config_content = "export default [];\n"

for pattern in workspace_dirs:
    for pkg_dir in glob.glob(pattern):
        if not os.path.isdir(pkg_dir):
            continue
            
        pkg_json_path = os.path.join(pkg_dir, "package.json")
        if not os.path.exists(pkg_json_path):
            continue
            
        try:
            with open(pkg_json_path, "r", encoding="utf-8") as f:
                data = json.load(f)
                
            lint_script = data.get("scripts", {}).get("lint", "")
            if "eslint" in lint_script:
                has_config = False
                for ext in ["js", "mjs", "cjs"]:
                    if os.path.exists(os.path.join(pkg_dir, f"eslint.config.{ext}")):
                        has_config = True
                        break
                        
                if not has_config:
                    config_path = os.path.join(pkg_dir, "eslint.config.mjs")
                    with open(config_path, "w", encoding="utf-8") as f:
                        f.write(config_content)
                    print(f"Created {config_path}")
        except Exception as e:
            print(f"Failed to process {pkg_dir}: {e}")
