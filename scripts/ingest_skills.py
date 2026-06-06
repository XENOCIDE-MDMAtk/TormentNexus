"""
Optimized bulk ingest of agent skills into SQLite skills registry.
Track B1 - skips expensive similarity checks for speed; deduplication done separately.
"""

import sqlite3
import re
from pathlib import Path

SKILLS_DIR = Path(r"C:\Users\hyper\workspace\.agent\skills")
DB_PATH = Path(r"C:\Users\hyper\workspace\tormentnexus\data\assimilation_state.db")


def extract_skill_info(skill_dir: Path) -> dict:
    """Extract skill metadata from SKILL.md file."""
    md_file = skill_dir / "SKILL.md"
    if not md_file.exists():
        return None

    content = md_file.read_text(encoding="utf-8", errors="replace")
    category = "general"
    description = ""
    frontmatter = content[:800]

    fm_match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
    if fm_match:
        fm_text = fm_match.group(1)
        m = re.search(r"category:\s*(.+)", fm_text, re.IGNORECASE)
        if m:
            category = m.group(1).strip().strip("\"'")
        m = re.search(r"description:\s*(.+)", fm_text, re.IGNORECASE)
        if m:
            description = m.group(1).strip().strip("\"'")[:200]

    if not description:
        for line in content.splitlines():
            line = line.strip()
            if line and not line.startswith("#") and not line.startswith("---"):
                description = line[:200]
                break

    return dict(
        name=skill_dir.name,
        description=description,
        category=category,
        frontmatter=frontmatter,
        content=content,
    )


def main():
    print("=== Skill Bulk Ingest (Optimized) ===\n")

    if not SKILLS_DIR.exists():
        print(f"ERROR: Skills directory not found: {SKILLS_DIR}")
        return

    skill_dirs = [d for d in SKILLS_DIR.iterdir() if d.is_dir()]
    print(f"Found {len(skill_dirs)} skill directories\n")

    # Ensure DB
    DB_PATH.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(DB_PATH))
    conn.executescript("""
        CREATE TABLE IF NOT EXISTS skills (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE NOT NULL,
            description TEXT DEFAULT '',
            category TEXT DEFAULT 'general',
            frontmatter TEXT DEFAULT '',
            content TEXT DEFAULT '',
            version INTEGER DEFAULT 1,
            similarity_score INTEGER DEFAULT 100,
            canonical_id INTEGER REFERENCES skills(id),
            status TEXT DEFAULT 'active',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name);
        CREATE INDEX IF NOT EXISTS idx_skills_status ON skills(status);
    """)
    conn.commit()

    inserted = 0
    skipped = 0
    errors = 0

    for skill_dir in skill_dirs:
        try:
            info = extract_skill_info(skill_dir)
            if not info:
                skipped += 1
                continue

            conn.execute(
                """
                INSERT OR IGNORE INTO skills (name, description, category, frontmatter, content)
                VALUES (?,?,?,?,?)
            """,
                (
                    info["name"],
                    info["description"],
                    info["category"],
                    info["frontmatter"],
                    info["content"],
                ),
            )
            inserted += 1

        except Exception as e:
            errors += 1
            print(f"  ERROR {skill_dir.name}: {e}")

    conn.commit()
    total = conn.execute(
        "SELECT COUNT(*) FROM skills WHERE status='active'"
    ).fetchone()[0]
    conn.close()

    print("=== INGEST COMPLETE ===")
    print(f"  Processed: {len(skill_dirs)}")
    print(f"  Inserted:  {inserted}")
    print(f"  Skipped:   {skipped}")
    print(f"  Errors:    {errors}")
    print(f"  DB total:  {total} active skills")


if __name__ == "__main__":
    main()
