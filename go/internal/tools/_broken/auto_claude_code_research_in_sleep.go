package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HandleInstall clones a git repository into a destination directory.
// Required arguments:
//   - repo_url: string, the HTTPS (or SSH) URL of the repository to clone.
//   - dest: string, the directory where the repository should be placed.
func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo_url")
	dest, _ :=getString(args, "dest")
	if repo == "" || dest == "" {
		return err("both 'repo_url' and 'dest' arguments must be provided")
}

	// Ensure the destination does not already exist.
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		return err(fmt.Sprintf("destination path already exists: %s", dest))
}

	cmd := exec.CommandContext(ctx, "git", "clone", repo, dest)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("git clone failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Successfully cloned %s into %s", repo, dest))
}

// HandleUninstall removes a previously installed ARIS repository directory.
// Required argument:
//   - dest: string, the directory to delete.
func HandleUninstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dest, _ :=getString(args, "dest")
	if dest == "" {
		return err("argument 'dest' must be provided")
}

	// Safety check: do not allow empty path which would delete everything.
	if strings.TrimSpace(dest) == "" {
		return err("invalid destination path")
}

	removeErr := os.RemoveAll(dest)
	if removeErr != nil {
		return err(fmt.Sprintf("failed to remove %s: %v", dest, removeErr))
}

	return ok(fmt.Sprintf("Removed directory %s", dest))
}

// HandleListSkills enumerates skill directories inside a project.
// Required argument:
//   - project_path: string, path to the root of the project (contains a 'skills' folder).
func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	if projectPath == "" {
		return err("argument 'project_path' must be provided")
}

	skillsPath := filepath.Join(projectPath, "skills")
	entries, readErr := os.ReadDir(skillsPath)
	if readErr != nil {
		return err(fmt.Sprintf("cannot read skills directory %s: %v", skillsPath, readErr))
}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())

	}
	if len(names) == 0 {
		return ok("No skill directories found.")
}

	return ok(strings.Join(names, "\n"))
}

}

// HandleStatus reports basic information about the ARIS installation.
// Required argument:
//   - project_path: string, path to the root of the project.
func HandleStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "project_path")
	if projectPath == "" {
		return err("argument 'project_path' must be provided")
}

	// Check for presence of key ARIS files/directories.
	paths := []struct {
		name string
		path string
	}{
		{"CLAUDE.md", filepath.Join(projectPath, "CLAUDE.md")},
		{".claude/skills", filepath.Join(projectPath, ".claude", "skills")},
		{".aris", filepath.Join(projectPath, ".aris")},
		{"research-wiki", filepath.Join(projectPath, "research-wiki")},
	}
	var sb strings.Builder
	sb.WriteString("ARIS installation status:\n")
	for _, p := range paths {
		if _, statErr := os.Stat(p.path); os.IsNotExist(statErr) {
			sb.WriteString(fmt.Sprintf("- %s: MISSING (%s)\n", p.name, p.path))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: PRESENT (%s)\n", p.name, p.path))

	}
	return ok(sb.String())
}
}