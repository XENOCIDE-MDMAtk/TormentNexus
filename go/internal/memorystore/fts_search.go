package memorystore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// FTSMemorySearch provides full-text search across L2 and L3 memory tiers
// using SQLite FTS5 virtual tables.  Content, tags, kind, and category are
// indexed for fast keyword and phrase queries.
type FTSMemorySearch struct {
	db *sql.DB
}

const ftsSchemaSQL = `
-- FTS5 virtual table for L2 (hot/warm) memories
CREATE VIRTUAL TABLE IF NOT EXISTS l2_memory_fts USING fts5(
    memory_id UNINDEXED,
    content,
    tags,
    kind,
    category,
    source_url,
    tokenize='porter unicode61'
);

-- Triggers to keep FTS index in sync with L2 vault table
CREATE TRIGGER IF NOT EXISTS l2_vault_ai AFTER INSERT ON l2_vault BEGIN
    INSERT INTO l2_memory_fts(memory_id, content, tags, kind, category, source_url)
    VALUES (new.id, new.content, new.tags, new.memory_kind, new.category, new.source_url);
END;

CREATE TRIGGER IF NOT EXISTS l2_vault_ad AFTER DELETE ON l2_vault BEGIN
    DELETE FROM l2_memory_fts WHERE memory_id = old.id;
END;

CREATE TRIGGER IF NOT EXISTS l2_vault_au AFTER UPDATE ON l2_vault BEGIN
    DELETE FROM l2_memory_fts WHERE memory_id = old.id;
    INSERT INTO l2_memory_fts(memory_id, content, tags, kind, category, source_url)
    VALUES (new.id, new.content, new.tags, new.memory_kind, new.category, new.source_url);
END;
`

// NewFTSMemorySearch initialises the FTS5 search index on the memory DB.
func NewFTSMemorySearch(db *sql.DB) (*FTSMemorySearch, error) {
	if _, err := db.Exec(ftsSchemaSQL); err != nil {
		return nil, fmt.Errorf("fts schema: %w", err)
	}
	return &FTSMemorySearch{db: db}, nil
}

// FTSMemorySearchResult combines a memory record with its BM25 relevance score.
type FTSMemorySearchResult struct {
	Record controlplane.L2VaultRecord `json:"record"`
	Score  float64                    `json:"score"`
	Tier   string                     `json:"tier"` // "l2" or "l3"
}

type SearchResponse struct {
	Results []FTSMemorySearchResult `json:"results"`
	Total   int                     `json:"total"`
	Offset  int                     `json:"offset"`
	Limit   int                     `json:"limit"`
}

