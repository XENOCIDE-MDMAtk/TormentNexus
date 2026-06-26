package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func HandleListPages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	client := http.DefaultClient
	resp, fetchErr := client.Get(devtoolsURL + "/json/list")
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch pages: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pages []map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pages)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode pages: %v", decodeErr))
}

	var result strings.Builder
	for i, page := range pages {
		if i > 0 {
			result.WriteString("\n")

		title, _ :=getString(page, "title")
		url, _ :=getString(page, "url")
		id, _ :=getString(page, "id")
		result.WriteString(fmt.Sprintf("Page %d: %s\nURL: %s\nID: %s", i+1, title, url, id))

	return ok(result.String())
}

}
}

func HandleBrowsePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	parsedURL, parseErr := url.ParseRequestURI(targetURL)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", parseErr))
}

	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	// First get available pages to find a target
	client := http.DefaultClient
	resp, fetchErr := client.Get(devtoolsURL + "/json/list")
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch pages: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pages []map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pages)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode pages: %v", decodeErr))
}

	var targetPage map[string]interface{}
	for _, page := range pages {
		if strings.HasPrefix(getString(page, "url"), "about:") {
			targetPage = page
			break
		}
	}

	if targetPage == nil {
		return err("no available blank page found")
}

	wsURL, _ :=getString(targetPage, "webSocketDebuggerUrl")
	if wsURL == "" {
		return err("no websocket debugger URL found for target page")
}

	// Navigate to the target URL
	navigateData := map[string]interface{}{
		"url": parsedURL.String(),
	}
	jsonData, _ := json.Marshal(navigateData)

	navigateURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/go"
	navResp, navErr := client.Post(navigateURL, "application/json", strings.NewReader(string(jsonData)))
	if navErr != nil {
		return err(fmt.Sprintf("failed to navigate: %v", navErr))
}

	defer navResp.Body.Close()

	if navResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(navResp.Body)
		return err(fmt.Sprintf("navigation failed: %s - %s", navResp.Status, string(body)))
}

	return ok(fmt.Sprintf("Successfully navigated to %s", parsedURL.String()))
}

func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "page_id")
	if pageID == "" {
		return err("page_id parameter is required")
}

	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	client := http.DefaultClient

	// Get the websocket URL for the page
	pageInfoURL := fmt.Sprintf("%s/json/list/%s", devtoolsURL, pageID)
	resp, fetchErr := client.Get(pageInfoURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch page info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pageInfo map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pageInfo)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode page info: %v", decodeErr))
}

	wsURL, _ :=getString(pageInfo, "webSocketDebuggerUrl")
	if wsURL == "" {
		return err("no websocket debugger URL found for target page")
}

	// Capture screenshot
	captureURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/screenshot"
	captureResp, captureErr := client.Get(captureURL)
	if captureErr != nil {
		return err(fmt.Sprintf("failed to capture screenshot: %v", captureErr))
}

	defer captureResp.Body.Close()

	if captureResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(captureResp.Body)
		return err(fmt.Sprintf("screenshot failed: %s - %s", captureResp.Status, string(body)))
}

	// Read the image data
	imageData, readErr := io.ReadAll(captureResp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read screenshot data: %v", readErr))
}

	// Save to temporary file
	tempFile, tempErr := os.CreateTemp("", "screenshot-*.png")
	if tempErr != nil {
		return err(fmt.Sprintf("failed to create temp file: %v", tempErr))
}

	defer tempFile.Close()

	_, writeErr := tempFile.Write(imageData)
	if write有了writeErr != nil {
		return err(fmt.Sprintf("failed to write screenshot: %v", writeErr))
}

	return ok(fmt.Sprintf("Screenshot saved to %s", tempFile.Name()))
}

func HandleGetConsoleMessages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "page_id")
	if pageID == "" {
		return err("page_id parameter is required")
}

	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	client := http.DefaultClient

	// Get the websocket URL for the page
	pageInfoURL := fmt.Sprintf("%s/json/list/%s", devtoolsURL, pageID)
	resp, fetchErr := client.Get(pageInfoURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch page info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pageInfo map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pageInfo)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode page info: %v", decodeErr))
}

	wsURL, _ :=getString(pageInfo, "webSocketDebuggerUrl")
	if wsURL == "" {
		return err("no websocket debugger URL found for target page")
}

	// Enable console messages
	enableURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/console/enable"
	enableResp, enableErr := client.Post(enableURL, "application/json", strings.NewReader("{}"))
	if enableErr != nil {
		return err(fmt.Sprintf("failed to enable console: %v", enableErr))
}

	defer enableResp.Body.Close()

	if enableResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(enableResp.Body)
		return err(fmt.Sprintf("console enable failed: %s - %s", enableResp.Status, string(body)))
}

	// Get console messages
	consoleURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/console/messages"
	consoleResp, consoleErr := client.Get(consoleURL)
	if consoleErr != nil {
		return err(fmt.Sprintf("failed to get console messages: %v", consoleErr))
}

	defer consoleResp.Body.Close()

	if consoleResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(consoleResp.Body)
		return err(fmt.Sprintf("console messages failed: %s - %s", consoleResp.Status, string(body)))
}

	var messages struct {
		Messages []map[string]interface{} `json:"messages"`
	}
	decodeMsgErr := json.NewDecoder(consoleResp.Body).Decode(&messages)
	if decodeMsgErr != nil {
		return err(fmt.Sprintf("failed to decode console messages: %v", decodeMsgErr))
}

	var result strings.Builder
	for i, msg := range messages.Messages {
		if i > 0 {
			result.WriteString("\n")

		text, _ :=getString(msg, "text")
		level, _ :=getString(msg, "level")
		source, _ :=getString(msg, "url")
		line, _ :=getInt(msg, "lineNumber")
		col, _ :=getInt(msg, "columnNumber")

		result.WriteString(fmt.Sprintf("Message %d [%s]: %s\nSource: %s:%d:%d", i+1, level, text, source, line, col))

	if len(messages.Messages) == 0 {
		return ok("No console messages found")
}

	return ok(result.String())
}

}
}

