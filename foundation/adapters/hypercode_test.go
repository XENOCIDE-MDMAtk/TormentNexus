package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHypercodeAdapterBuildsStatusWithoutPanicking(t *testing.T) {
	dir := t.TempDir()
	hypercodeDir := filepath.Join(dir, "..", "hypercode")
	if err := os.MkdirAll(hypercodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hypercodeDir, "README.md"), []byte("# Hypercode"), 0o644); err != nil {
		t.Fatal(err)
	}
	adapter := NewHypercodeAdapter(dir)
	status := adapter.Status()
	if !status.Assimilated {
		t.Fatal("expected assimilated hypercode adapter")
	}
	if status.MemoryContext == "" {
		t.Fatal("expected memory context")
	}
	if status.Provider.CurrentProvider == "" {
		t.Fatal("expected provider status")
	}
	if status.HypercodeRepoPath == "" {
		t.Fatal("expected discovered hypercode repo path")
	}
	if adapter.RouteMCP("list tools") == "" {
		t.Fatal("expected routed MCP string")
	}
	if adapter.BuildSystemContext() == "" {
		t.Fatal("expected system context")
	}
}
