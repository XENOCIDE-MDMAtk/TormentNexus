package interop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"testing"

	"github.com/borghq/borg-go/internal/lockfile"
	"github.com/borghq/borg-go/internal/lockfile"
)

func TestResolveTRPCBasesPrefersLockedAndConfiguredBases(t *testing.T) {
	tempDir := t.TempDir()
	mainLockPath := filepath.Join(tempDir, "lock")
	if err := lockfile.Write(mainLockPath, lockfile.Record{
		Host: "0.0.0.0",
		Port: 4100,
	}); err != nil {
		t.Fatalf("failed to seed lock file: %v", err)
	}

	t.Setenv("BORG_TRPC_UPSTREAM", "http://127.0.0.1:4200/trpc")

	bases := ResolveTRPCBases(mainLockPath)
	if len(bases) < 2 {
		t.Fatalf("expected locked and configured bases, got %+v", bases)
	}
	if bases[0] != "http://127.0.0.1:4100/trpc" {
		t.Fatalf("expected locked base first, got %+v", bases)
	}
	if !slices.Contains(bases, "http://127.0.0.1:4200/trpc") {
		t.Fatalf("expected configured base in list, got %+v", bases)
	}
}

func TestCallTRPCProcedureReturnsUnwrappedJSONData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trpc/session.list" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"data": map[string]any{
					"json": []map[string]any{
						{"id": "sess_1", "status": "running"},
					},
				},
			},
		})
	}))
	defer server.Close()

	t.Setenv("BORG_TRPC_UPSTREAM", server.URL+"/trpc")

	result, err := CallTRPCProcedure(context.Background(), filepath.Join(t.TempDir(), "missing-lock"), "session.list", nil)
	if err != nil {
		t.Fatalf("expected no bridge error, got %v", err)
	}
	if result.BaseURL != server.URL+"/trpc" {
		t.Fatalf("expected test server base url, got %q", result.BaseURL)
	}
	if string(result.Data) != `[{"id":"sess_1","status":"running"}]` {
		t.Fatalf("expected unwrapped tRPC payload, got %s", string(result.Data))
	}
}
