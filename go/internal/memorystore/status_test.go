package memorystore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStatusReturnsEmptyShapeWhenStoreMissing(t *testing.T) {
	tempDir := t.TempDir()

	status, err := ReadStatus(tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedPath := filepath.Join(tempDir, ".borg", "sectioned_memory.json")
	if status.Exists {
		t.Fatalf("expected missing store")
	}
	if status.StorePath != expectedPath {
		t.Fatalf("expected store path %q, got %q", expectedPath, status.StorePath)
	}
	if status.DefaultSectionCount != 5 {
		t.Fatalf("expected 5 default sections, got %d", status.DefaultSectionCount)
	}
	if len(status.MissingSections) != 5 {
		t.Fatalf("expected all default sections missing, got %+v", status.MissingSections)
	}
}

func TestReadStatusSummarizesSectionedStore(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, ".borg", "sectioned_memory.json")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		t.Fatalf("failed to create store dir: %v", err)
	}
	if err := os.WriteFile(storePath, []byte(`{"sections":[{"section":"project_context","entries":[{"createdAt":"2026-03-10T10:00:00Z"},{"createdAt":"2026-03-10T11:00:00Z"}]},{"section":"commands","entries":[{"createdAt":"2026-03-09T09:00:00Z"}]}]}`), 0o644); err != nil {
		t.Fatalf("failed to write store: %v", err)
	}

	status, err := ReadStatus(tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !status.Exists {
		t.Fatalf("expected store to exist")
	}
	if status.TotalEntries != 3 {
		t.Fatalf("expected 3 total entries, got %d", status.TotalEntries)
	}
	if status.PresentDefaultSectionCount != 2 {
		t.Fatalf("expected 2 present default sections, got %d", status.PresentDefaultSectionCount)
	}
	if status.PopulatedSectionCount != 2 {
		t.Fatalf("expected 2 populated sections, got %d", status.PopulatedSectionCount)
	}
	if status.LastUpdatedAt != "2026-03-10T11:00:00Z" {
		t.Fatalf("expected latest timestamp, got %q", status.LastUpdatedAt)
	}
}

func TestReadStatusFallsBackToLegacyStore(t *testing.T) {
	tempDir := t.TempDir()
	legacyPath := filepath.Join(tempDir, ".borg", "claude_mem.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("failed to create legacy dir: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte(`{"sections":[{"section":"general","entries":[{"createdAt":"2026-03-10T11:00:00Z"}]}]}`), 0o644); err != nil {
		t.Fatalf("failed to write legacy store: %v", err)
	}

	status, err := ReadStatus(tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !status.Exists {
		t.Fatalf("expected legacy store to be read")
	}
	if status.SectionCount != 1 {
		t.Fatalf("expected 1 section, got %d", status.SectionCount)
	}
	if status.StorePath != filepath.Join(tempDir, ".borg", "sectioned_memory.json") {
		t.Fatalf("expected canonical store path, got %q", status.StorePath)
	}
}
