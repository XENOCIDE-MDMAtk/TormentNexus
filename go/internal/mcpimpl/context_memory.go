package mcpimpl

import (
	"context"
)

func HandleStoreContext_context_memory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return ok("stored " + key + " = " + value)
}

func HandleRetrieveContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok("retrieved " + key + " = placeholder")
}