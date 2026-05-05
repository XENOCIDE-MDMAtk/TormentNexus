package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/borghq/borg-go/internal/controlplane"
	_ "modernc.org/sqlite"
)

type VectorStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewVectorStore(dbPath string) (*VectorStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent access support
	if dbPath != ":memory:" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
		}
		if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set busy timeout: %w", err)
		}
	}

	// Initialize schema from foundation
	if _, err := db.Exec(controlplane.VectorSchemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init vector schema: %w", err)
	}

	return &VectorStore{db: db}, nil
}

func (s *VectorStore) Close() error {
	return s.db.Close()
}

func (s *VectorStore) Commit(ctx context.Context, entry controlplane.L2VaultRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Insert into regular table
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO l2_vault (id, session_id, memory_type, content, importance, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			content = excluded.content,
			importance = excluded.importance,
			created_at = excluded.created_at
	`, entry.ID, entry.SessionID, string(entry.Type), entry.Content, entry.Importance, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("memorystore commit insert: %w", err)
	}

	// If we have an embedding, insert into virtual table
	if len(entry.Embedding) > 0 {
		embeddingJSON, _ := json.Marshal(entry.Embedding)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO vec_l2_vault (rowid, embedding)
			SELECT rowid, ? FROM l2_vault WHERE id = ?
		`, string(embeddingJSON), entry.ID)
		if err != nil {
			return fmt.Errorf("memorystore commit embedding: %w", err)
		}
	}
	return nil
}

func (s *VectorStore) SemanticSearch(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Fallback to LIKE search (real sqlite-vec search when embeddings available)
	queryStr := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, memory_type, content, importance, created_at
		FROM l2_vault
		WHERE content LIKE ?
		ORDER BY importance DESC, created_at DESC
		LIMIT ?
	`, queryStr, limit)
	if err != nil {
		return nil, fmt.Errorf("memorystore search: %w", err)
	}
	defer rows.Close()

	var results []controlplane.L2VaultRecord
	for rows.Next() {
		var r controlplane.L2VaultRecord
		var mType string
		if err := rows.Scan(&r.ID, &r.SessionID, &mType, &r.Content, &r.Importance, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Type = controlplane.MemoryType(mType)
		results = append(results, r)
	}
	return results, nil
}
