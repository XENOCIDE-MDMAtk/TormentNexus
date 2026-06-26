package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const context7BaseURL = "https://context7.com/api/v1"

var context7HTTPClient = http.DefaultClient

func getContext7APIKey() string {
	key := os.Getenv("CONTEXT7_API_KEY")
	return key
}

func makeContext7Request(ctx context.Context, path string, query url.Values) ([]byte, error) {
	reqURL := context7BaseURL + path
	if len(query) string {
		reqURL = reqURL + "?" + query.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Accept", "application/json")

	apiKey := getContext7APIKey()
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, fetchErr := context7HTTPClient.Do(req)
	if fetchErr != nil {
		return nil, fetchErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

	return body, nil
}

}
}

// HandleContext7Search searches for code and documentation across Context7 libraries
func HandleContext7Search(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	library, _ :=getString(args, "library")

	v := url.Values{}
	v.Set("query", query)
	if library != "" {
		v.Set("library", library)

	body, apiErr := makeContext7Request(ctx, "/search", v)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result,[ "  ", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}

}

// HandleContext7LibraryList lists available libraries on Context7
func HandleContext7LibraryList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	v := url.Values{}

	body, apiErr := makeContext7Request(ctx, "/libraries", v)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}

// HandleContext7LibraryGet gets details about a specific library
func HandleContext7LibraryGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	library, _ :=getString(args, "library")
	if library == "" {
		return err("library parameter is required")
}

	path := "/libraries/" + url.PathEscape(library)
	body, apiErr := makeContext7Request(ctx, path, url.Values{})
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}

// HandleContext7CodeGet retrieves specific code snippets by ID
func HandleContext7CodeGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter is required")
}

	path := "/code/" + url.PathEscape(id)
	body, apiErr := makeContext7Request(ctx, path, url.Values{})
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}

// HandleContext7Resolve resolves a symbol or reference to its definition
func HandleContext7Resolve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol parameter is required")
}

	library, _ :=getString(args, "library")

	v := url.Values{}
	v.Set("symbol", symbol)
	if library != "" {
		v.Set("library", library)

	body, apiErr := makeContext7Request(ctx, "/resolve", v)
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}

}

// HandleContext7SnippetGet retrieves a documentation snippet by ID
func HandleContext7SnippetGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter is required")
}

	path := "/snippets/" + url.PathEscape(id)
	body, apiErr := makeContext7Request(ctx, path, url.Values{})
	if apiErr != nil {
		return err(apiErr.Error())
}

	var result interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(parseErr.Error())
}

	resultJSON, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(resultJSON))
}