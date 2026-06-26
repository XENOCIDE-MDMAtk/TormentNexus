package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Global in-memory storage for codebase memories
var memoryStore = make(map[string]string)

// HandleStoreMemory stores a memory with a given key and value
func HandleStoreMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")

	if key == "" {
		return err("key is required")
}

	if value == "" {
		return err("value is required")
}

	memoryStore[key] = value
	return ok(fmt.Sprintf("Stored memory with key: %s", key))
}

// HandleGetMemory retrieves a memory by key
func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")

	if key == "" {
		return err("key is required")
}

	value, exists := memoryStore[key]
	if !exists {
		return err(fmt.Sprintf("memory with key '%s' not found", key))
}

	return ok(value)
}

// HandleSearchMemories searches memories by query string
func HandleSearchMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")

	if query == "" {
		return err("query is required")
}

	var results []string
	for key, value := range memoryStore {
		if strings.Contains(strings.ToLower(value), strings.ToLower(query)) {
			results = append(results, fmt.Sprintf("%s: %s", key, value))

	}

	if len(results) == 0 {
		return ok("No memories found matching query")
}

	sort.Strings(results)
	return ok(strings.Join(results, "\n"))
}

}

// HandleListMemories lists all stored memory keys
func HandleListMemories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if len(memoryStore) == 0 {
		return ok("No memories stored")
}

	var keys []string
	for key := range memoryStore {
		keys = append(keys, key)

	sort.Strings(keys)
	return ok(strings.Join(keys, "\n"))
}

}

// HandleDeleteMemory deletes a memory by key
func HandleDeleteMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")

	if key == "" {
		return err("key is required")
}

	if _, exists := memoryStore[key]; !exists {
		return err(fmt.Sprintf("memory with key '%s' not found", key))
}

	delete(memoryStore, key)
	return ok(fmt.Sprintf("Deleted memory with key: %s", key))
}