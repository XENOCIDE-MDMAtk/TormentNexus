package sessionimport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerDiscoversExpandedImportSources(t *testing.T) {
	workspaceRoot := t.TempDir()
	homeDir := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	mustWriteFile := func(path string, content []byte) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create parent dirs for %s: %v", path, err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	mustWriteFile(filepath.Join(workspaceRoot, ".llm", "logs.db"), []byte("SQLite format 3\x00fake llm db"))
	mustWriteFile(filepath.Join(homeDir, ".prism-mcp", "data.db"), []byte("SQLite format 3\x00fake prism db"))
	mustWriteFile(filepath.Join(homeDir, ".gemini", "antigravity", "brain", "session-history.jsonl"), []byte("{\"model\":\"gemini\"}\n"))
	mustWriteFile(filepath.Join(homeDir, "AppData", "Roaming", "Code", "User", "globalStorage", "emptyWindowChatSessions", "session.json"), []byte("{\"tool\":\"copilot\"}"))

	scanner := NewScanner(workspaceRoot, homeDir, 20)
	candidates, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	byTool := map[string][]Candidate{}
	for _, candidate := range candidates {
		byTool[candidate.SourceTool] = append(byTool[candidate.SourceTool], candidate)
	}

	if len(byTool["llm-cli"]) == 0 {
		t.Fatalf("expected llm-cli candidate, got %+v", candidates)
	}
	if byTool["llm-cli"][0].SessionFormat != "db" {
		t.Fatalf("expected llm-cli db format, got %+v", byTool["llm-cli"][0])
	}
	if len(byTool["prism-mcp"]) == 0 {
		t.Fatalf("expected prism-mcp candidate, got %+v", candidates)
	}
	if len(byTool["antigravity"]) == 0 {
		t.Fatalf("expected antigravity candidate, got %+v", candidates)
	}
	if len(byTool["vscode-extensions"]) == 0 {
		t.Fatalf("expected vscode-extensions candidate, got %+v", candidates)
	}
}

func TestValidateCandidateAcceptsDatabaseArtifacts(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "logs.db")
	if err := os.WriteFile(dbPath, []byte("SQLite format 3\x00fake"), 0o644); err != nil {
		t.Fatalf("failed to seed db candidate: %v", err)
	}

	result := ValidateCandidate(Candidate{
		SourceTool:    "llm-cli",
		SourcePath:    dbPath,
		SessionFormat: "db",
		EstimatedSize: 20,
	})

	if !result.Valid {
		t.Fatalf("expected db candidate to validate, got %+v", result)
	}
	if result.SourceType != "database-log" {
		t.Fatalf("expected database-log source type, got %+v", result)
	}
}
