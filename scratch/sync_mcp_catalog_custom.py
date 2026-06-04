import sqlite3

MAIN_DB = "tormentnexus.db"
CATALOG_DB = "catalog.db"

def main():
    print("=== SYNCING MCP CATALOG TABLES ===")
    conn_main = sqlite3.connect(MAIN_DB)
    c_main = conn_main.cursor()
    
    conn_cat = sqlite3.connect(CATALOG_DB)
    c_cat = conn_cat.cursor()
    
    # 1. published_mcp_servers
    print("Syncing published_mcp_servers...")
    # Select columns that exist in main (21 columns)
    cols = ['uuid', 'canonical_id', 'display_name', 'description', 'author', 'repository_url', 
            'homepage_url', 'icon_url', 'transport', 'install_method', 'auth_model', 'status', 
            'confidence', 'tags', 'categories', 'stars', 'last_seen_at', 'last_verified_at', 
            'created_at', 'updated_at', 'favicon_url']
    
    cols_str = ", ".join(cols)
    placeholders = ", ".join(["?"] * len(cols))
    
    c_main.execute(f"SELECT {cols_str} FROM published_mcp_servers")
    rows = c_main.fetchall()
    print(f"Read {len(rows)} servers from main.")
    
    c_cat.executemany(f"""
        INSERT OR IGNORE INTO published_mcp_servers ({cols_str})
        VALUES ({placeholders})
    """, rows)
    conn_cat.commit()
    
    # 2. published_mcp_server_sources
    print("Syncing published_mcp_server_sources...")
    c_main.execute("SELECT uuid, server_uuid, source_name, source_url, raw_payload, first_seen_at, last_seen_at FROM published_mcp_server_sources")
    rows_sources = c_main.fetchall()
    c_cat.executemany("""
        INSERT OR IGNORE INTO published_mcp_server_sources (uuid, server_uuid, source_name, source_url, raw_payload, first_seen_at, last_seen_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """, rows_sources)
    conn_cat.commit()

    # 3. published_mcp_config_recipes
    print("Syncing published_mcp_config_recipes...")
    c_main.execute("SELECT uuid, server_uuid, recipe_version, template, required_secrets, required_env, confidence, explanation, is_active, generated_by, created_at, updated_at FROM published_mcp_config_recipes")
    rows_recipes = c_main.fetchall()
    c_cat.executemany("""
        INSERT OR IGNORE INTO published_mcp_config_recipes (uuid, server_uuid, recipe_version, template, required_secrets, required_env, confidence, explanation, is_active, generated_by, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, rows_recipes)
    conn_cat.commit()

    # 4. published_mcp_validation_runs
    print("Syncing published_mcp_validation_runs...")
    c_main.execute("SELECT uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at FROM published_mcp_validation_runs")
    rows_val = c_main.fetchall()
    c_cat.executemany("""
        INSERT OR IGNORE INTO published_mcp_validation_runs (uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count, findings_summary, performed_by, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, rows_val)
    conn_cat.commit()

    print("Sync complete!")
    c_cat.execute("SELECT count(*) FROM published_mcp_servers")
    print(f"Catalog total servers: {c_cat.fetchone()[0]}")
    c_cat.execute("SELECT count(*) FROM published_mcp_config_recipes")
    print(f"Catalog total recipes: {c_cat.fetchone()[0]}")

    conn_main.close()
    conn_cat.close()

if __name__ == "__main__":
    main()
