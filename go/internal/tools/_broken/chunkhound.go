package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a shared HTTP client with a 30-second timeout.
var Client = http.DefaultClient

// HandleSearch performs a search query against the ChunkHound API.
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	baseURL := "https://api.chunkhound.example/search"
	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))

	fullURL := baseURL + "?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := Client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch data: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	// Format the result as a readable string
	resultStr := fmt.Sprintf("Found %d results for '%s'", limit, query)
	if data, found := result["data"]; found {
		resultStr += fmt.Sprintf("\nData: %v", data)

	return ok(resultStr)
}

}

// HandleGetChunk retrieves a specific chunk by ID.
func HandleGetChunk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chunkID, _ :=getString(args, "chunk_id")
	if chunkID == "" {
		return err("chunk_id parameter is required")
}

	baseURL := "https://api.chunkhound.example/chunks/" + url.QueryEscape(chunkID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := Client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch chunk: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return err(fmt.Sprintf("chunk with ID '%s' not found", chunkID))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status: %d", resp.StatusCode))
}

	var chunkData map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&chunkData)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse chunk data: %v", parseErr))
}

	content, _ := chunkData["content"].(string)
	return ok(fmt.Sprintf("Chunk ID: %s\nContent: %s", chunkID, content))
}

// HandleListSources lists available data sources.
func HandleListSources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := "https://api.chunkhound.example/sources"

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := Client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch sources: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status: %d", resp.StatusCode))
}

	var sources []map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&sources)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse sources: %v", parseErr))
}

	if len(sources) == 0 {
		return ok("No sources found.")
}

	var sb strings.Builder
	sb.WriteString("Available Sources:\n")
	for i, src := range sources {
		name, _ := src["name"].(string)
		id, _ := src["id"].(string)
		sb.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, name, id))

	return ok(sb.String())
}

}

// HandleStats retrieves general statistics about the index.
func HandleStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := "https://api.chunkhound.example/stats"

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := Client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch stats: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status: %d", resp.StatusCode))
}

	var stats map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&stats)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse stats: %v", parseErr))
}

	totalChunks, _ := stats["total_chunks"].(float64)
	totalSources, _ := stats["total_sources"].(float64)
	lastUpdated, _ := stats["last_updated"].(string)

	return ok(fmt.Sprintf("Total Chunks: %.0f\nTotal Sources: %.0f\nLast Updated: %s", totalChunks, totalSources, lastUpdated))
}