package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HandleSearch searches a codebase using semble and returns relevant code snippets.
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	path, _ :=getString(args, "path")
	topK, _ :=getInt(args, "top_k")
	content, _ :=getString(args, "content")

	if query == "" {
		return err("query is required")
}

	// Build semble search command
	cmdArgs := []string{"search", query}
	if path != "" {
		cmdArgs = append(cmdArgs, path)

	if topK > 0 {
		cmdArgs = append(cmdArgs, "--top-k", strconv.Itoa(topK))

	if content != "" {
		cmdArgs = append(cmdArgs, "--content", content)

	cmd := exec.CommandContext(ctx, "semble", cmdArgs...)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble search failed: %s: %s", cmdErr.Error(), string(out)))
}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return ok("No results found for query: " + query)
}

	return ok(result)
}

}
}
}

// HandleFindRelated finds code related to a specific file and line number.
func HandleFindRelated(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	lineNum, _ :=getInt(args, "line_number")
	path, _ :=getString(args, "path")

	if filePath == "" {
		return err("file_path is required")
}

	if lineNum <= 0 {
		return err("line_number must be a positive integer")
}

	cmdArgs := []string{"find-related", filePath, strconv.Itoa(lineNum)}
	if path != "" {
		cmdArgs = append(cmdArgs, path)

	cmd := exec.CommandContext(ctx, "semble", cmdArgs...)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble find-related failed: %s: %s", cmdErr.Error(), string(out)))
}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return ok("No related code found for " + filePath + ":" + strconv.Itoa(lineNum))
}

	return ok(result)
}

}

// HandleSavings shows token savings statistics from semble usage.
func HandleSavings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "semble", "savings")
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble savings failed: %s: %s", cmdErr.Error(), string(out)))
}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return ok("No savings data available yet.")
}

	return ok(result)
}

// HandleIndexInfo returns information about the semble index for a given path.
func HandleIndexInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")

	if path == "" {
		// Default to current directory
		path = "."
	}

	// Resolve to absolute path for display
	absPath, pathErr := filepath.Abs(path)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to resolve path: %s", pathErr.Error()))
}

	// Check if semble is available
	whichCmd := exec.CommandContext(ctx, "which", "semble")
	_, whichErr := whichCmd.CombinedOutput()
	if whichErr != nil {
		// Try uvx fallback
		whichCmd = exec.CommandContext(ctx, "which", "uvx")
		_, uvxErr := whichCmd.CombinedOutput()
		if uvxErr != nil {
			return err("semble is not installed. Install with: uv tool install semble")

	}

	// Check if the path exists
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		return err(fmt.Sprintf("path does not exist: %s", absPath))
}

	if !info.IsDir() {
		return err(fmt.Sprintf("path is not a directory: %s", absPath))
}

	// Build info about the index
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Index path: %s\n", absPath))
	sb.WriteString(fmt.Sprintf("Content type: %s\n", contentOrDefault(content)))

	// Check for .sembleignore
	ignorePath := filepath.Join(absPath, ".sembleignore")
	if _, ignoreErr := os.Stat(ignorePath); ignoreErr == nil {
		sb.WriteString("Custom ignore file: .sembleignore found\n")

	// Check for .gitignore
	gitignorePath := filepath.Join(absPath, ".gitignore")
	if _, gitErr := os.Stat(gitignorePath); gitErr == nil {
		sb.WriteString("Git ignore file: .gitignore found\n")

	// Check cache location
	cacheDir := os.Getenv("SEMBLE_CACHE_LOCATION")
	if cacheDir == "" {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			cacheDir = filepath.Join(homeDir, ".cache", "semble")

	}
	if cacheDir != "" {
		sb.WriteString(fmt.Sprintf("Cache location: %s\n", cacheDir))

	return ok(sb.String())
}

}
}
}
}
}

// HandleSearchRemote searches a remote git repository using semble.
func HandleSearchRemote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	gitURL, _ :=getString(args, "git_url")
	topK, _ :=getInt(args, "top_k")
	content, _ :=getString(args, "content")

	if query == "" {
		return err("query is required")
}

	if gitURL == "" {
		return err("git_url is required")
}

	// Validate it looks like a git URL
	if !strings.HasPrefix(gitURL, "https://") && !strings.HasPrefix(gitURL, "git@") && !strings.HasPrefix(gitURL, "http://") {
		return err("git_url must be a valid git URL (https://..., git@..., or http://...)")
}

	// Build semble search command with git URL
	cmdArgs := []string{"search", query, gitURL}
	if topK > 0 {
		cmdArgs = append(cmdArgs, "--top-k", strconv.Itoa(topK))

	if content != "" {
		cmdArgs = append(cmdArgs, "--content", content)

	// Use a longer timeout for remote repos since they need cloning
	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "semble", cmdArgs...)
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble remote search failed: %s: %s", cmdErr.Error(), string(out)))
}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return ok("No results found for query: " + query)
}

	return ok(result)
}

}
}

// contentOrDefault returns the content type or "code" as default.
func contentOrDefault(content string) string {
	if content == "" {
		return "code"
	}
	return content
}