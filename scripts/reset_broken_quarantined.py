#!/usr/bin/env python3
import pathlib
import sqlite3
import os

DB_PATH = "data/assimilation_state.db"
BROKEN_DIR = pathlib.Path("go/internal/tools/_broken")

def main():
    if not BROKEN_DIR.exists():
        print(f"ERROR: Directory {BROKEN_DIR} does not exist")
        return

    conn = sqlite3.connect(DB_PATH)
    c = conn.cursor()

    broken_files = list(BROKEN_DIR.glob("*.go"))
    print(f"Found {len(broken_files)} Go files in {BROKEN_DIR}.")

    reset_count = 0
    deleted_files = 0
    not_in_db_count = 0

    for fpath in broken_files:
        stem = fpath.stem
        
        # Check DB status
        c.execute("SELECT status FROM mcp_servers WHERE name=? OR package_name=?", (stem, stem))
        row = cursor = c.fetchone()
        
        if row:
            # Reset database status to pending
            c.execute(
                "UPDATE mcp_servers SET status='pending', notes=?, go_file=NULL, tools_exposed=NULL WHERE name=? OR package_name=?",
                ("reset broken quarantined tool", stem, stem)
            )
            reset_count += c.rowcount
        else:
            not_in_db_count += 1
            
        # Delete the file
        try:
            os.remove(fpath)
            deleted_files += 1
        except Exception as e:
            print(f"Failed to delete {fpath.name}: {e}")

    conn.commit()
    conn.close()

    print("\n=== Reset Quarantined Tools Summary ===")
    print(f"DB records reset to pending: {reset_count}")
    print(f"Files deleted from _broken:  {deleted_files}")
    print(f"Files not found in DB:       {not_in_db_count}")

if __name__ == "__main__":
    main()
