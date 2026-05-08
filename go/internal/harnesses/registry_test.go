package harnesses

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/borghq/borg-go/internal/controlplane"
)

func TestListBuildsHarnessDefinitions(t *testing.T) {
	workspaceRoot := t.TempDir()
	borgPath := filepath.Join(workspaceRoot, "submodules", "borg")
	if err := os.MkdirAll(borgPath, 0o755); err != nil {
		t.Fatalf("failed to create borg path: %v", err)
	}
	toolsDir := filepath.Join(borgPath, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create borg tools path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(`
package tools

func demo() {
	_ = Tool{Name: "run_shell_command"}
	_ = Tool{Name: "read_file"}
}
`), 0o644); err != nil {
		t.Fatalf("failed to seed borg tool registry: %v", err)
	}

	definitions := List(workspaceRoot, []controlplane.Tool{
		{Type: "codex", Available: true},
		{Type: "opencode", Available: true},
		{Type: "claude-code", Available: true},
		{Type: "copilot", Available: true},
	})

	if len(definitions) != 49 {
		t.Fatalf("expected 49 harness definitions, got %d", len(definitions))
	}
	if !definitions[0].Primary || !definitions[0].Installed {
		t.Fatalf("expected borg to be primary and installed, got %+v", definitions[0])
	}
	if definitions[0].ToolCallCount != 2 {
		t.Fatalf("expected 2 borg tool calls, got %+v", definitions[0])
	}
	if definitions[0].ToolInventoryStatus != "source-backed" || definitions[0].IntegrationLevel != "source-backed" {
		t.Fatalf("expected borg to be source-backed, got %+v", definitions[0])
	}
	installed := map[string]bool{}
	for _, definition := range definitions {
		installed[definition.ID] = definition.Installed
	}
	if !installed["opencode"] {
		t.Fatalf("expected opencode to be installed")
	}
	if !installed["claude-code"] {
		t.Fatalf("expected claude-code harness to be installed")
	}
	if !installed["copilot"] {
		t.Fatalf("expected copilot harness to be installed")
	}

	summary := Summarize(definitions)
	if summary.SourceBackedHarnessCount != 1 || summary.SourceBackedToolCount != 2 {
		t.Fatalf("expected one source-backed harness with two tools, got %+v", summary)
	}
	if summary.MetadataOnlyHarnessCount != 47 {
		t.Fatalf("expected forty-seven metadata-only harnesses, got %+v", summary)
	}
	if summary.OperatorDefinedHarnessCount != 1 {
		t.Fatalf("expected one operator-defined harness, got %+v", summary)
	}
}
package harnesses

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/borghq/borg-go/internal/controlplane"
)

func TestListBuildsHarnessDefinitions(t *testing.T) {
	workspaceRoot := t.TempDir()
	borgPath := filepath.Join(workspaceRoot, "submodules", "borg")
	if err := os.MkdirAll(borgPath, 0o755); err != nil {
		t.Fatalf("failed to create borg path: %v", err)
	}
	toolsDir := filepath.Join(borgPath, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed to create borg tools path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "registry.go"), []byte(`
package tools

func demo() {
	_ = Tool{Name: "run_shell_command"}
	_ = Tool{Name: "read_file"}
}
`), 0o644); err != nil {
		t.Fatalf("failed to seed borg tool registry: %v", err)
	}

	definitions := List(workspaceRoot, []controlplane.Tool{
		{Type: "codex", Available: true},
		{Type: "opencode", Available: true},
		{Type: "claude-code", Available: true},
		{Type: "copilot", Available: true},
	})

	if len(definitions) != 49 {
		t.Fatalf("expected 49 harness definitions, got %d", len(definitions))
	}
	if !definitions[0].Primary || !definitions[0].Installed {
		t.Fatalf("expected borg to be primary and installed, got %+v", definitions[0])
	}
	if definitions[0].ToolCallCount != 2 {
		t.Fatalf("expected 2 borg tool calls, got %+v", definitions[0])
	}
	if definitions[0].ToolInventoryStatus != "source-backed" || definitions[0].IntegrationLevel != "source-backed" {
		t.Fatalf("expected borg to be source-backed, got %+v", definitions[0])
	}
	installed := map[string]bool{}
	for _, definition := range definitions {
		installed[definition.ID] = definition.Installed
	}
	if !installed["opencode"] {
		t.Fatalf("expected opencode to be installed")
	}
	if !installed["claude-code"] {
		t.Fatalf("expected claude-code harness to be installed")
	}
	if !installed["copilot"] {
		t.Fatalf("expected copilot harness to be installed")
	}

	summary := Summarize(definitions)
	if summary.SourceBackedHarnessCount != 1 || summary.SourceBackedToolCount != 2 {
		t.Fatalf("expected one source-backed harness with two tools, got %+v", summary)
	}
	if summary.MetadataOnlyHarnessCount != 47 {
		t.Fatalf("expected forty-seven metadata-only harnesses, got %+v", summary)
	}
	if summary.OperatorDefinedHarnessCount != 1 {
		t.Fatalf("expected one operator-defined harness, got %+v", summary)
	}
}
