package mcp

import (
	"context"
	"fmt"
	"testing"
)

func TestDecisionSystemSearchTools(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	results, err := ds.SearchTools(context.Background(), "bash")
	if err != nil {
		t.Fatalf("SearchTools failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results for 'bash' query")
	}

	// Should find bash and shell (both match)
	found := false
	for _, r := range results {
		if r.OriginalName == "bash" || r.OriginalName == "shell" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find bash or shell tool")
	}
}

func TestDecisionSystemLoadUnload(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	// Load a tool
	lt, err := ds.LoadTool(context.Background(), "borg__bash")
	if err != nil {
		t.Fatalf("LoadTool failed: %v", err)
	}
	if lt.OriginalName != "bash" {
		t.Errorf("expected OriginalName=bash, got %s", lt.OriginalName)
	}

	// Check it's in loaded list
	loaded := ds.ListLoadedTools()
	found := false
	for _, l := range loaded {
		if l.Name == "borg__bash" {
			found = true
			break
		}
	}
	if !found {
		t.Error("borg__bash should be in loaded list")
	}

	// Unload
	if err := ds.UnloadTool("borg__bash"); err != nil {
		t.Fatalf("UnloadTool failed: %v", err)
	}

	loaded = ds.ListLoadedTools()
	for _, l := range loaded {
		if l.Name == "borg__bash" {
			t.Error("borg__bash should not be in loaded list after unload")
		}
	}
}

func TestDecisionSystemListAll(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	overview := ds.ListAllTools()
	if overview == nil {
		t.Fatal("ListAllTools returned nil")
	}

	if overview.TotalKnown == 0 {
		t.Error("expected some known tools")
	}

	if len(overview.AlwaysOnTools) == 0 {
		t.Error("expected always-on tools from builtins")
	}
}

func TestDecisionSystemAlwaysOnAutoLoad(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	// Always-on tools should be auto-loaded
	loaded := ds.ListLoadedTools()
	alwaysOnCount := 0
	for _, l := range loaded {
		if l.IsAlwaysOn {
			alwaysOnCount++
		}
	}

	if alwaysOnCount == 0 {
		t.Error("expected always-on tools to be auto-loaded")
	}
}

func TestDecisionSystemCatalogPersistence(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries([]ToolEntry{
		{
			Name:           "test__tool1",
			OriginalName:   "tool1",
			Server:         "test",
			AdvertisedName: "test__tool1",
			Description:    "A test tool",
			AlwaysOn:       false,
		},
	})

	tmpDir := t.TempDir()
	path := tmpDir + "/catalog.json"

	if err := ds.SaveCatalog(path); err != nil {
		t.Fatalf("SaveCatalog failed: %v", err)
	}

	ds2 := NewDecisionSystem(DefaultDecisionConfig(), nil)
	if err := ds2.LoadCatalog(path); err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	overview := ds2.ListAllTools()
	if overview.TotalKnown < 1 {
		t.Errorf("expected at least 1 tool after loading catalog, got %d", overview.TotalKnown)
	}
}

func TestDecisionSystemSearchAndCallNoMatch(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	// SearchAndCall with no aggregator should fail gracefully for ambiguous queries
	_, err := ds.SearchAndCall(context.Background(), "xyzzy_nonexistent_tool_12345", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool query")
	}
}

func TestDecisionSystemObservability(t *testing.T) {
	ds := NewDecisionSystem(DefaultDecisionConfig(), nil)
	ds.AddCatalogEntries(BuiltinTools())

	// Perform a search to generate events
	_, _ = ds.SearchTools(context.Background(), "file")

	events := ds.GetEvents(10)
	if len(events) == 0 {
		t.Error("expected observability events after search")
	}

	found := false
	for _, e := range events {
		if e.Type == "search" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'search' event type")
	}
}

func TestDecisionSystemLRUEviction(t *testing.T) {
	cfg := DefaultDecisionConfig()
	cfg.SoftCap = 3
	cfg.HardCap = 5
	ds := NewDecisionSystem(cfg, nil)

	// Add more tools than the cap
	tools := []ToolEntry{}
	for i := 0; i < 10; i++ {
		tools = append(tools, ToolEntry{
			Name:           fmt.Sprintf("srv__tool%d", i),
			OriginalName:   fmt.Sprintf("tool%d", i),
			Server:         "srv",
			AdvertisedName: fmt.Sprintf("srv__tool%d", i),
			Description:    fmt.Sprintf("Tool number %d", i),
		})
	}
	ds.AddCatalogEntries(tools)

	// Load more than hard cap
	for i := 0; i < 6; i++ {
		_, err := ds.LoadTool(context.Background(), fmt.Sprintf("srv__tool%d", i))
		if err != nil {
			t.Logf("LoadTool %d: %v", i, err)
		}
	}

	loaded := ds.ListLoadedTools()
	if len(loaded) > cfg.HardCap {
		t.Errorf("expected at most %d loaded tools, got %d", cfg.HardCap, len(loaded))
	}
}

func TestBuiltinToolsNotEmpty(t *testing.T) {
	tools := BuiltinTools()
	if len(tools) == 0 {
		t.Fatal("BuiltinTools() returned empty list")
	}

	// Check critical tools exist
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.OriginalName] = true
	}

	required := []string{"bash", "read_file", "write_file", "edit_file", "search_files", "list_directory", "find_files"}
	for _, name := range required {
		if !names[name] {
			t.Errorf("missing required builtin tool: %s", name)
		}
	}

	// Check cross-harness aliases exist
	aliases := []string{"shell", "cat", "grep", "ls", "find", "run_command", "execute_command"}
	for _, name := range aliases {
		if !names[name] {
			t.Errorf("missing cross-harness alias tool: %s", name)
		}
	}
}
