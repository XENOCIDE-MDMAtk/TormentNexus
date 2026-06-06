package tools

/**
 * @file anysiteio_anysite_mcp_server.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Anysite MCP Server.
 * Replaces: https://github.com/anysiteio/anysite-mcp-server
 * Tools exposed: discover, execute, get_page, query_cache, export_data
 *
 * WHY: The entire web as a database - query structured data from LinkedIn,
 * Instagram, Twitter/X, Reddit, YouTube, SEC EDGAR, Y Combinator, Crunchbase,
 * and any URL through five universal meta-tools.
 */

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const anysiteBaseURL = "https://api.anysite.io"

func anysiteAPIKey() string {
	// Try ANYSITE_ACCESS_TOKEN first (from local dev), then ANYSITE_API_KEY
	if key := os.Getenv("ANYSITE_ACCESS_TOKEN"); key != "" {
		return key
	}
	return os.Getenv("ANYSITE_API_KEY")
}

func anysiteRequest(ctx context.Context, method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	apiKey := anysiteAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("ANYSITE_ACCESS_TOKEN or ANYSITE_API_KEY environment variable is not set")
	}

	baseURL := os.Getenv("ANYSITE_API_URL")
	if baseURL == "" {
		baseURL = anysiteBaseURL
	}

	url := baseURL + endpoint
	var bodyReader io.Reader
	if payload != nil {
		data, _ := json.Marshal(payload)
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Anysite API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anysite response: %v", err)
	}
	return result, nil
}

