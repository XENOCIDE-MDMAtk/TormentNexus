package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveClientTargets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "borg-mcp-sync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cwd, _ := os.Getwd()
	appData := filepath.Join(tempDir, "AppData", "Roaming")
	_ = os.MkdirAll(appData, 0755)

	// Create a dummy VSCode config
	// On Windows ResolveClientTargets uses appData candidate
	vscodePath := filepath.Join(appData, "Code", "User", "settings.json")
	_ = os.MkdirAll(filepath.Dir(vscodePath), 0755)
	_ = os.WriteFile(vscodePath, []byte(`{"mcp.servers": {}}`), 0644)

	// Inject custom appData (as if Roaming)
	targets := ResolveClientTargets(tempDir, appData, cwd)

	if len(targets) != 3 {
		t.Errorf("expected 3 client targets, got %d", len(targets))
	}

	foundVSCode := false
	for _, target := range targets {
		if target.Client == VSCode && target.Exists {
			foundVSCode = true
		}
	}

	if !foundVSCode {
		t.Errorf("expected VSCode target to be found in %s. Targets: %+v", vscodePath, targets)
	}
}

func TestSyncToClient(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "borg-mcp-sync-write-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "claude_desktop_config.json")
	servers := map[string]McpServerConfig{
		"test-server": {
			Command: "echo",
			Args:    []string{"hello"},
		},
	}

	result, err := SyncToClient(ClaudeDesktop, targetPath, servers)
	if err != nil {
		t.Fatalf("SyncToClient failed: %v", err)
	}

	if !result.Written || result.ServerCount != 1 {
		t.Errorf("unexpected sync result: %+v", result)
	}

	// Read it back
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if !strings.Contains(string(data), "test-server") || !strings.Contains(string(data), "mcpServers") {
		t.Errorf("written config missing test-server or mcpServers key: %s", string(data))
	}
}
