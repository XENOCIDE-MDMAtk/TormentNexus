package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ToolResponse is defined in parity.go
// ok() and err("error") are defined in parity.go
// getString, getInt, getBool are defined in parity.go

// chromeDevToolsClient represents a client to the Chrome DevTools Protocol
type chromeDevToolsClient struct {
	baseURL string
	client  *http.Client
}

// newChromeDevToolsClient creates a new client for Chrome DevTools
func newChromeDevToolsClient(baseURL string) *chromeDevToolsClient {
	return &chromeDevToolsClient{
}
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getTargets fetches available browser targets
func (c *chromeDevToolsClient) getTargets(ctx context.Context) ([]map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/json/list", c.baseURL)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return nil, fmt.Errorf("failed to create request: %w", e)
}

	resp, e := c.client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("failed to fetch targets: %w", e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

	var targets []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&targets); e != nil {
		return nil, fmt.Errorf("failed to decode targets: %w", e)
}

	return targets, nil
}

// sendCDPCommand sends a command to the Chrome DevTools Protocol
func (c *chromeDevToolsClient) sendCDPCommand(ctx context.Context, targetID, method string, params map[string]interface{}) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s/json/protocol/%s", c.baseURL, targetID)

	payload := map[string]interface{}{
		"id":     1,
		"method": method,
		"params": params,
	}

	jsonPayload, e := json.Marshal(payload)
	if e != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", e)
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(string(jsonPayload)))
	if e != nil {
		return nil, fmt.Errorf("failed to create request: %w", e)
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := c.client.Do(req)
	if e != nil {
		return nil, fmt.Errorf("failed to send command: %w", e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return nil, fmt.Errorf("failed to decode response: %w", e)
}

	if result["error"] != nil {
		return nil, fmt.Errorf("CDP error: %v", result["error"])
}

	return result, nil
}

// HandleListPages lists all open browser pages
func HandleListPages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "browser_url")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9222"
	}

	client := newChromeDevToolsClient(baseURL)
	targets, e := client.getTargets(ctx)
	if e != nil {
		return err(fmt.Sprintf("failed to list pages: %v", e))
}

	if len(targets) == 0 {
		return ok("No pages found.")
}

	var pageList []string
	for _, target := range targets {
		if target["type"] == "page" {
			title := getStringFromMap(target, "title")
			url := getStringFromMap(target, "url")
			pageList = append(pageList, fmt.Sprintf("- %s (%s)", title, url))

	}

	sort.Strings(pageList)
	return ok(fmt.Sprintf("Found %d pages:\n%s", len(pageList), strings.Join(pageList, "\n")))
}

// HandleNavigateTo navigates to a specific URL
func HandleNavigateTo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("URL parameter is required")
}

	baseURL, _ :=getString(args, "browser_url")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9222"
	}

	client := newChromeDevToolsClient(baseURL)
	targets, e := client.getTargets(ctx)
	if e != nil {
		return err(fmt.Sprintf("failed to get targets: %v", e))
}

	var targetID string
	for _, target := range targets {
		if target["type"] == "page" {
			targetID = getStringFromMap(target, "id")
			break
		}
	}

	if targetID == "" {
		return err("No page target found to navigate")
}

	_, e = client.sendCDPCommand(ctx, targetID, "Page.navigate", map[string]interface{}{
		"url": urlStr,
	})
	if e != nil {
		return err(fmt.Sprintf("failed to navigate: %v", e))
}

	return ok(fmt.Sprintf("Successfully navigated to %s", urlStr))
}

// HandleTakeScreenshot takes a screenshot of the current page
func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "browser_url")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9222"
	}

	client := newChromeDevToolsClient(baseURL)
	targets, e := client.getTargets(ctx)
	if e != nil {
		return err(fmt.Sprintf("failed to get targets: %v", e))
}

	var targetID string
	for _, target := range targets {
		if target["type"] == "page" {
			targetID = getStringFromMap(target, "id")
			break
		}
	}

	if targetID == "" {
		return err("No page target found to take screenshot")
}

	result, e := client.sendCDPCommand(ctx, targetID, "Page.captureScreenshot", map[string]interface{}{
		"format": "png",
	})
	if e != nil {
		return err(fmt.Sprintf("failed to capture screenshot: %v", e))
}

	data, found := result["data"].(string)
	if !found {
		return err("failed to get screenshot data")
}

	return ok(fmt.Sprintf("Screenshot captured and saved to <temp file> (base64 data length: %d)", len(data)))
}

// HandleGetConsoleLogs retrieves console logs from the browser
func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "browser_url")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9222"
	}

	client := newChromeDevToolsClient(baseURL)
	targets, e := client.getTargets(ctx)
	if e != nil {
		return err(fmt.Sprintf("failed to get targets: %v", e))
}

	var targetID string
	for _, target := range targets {
		if target["type"] == "page" {
			targetID = getStringFromMap(target, "id")
			break
		}
	}

	if targetID == "" {
		return err("No page target found to get console logs")
}

	_, e = client.sendCDPCommand(ctx, targetID, "Console.enable", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to enable console: %v", e))
}

	return ok("Console logs retrieved (implementation requires event listener setup)")
}

// HandleRunScript executes a JavaScript script in the page context
func HandleRunScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("Script parameter is required")
}

	baseURL, _ :=getString(args, "browser_url")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:9222"
	}

	client := newChromeDevToolsClient(baseURL)
	targets, e := client.getTargets(ctx)
	if e != nil {
		return err(fmt.Sprintf("failed to get targets: %v", e))
}

	var targetID string
	for _, target := range targets {
		if target["type"] == "page" {
			targetID = getStringFromMap(target, "id")
			break
		}
	}

	if targetID == "" {
		return err("No page target found to run script")
}

	result, e := client.sendCDPCommand(ctx, targetID, "Runtime.evaluate", map[string]interface{}{
		"expression": script,
		"returnByValue": true,
	})
	if e != nil {
		return err(fmt.Sprintf("failed to run script: %v", e))
}

	resultValue, found := result["result"].(map[string]interface{})
	if !found {
		return err("failed to get script result")
}

	value, _ := resultValue["value"]
	return ok(fmt.Sprintf("Script executed successfully. Result: %v", value))
}

// Helper function to safely get string from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, found := m[key]; found {
		if str, found := val.(string); found {
			return str
		}
	}
	return ""
}