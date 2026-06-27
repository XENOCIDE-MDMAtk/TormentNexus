package skillregistry

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// IndexSkillsToCatalog indexes the currently loaded skills into the given catalog.db
// It creates a published_skills table if it doesn't exist, to enable unified search
// across MCP servers and internal skills.
func (sr *SkillRegistry) IndexSkillsToCatalog(catalogDBPath string) error {
	db, err := sql.Open("sqlite", catalogDBPath)
	if err != nil {
		return fmt.Errorf("failed to open catalog.db: %w", err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS published_skills (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			content TEXT,
			use_count INTEGER DEFAULT 0,
			successes INTEGER DEFAULT 0,
			failures INTEGER DEFAULT 0,
			is_retired BOOLEAN DEFAULT 0
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create published_skills table: %w", err)
	}

	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Upsert all skills
	stmt, err := db.Prepare(`
		INSERT INTO published_skills (id, name, description, content, use_count, successes, failures, is_retired)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			description=excluded.description,
			content=excluded.content,
			use_count=excluded.use_count,
			successes=excluded.successes,
			failures=excluded.failures,
			is_retired=excluded.is_retired;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for id, skill := range sr.skills {
		// Assuming we extract the properties if they exist on the struct
		// Using dummy metrics for the ones missing from SkillInfo since they are in SkillLoaded
		// If SkillInfo has no metrics, we just insert zeroes.
		_, err := stmt.Exec(
			strings.ToLower(id),
			skill.Name,
			skill.Description,
			skill.Content,
			0, // use_count
			0, // successes
			0, // failures
			false, // is_retired
		)
		if err != nil {
			fmt.Printf("[CatalogIndexer] Failed to index skill %s: %v\n", id, err)
		}
	}

	return nil
}
