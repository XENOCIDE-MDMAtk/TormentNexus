package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Client for communicating with BrowserTools Server
var http.DefaultClient = http.DefaultClient

// Server discovery state
type serverState struct {
	host       string
	port       int
	discovered bool
}

var server = serverState{
	host:       "127.0.0.1",
	port:       3025,
	discovered: false,
}

func getDefaultServerPort() int {
	// Check environment variable
	if envPort := os.Getenv("BROWSER_TOOLS_PORT"); envPort != "" {
		if p, e := strconv.Atoi(envPort); e == nil && p > 0 {
			return p
		}
	}

	// Try to read from .port file
	if portFile, e := os.ReadFile(".port"); e == nil {
		if p, e := strconv.Atoi(string(portFile)); e == nil && p > 0 {
			return p
		}
	}

	return 3025
}

func getDefaultServerHost() string {
	if envHost := os.Getenv("BROWSER_TOOLS_HOST"); envHost != "" {
		return envHost
	}
	return "127.0.0.1"
}

func discoverServer() bool {
	hosts := []string{getDefaultServerHost(), "127.0.0.1", "localhost"}
	ports := []int{getDefaultServerPort()}

	// Add fallback ports
	for p := 3025; p <= 3035; p++ {
		if p != ports[0] {
			ports = append(ports, p)

	}

	for _, host := range hosts {
		for _, port := range ports {
			url := fmt.Sprintf("http://%s:%d/.identity", host, port)
			resp, e := http.DefaultClient.Get(url)
			if e != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				continue
			}

			var identity struct {
				Signature string `json:"signature"`
			}

			if e := json.NewDecoder(resp.Body).Decode(&identity); e != nil {
				continue
			}

			if identity.Signature == "mcp-browser-connector-24x7" {
				server.host = host
				server.port = port
				server.discovered = true
				return true
			}
		}
	}

	return false
}

}

func withServerConnection(apiCall func() (ToolResponse, error)) (ToolResponse, error) {
	if !server.discovered {
		if !discoverServer() {
			return err("Failed to discover browser connector server. Please ensure it's running.")

	}

	response, apiErr := apiCall()
	if apiErr != nil {
		server.discovered = false
		if discoverServer() {
			response, retryErr := apiCall()
			if retryErr != nil {
				return err(fmt.Sprintf("Error after reconnection attempt: %s", retryErr.Error()))
}

			return response, nil
		}
		return err(fmt.Sprintf("Failed to reconnect to server: %s", apiErr.Error()))
}

	return response, nil
}

}

func makeGetRequest(path string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d%s", server.host, server.port, path)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		data, e := makeGetRequest("/console-logs")
		if e != nil {
			return err(fmt.Sprintf("Failed to get console logs: %s", e.Error()))
}

		return ok(string(data))
	})

func HandleGetConsoleErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		data, e := makeGetRequest("/console-errors")
		if e != nil {
			return err(fmt.Sprintf("Failed to get console errors: %s", e.Error()))
}

		return ok(string(data))
	})

func HandleGetNetworkErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		data, e := makeGetRequest("/network-errors")
		if e != nil {
			return err(fmt.Sprintf("Failed to get network errors: %s", e.Error()))
}

		return ok(string(data))
	})

func HandleGetNetworkLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		data, e := makeGetRequest("/network-success")
		if e != nil {
			return err(fmt.Sprintf("Failed to get network logs: %s", e.Error()))
}

		return ok(string(data))
	})

func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		url := fmt.Sprintf("http://%s:%d/capture-screenshot", server.host, server.port)
		resp, e := http.DefaultClient.Post(url, "application/json", nil)
		if e != nil {
			return err(fmt.Sprintf("Failed to take screenshot: %s", e.Error()))
}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var result struct {
				Error string `json:"error"`
			}
			if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
				return err(fmt.Sprintf("Error taking screenshot: %s", resp.Status))
}

			return err(fmt.Sprintf("Error taking screenshot: %s", result.Error))
}

		return ok("Successfully saved screenshot")
	})

func HandleGetSelectedElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return withServerConnection(func() (ToolResponse, error) {
}
		data, e := makeGetRequest("/selected-element")
		if e != nil {
			return err(fmt.Sprintf("Failed to get selected element: %s", e.Error()))
}

		return ok(string(data))
	})

func init() {
	// Initialize server port from file if exists
	server.port = getDefaultServerPort()
}