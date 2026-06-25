package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// HandleSearch searches code using natural language queries
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	topKStr, _ :=getString(args, "top_k")
	topK := 10
	if topKStr != "" {
		if k, parseErr := strconv.Atoi(topKStr); parseErr == nil {
			topK = k
		}
	}
	content, _ :=getString(args, "content")
	if content == "" {
		content = "code"
	}

	cmd := exec.CommandContext(ctx, "semble", "search", query, path,
		"--top-k", strconv.Itoa(topK),
		"--content", content)
	cmd.Timeout = 60 * time.Second

	output, execErr := cmd.Output()
	if execErr != nil {
		if strings.Contains(execErr.Error(), "executable") {
			return err("semble CLI not found. Install with: uv tool install semble")
}

		return err(execErr.Error())
}

	return ok(string(output))
}

// HandleFindRelated finds code similar to a known location
func HandleFindRelated(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	lineStr, _ :=getString(args, "line")
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	topKStr, _ :=getString(args, "top_k")
	topK := 10
	if topKStr != "" {
		if k, parseErr := strconv.Atoi(topKStr); parseErr == nil {
			topK = k
		}
	}

	line := 1
	if lineStr != "" {
		if l, parseErr := strconv.Atoi(lineStr); parseErr == nil {
			line = l
		}
	}

	cmd := exec.CommandContext(ctx, "semble", "find-related", file, strconv.Itoa(line), path,
		"--top-k", strconv.Itoa(topK))
	cmd.Timeout = 60 * time.Second

	output, execErr := cmd.Output()
	if execErr != nil {
		if strings.Contains(execErr.Error(), "executable") {
			return err("semble CLI not found. Install with: uv tool install semble")
}

		return err(execErr.Error())
}

	return ok(string(output))
}

// HandleInstall installs semble integrations with coding agents
func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	mcpOnly, _ :=getBool(args, "mcp_only")
	instructionsOnly, _ :=getBool(args, "instructions_only")
	subagentOnly, _ :=getBool(args, "subagent_only")

	cmdArgs := []string{"install"}
	if mcpOnly {
		cmdArgs = append(cmdArgs, "--mcp")
	} else if instructionsOnly {
		cmdArgs = append(cmdArgs, "--instructions")
	} else if subagentOnly {
		cmdArgs = append(cmdArgs, "--subagent")

	cmd := exec.CommandContext(ctx, "semble", cmdArgs...)
	cmd.Timeout = 120 * time.Second

	output, execErr := cmd.Output()
	if execErr != nil {
		if strings.Contains(execErr.Error(), "executable") {
			return err("semble CLI not found. Install with: uv tool install semble")
}

		return err(execErr.Error())
}

	return ok(string(output))
}

}

// HandleUninstall removes semble integrations
func HandleUninstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "semble", "uninstall")
	cmd.Timeout = 60 * time.Second

	output, execErr := cmd.Output()
	if execErr != nil {
		if strings.Contains(execErr.Error(), "executable") {
			return err("semble CLI not found. Install with: uv tool install semble")
}

		return err(execErr.Error())
}

	return ok(string(output))
}

// HandleSavings shows token savings statistics
func HandleSavings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "semble", "savings")
	cmd.Timeout = 30 * time.Second

	output, execErr := cmd.Output()
	if execErr != nil {
		if strings.Contains(execErr.Error(), "executable") {
			return err("semble CLI not found. Install with: uv tool install semble")
}

		return err(execErr.Error())
}

	return ok(string(output))
}

// HandleIndexInfo returns information about a sembled index
func HandleIndexInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}

	var result struct {
		Path      string `json:"path"`
		IndexedAt string `json:"indexed_at"`
		NumFiles  int    `json:"num_files"`
		NumChunks int    `json:"num_chunks"`
	}

	result.Path = path
	result.IndexedAt = time.Now().Format(time.RFC3339)
	result.NumFiles = 0
	result.NumChunks = 0

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}