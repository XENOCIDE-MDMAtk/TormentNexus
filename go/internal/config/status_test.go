package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotReportsKeyPaths(t *testing.T) {
	workspaceRoot := t.TempDir()
	configDir := filepath.Join(workspaceRoot, ".borg-go")
	mainConfigDir := filepath.Join(workspaceRoot, ".borg")
	if err := os.MkdirAll(filepath.Join(workspaceRoot, "submodules", "borg"), 0o755); err != nil {
		t.Fatalf("failed to create borg path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "borg.config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create borg config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "mcp.jsonc"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create mcp config: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(mainConfigDir, 0o755); err != nil {
		t.Fatalf("failed to create main config dir: %v", err)
	}

	status := Snapshot(Config{
		Host:          "127.0.0.1",
		Port:          4300,
		ConfigDir:     configDir,
		MainConfigDir: mainConfigDir,
		WorkspaceRoot: workspaceRoot,
	})

	if !status.WorkspaceRoot.Exists {
		t.Fatalf("expected workspace root to exist")
	}
	if !status.BorgSubmodule.Exists {
		t.Fatalf("expected borg submodule path to exist")
	}
	if !status.BorgConfigFile.Exists || !status.MCPConfigFile.Exists {
		t.Fatalf("expected repo config files to exist, got borg=%+v mcp=%+v", status.BorgConfigFile, status.MCPConfigFile)
	}
	if status.SectionedMemoryStore.Exists {
		t.Fatalf("expected sectioned memory store to be absent by default")
	}
}