// HandleAnysiteDiscover implements the discover tool.
// Lists endpoints in a category with their params and LLM hints.
// Replaces: https://github.com/anysiteio/anysite-mcp-server
// Tool: discover
func HandleAnysiteDiscover(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ := getString(args, "source")
	if source == "" {
		return err("source parameter is required (e.g., 'linkedin', 'twitter', 'crunchbase')")
	}

	category, _ := getString(args, "category")
	if category == "" {
		return err("category parameter is required (e.g., 'search', 'profile', 'posts')")
	}

	payload := map[string]interface{}{
		"source":   source,
		"category": category,
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/discover", payload)
	if err != nil {
		return err(fmt.Sprintf("Discover request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAnysiteExecute implements the execute tool.
// Fetches data from a source/category/endpoint. Returns first items + cache_key.
// Replaces: https://github.com/anysiteio/anysite-mcp-server
// Tool: execute
func HandleAnysiteExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ := getString(args, "source")
	if source == "" {
		return err("source parameter is required")
	}

	category, _ := getString(args, "category")
	if category == "" {
		return err("category parameter is required")
	}

	endpoint, _ := getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint parameter is required")
	}

	// Extract params if provided
	payload := map[string]interface{}{
		"source":   source,
		"category": category,
		"endpoint": endpoint,
	}

	if params, ok := args["params"].(map[string]interface{}); ok {
		payload["params"] = params
	}

	// Add optional pagination
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}
	if offset := getInt(args, "offset"); offset > 0 {
		payload["offset"] = offset
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/execute", payload)
	if err != nil {
		return err(fmt.Sprintf("Execute request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAnysiteGetPage implements the get_page tool.
// Paginates through cached items without re-fetching from the source.
// Replaces: https://github.com/anysiteio/anysite-mcp-server
// Tool: get_page
func HandleAnysiteGetPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cacheKey, _ := getString(args, "cache_key")
	if cacheKey == "" {
		return err("cache_key parameter is required")
	}

	payload := map[string]interface{}{
		"cache_key": cacheKey,
	}

	if offset := getInt(args, "offset"); offset >= 0 {
		payload["offset"] = offset
	}
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/get_page", payload)
	if err != nil {
		return err(fmt.Sprintf("Get page request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAnysiteQueryCache implements the query_cache tool.
// Filters, sorts, and aggregates cached items locally.
// Replaces: https://github.com/anysiteio/anysite-mcp-server
// Tool: query_cache
func HandleAnysiteQueryCache(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cacheKey, _ := getString(args, "cache_key")
	if cacheKey == "" {
		return err("cache_key parameter is required")
	}

	payload := map[string]interface{}{
		"cache_key": cacheKey,
	}

	// Add optional query parameters
	if conditions, ok := args["conditions"].([]interface{}); ok {
		payload["conditions"] = conditions
	}
	if sortBy, _ := getString(args, "sort_by"); sortBy != "" {
		payload["sort_by"] = sortBy
	}
	if sortOrder, _ := getString(args, "sort_order"); sortOrder != "" {
		payload["sort_order"] = sortOrder
	}
	if aggregate, ok := args["aggregate"].(map[string]interface{}); ok {
		payload["aggregate"] = aggregate
	}
	if groupBy, _ := getString(args, "group_by"); groupBy != "" {
		payload["group_by"] = groupBy
	}
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}
	if offset := getInt(args, "offset"); offset >= 0 {
		payload["offset"] = offset
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/query_cache", payload)
	if err != nil {
		return err(fmt.Sprintf("Query cache request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAnysiteExportData implements the export_data tool.
// Dumps the full cached dataset to json, jsonl, or csv format.
// Replaces: https://github.com/anysiteio/anysite-mcp-server
// Tool: export_data
func HandleAnysiteExportData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cacheKey, _ := getString(args, "cache_key")
	if cacheKey == "" {
		return err("cache_key parameter is required")
	}

	format, _ := getString(args, "format")
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "jsonl" && format != "csv" {
		return err("format must be one of: json, jsonl, csv")
	}

	payload := map[string]interface{}{
		"cache_key": cacheKey,
		"format":    format,
	}

	// Optional: limit and offset for partial export
	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}
	if offset := getInt(args, "offset"); offset >= 0 {
		payload["offset"] = offset
	}
	if listUnpack, ok := args["list_unpack"].(bool); ok {
		payload["list_unpack"] = listUnpack
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/export", payload)
	if err != nil {
		return err(fmt.Sprintf("Export data request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleAnysiteSearch provides a simplified search interface combining discover+execute.
// Tool: anysite_search
func HandleAnysiteSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}

	source, _ := getString(args, "source")
	if source == "" {
		source = "web" // default to web search
	}

	payload := map[string]interface{}{
		"query":  query,
		"source": source,
	}

	if limit := getInt(args, "limit"); limit > 0 {
		payload["limit"] = limit
	}

	result, err := anysiteRequest(ctx, "POST", "/v1/search", payload)
	if err != nil {
		return err(fmt.Sprintf("Search request failed: %v", err))
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// anysiteSourceCache stores discovered sources to avoid repeated API calls
var anysiteSourceCache = make(map[string]interface{})

// HandleAnysiteListSources returns available data sources.
// Tool: anysite_list_sources
func HandleAnysiteListSources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Check cache first
	if cached, ok := anysiteSourceCache["sources"]; ok {
		out, _ := json.MarshalIndent(cached, "", "  ")
		return ok(string(out))
	}

	result, err := anysiteRequest(ctx, "GET", "/v1/sources", nil)
	if err != nil {
		// Return default sources list if API doesn't support this endpoint
		sources := []map[string]string{
			{"name": "linkedin", "description": "Professional profiles, companies, posts"},
			{"name": "twitter", "description": "Tweets, user profiles, engagement"},
			{"name": "instagram", "description": "Posts, stories, user profiles"},
			{"name": "reddit", "description": "Posts, comments, subreddits"},
			{"name": "youtube", "description": "Videos, channels, comments"},
			{"name": "crunchbase", "description": "Companies, funding, people"},
			{"name": "sec_edgar", "description": "SEC filings, financial data"},
			{"name": "ycombinator", "description": "Startup directory, companies"},
			{"name": "web", "description": "Universal web parser for any URL"},
			{"name": "github", "description": "Repositories, users, code (via AI parser)"},
			{"name": "amazon", "description": "Products, reviews (via AI parser)"},
			{"name": "google_maps", "description": "Places, reviews (via AI parser)"},
			{"name": "g2", "description": "Software reviews (via AI parser)"},
			{"name": "duckduckgo", "description": "Web search results"},
		}
		out, _ := json.MarshalIndent(map[string]interface{}{"sources": sources}, "", "  ")
		return ok(string(out))
	}

	anysiteSourceCache["sources"] = result
	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// Helper for string parsing with defaults
func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultVal
}
