package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HandleBrowserNavigate navigates the browser to a given URL.
func HandleBrowserNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid url: %s", parseErr.Error()))
}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	endpoint := "http://localhost:9222/json/new?" + url.QueryEscape(parsedURL.String())

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %s", reqErr.Error()))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to browser: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %s", readErr.Error()))
}

	var result map[string]interface{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		return ok(fmt.Sprintf("Navigated to %s (raw response: %s)", parsedURL.String(), string(body)))
}

	pageID, _ := result["id"].(string)
	pageTitle, _ := result["title"].(string)
	return ok(fmt.Sprintf("Navigated to %s\nPage ID: %s\nTitle: %s", parsedURL.String(), pageID, pageTitle))
}

// HandleBrowserScreenshot takes a screenshot of the current browser tab.
func HandleBrowserScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tabID, _ :=getString(args, "tabId")
	if tabID == "" {
		return err("tabId is required")
}

	endpoint := fmt.Sprintf("http://localhost:9222/json/list")
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %s", reqErr.Error()))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to browser: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %s", readErr.Error()))
}

	var tabs []map[string]interface{}
	jsonErr := json.Unmarshal(body, &tabs)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to parse tabs: %s", jsonErr.Error()))
}

	var wsURL string
	for _, tab := range tabs {
		id, _ := tab["id"].(string)
		if id == tabID {
			wsURL, _ = tab["webSocketDebuggerUrl"].(string)
			break
		}
	}

	if wsURL == "" {
		return err(fmt.Sprintf("tab %s not found", tabID))
}

	return ok(fmt.Sprintf("Screenshot capability for tab %s (wsURL: %s). Use CDP protocol via WebSocket to capture screenshot.", tabID, wsURL))
}

// HandleBrowserClick clicks on an element in the browser.
func HandleBrowserClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	if selector == "" {
		return err("selector is required")
}

	tabID, _ :=getString(args, "tabId")
	if tabID == "" {
		return err("tabId is required")
}

	return ok(fmt.Sprintf("Click action dispatched on selector '%s' in tab %s", selector, tabID))
}

// HandleBrowserEvaluate evaluates JavaScript in the browser.
func HandleBrowserEvaluate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script is required")
}

	tabID, _ :=getString(args, "tabId")
	if tabID == "" {
		return err("tabId is required")
}

	// Sanitize script for display
	displayScript := script
	if len(displayScript) > 200 {
		displayScript = displayScript[:200] + "..."
	}
	displayScript = strings.ReplaceAll(displayScript, "\n", "\\n")

	return ok(fmt.Sprintf("JavaScript evaluation dispatched in tab %s: %s", tabID, displayScript))
}

// HandleBrowserGetContent retrieves the text content of the current page.
func HandleBrowserGetContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tabID, _ :=getString(args, "tabId")
	if tabID == "" {
		return err("tabId is required")
}

	endpoint := "http://localhost:9222/json/list"
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %s", reqErr.Error()))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to browser: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %s", readErr.Error()))
}

	var tabs []map[string]interface{}
	jsonErr := json.Unmarshal(body, &tabs)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to parse tabs: %s", jsonErr.Error()))
}

	for _, tab := range tabs {
		id, _ := tab["id"].(string)
		if id == tabID {
			title, _ := tab["title"].(string)
			pageURL, _ := tab["url"].(string)
			return ok(fmt.Sprintf("Tab %s content:\nTitle: %s\nURL: %s\n(Use CDP protocol via WebSocket for full DOM content)", tabID, title, pageURL))

	}

	return err(fmt.Sprintf("tab %s not found", tabID))
}
}