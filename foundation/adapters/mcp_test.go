package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPAdapterStatusAndToolHints(t *testing.T) {
	home := t.TempDir()
	hypercodeDir := filepath.Join(home, ".hypercode")
	if err := os.MkdirAll(hypercodeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := `{"mcpServers":{"demo":{"command":"cmd","args":["/c","echo demo"],"env":{"FOO":"BAR"}}}}`
	if err := os.WriteFile(filepath.Join(hypercodeDir, "mcp.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	adapter := NewMCPAdapter(t.TempDir())
	adapter.homeDir = home
	status := adapter.Status()
	if len(status.Servers) != 1 {
		t.Fatalf("unexpected server status: %#v", status)
	}
	if status.Servers[0].Name != "demo" {
		t.Fatalf("unexpected server name: %#v", status.Servers[0])
	}
	tools, err := adapter.ListTools()
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) < 2 {
		t.Fatalf("expected MCP tool hints, got %#v", tools)
	}
	if !strings.Contains(adapter.RouteCall("demo", "list"), "demo:list") {
		t.Fatalf("unexpected route call: %s", adapter.RouteCall("demo", "list"))
	}
	call, err := adapter.CallTool(MCPCallRequest{ServerName: "demo", ToolName: "list-tools", Arguments: map[string]interface{}{"limit": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if call.ServerName != "demo" || call.ToolName != "list-tools" {
		t.Fatalf("unexpected MCP call result: %#v", call)
	}
}
