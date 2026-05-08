package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/borghq/borg-go/internal/config"
)

func TestSubmoduleUpdateAllFallsBackToNativeGitReport(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	workspace := t.TempDir()
	init := exec.Command("git", "init")
	init.Dir = workspace
	if output, err := init.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v (%s)", err, string(output))
	}

	t.Setenv("BORG_TRPC_UPSTREAM", "http://127.0.0.1:1/trpc")
	cfg := config.Default()
	cfg.WorkspaceRoot = workspace
	cfg.ConfigDir = filepath.Join(workspace, ".borg-go")
	cfg.MainConfigDir = filepath.Join(workspace, ".borg")
	server := New(cfg, stubDetector{})

	req := httptest.NewRequest(http.MethodPost, "/api/submodules/update-all", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, needle := range []string{
		`"fallback":"go-local-git-orchestration"`,
		`"procedure":"submodule.updateAll"`,
		`"total":0`,
		`"successful":0`,
		`"failed":0`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected response to contain %s, got %s", needle, body)
		}
	}
}
