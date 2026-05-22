package config

import (
	"os"
	"path/filepath"
)

type PathStatus struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type Status struct {
	Host                 string     `json:"host"`
	Port                 int        `json:"port"`
	BaseURL              string     `json:"baseUrl"`
	WorkspaceRoot        PathStatus `json:"workspaceRoot"`
	ConfigDir            PathStatus `json:"configDir"`
	MainConfigDir        PathStatus `json:"mainConfigDir"`
	HypercodeConfigFile       PathStatus `json:"hypercodeConfigFile"`
	MCPConfigFile        PathStatus `json:"mcpConfigFile"`
	GoLockPath           PathStatus `json:"goLockPath"`
	MainLockPath         PathStatus `json:"mainLockPath"`
	ImportedInstructions PathStatus `json:"importedInstructions"`
	SectionedMemoryStore PathStatus `json:"sectionedMemoryStore"`
	LegacyMemoryStore    PathStatus `json:"legacyMemoryStore"`
	HypercodeSubmodule        PathStatus `json:"hypercodeSubmodule"`
}

func Snapshot(cfg Config) Status {
	return Status{
		Host:                 cfg.Host,
		Port:                 cfg.Port,
		BaseURL:              cfg.BaseURL(),
		WorkspaceRoot:        buildPathStatus(cfg.WorkspaceRoot),
		ConfigDir:            buildPathStatus(cfg.ConfigDir),
		MainConfigDir:        buildPathStatus(cfg.MainConfigDir),
		HypercodeConfigFile:       buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "hypercode.config.json")),
		MCPConfigFile:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "mcp.jsonc")),
		GoLockPath:           buildPathStatus(cfg.LockPath()),
		MainLockPath:         buildPathStatus(cfg.MainLockPath()),
		ImportedInstructions: buildPathStatus(cfg.ImportedInstructionsPath()),
		SectionedMemoryStore: buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".hypercode", "sectioned_memory.json")),
		LegacyMemoryStore:    buildPathStatus(filepath.Join(cfg.WorkspaceRoot, ".hypercode", "claude_mem.json")),
		HypercodeSubmodule:        buildPathStatus(filepath.Join(cfg.WorkspaceRoot, "submodules", "hypercode")),
	}
}

func buildPathStatus(path string) PathStatus {
	_, err := os.Stat(path)
	return PathStatus{
		Path:   path,
		Exists: err == nil,
	}
}
