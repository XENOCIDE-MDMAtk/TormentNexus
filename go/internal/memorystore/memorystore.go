package memorystore

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
)

type Manager struct {
	path string
	vs   *VectorStore
}

func NewManager(path string) *Manager {
	// The path here is for the JSON file, but we'll use a SQLite DB next to it
	dbPath := filepath.Join(filepath.Dir(path), "memory.db")
	vs, _ := NewVectorStore(dbPath)
	m := &Manager{path: path, vs: vs}
	m.startSleepCycleEngine()
	return m
}

func (m *Manager) startSleepCycleEngine() {
	if m.vs == nil {
		return
	}
	go func() {
		ctx := context.Background()
		limbo, limboErr := NewLimboVault(m.vs.db)
		if limboErr != nil {
			fmt.Printf("SleepCycle: L4 limbo unavailable: %v\n", limboErr)
			limbo = nil
		}

		// Run initial cycle on boot
		_ = m.vs.ForgettingCurveDecay(ctx)
		_ = m.vs.ConsolidateMemories(ctx)
		_ = m.vs.MentalModelReflection(ctx)
		if limbo != nil {
			_ = BuryOrphanedMemories(ctx, m.vs.db, limbo)
			_ = DreamCycle(ctx, m.vs.db)
		}

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			_ = m.vs.ForgettingCurveDecay(ctx)
			_ = m.vs.ConsolidateMemories(ctx)
			_ = m.vs.MentalModelReflection(ctx)
			if limbo != nil {
				_ = BuryOrphanedMemories(ctx, m.vs.db, limbo)
				_ = DreamCycle(ctx, m.vs.db)
				// Hard-delete limbo entries older than 90 days
				n, _ := limbo.HardDelete(ctx, 90*24*time.Hour)
				if n > 0 {
					fmt.Printf("SleepCycle: purged %d old limbo entries\n", n)
				}
			}
		}
	}()
}

func (m *Manager) Close() error {
	if m.vs != nil {
		return m.vs.Close()
	}
	return nil
}

func (m *Manager) GetAll() ([]map[string]interface{}, error) {
	if m.vs == nil {
		return []map[string]interface{}{}, nil
	}

	// We'll return the L2 vault entries as a generic map for compatibility
	results, err := m.vs.SemanticSearch(context.Background(), "", 1000)
	if err != nil {
		return nil, err
	}

	var genericResults []map[string]interface{}
	for _, r := range results {
		genericResults = append(genericResults, map[string]interface{}{
			"id":               r.ID,
			"session_id":       r.SessionID,
			"type":             string(r.Type),
			"content":          r.Content,
			"importance":       r.Importance,
			"heat_score":       r.HeatScore,
			"last_accessed_at": r.LastAccessedAt,
			"created_at":       r.CreatedAt,
		})
	}
	return genericResults, nil
}

func (m *Manager) GetMemories() []string {
	all, _ := m.GetAll()
	var contents []string
	for _, item := range all {
		if content, ok := item["content"].(string); ok {
			contents = append(contents, content)
		}
	}
	return contents
}

func (m *Manager) AddMemory(mem string) {
	if m.vs == nil {
		return
	}

	entry := controlplane.L2VaultRecord{
		ID:         fmt.Sprintf("mem-%d", SystemNowUnixNano()),
		SessionID:  "manual",
		Type:       controlplane.MemoryLongTerm,
		Content:    mem,
		Importance: 0.5,
		CreatedAt:  controlplane.Now(),
	}
	_ = m.vs.Commit(context.Background(), entry)
}

// SystemNowUnixNano is a helper to get unique IDs
func SystemNowUnixNano() int64 {
	return controlplane.Now().UnixNano()
}
