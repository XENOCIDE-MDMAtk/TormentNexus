package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// HandleCreateThought creates a new thought entry with optional tags
func HandleCreateThought(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content is a required field")
}

	tags, _ :=getString(args, "tags")
	var tagList []string
	if tags != "" {
		rawTags := strings.Split(tags, ",")
		for _, rawTag := range rawTags {
			trimmed := strings.TrimSpace(rawTag)
			if trimmed != "" {
				tagList = append(tagList, trimmed)

		}
	}

	thoughtID := fmt.Sprintf("thought_%d", time.Now().UnixNano())
	responseText := fmt.Sprintf("Successfully created thought:\nID: %s\nContent: %s\nTags: %s", thoughtID, content, strings.Join(tagList, ", "))
	return ok(responseText)
}

}

// HandleListThoughts lists existing thoughts with optional tag filtering and limit
func HandleListThoughts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tagFilter, _ :=getString(args, "tag")
	limitInput, _ :=getInt(args, "limit")
	limit := 10
	if limitInput > 0 {
		limit = limitInput
	}

	// Simulated thought storage
	allThoughts := []struct {
		ID      string
		Content string
		Tags    []string
	}{
		{"thought_1", "Initial idea for the project", []string{"project", "idea"}},
		{"thought_2", "Go is great for MCP servers", []string{"go", "mcp", "programming"}},
		{"thought_3", "Need to test edge cases", []string{"testing", "quality"}},
		{"thought_4", "MCP protocol documentation", []string{"mcp", "docs"}},
		{"thought_5", "Deploy thoughtbox to production", []string{"deployment", "project"}},
	}

	var matchedThoughts []struct {
		ID      string
		Content string
		Tags    []string
	}

	for _, thought := range allThoughts {
		if tagFilter == "" {
			matchedThoughts = append(matchedThoughts, thought)
		} else {
			for _, tag := range thought.Tags {
				if tag == tagFilter {
					matchedThoughts = append(matchedThoughts, thought)
					break
				}
			}
		}
		if len(matchedThoughts) >= limit {
			break
		}
	}

	var outputLines []string
	outputLines = append(outputLines, fmt.Sprintf("Found %d thoughts (limit: %d, tag filter: %q):", len(matchedThoughts), limit, tagFilter))
	for _, thought := range matchedThoughts {
		outputLines = append(outputLines, fmt.Sprintf("- [%s] %s (tags: %s)", thought.ID, thought.Content, strings.Join(thought.Tags, ", ")))

	return ok(strings.Join(outputLines, "\n"))
}

}

// HandleDeleteThought deletes a thought by its unique ID
func HandleDeleteThought(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	thoughtID, _ :=getString(args, "thought_id")
	if thoughtID == "" {
		return err("thought_id is a required field")
}

	// Simulated valid thought IDs
	validIDs := map[string]bool{
		"thought_1": true,
		"thought_2": true,
		"thought_3": true,
		"thought_4": true,
		"thought_5": true,
	}

	if !validIDs[thoughtID] {
		return err(fmt.Sprintf("thought with ID %q does not exist", thoughtID))
}

	return ok(fmt.Sprintf("Successfully deleted thought with ID: %s", thoughtID))
}