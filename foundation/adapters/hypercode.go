package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robertpelloni/hypercode/hypercode"
)

type HypercodeStatus struct {
	Assimilated       bool           `json:"assimilated"`
	HypercodeCoreURL       string         `json:"hypercodeCoreUrl,omitempty"`
	MemoryContext     string         `json:"memoryContext,omitempty"`
	Provider          ProviderStatus `json:"provider"`
	MCPServerNames    []string       `json:"mcpServerNames,omitempty"`
	MCPConfigPath     string         `json:"mcpConfigPath,omitempty"`
	HypercodeRepoPath string         `json:"hypercodeRepoPath,omitempty"`
	Warnings          []string       `json:"warnings,omitempty"`
}

type HypercodeAdapter struct {
	hypercodeAdapter *hypercode.Adapter
	workingDir  string
	homeDir     string
}

func NewHypercodeAdapter(workingDir string) *HypercodeAdapter {
	homeDir, _ := os.UserHomeDir()
	return &HypercodeAdapter{
		hypercodeAdapter: hypercode.NewAdapter(),
		workingDir:  workingDir,
		homeDir:     homeDir,
	}
}

func (a *HypercodeAdapter) Status() HypercodeStatus {
	status := HypercodeStatus{
		Assimilated:   a.hypercodeAdapter != nil && a.hypercodeAdapter.Assimilated,
		MemoryContext: a.MemoryContext(),
		Provider:      BuildProviderStatus(),
	}
	if a.hypercodeAdapter != nil {
		status.HypercodeCoreURL = a.hypercodeAdapter.HypercodeCoreURL
	}
	if repoPath, ok := a.findHypercodeRepo(); ok {
		status.HypercodeRepoPath = repoPath
	} else {
		status.Warnings = append(status.Warnings, "adjacent hypercode repo not found")
	}
	if configPath, names, err := a.listMCPServers(); err == nil {
		status.MCPConfigPath = configPath
		status.MCPServerNames = names
	} else {
		status.Warnings = append(status.Warnings, err.Error())
	}
	return status
}

func (a *HypercodeAdapter) MemoryContext() string {
	if a.hypercodeAdapter == nil {
		return ""
	}
	return a.hypercodeAdapter.GetMemoryContext()
}

func (a *HypercodeAdapter) RouteMCP(request string) string {
	if a.hypercodeAdapter == nil {
		return request
	}
	return a.hypercodeAdapter.RouteMCP(request)
}

func (a *HypercodeAdapter) BuildSystemContext() string {
	status := a.Status()
	parts := []string{
		"[Hypercode Adapter]",
		fmt.Sprintf("Assimilated: %t", status.Assimilated),
	}
	if status.HypercodeCoreURL != "" {
		parts = append(parts, fmt.Sprintf("Hypercode Core URL: %s", status.HypercodeCoreURL))
	}
	if status.MemoryContext != "" {
		parts = append(parts, status.MemoryContext)
	}
	if len(status.Provider.Available) > 0 {
		parts = append(parts, BuildProviderContext())
	}
	if len(status.MCPServerNames) > 0 {
		parts = append(parts, fmt.Sprintf("Configured MCP servers: %s", strings.Join(status.MCPServerNames, ", ")))
	}
	if status.HypercodeRepoPath != "" {
		parts = append(parts, fmt.Sprintf("Hypercode repo: %s", status.HypercodeRepoPath))
	}
	if len(status.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("Warnings: %s", strings.Join(status.Warnings, "; ")))
	}
	return strings.Join(parts, "\n")
}

func (a *HypercodeAdapter) listMCPServers() (string, []string, error) {
	configPath, config, err := ParseMCPConfig(a.homeDir)
	if err != nil {
		return configPath, nil, fmt.Errorf("mcp config unavailable: %w", err)
	}
	names := make([]string, 0, len(config.MCPServers))
	for name := range config.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)
	return configPath, names, nil
}

func (a *HypercodeAdapter) findHypercodeRepo() (string, bool) {
	candidates := []string{}
	if a.workingDir != "" {
		candidates = append(candidates,
			filepath.Join(a.workingDir, "..", "hypercode"),
			filepath.Join(a.workingDir, "../hypercode"),
		)
	}
	if a.homeDir != "" {
		candidates = append(candidates, filepath.Join(a.homeDir, "workspace", "hypercode"))
	}
	for _, candidate := range candidates {
		clean := filepath.Clean(candidate)
		if stat, err := os.Stat(filepath.Join(clean, "README.md")); err == nil && !stat.IsDir() {
			return clean, true
		}
	}
	return "", false
}
