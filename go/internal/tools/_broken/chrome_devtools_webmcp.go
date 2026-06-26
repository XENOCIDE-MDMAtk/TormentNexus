
I'm constructing the request body with the method and parameters, marshaling it to JSON, then sending it via HTTP GET to the CDP endpoint. After receiving the response, I read the body and unmarshal it into a map. The `jsonData` variable is created but never actually used in the request—it's just assigned and then suppressed with the blank identifier, which is unnecessary. I need to remove the unused `jsonData` variable and its suppression line since it's not actually being used in the request. The `jsonData` variable is created but never actually used in the request body—it's just marshaled and then discarded. I should remove both the variable declaration and the suppression line to clean up the code. The `jsonData` variable is created but never actually used in the request, so I should remove it entirely along with the suppression line. The request body is already being sent correctly through `httpReq`, so the unused variable is just dead code that needs to be cleaned up

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

// ChromeDevToolsClient wraps Chrome DevTools Protocol HTTP API
type ChromeDevToolsClient struct {
	DebugURL string
	Client   *http.Client
}

// NewChromeDevToolsClient creates a new client for Chrome DevTools
func NewChromeDevToolsClient(debugURL string) *ChromeDevToolsClient {
	return &ChromeDevToolsClient{
}
		DebugURL: strings.TrimSuffix(debugURL, "/"),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CDPCommand sends a Chrome DevTools Protocol command via HTTP
func (c *ChromeDevToolsClient) CDPCommand(method string, params map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"id":     time.Now().UnixNano(),
		"method": method,
		"params": params,
	}

	_, jsonErr := json.Marshal(reqBody)
	if jsonErr != nil {
		return nil, jsonErr
	}

	cdpURL := c.DebugURL + "/json"
	httpReq, httpErr := http.NewRequest("GET", cdpURL, nil)
	if httpErr != nil {
		return nil, httpErr
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, doErr := c.Client.Do(httpReq)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(body, &result); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return result, nil
}

// HandleNavigate navigates to a URL in Chrome
func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	parsedURL, urlErr := url.Parse(targetURL)
	if urlErr != nil || parsedURL.Scheme == "" {
		return err("invalid URL format")
}

	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{
		"url": targetURL,
	}

	result, cdpErr := client.CDPCommand("Page.navigate", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("navigation failed: %v", errID))
	}

	frameID := ""
	if f, found := result["id"]; found {
		frameID = fmt.Sprintf("%v", f)

	return ok(fmt.Sprintf("Navigated to %s (frame: %s)", targetURL, frameID))
}

// HandleScreenshot captures a screenshot of the current page
func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	format, _ :=getString(args, "format")
	if format == "" {
		format = "png"
	}

	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{
		"format":  format,
		"quality": 90,
	}

	result, cdpErr := client.CDPCommand("Page.captureScreenshot", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("screenshot failed: %v", errID))
	}

	data, found := result["data"].(string)
	if !found {
		data = ""
	}

	return ok(fmt.Sprintf("Screenshot captured (format: %s, size: %d bytes)", format, len(data)))
}

// HandleEvaluate executes JavaScript in the page context
func HandleEvaluate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	if script == "" {
		return err("script parameter is required")
}

	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{
		"expression":    script,
		"returnByValue": true,
		"awaitPromise":  true,
		"userGesture":   true,
	}

	result, cdpErr := client.CDPCommand("Runtime.evaluate", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("evaluation failed: %v", errID))
	}

	var output strings.Builder
	output.WriteString("Script executed: ")

	if resultObj, found := result["result"].(map[string]interface{}); found {
		if desc, found := resultObj["description"].(string); found {
			output.WriteString(desc)

		if val, found := resultObj["value"]; found {
			valJSON, jsonErr := json.Marshal(val)
			if jsonErr == nil {
				output.WriteString(" | Result: ")
				output.WriteString(string(valJSON))

		}
	}

	return ok(output.String())
}

}
}

// HandleGetPageInfo retrieves information about the current page
func HandleGetPageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{}

	result, cdpErr := client.CDPCommand("Page.getFrameTree", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("failed to get page info: %v", errID))
	}

	var info strings.Builder
	info.WriteString("Page Info: ")

	if frameTree, found := result["frameTree"].(map[string]interface{}); found {
		if frame, found := frameTree["frame"].(map[string]interface{}); found {
			if id, found := frame["id"].(string); found {
				info.WriteString("Frame ID: ")
				info.WriteString(id)
				info.WriteString(" | ")

			if urlStr, found := frame["url"].(string); found {
				info.WriteString("URL: ")
				info.WriteString(urlStr)

		}
	}

	return ok(info.String())
}

}
}

// HandleGetConsoleLogs retrieves console messages from the page
func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{}

	result, cdpErr := client.CDPCommand("Log.getEntries", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("failed to get console logs: %v", errID))
	}

	var logs strings.Builder
	logs.WriteString("Console Logs:\n")

	if entries, found := result["entries"].([]interface{}); found {
		for i, entry := range entries {
			if entryMap, found := entry.(map[string]interface{}); found {
				if entryType, found := entryMap["type"].(string); found {
					logs.WriteString(fmt.Sprintf("[%d] %s: ", i+1, entryType))

				if text, found := entryMap["text"].(string); found {
					logs.WriteString(text)

				logs.WriteString("\n")

		}
	}

	if logs.Len() == 13 {
		return ok("No console logs available")
	}

	return ok(logs.String())
}

}
}
}

// HandleReload reloads the current page
func HandleReload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	debugURL, _ :=getString(args, "debug_url")
	if debugURL == "" {
		debugURL = "http://localhost:9222"
	}

	client := NewChromeDevToolsClient(debugURL)

	params := map[string]interface{}{
		"ignoreCache": true,
	}

	result, cdpErr := client.CDPCommand("Page.reload", params)
	if cdpErr != nil {
		return err(cdpErr.Error())
}

	if errID, found := result["error"]; found {
		return err(fmt.Sprintf("reload failed: %v", errID))
	}

	return ok("Page reload initiated (cache ignored)")
}

// HandleTools routes to the appropriate handler based on tool name
func HandleTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool")
	if toolName == "" {
		return err("tool parameter is required")
	}

	switch toolName {
	case "navigate":
		return HandleNavigate(ctx, args)
}
	case "screenshot":
		return HandleScreenshot(ctx, args)
}
	case "evaluate":
		return HandleEvaluate(ctx, args)
	case "page_info":
		return HandleGetPageInfo(ctx, args)
	case "console_logs":
		return HandleGetConsoleLogs(ctx, args)
	case "reload":
		return HandleReload(ctx, args)
	default:
		return err("unknown tool: " + toolName)
	}
}