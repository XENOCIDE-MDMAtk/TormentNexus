package lockfile

import (
	"path/filepath"
	"testing"
)

func TestWriteAndRead(t *testing.T) {
	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, "lock")
	expected := Record{
		Host:      "127.0.0.1",
		Port:      4300,
		Version:   "0.0.1-experimental",
		StartedAt: "2026-03-28T00:00:00Z",
	}

	if err := Write(lockPath, expected); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	actual, err := Read(lockPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if actual != expected {
		t.Fatalf("expected %+v, got %+v", expected, actual)
	}
}
