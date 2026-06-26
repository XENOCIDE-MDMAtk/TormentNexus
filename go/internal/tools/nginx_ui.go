package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// runCommand executes a shell command and returns combined output or error.
func runCommand(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return string(output), fmt.Errorf("command failed: %w, output: %s", execErr, string(output))
}

	return string(output), nil
}

// HandleTestConfig tests the Nginx configuration syntax.
func HandleTestConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	output, execErr := runCommand("nginx", "-t")
	if execErr != nil {
		return err(fmt.Sprintf("Nginx configuration test failed: %s", output))
}

	return ok("Nginx configuration is valid.\n" + output)
}

// HandleReload reloads Nginx to apply configuration changes.
func HandleReload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_, execErr := runCommand("nginx", "-s", "reload")
	if execErr != nil {
		return err(fmt.Sprintf("Failed to reload Nginx: %v", execErr))
}

	return ok("Nginx reloaded successfully.")
}

// HandleReadConfig reads the content of a configuration file.
func HandleReadConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	
	content, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read file: %v", readErr))
}

	return ok(string(content))
}

// HandleWriteConfig writes content to a configuration file.
func HandleWriteConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")

	if path == "" {
		return err("path is required")
}

	if content == "" {
		return err("content is required")
}

	cleanPath := filepath.Clean(path)
	
	writeErr := os.WriteFile(cleanPath, []byte(content), 0644)
	if writeErr != nil {
		return err(fmt.Sprintf("failed to write file: %v", writeErr))
}

	return ok(fmt.Sprintf("File written successfully to %s", cleanPath))
}

// HandleGetVersion retrieves the Nginx version.
func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	output, execErr := runCommand("nginx", "-v")
	// nginx -v outputs to stderr, runCommand captures CombinedOutput
	if execErr != nil {
		// nginx -v returns exit code 0 but sometimes exec behaves weirdly if stderr is used heavily
		// However, runCommand checks exit code. If exit code is non-zero, it errors.
		// Usually nginx -v is fine.
		return err(fmt.Sprintf("Failed to get Nginx version: %v", execErr))
}

	return ok(strings.TrimSpace(output))
}