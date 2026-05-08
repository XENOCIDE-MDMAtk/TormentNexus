package vault

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestScratchpadLifecycle(t *testing.T) {
	sp, err := NewScratchpad()
	if err != nil { t.Fatalf("create: %v", err) }
	defer sp.Close()
	if err := sp.InitSession("test-agent", "test-project", "Build a REST API", ""); err != nil { t.Fatalf("init: %v", err) }
	if sp.SessionID() == "" { t.Error("session ID should be set") }
	sp.Append("assistant", "I'll build a REST API with Go.")
	sp.AppendTool("tool", "Reading main.go", "read_file", `{"path":"main.go"}`, "package main...")
	sp.Append("assistant", "The file exists.")
	entries, err := sp.GetAll()
	if err != nil { t.Fatalf("getall: %v", err) }
	if len(entries) != 4 { t.Errorf("expected 4 entries, got %d", len(entries)) }
	if entries[0].Role != "user" { t.Errorf("entry 0 role: %q", entries[0].Role) }
	if entries[2].ToolName != "read_file" { t.Errorf("entry 2 tool: %q", entries[2].ToolName) }
	sess, drainEntries, err := sp.Drain()
	if err != nil { t.Fatalf("drain: %v", err) }
	if sess.Status != StatusCompleted { t.Errorf("status: %q", sess.Status) }
	if sess.CompletedAt == nil { t.Error("completed_at should be set") }
	if len(drainEntries) != 4 { t.Errorf("drain entries: got %d", len(drainEntries)) }
}

func TestScratchpadWithL2Context(t *testing.T) {
	sp, _ := NewScratchpad(); defer sp.Close()
	ctx := "[Relevant Memory]\n- Previously used REST patterns"
	sp.InitSession("agent", "proj", "Build API", ctx)
	entries, _ := sp.GetAll()
	if len(entries) != 2 { t.Fatalf("expected 2, got %d", len(entries)) }
	if entries[0].Role != "system" { t.Errorf("entry 0: %q", entries[0].Role) }
	if !strings.Contains(entries[0].Content, "REST") { t.Error("L2 context not injected") }
}

func TestScratchpadNoSessionAppend(t *testing.T) {
	sp, _ := NewScratchpad(); defer sp.Close()
	err := sp.Append("assistant", "should fail")
	if err == nil { t.Error("expected error") }
}

func openTestVault(t *testing.T) *Vault {
	t.Helper()
	v, err := OpenVault(":memory:")
	if err != nil { t.Fatalf("open vault: %v", err) }
	t.Cleanup(func() { v.Close() })
	return v
}

func TestVaultCommitAndGet(t *testing.T) {
	v := openTestVault(t)
	entry := VaultEntry{MemoryType: MemoryHeuristic, SessionID: "sess-1", AgentName: "test", Project: "borg", Content: "Always use context.Context as first parameter", Importance: 0.8}
	id, err := v.Commit(entry)
	if err != nil { t.Fatalf("commit: %v", err) }
	got, err := v.Get(id)
	if err != nil { t.Fatalf("get: %v", err) }
	if got == nil { t.Fatal("not found") }
	if got.Content != entry.Content { t.Errorf("content: %q", got.Content) }
	if got.MemoryType != MemoryHeuristic { t.Errorf("type: %q", got.MemoryType) }
}

func TestVaultSearch(t *testing.T) {
	v := openTestVault(t)
	entries := []VaultEntry{
		{MemoryType: MemoryHeuristic, SessionID: "s1", Content: "Use chi router for REST APIs in Go", Importance: 0.8},
		{MemoryType: MemoryHeuristic, SessionID: "s2", Content: "Always handle errors in Go", Importance: 0.7},
		{MemoryType: MemoryRaw, SessionID: "s3", Content: "Raw transcript of database work", Importance: 0.3},
	}
	for _, e := range entries {
		id, _ := v.Commit(e)
		var vec []float32
		if strings.Contains(e.Content, "REST") { vec = []float32{0.9, 0.1, 0, 0} } else if strings.Contains(e.Content, "error") { vec = []float32{0.7, 0.3, 0, 0} } else { vec = []float32{0.1, 0, 0.8, 0.1} }
		v.stmtStoreEmbed.Exec(id, "all-MiniLM-L6-v2", 4, encodeVaultVector(vec))
	}
	hits, err := v.Search([]float32{1, 0, 0, 0}, 5, 0.1, true)
	if err != nil { t.Fatalf("search: %v", err) }
	if len(hits) == 0 { t.Fatal("no results") }
	if !strings.Contains(hits[0].Entry.Content, "chi") { t.Errorf("top hit: %q", hits[0].Entry.Content) }
}

func TestVaultRecentEntries(t *testing.T) {
	v := openTestVault(t)
	for i := 0; i < 5; i++ {
		v.Commit(VaultEntry{MemoryType: MemoryHeuristic, SessionID: fmt.Sprintf("s%d", i), Content: fmt.Sprintf("Lesson %d", i), Importance: 0.5})
		time.Sleep(1 * time.Millisecond)
	}
	recent, err := v.RecentEntries(3, MemoryHeuristic)
	if err != nil { t.Fatalf("recent: %v", err) }
	if len(recent) != 3 { t.Fatalf("expected 3, got %d", len(recent)) }
	if !strings.Contains(recent[0].Content, "Lesson 4") { t.Errorf("first: %q", recent[0].Content) }
}

func TestBridgeFullCycle(t *testing.T) {
	v := openTestVault(t)
	hID, _ := v.Commit(VaultEntry{MemoryType: MemoryHeuristic, SessionID: "past", AgentName: "test", Project: "borg", Content: "Always use structured logging with slog in Go services", Importance: 0.9})
	vec := []float32{0.8, 0.2, 0, 0}
	v.stmtStoreEmbed.Exec(hID, "all-MiniLM-L6-v2", 4, encodeVaultVector(vec))
	mockEmbed := func(text string) ([]float32, error) { return []float32{0.9, 0.1, 0, 0}, nil }
	bridge := NewBridge(v, mockEmbed)

	sp, err := bridge.InjectContext("test-agent", "borg", "Build a logging service")
	if err != nil { t.Fatalf("inject: %v", err) }
	entries, _ := sp.GetAll()
	if len(entries) < 2 { t.Fatalf("expected 2+ entries, got %d", len(entries)) }
	if entries[0].Role != "system" { t.Errorf("role: %q", entries[0].Role) }
	if !strings.Contains(entries[0].Content, "slog") { t.Errorf("no slog: %q", entries[0].Content) }

	sp.Append("assistant", "I'll use slog")
	sp.AppendTool("tool", "Writing", "write_file", `{"path":"logger.go"}`, "written")

	if err := bridge.CommitSession(sp, "Use slog for structured logging"); err != nil { t.Fatalf("commit: %v", err) }
	recent, _ := v.RecentEntries(10, "")
	if len(recent) < 2 { t.Fatalf("expected 2+ L2 entries, got %d", len(recent)) }
	hasRaw, hasHeuristic := false, false
	for _, e := range recent {
		if e.MemoryType == MemoryRaw { hasRaw = true }
		if e.MemoryType == MemoryHeuristic { hasHeuristic = true }
	}
	if !hasRaw { t.Error("no raw") }
	if !hasHeuristic { t.Error("no heuristic") }
}
