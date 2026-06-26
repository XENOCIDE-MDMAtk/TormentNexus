package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client for Supabase API
var supabaseClient = http.DefaultClient

// Helper to build query strings safely
func buildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)

	return values.Encode()
}

}

// HandleListTables lists tables in a specific schema
func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	projectRef, _ :=getString(args, "project_ref")
	schema, _ :=getString(args, "schema")
	if schema == "" {
		schema = "public"
	}

	if apiKey == "" || projectRef == "" {
		return err("missing required arguments: api_key and project_ref")
}

	baseURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/tables", projectRef)
	queryParams := map[string]string{
		"select": "*",
		"schema": schema,
	}
	if len(queryParams) > 0 {
		baseURL += "?" + buildQuery(queryParams)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "count=exact")

	resp, fetchErr := supabaseClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch tables: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	// Parse JSON to ensure it's valid and format nicely
	var result interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	prettyJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format JSON: %v", marshalErr))
}

	return ok(string(prettyJSON))
}

}

// HandleGetRow retrieves a specific row from a table
func HandleGetRow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	projectRef, _ :=getString(args, "project_ref")
	tableName, _ :=getString(args, "table_name")
	id, _ :=getString(args, "id")

	if apiKey == "" || projectRef == "" || tableName == "" || id == "" {
		return err("missing required arguments: api_key, project_ref, table_name, and id")
}

	baseURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/%s?id=eq.%s", projectRef, tableName, id)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, fetchErr := supabaseClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch row: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	prettyJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format JSON: %v", marshalErr))
}

	return ok(string(prettyJSON))
}

// HandleInsertRow inserts a new row into a table
func HandleInsertRow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	projectRef, _ :=getString(args, "project_ref")
	tableName, _ :=getString(args, "table_name")
	dataJSON, _ :=getString(args, "data")

	if apiKey == "" || projectRef == "" || tableName == "" || dataJSON == "" {
		return err("missing required arguments: api_key, project_ref, table_name, and data")
}

	// Validate JSON input
	var data map[string]interface{}
	parseErr := json.Unmarshal([]byte(dataJSON), &data)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid JSON in data argument: %v", parseErr))
}

	baseURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/%s", projectRef, tableName)

	reqBody, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal data: %v", marshalErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(string(reqBody)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, fetchErr := supabaseClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to insert row: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result interface{}
	parseErr = json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	prettyJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to format JSON: %v", marshalErr))
}

	return ok(string(prettyJSON))
}

// HandleCountRows counts rows in a table with optional filters
func HandleCountRows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	projectRef, _ :=getString(args, "project_ref")
	tableName, _ :=getString(args, "table_name")
	filter, _ :=getString(args, "filter") // e.g., "status=eq.active"

	if apiKey == "" || projectRef == "" || tableName == "" {
		return err("missing required arguments: api_key, project_ref, and table_name")
}

	baseURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/%s", projectRef, tableName)
	
	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("select", "count")
	queryParams.Set("count", "exact")
	
	if filter != "" {
		// Simple parsing of filter string like "status=eq.active"
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) == 2 {
			queryParams.Set(parts[0], parts[1])

	}

	if len(queryParams) > 0 {
		baseURL += "?" + queryParams.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Prefer", "count=exact")

	resp, fetchErr := supabaseClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to count rows: %v", fetchErr))
}

	defer resp.Body.Close()

	countStr := resp.Header.Get("Content-Range")
	if countStr == "" {
		// Fallback to reading body if header is missing
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return err(fmt.Sprintf("failed to read response: %v", readErr))
}

		return err(fmt.Sprintf("count not found in headers, response: %s", string(body)))
}

	// Content-Range format: "*/123" or "0-10/123"
	parts := strings.Split(countStr, "/")
	if len(parts) != 2 {
		return err(fmt.Sprintf("unexpected Content-Range format: %s", countStr))
}

	count, parseErr := strconv.Atoi(parts[1])
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse count: %v", parseErr))
}

	return ok(fmt.Sprintf("Total rows: %d", count))
}
}
}