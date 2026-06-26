package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HandleSearch searches a codebase using semble for fast, token-efficient code retrieval.
// It supports local paths and git URLs, with configurable content type and result count.
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	path, _ :=getString(args, "path")
	if path == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return err(fmt.Sprintf("failed to get current directory: %s", cwdErr.Error()))
}

		path = cwd
	}

	content, _ :=getString(args, "content")
	if content == "" {
		content = "code"
	}

	topKStr, _ :=getString(args, "top_k")
	topK := 10
	if topKStr != "" {
		parsed, parseErr := strconv.Atoi(topKStr)
		if parseErr == nil && parsed > 0 {
			topK = parsed
		}
	}

	// Check if semble is available
	semblePath, lookErr := exec.LookPath("semble")
	if lookErr != nil {
		return err("semble is not installed. Install it with: uv tool install semble")
}

	// Build command arguments
	cmdArgs := []string{"search", query, path, "--top-k", strconv.Itoa(topK), "--content", content}

	cmd := exec.CommandContext(ctx, semblePath, cmdArgs...)
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble search failed: %s\n%s", cmdErr.Error(), string(output)))
}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return ok("No results found for query: " + query)
}

	return ok(result)
}

// HandleFindRelated finds code snippets related to a specific file and line number.
// Useful for tracing dependencies, usages, and similar patterns.
func HandleFindRelated(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("file_path is required")
}

	lineNum, _ :=getInt(args, "line_number")
	if lineNum <= 0 {
		return err("line_number must be a positive integer")
}

	path, _ :=getString(args, "path")
	if path == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return err(fmt.Sprintf("failed to get current directory: %s", cwdErr.Error()))
}

		path = cwd
	}

	semblePath, lookErr := exec.LookPath("semble")
	if lookErr != nil {
		return err("semble is not installed. Install it with: uv tool install semble")
}

	cmdArgs := []string{"find-related", filePath, strconv.Itoa(lineNum), path}

	cmd := exec.CommandContext(ctx, semblePath, cmdArgs...)
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble find-related failed: %s\n%s", cmdErr.Error(), string(output)))
}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return ok(fmt.Sprintf("No related code found for %s:%d", filePath, lineNum))
}

	return ok(result)
}

// HandleSavings displays token savings statistics from semble usage.
// Shows how many tokens have been saved compared to grep+read workflows.
func HandleSavings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	semblePath, lookErr := exec.LookPath("semble")
	if lookErr != nil {
		return err("semble is not installed. Install it with: uv tool install semble")
}

	cmd := exec.CommandContext(ctx, semblePath, "savings")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(fmt.Sprintf("semble savings failed: %s\n%s", cmdErr.Error(), string(output)))
}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return ok("No savings data available yet. Run some searches first.")
}

	return ok(result)
}