func HandleGetNetworkRequests(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "page_id")
	if pageID == "" {
		return err("page_id parameter is required")
}

	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	client := http.DefaultClient

	// Get the websocket URL for the page
	pageInfoURL := fmt.Sprintf("%s/json/list/%s", devtoolsURL, pageID)
	resp, fetchErr := client.Get(pageInfoURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch page info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pageInfo map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pageInfo)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode page info: %v", decodeErr))
}

	wsURL, _ :=getString(pageInfo, "webSocketDebuggerUrl")
	if wsURL == "" {
		return err("no websocket debugger URL found for target page")
}

	// Enable network monitoring
	enableURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/network/enable"
	enableResp, enableErr := client.Post(enableURL, "application/json", strings.NewReader("{}"))
	if enableErr != nil {
		return err(fmt.Sprintf("failed to enable network: %v", enableErr))
}

	defer enableResp.Body.Close()

	if enableResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(enableResp.Body)
		return err(fmt.Sprintf("network enable failed: %s - %s", enableResp.Status, string(body)))
}

	// Get network requests
	networkURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/network/requests"
	networkResp, networkErr := client.Get(networkURL)
	if networkErr != nil {
		return err(fmt.Sprintf("failed to get network requests: %v", networkErr))
}

	defer networkResp.Body.Close()

	if networkResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(networkResp.Body)
		return err(fmt.Sprintf("network requests failed: %s - %s", networkResp.Status, string(body)))
}

	var requests struct {
		Requests []map[string]interface{} `json:"requests"`
	}
	decodeReqErr := json.NewDecoder(networkResp.Body).Decode(&requests)
	if decodeReqErr != nil {
		return err(fmt.Sprintf("failed to decode network requests: %v", decodeReqErr))
}

	var result strings.Builder
	for i, req := range requests.Requests {
		if i > 0 {
			result.WriteString("\n")

		method, _ :=getString(req, "method")
		url, _ :=getString(req, "url")
		status, _ :=getInt(req, "status")
		mime, _ :=getString(req, "mimeType")
		size, _ :=getInt(req, "encodedDataLength")

		result.WriteString(fmt.Sprintf("Request %d: %s %s\nStatus: %d\nType: %s\nSize: %d bytes", i+1, method, url, status, mime, size))

	if len(requests.Requests) == 0 {
		return ok("No network requests found")
}

	return ok(result.String())
}

}
}

func HandleGetPerformanceMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageID, _ :=getString(args, "page_id")
	if pageID == "" {
		return err("page_id parameter is required")
}

	// Default Chrome DevTools URL
	devtoolsURL := "http://127.0.0.1:9222"
	if val, exists := os.LookupEnv("CHROME_DEVTOOLS_URL"); exists {
		devtoolsURL = val
	}

	client := http.DefaultClient

	// Get the websocket URL for the page
	pageInfoURL := fmt.Sprintf("%s/json/list/%s", devtoolsURL, pageID)
	resp, fetchErr := client.Get(pageInfoURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch page info: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("devtools API error: %s - %s", resp.Status, string(body)))
}

	var pageInfo map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&pageInfo)
	if decodeErr != nil {
		return err(fmt.Sprintf("failed to decode page info: %v", decodeErr))
}

	wsURL, _ :=getString(pageInfo, "webSocketDebuggerUrl")
	if wsURL == "" {
		return err("no websocket debugger URL found for target page")
}

	// Start performance tracing
	traceURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/performance/start"
	traceResp, traceErr := client.Post(traceURL, "application/json", strings.NewReader("{}"))
	if traceErr != nil {
		return err(fmt.Sprintf("failed to start performance tracing: %v", traceErr))
}

	defer traceResp.Body.Close()

	if traceResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(traceResp.Body)
		return err(fmt.Sprintf("performance tracing start failed: %s - %s", traceResp.Status, string(body)))
}

	// Wait a moment for data collection
	time.Sleep(2 * time.Second)

	// Stop performance tracing and get metrics
	metricsURL := strings.Replace(wsURL, "ws://", "http://", 1) + "/performance/metrics"
	metricsResp, metricsErr := client.Get(metricsURL)
	if metricsErr != nil {
		return err(fmt.Sprintf("failed to get performance metrics: %v", metricsErr))
}

	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(metricsResp.Body)
		return err(fmt.Sprintf("performance metrics failed: %s - %s", metricsResp.Status, string(body)))
}

	var metrics struct {
		Metrics []map[string]interface{} `json:"metrics"`
	}
	decodeMetricsErr := json.NewDecoder(metricsResp.Body).Decode(&metrics)
	if decodeMetricsErr != nil {
		return err(fmt.Sprintf("failed to decode performance metrics: %v", decodeMetricsErr))
}

	var result strings.Builder
	for _, metric := range metrics.Metrics {
		name, _ :=getString(metric, "name")
		value := getFloat(metric, "value")
		result.WriteString(fmt.Sprintf("%s: %.2f\n", name, value))

	if len(metrics.Metrics) == 0 {
		return ok("No performance metrics found")
}

	return ok(result.String())
}

}

// Helper function to get float value from map
func getFloat(m map[string]interface{}, key string) float64 {
	if val, found := m[key]; found {
		switch v := val.(type) {
		case float64:
			return v
}
		case int:
			return float64(v)
}
		case int64:
			return float64(v)
}
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	}
	return 0
}