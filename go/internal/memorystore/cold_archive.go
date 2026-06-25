package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

// L3ColdArchive implements the L3 cold storage tier for long-term compressed
// memory.  Memories that have decayed below a heat threshold are moved here
// automatically during maintenance cycles.  They can be promoted back to L2
// on semantic-match recall.
type L3ColdArchive struct {
	db *sql.DB
}

// ColdArchiveSchemaSQL is the DDL for the L3 cold archive table.
const ColdArchiveSchemaSQL = `
CREATE TABLE IF NOT EXISTS l3_cold_archive (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL DEFAULT '',
    kind        TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT '',
    tags        TEXT NOT NULL DEFAULT '',
    source_url  TEXT NOT NULL DEFAULT '',
    content     TEXT NOT NULL,
    importance  REAL NOT NULL DEFAULT 0.0,
    heat_score  REAL NOT NULL DEFAULT 0.0,
    archived_at TEXT NOT NULL DEFAULT (datetime('now')),
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    metadata    TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_cold_archive_heat   ON l3_cold_archive(heat_score);
CREATE INDEX IF NOT EXISTS idx_cold_archive_kind   ON l3_cold_archive(kind);
CREATE INDEX IF NOT EXISTS idx_cold_archive_created ON l3_cold_archive(created_at);
`

// NewColdArchive opens (or creates) the L3 cold archive DB.
func NewColdArchive(dbPath string) (*L3ColdArchive, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cold archive open: %w", err)
	}
	if dbPath != ":memory:" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("cold archive WAL: %w", err)
		}
	}
	if _, err := db.Exec(ColdArchiveSchemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("cold archive schema: %w", err)
	}
	return &L3ColdArchive{db: db}, nil
}

func (a *L3ColdArchive) Close() error {
	return a.db.Close()
}

// Archive moves a memory from L2 (hot/warm) to L3 (cold).
func (a *L3ColdArchive) Archive(ctx context.Context, record controlplane.L2VaultRecord) error {
	meta, _ := json.Marshal(map[string]any{
		"heat_decayed": true,
		"archived_at":  time.Now().UTC().Format(time.RFC3339),
	})
	_, err := a.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO l3_cold_archive
			(id, session_id, kind, category, tags, source_url,
			 content, importance, heat_score, archived_at, created_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?, ?)
	`, record.ID, record.SessionID, record.Kind, record.Category,
		record.Tags, record.SourceURL,
		record.Content, record.Importance, record.HeatScore,
		record.CreatedAt.Format(time.RFC3339), string(meta))
	return err
}

// Promote moves a cold record back to L2 (e.g. on semantic match).
func (a *L3ColdArchive) Promote(ctx context.Context, id string) (*controlplane.L2VaultRecord, error) {
	row := a.db.QueryRowContext(ctx, `
		SELECT id, session_id, kind, category, tags, source_url,
		       content, importance, heat_score, created_at
		FROM l3_cold_archive WHERE id = ?
	`, id)

	var r controlplane.L2VaultRecord
	var createdAtStr string
	err := row.Scan(&r.ID, &r.SessionID, &r.Kind, &r.Category,
		&r.Tags, &r.SourceURL, &r.Content,
		&r.Importance, &r.HeatScore, &createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("promote find: %w", err)
	}

	r.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	r.LastAccessedAt = time.Now()
	r.Type = controlplane.MemoryLongTerm

	// Boost heat on promotion so it doesn't immediately re-archive
	r.HeatScore = 25.0

	// Delete from cold archive
	_, _ = a.db.ExecContext(ctx, `DELETE FROM l3_cold_archive WHERE id = ?`, id)

	return &r, nil
}

// SearchCold performs a keyword search across the cold archive.
func (a *L3ColdArchive) SearchCold(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	like := "%" + query + "%"
	rows, err := a.db.QueryContext(ctx, `
		SELECT id, session_id, kind, category, tags, source_url,
		       content, importance, heat_score, created_at
		FROM l3_cold_archive
		WHERE content LIKE ? OR tags LIKE ? OR kind LIKE ?
		ORDER BY heat_score DESC, importance DESC
		LIMIT ?
	`, like, like, like, limit)
	if err != nil {
		return nil, fmt.Errorf("cold search: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var createdAtStr string
		if err := rows.Scan(&r.ID, &r.SessionID, &r.Kind, &r.Category,
			&r.Tags, &r.SourceURL, &r.Content,
			&r.Importance, &r.HeatScore, &createdAtStr); err != nil {
			continue
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		r.LastAccessedAt = time.Now()
		r.Type = controlplane.MemoryArchive
		results = append(results, r)
	}
	return results, nil
}

// Count returns the total number of archived memories.
func (a *L3ColdArchive) Count(ctx context.Context) (int, error) {
	var count int
	err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM l3_cold_archive`).Scan(&count)
	return count, err
}

// GC purges records that have been in cold storage past the given TTL.
func (a *L3ColdArchive) GC(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan).Format(time.RFC3339)
	res, err := a.db.ExecContext(ctx, `DELETE FROM l3_cold_archive WHERE archived_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}
