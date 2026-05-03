package memorystore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/borghq/borg-go/internal/controlplane"

	_ "modernc.org/sqlite"
)

type VectorStore struct {
	db *sql.DB
}

func NewVectorStore(dbPath string) (*VectorStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize schema from foundation
	if _, err := db.Exec(controlplane.VectorSchemaSQL); err != nil {
		return nil, fmt.Errorf("failed to init vector schema: %w", err)
	}

	return &VectorStore{db: db}, nil
}

func (s *VectorStore) Close() error {
	return s.db.Close()
}

func (s *VectorStore) Commit(ctx context.Context, entry controlplane.L2VaultRecord) error {
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
		return err
	}

	// If we have an embedding, insert into virtual table
	if len(entry.Embedding) > 0 {
		embeddingJSON, _ := json.Marshal(entry.Embedding)
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO vec_l2_vault (rowid, embedding)
			SELECT rowid, ? FROM l2_vault WHERE id = ?
		`, string(embeddingJSON), entry.ID)
		return err
	}

	return nil
}

func (s *VectorStore) SemanticSearch(ctx context.Context, query string, limit int) ([]controlplane.L2VaultRecord, error) {
	// NOTE: This implementation assumes we have a way to generate embeddings in Go.
	// For now, if we don't have embeddings, we fall back to LIKE search.
	
	// Real implementation with sqlite-vec would look like:
	/*
	rows, err := s.db.QueryContext(ctx, `
		SELECT v.id, v.session_id, v.memory_type, v.content, v.importance, v.created_at
		FROM l2_vault v
		JOIN vec_l2_vault vec ON v.rowid = vec.rowid
		WHERE vec.embedding MATCH ? -- query embedding
		ORDER BY distance
		LIMIT ?
	`, queryEmbedding, limit)
	*/

	// Fallback to LIKE search
	queryStr := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, session_id, memory_type, content, importance, created_at
		FROM l2_vault
		WHERE content LIKE ?
		ORDER BY importance DESC, created_at DESC
		LIMIT ?
	`, queryStr, limit)

	if err != nil {
		return nil, err
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
