package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type chromeTarget struct {
	Description          string `json:"description"`
	DevtoolsFrontendURL  string `json:"devtoolsFrontendUrl"`
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func HandleListTabs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9222
	}

	endpoint := fmt.Sprintf("http://%s:%d/json", host, port)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to Chrome DevTools at %s: %v", endpoint, fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var targets []chromeTarget
	if parseErr := json.Unmarshal(body, &targets); parseErr != nil {
		return err(fmt.Sprintf("failed to parse targets: %v", parseErr))
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d target(s):\n\n", len(targets)))
	for i, t := range targets {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, t.Title))
		sb.WriteString(fmt.Sprintf("   ID: %s\n", t.ID))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", t.Type))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", t.URL))
		if t.WebSocketDebuggerURL != "" {
			sb.WriteString(fmt.Sprintf("   WebSocket: %s\n", t.WebSocketDebuggerURL))

		sb.WriteString("\n")

	return ok(sb.String())
}

}
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9222
	}

	endpoint := fmt.Sprintf("http://%s:%d/json/version", host, port)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to Chrome DevTools at %s: %v", endpoint, fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	var versionInfo map[string]interface{}
	if parseErr := json.Unmarshal(body, &versionInfo); parseErr != nil {
		return err(fmt.Sprintf("failed to parse version info: %v", parseErr))
}

	var sb strings.Builder
	sb.WriteString("Chrome DevTools Version Info:\n\n")
	if browser, found := versionInfo["Browser"]; found {
		sb.WriteString(fmt.Sprintf("Browser: %v\n", browser))

	if protocol, found := versionInfo["Protocol-Version"]; found {
		sb.WriteString(fmt.Sprintf("Protocol Version: %v\n", protocol))

	if userAgent, found := versionInfo["User-Agent"]; found {
		sb.WriteString(fmt.Sprintf("User-Agent: %v\n", userAgent))

	if v8, found := versionInfo["V8-Version"]; found {
		sb.WriteString(fmt.Sprintf("V8 Version: %v\n", v8))

	if wsURL, found := versionInfo["webSocketDebuggerUrl"]; found {
		sb.WriteString(fmt.Sprintf("WebSocket Debugger URL: %v\n", wsURL))

	return ok(sb.String())
}

}
}
}
}
}

func HandleCloseTab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9222
	}
	targetID, _ :=getString(args, "targetId")
	if targetID == "" {
		return err("targetId is required")
}

	endpoint := fmt.Sprintf("http://%s:%d/json/close/%s", host, port, targetID)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to close tab %s: %v", targetID, fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	result := strings.TrimSpace(string(body))
	if result == "Target is closing" {
		return ok(fmt.Sprintf("Successfully closed tab %s", targetID))
}

	return ok(fmt.Sprintf("Close response for tab %s: %s", targetID, result))
}

func HandleActivateTab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9222
	}
	targetID, _ :=getString(args, "targetId")
	if targetID == "" {
		return err("targetId is required")
}

	endpoint := fmt.Sprintf("http://%s:%d/json/activate/%s", host, port, targetID)

	client := http.Client{Timeout: 30 * time.Second}
	req, reqErr := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to activate tab %s: %v", targetID, fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	result := strings.TrimSpace(string(body))
	if result == "Target activated" {
		return ok(fmt.Sprintf("Successfully activated tab %s", targetID))
}

	return ok(fmt.Sprintf("Activate response for tab %s: %s", targetID, result))
}