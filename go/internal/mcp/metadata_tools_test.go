package mcp

import "testing"

func TestMetadataToolsNormalizeAndConvertToToolEntries(t *testing.T) {
	raw := []any{
		map[string]any{"name": "search_tools", "description": "Search tools", "inputSchema": map[string]any{"type": "object"}, "alwaysOn": true},
		map[string]any{"name": "", "description": "ignored"},
	}
	tools := MetadataToolsFromAny(raw)
	if len(tools) != 1 {
		t.Fatalf("expected 1 normalized metadata tool, got %#v", tools)
	}
	if !tools[0].AlwaysOn || tools[0].Name != "search_tools" {
		t.Fatalf("unexpected normalized metadata tool: %#v", tools[0])
	}

	encoded := MetadataToolsToAny(tools)
	if len(encoded) != 1 {
		t.Fatalf("expected one encoded metadata tool, got %#v", encoded)
	}
	encodedTool, _ := encoded[0].(map[string]any)
	if encodedTool["alwaysOn"] != true {
		t.Fatalf("expected alwaysOn to survive encoding, got %#v", encodedTool)
	}

	entry := ToolEntryFromMetadata("core", tools[0])
	if entry.Name != "core__search_tools" || entry.OriginalName != "search_tools" || !entry.AlwaysOn {
		t.Fatalf("unexpected tool entry from metadata: %#v", entry)
	}
}