// HandleIndexInfo returns information about the semble index for a given path,
// including file count, index size, and content type coverage.
func HandleIndexInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return err(fmt.Sprintf("failed to get current directory: %s", cwdErr.Error()))
}

		path = cwd
	}

	// Resolve to absolute path
	absPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return err(fmt.Sprintf("failed to resolve path: %s", absErr.Error()))
}

	// Check if the path exists
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		return err(fmt.Sprintf("path does not exist: %s", absPath))
}

	if !info.IsDir() {
		return err(fmt.Sprintf("path is not a directory: %s", absPath))
}

	// Check for .gitignore and .sembleignore
	hasGitignore := false
	hasSembleignore := false

	if _, gitErr := os.Stat(filepath.Join(absPath, ".gitignore")); gitErr == nil {
		hasGitignore = true
	}
	if _, sembleErr := os.Stat(filepath.Join(absPath, ".sembleignore")); sembleErr == nil {
		hasSembleignore = true
	}

	// Count source files by walking the directory
	fileCount := 0
	codeExts := map[string]bool{
		".py": true, ".js": true, ".ts": true, ".go": true, ".rs": true,
		".java": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".rb": true, ".php": true, ".swift": true, ".kt": true, ".scala": true,
		".lua": true, ".sh": true, ".bash": true, ".zsh": true, ".fish": true,
		".ex": true, ".exs": true, ".hs": true, ".zig": true, ".cs": true,
		".m": true, ".mm": true, ".r": true, ".R": true, ".jl": true,
		".pl": true, ".pm": true, ".t": true, ".tsx": true, ".jsx": true,
		".vue": true, ".svelte": true, ".dart": true, ".clj": true,
		".cljs": true, ".erl": true, ".hrl": true, ".ml": true, ".mli": true,
		".fs": true, ".fsx": true, ".vim": true, ".el": true, ".lisp": true,
	}

	// Well-known directories to skip
	skipDirs := map[string]bool{
		"node_modules": true, ".venv": true, "venv": true, "dist": true,
		"build": true, "__pycache__": true, ".git": true, ".hg": true,
		".svn": true, "target": true, ".tox": true, ".mypy_cache": true,
		".pytest_cache": true, ".ruff_cache": true, "vendor": true,
		".next": true, ".nuxt": true, "coverage": true, ".coverage": true,
	}

	// Parse .gitignore patterns if present
	var ignorePatterns []*regexp.Regexp
	if hasGitignore {
		gitignoreData, readErr := os.ReadFile(filepath.Join(absPath, ".gitignore"))
		if readErr == nil {
			lines := strings.Split(string(gitignoreData), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				// Simple pattern conversion: treat as suffix or prefix match
				pattern := regexp.QuoteMeta(line)
				pattern = strings.ReplaceAll(pattern, `\*`, ".*")
				pattern = strings.ReplaceAll(pattern, `\?`, ".")
				re, reErr := regexp.Compile(pattern)
				if reErr == nil {
					ignorePatterns = append(ignorePatterns, re)

			}
		}
	}

	shouldIgnore := func(name string) bool {
		if skipDirs[name] || strings.HasPrefix(name, ".") {
			return true
		}
		for _, pat := range ignorePatterns {
			if pat.MatchString(name) {
				return true
			}
		}
		return false
	}

	walkErr := filepath.WalkDir(absPath, func(walkPath string, d os.DirEntry, walkDirErr error) error {
		if walkDirErr != nil {
			return nil
		}
		if d.IsDir() {
			if shouldIgnore(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if codeExts[ext] {
			fileCount++
		}
		return nil
	})
	if walkErr != nil {
		// Non-fatal, continue with what we have
	}

	// Check for cached index
	cacheDir := os.Getenv("SEMBLE_CACHE_LOCATION")
	if cacheDir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr == nil {
			cacheDir = filepath.Join(home, ".cache", "semble")

	}

	var indexStatus string
	if cacheDir != "" {
		// Generate a simple hash-like identifier from the path for the index name
		indexName := strings.ReplaceAll(absPath, string(filepath.Separator), "_")
		indexPath := filepath.Join(cacheDir, "indexes", indexName)
		if _, idxErr := os.Stat(indexPath); idxErr == nil {
			indexStatus = "cached"
		} else {
			indexStatus = "not built (will be built on first search)"
		}
	} else {
		indexStatus = "unknown (cache directory not available)"
	}

	var sb strings.Builder
	sb.WriteString("Semble Index Info\n")
	sb.WriteString("=================\n\n")
	sb.WriteString(fmt.Sprintf("Path:            %s\n", absPath))
	sb.WriteString(fmt.Sprintf("Index status:    %s\n", indexStatus))
	sb.WriteString(fmt.Sprintf("Source files:    %d\n", fileCount))
	sb.WriteString(fmt.Sprintf(".gitignore:      %v\n", hasGitignore))
	sb.WriteString(fmt.Sprintf(".sembleignore:   %v\n", hasSembleignore))

	return ok(sb.String())
}

}
}

// Helper: count files (used internally, keeping the main handler clean)
func countSourceFiles(absPath string, codeExts map[string]bool, skipDirs map[string]bool, ignorePatterns []*regexp.Regexp) int {
	count := 0
	filepath.WalkDir(absPath, func(walkPath string, d os.DirEntry, walkDirErr error) error {
		if walkDirErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			for _, pat := range ignorePatterns {
				if pat.MatchString(name) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if codeExts[ext] {
			count++
		}
		return nil
	})
	return count
}

// unused: kept for potential future use
var _ = sort.Strings

// HandleSearchAPI searches a remote semble-compatible API endpoint.
// This provides an alternative to the CLI for environments where semble
// is accessible via HTTP.
func HandleSearchAPI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL, _ :=getString(args, "api_url")
	if apiURL == "" {
		apiURL = "http://localhost:8765"
	}

	repoPath, _ :=getString(args, "repo_path")
	if repoPath == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return err(fmt.Sprintf("failed to get current directory: %s", cwdErr.Error()))
}

		repoPath = cwd
	}

	topK, _ :=getInt(args, "top_k")
	if topK <= 0 {
		topK = 10
	}

	content, _ :=getString(args, "content")
	if content == "" {
		content = "code"
	}

	// Build request body
	reqBody := map[string]interface{}{
		"query":   query,
		"path":    repoPath,
		"top_k":   topK,
		"content": content,
	}

	jsonData, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %s", jsonErr.Error()))
}

	endpoint := strings.TrimRight(apiURL, "/") + "/search"
	req, reqErr := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(jsonData)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %s", reqErr.Error()))
}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("request failed: %s", respErr.Error()))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %s", readErr.Error()))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}