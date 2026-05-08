package memory

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestStoreAndRetrieve(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())

	id, err := mm.Store(Memory{
		Kind:    KindFact,
		Tier:    TierLongTerm,
		Content: "The project uses Go 1.22 and TypeScript",
		Tags:    []string{"project", "config"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected non-empty ID")
	}

	results, err := mm.Retrieve(MemoryQuery{Query: "Go TypeScript project", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected to find the stored memory")
	}
	if results[0].Content != "The project uses Go 1.22 and TypeScript" {
		t.Errorf("unexpected content: %s", results[0].Content)
	}
}

func TestRetrieveByKind(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())

	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "fact one"})
	mm.Store(Memory{Kind: KindDecision, Tier: TierLongTerm, Content: "decision one"})
	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "fact two"})

	results, _ := mm.Retrieve(MemoryQuery{Kind: KindFact, Limit: 10})
	if len(results) != 2 {
		t.Errorf("expected 2 fact memories, got %d", len(results))
	}
}

func TestRetrieveByTag(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())

	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "tagged", Tags: []string{"important", "project"}})
	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "untagged"})

	results, _ := mm.Retrieve(MemoryQuery{Tags: []string{"important"}, Limit: 10})
	if len(results) != 1 {
		t.Errorf("expected 1 tagged memory, got %d", len(results))
	}
}

func TestDelete(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())

	id, _ := mm.Store(Memory{Kind: KindFact, Tier: TierShortTerm, Content: "to be deleted"})
	if !mm.Delete(id) {
		t.Fatal("expected delete to succeed")
	}
	if mm.Delete(id) {
		t.Fatal("expected second delete to fail")
	}
	if mm.Count() != 0 {
		t.Error("expected 0 memories after delete")
	}
}

func TestDemoteAndPromote(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())

	id, _ := mm.Store(Memory{Kind: KindContext, Tier: TierShortTerm, Content: "session context"})

	// Manually age it
	mm.mu.Lock()
	mm.memories[id].CreatedAt = time.Now().Add(-2 * time.Hour)
	mm.mu.Unlock()

	demoted := mm.Demote(1 * time.Hour)
	if demoted != 1 {
		t.Fatalf("expected 1 demotion, got %d", demoted)
	}

	m, ok := mm.Get(id)
	if !ok {
		t.Fatal("memory should exist")
	}
	if m.Tier != TierMediumTerm {
		t.Errorf("expected tier=medium, got %s", m.Tier)
	}

	// Access it enough to promote
	for i := 0; i < 10; i++ {
		mm.Get(id)
	}

	promoted := mm.Promote(5)
	if promoted != 1 {
		t.Fatalf("expected 1 promotion, got %d", promoted)
	}

	m, _ = mm.Get(id)
	if m.Tier != TierLongTerm {
		t.Errorf("expected tier=long, got %s", m.Tier)
	}
}

func TestPrune(t *testing.T) {
	cfg := DefaultMemoryManagerConfig()
	cfg.MaxShortTerm = 3
	mm := NewMemoryManager(cfg)

	for i := 0; i < 10; i++ {
		mm.Store(Memory{
			Kind:           KindContext,
			Tier:           TierShortTerm,
			Content:        fmt.Sprintf("context item %d with unique content", i),
			RelevanceScore: float64(i),
		})
	}

	// Force different scores so prune can differentiate
	mm.mu.Lock()
	count := 0
	for _, m := range mm.memories {
		m.RelevanceScore = float64(count)
		count++
	}
	mm.mu.Unlock()

	pruned := mm.Prune()
	if pruned == 0 {
		t.Error("expected some pruning")
	}

	counts := mm.CountByTier()
	if counts[TierShortTerm] > 3 {
		t.Errorf("expected at most 3 short-term memories, got %d", counts[TierShortTerm])
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/memories.json"

	mm := NewMemoryManager(MemoryManagerConfig{
		StorePath:     path,
		MaxShortTerm:  100,
		MaxMediumTerm: 500,
		MaxLongTerm:   5000,
	})

	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "persistent fact", Tags: []string{"test"}})
	mm.Store(Memory{Kind: KindDecision, Tier: TierLongTerm, Content: "persistent decision"})

	if err := mm.Save(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist")
	}

	// Load into new manager
	mm2 := NewMemoryManager(MemoryManagerConfig{
		StorePath:     path,
		MaxShortTerm:  100,
		MaxMediumTerm: 500,
		MaxLongTerm:   5000,
	})

	if mm2.Count() != 2 {
		t.Errorf("expected 2 memories after load, got %d", mm2.Count())
	}

	results, _ := mm2.Retrieve(MemoryQuery{Query: "persistent", Limit: 5})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestStats(t *testing.T) {
	mm := NewMemoryManager(DefaultMemoryManagerConfig())
	mm.Store(Memory{Kind: KindFact, Tier: TierLongTerm, Content: "fact"})
	mm.Store(Memory{Kind: KindContext, Tier: TierShortTerm, Content: "ctx"})

	stats := mm.Stats()
	if stats["total"] != 2 {
		t.Errorf("expected total=2, got %v", stats["total"])
	}
}
