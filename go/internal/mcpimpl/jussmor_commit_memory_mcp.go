package mcpimpl

import "context"

func HandleAddMemory_jussmor_commit_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	_ = content // placeholder - implement actual storage
	return success("memory stored")
}

func HandleSearchMemory_jussmor_commit_memory_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("found 0 results for: " + query)
}