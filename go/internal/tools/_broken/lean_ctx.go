package tools

import (
	"context"
	"fmt"
	"os/exec"
)

// execLeanCtx runs the lean-ctx binary with the provided arguments and returns its stdout.
func execLeanCtx(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "lean-ctx", args...)
	outputBytes, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return "", fmt.Errorf("lean-ctx execution failed: %w, output: %s", execErr, string(outputBytes))
}

	return string(outputBytes), nil
}

// HandleCtxRead implements the MCP tool `ctx_read`.
// Arguments:
//   - path (string): file or directory path to read.
//   - mode (string, optional): compression mode (e.g., "map", "signatures").
func HandleCtxRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing required argument: path")
}

	mode, _ :=getString(args, "mode")

	cliArgs := []string{"read"}
	if mode != "" {
		cliArgs = append(cliArgs, "--mode", mode)

	cliArgs = append(cliArgs, path)

	output, execErr := execLeanCtx(ctx, cliArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(TextContent{Content: output})
}

}

// HandleCtxShell implements the MCP tool `ctx_shell`.
// Arguments:
//   - command (string): the shell command to execute via lean-ctx.
func HandleCtxShell(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("missing required argument: command")
}

	cliArgs := []string{"-c", command}
	output, execErr := execLeanCtx(ctx, cliArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(TextContent{Content: output})
}

// HandleCtxSearch implements the MCP tool `ctx_search`.
// Arguments:
//   - query (string): search query.
//   - path (string, optional): directory to limit the search.
func HandleCtxSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
}

	path, _ :=getString(args, "path")

	cliArgs := []string{"search", query}
	if path != "" {
		cliArgs = append(cliArgs, "--path", path)

	output, execErr := execLeanCtx(ctx, cliArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(TextContent{Content: output})
}

}

// HandleCtxTree implements the MCP tool `ctx_tree`.
// Arguments:
//   - path (string, optional): directory to list; defaults to current directory.
func HandleCtxTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	cliArgs := []string{"ls", path}
	output, execErr := execLeanCtx(ctx, cliArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(TextContent{Content: output})
}

// HandleCtxPreload implements the MCP tool `ctx_preload`.
// Arguments:
//   - task (string): description of the task to warm the cache.
func HandleCtxPreload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	if task == "" {
		return err("missing required argument: task")
}

	cliArgs := []string{"preload", task}
	output, execErr := execLeanCtx(ctx, cliArgs)
	if execErr != nil {
		return err(execErr.Error())
}

	return ok(TextContent{Content: output})
}