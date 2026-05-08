package sessionimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildManifestReflectsCandidates(t *testing.T) {
	manifest := BuildManifest([]ValidationResult{{SourceTool: "claude-code", Valid: true}, {SourceTool: "llm-cli", Valid: false}})
	if manifest.Count != 2 {
		t.Fatalf("expected manifest count 2, got %#v", manifest)
	}
	if len(manifest.Candidates) != 2 {
		t.Fatalf("expected two manifest candidates, got %#v", manifest)
	}
	if manifest.GeneratedAt == "" {
		t.Fatalf("expected generated timestamp, got %#v", manifest)
	}
}

func TestExportSessionsWritesExportsAndManifest(t *testing.T) {
	workspace := t.TempDir()
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	sourceDir := filepath.Join(workspace, ".claude")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	sourcePath := filepath.Join(sourceDir, "session-history.json")
	sourceContent := []byte(`{"messages":[{"role":"user","content":"hello"}]}`)
	if err := os.WriteFile(sourcePath, sourceContent, 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	outputDir := filepath.Join(workspace, "custom-export")
	result, err := ExportSessions(workspace, home, outputDir)
	if err != nil {
		t.Fatalf("ExportSessions returned error: %v", err)
	}
	if result.OutputPath != outputDir {
		t.Fatalf("expected output path %q, got %#v", outputDir, result)
	}
	if result.ExportedCount < 1 {
		t.Fatalf("expected at least one exported file, got %#v", result)
	}
	if result.TotalSize != int64(len(sourceContent)) {
		t.Fatalf("expected total size %d, got %#v", len(sourceContent), result)
	}
	if len(result.Records) != 1 {
		t.Fatalf("expected one export record, got %#v", result)
	}

	record := result.Records[0]
	if record.SourceTool != "claude-code" || record.Format != "json" {
		t.Fatalf("unexpected export record: %#v", record)
	}

	exportedPath := filepath.Join(outputDir, record.SessionID+"."+record.Format)
	exportedContent, err := os.ReadFile(exportedPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}
	if string(exportedContent) != string(sourceContent) {
		t.Fatalf("expected exported content to match source, got %q", string(exportedContent))
	}

	manifestPath := filepath.Join(outputDir, "export-manifest.json")
	manifestRaw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read export manifest: %v", err)
	}
	var manifest struct {
		ExportedCount int      `json:"exportedCount"`
		Manifest      Manifest `json:"manifest"`
	}
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		t.Fatalf("failed to decode export manifest: %v", err)
	}
	if manifest.ExportedCount != 1 || manifest.Manifest.Count != 1 {
		t.Fatalf("unexpected manifest payload: %#v", manifest)
	}
}

func TestExportSessionsUsesDefaultOutputDirWhenBlank(t *testing.T) {
	workspace := t.TempDir()
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("failed to create home dir: %v", err)
	}

	result, err := ExportSessions(workspace, home, "")
	if err != nil {
		t.Fatalf("ExportSessions returned error: %v", err)
	}
	want := filepath.Join(workspace, ".borg", "exports")
	if result.OutputPath != want {
		t.Fatalf("expected default output path %q, got %#v", want, result)
	}
	if _, err := os.Stat(filepath.Join(want, "export-manifest.json")); err != nil {
		t.Fatalf("expected manifest in default output dir: %v", err)
	}
}