// Search executes a full-text query across L2 and (optionally) L3 memories.
// Query syntax uses FTS5 standard: words, phrases ("in quotes"), prefix (term*),
// and NEAR/AND/OR operators.
func (f *FTSMemorySearch) Search(ctx context.Context, query string, includeCold bool, limit int, offset int) (*SearchResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var results []FTSMemorySearchResult
	total := 0

	// Get total count first
	_ = f.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM l2_memory_fts WHERE l2_memory_fts MATCH ?
	`, query).Scan(&total)

	// Search L2 vault via FTS5
	rows, err := f.db.QueryContext(ctx, `
		SELECT f.memory_id, v.session_id, v.memory_type, v.memory_kind, v.category,
		       v.tags, v.source_url, v.content, v.importance, v.heat_score,
		       v.last_accessed_at, v.created_at,
		       bm25(l2_memory_fts, 0, 5.0, 2.0, 1.0, 1.0, 0.5) AS score
		FROM l2_memory_fts f
		JOIN l2_vault v ON v.id = f.memory_id
		WHERE l2_memory_fts MATCH ?
		ORDER BY score
		LIMIT ? OFFSET ?
	`, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fts l2 search: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r controlplane.L2VaultRecord
		var lastAccessed, createdAt sql.NullString
		var score float64
		if err := rows.Scan(&r.ID, &r.SessionID, &r.Type, &r.Kind, &r.Category,
			&r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore,
			&lastAccessed, &createdAt, &score); err != nil {
			continue
		}
		if lastAccessed.Valid {
			r.LastAccessedAt, _ = time.Parse(time.RFC3339, lastAccessed.String)
		}
		if createdAt.Valid {
			r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		results = append(results, FTSMemorySearchResult{Record: r, Score: score, Tier: "l2"})
	}

	// Optionally search L3 cold archive (keyword fallback since FTS5 not on L3)
	if includeCold {
		coldRows, err := f.db.QueryContext(ctx, `
			SELECT id, session_id, kind, category, tags, source_url,
			       content, importance, heat_score, created_at
			FROM l3_cold_archive
			WHERE content LIKE ? OR tags LIKE ? OR kind LIKE ?
			ORDER BY heat_score DESC, importance DESC
			LIMIT ? OFFSET ?
		`, "%"+query+"%", "%"+query+"%", "%"+query+"%", limit, offset)
		if err == nil {
			defer coldRows.Close()
			for coldRows.Next() {
				var r controlplane.L2VaultRecord
				var createdAtStr string
				if err := coldRows.Scan(&r.ID, &r.SessionID, &r.Kind, &r.Category,
					&r.Tags, &r.SourceURL, &r.Content, &r.Importance, &r.HeatScore, &createdAtStr); err != nil {
					continue
				}
				r.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
				r.LastAccessedAt = time.Now()
				r.Type = controlplane.MemoryArchive
				results = append(results, FTSMemorySearchResult{Record: r, Score: 0, Tier: "l3"})
			}
		}
	}

	return &SearchResponse{Results: results, Total: total, Offset: offset, Limit: limit}, nil
}

// RebuildIndex drops and re-creates the FTS index from scratch.
// Call this after bulk imports or schema migrations.
func (f *FTSMemorySearch) RebuildIndex(ctx context.Context) error {
	if _, err := f.db.ExecContext(ctx, `INSERT INTO l2_memory_fts(l2_memory_fts) VALUES('rebuild')`); err != nil {
		return fmt.Errorf("fts rebuild: %w", err)
	}
	return nil
}

// Optimize merges FTS5 index segments for better performance.
func (f *FTSMemorySearch) Optimize(ctx context.Context) error {
	if _, err := f.db.ExecContext(ctx, `INSERT INTO l2_memory_fts(l2_memory_fts) VALUES('optimize')`); err != nil {
		return fmt.Errorf("fts optimize: %w", err)
	}
	return nil
}

// rebuildFTSIndex truncates and re-populates the FTS index from the l2_vault table.
func rebuildFTSIndex(db *sql.DB) {
	// Remove all rows from the FTS table
	_, _ = db.Exec(`DELETE FROM l2_memory_fts`)
	// Re-insert all memories from l2_vault
	rows, err := db.Query(`SELECT id, content, tags, memory_kind, category, source_url FROM l2_vault`)
	if err != nil {
		fmt.Printf("Warning: rebuildFTSIndex query: %v\n", err)
		return
	}
	defer rows.Close()

	stmt, err := db.Prepare(`INSERT INTO l2_memory_fts(memory_id, content, tags, kind, category, source_url) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fmt.Printf("Warning: rebuildFTSIndex prepare: %v\n", err)
		return
	}
	defer stmt.Close()

	var count int
	for rows.Next() {
		var id, content, tags, kind, category, sourceURL string
		if err := rows.Scan(&id, &content, &tags, &kind, &category, &sourceURL); err != nil {
			continue
		}
		if _, err := stmt.Exec(id, content, tags, kind, category, sourceURL); err != nil {
			continue
		}
		count++
	}
	if count > 0 {
		fmt.Printf("FTS5 index rebuilt: %d memories indexed\n", count)
	}
}
