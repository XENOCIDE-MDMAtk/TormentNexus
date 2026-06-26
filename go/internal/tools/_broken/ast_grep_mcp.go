package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var http.DefaultClient = http.Client{Timeout: 30 * time.Second}

func HandleAstGrepSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return ok(""), err("pattern is required")
}

	rootDir, _ :=getString(args, "root_dir")
	if rootDir == "" {
		return ok(""), err("root_dir is required")
}

	_, statErr := os.Stat(rootDir)
	if statErr != nil {
		return ok(""), err(fmt.Sprintf("root_dir does not exist: %v", statErr))
}

	cmd := exec.CommandContext(ctx, "ast-grep", "-p", pattern, rootDir)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return ok(""), err(fmt.Sprintf("ast-grep failed: %v\nOutput: %s", runErr, string(output)))
}

	return ok(string(output))
}

func HandleAstGrepReplace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return ok(""), err("pattern is required")
}

	replacement, _ :=getString(args, "replacement")
	if replacement == "" {
		return ok(""), err("replacement is required")
}

	rootDir, _ :=getString(args, "root_dir")
	if rootDir == "" {
		return ok(""), err("root_dir is required")
}

	_, statErr := os.Stat(rootDir)
	if statErr != nil {
		return ok(""), err(fmt.Sprintf("root_dir does not exist: %v", statErr))
}

	cmd := exec.CommandContext(ctx, "ast-grep", "-p", pattern, "-r", replacement, rootDir)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return ok(""), err(fmt.Sprintf("ast-grep replace failed: %v\nOutput: %s", runErr, string(output)))
}

	return ok(string(output))
}

func HandleAstGrepVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "ast-grep", "--version")
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return ok(""), err(fmt.Sprintf("failed to get ast-grep version: %v", runErr))
}

	return ok(strings.TrimSpace(string(output)), nil)
}

func HandleAstGrepInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	installPath, _ :=getString(args, "install_path")
	if installPath == "" {
		installPath = "/usr/local/bin"
	}

	_, e := exec.LookPath("ast-grep")
	if e == nil {
		return ok("ast-grep is already installed")
	}

	downloadURL := "https://github.com/ast-grep/ast-grep/releases/latest/download/ast-grep"
	resp, fetchErr := http.DefaultClient.Get(downloadURL)
	if fetchErr != nil {
		return ok(""), err(fmt.Sprintf("failed to download ast-grep: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ok(""), err(fmt.Sprintf("failed to download ast-grep: HTTP %d", resp.StatusCode))
}

	exePath := filepath.Join(installPath, "ast-grep")
	exeFile, createErr := os.Create(exePath)
	if createErr != nil {
		return ok(""), err(fmt.Sprintf("failed to create executable: %v", createErr))
}

	defer exeFile.Close()

	_, copyErr := io.Copy(exeFile, resp.Body)
	if copyErr != nil {
		return ok(""), err(fmt.Sprintf("failed to write executable: %v", copyErr))
}

	chmodErr := os.Chmod(exePath, 0755)
	if chmodErr != nil {
		return ok(""), err(fmt.Sprintf("failed to make executable: %v", chmodErr))
}

	return ok("ast-grep installed successfully")
}

func HandleAstGrepListPatterns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	patterns := []string{
		"console.log($x)",
		"import $x from '$y'",
		"function $x($y) { $z }",
		"class $x { $y }",
		"const [$x, $y] = $z",
	}

	return ok(strings.Join(patterns, "\n"))
}

func HandleAstGrepValidatePattern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ :=getString(args, "pattern")
	if pattern == "" {
		return ok(""), err("pattern is required")
}

	if !strings.Contains(pattern, "$") {
		return ok(""), err("pattern must contain at least one placeholder ($x)")
}

	return ok("pattern is valid")
}