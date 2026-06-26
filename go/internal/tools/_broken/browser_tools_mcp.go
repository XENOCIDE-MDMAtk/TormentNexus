package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBrowserToolsHost = "127.0.0.1"
	defaultBrowserToolsPort = 3025
)

// BrowserToolsClient handles communication with the Browser Tools Server
type BrowserToolsClient struct {
	host           string
	port           int
	client         *http.Client
	serverDiscovered bool
}

// NewBrowserToolsClient creates a new client with auto-discovery
func NewBrowserToolsClient() *BrowserToolsClient {
	return &BrowserToolsClient{
}
		host:   defaultBrowserToolsHost,
		port:   defaultBrowserToolsPort,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// getDefaultServerPort returns the port from env, .port file, or default
func getDefaultServerPort() int {
	if envPort := os.Getenv("BROWSER_TOOLS_PORT"); envPort != "" {
		if port, parseErr := strconv.Atoi(envPort); parseErr == nil && port > 0 {
			return port
		}
	}
	
	exePath, pathErr := os.Executable()
	if pathErr == nil {
		portFile := filepath.Join(filepath.Dir(exePath), ".port")
		if data, readErr := os.ReadFile(portFile); readErr == nil {
			if port, parseErr := strconv.Atoi(strings.TrimSpace(string(data))); parseErr == nil && port > 0 {
				return port
			}
		}
	}
	
	return defaultBrowserToolsPort
}

// getDefaultServerHost returns the host from env or default
func getDefaultServerHost() string {
	if host := os.Getenv("BROWSER_TOOLS_HOST"); host != "" {
		return host
	}
	return defaultBrowserToolsHost
}

// discoverServer attempts to locate the Browser Tools Server
func (c *BrowserToolsClient) discoverServer() bool {
	if c.serverDiscovered {
		return true
	}
	
	hosts := []string{getDefaultServerHost(), "127.0.0.1", "localhost"}
	defaultPort := getDefaultServerPort()
	ports := []int{defaultPort}
	
	for p := 3025; p <= 3035; p++ {
		if p != defaultPort {
			ports = append(ports, p)

	}
	
	for _, host := range hosts {
		for _, port := range ports {
			discClient := http.DefaultClient
			resp, fetchErr := discClient.Get(fmt.Sprintf("http://%s:%d/.identity", host, port))
			if fetchErr != nil {
				continue
			}
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusOK {
				var identity struct {
					Signature string `json:"signature"`
				}
				if decodeErr := json.NewDecoder(resp.Body).Decode(&identity); decodeErr == nil {
					if identity.Signature == "mcp-browser-connector-24x7" {
						c.host = host
						c.port = port
						c.serverDiscovered = true
						return true
					}
				}
			}
		}
	}
	return false
}

// withServerConnection ensures server is connected before making API calls
func (c *BrowserToolsClient) withServerConnection(apiCall func() (string, error)) (string, error) {
	if !c.serverDiscovered {
		if !c.discoverServer() {
			return "", fmt.Errorf("failed to discover browser connector server. Please ensure it's running")

	}
	
	result, apiErr := apiCall()
	if apiErr != nil {
		c.serverDiscovered = false
		if c.discoverServer() {
			return apiCall()
}

		return "", apiErr
	}
	
	return result, nil
}

}

// getConsoleLogs retrieves browser console logs
func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		resp, fetchErr := client.client.Get(fmt.Sprintf("http://%s:%d/console-logs", client.host, client.port))
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// getConsoleErrors retrieves browser console errors
func HandleGetConsoleErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		resp, fetchErr := client.client.Get(fmt.Sprintf("http://%s:%d/console-errors", client.host, client.port))
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// getNetworkErrors retrieves network error logs
func HandleGetNetworkErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		resp, fetchErr := client.client.Get(fmt.Sprintf("http://%s:%d/network-errors", client.host, client.port))
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// getNetworkLogs retrieves all network logs (successful requests)
func HandleGetNetworkLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		resp, fetchErr := client.client.Get(fmt.Sprintf("http://%s:%d/network-success", client.host, client.port))
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// takeScreenshot captures a screenshot of the current browser tab
func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		req, reqErr := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/capture-screenshot", client.host, client.port), nil)
		if reqErr != nil {
			return "", reqErr
		}
		
		resp, fetchErr := client.client.Do(req)
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		if resp.StatusCode != http.StatusOK {
			var errResp struct {
				Error string `json:"error"`
			}
			if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error != "" {
				return "", fmt.Errorf(errResp.Error)
}

			return "", fmt.Errorf("screenshot failed with status %d", resp.StatusCode)
}

		var result struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Path    string `json:"path"`
		}
		if jsonErr := json.Unmarshal(body, &result); jsonErr == nil {
			if result.Message != "" {
				return result.Message, nil
			}
			if result.Path != "" {
				return fmt.Sprintf("Screenshot saved to: %s", result.Path), nil
			}
			if result.Success {
				return "Successfully saved screenshot", nil
			}
		}
		
		return "Successfully saved screenshot", nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// getSelectedElement retrieves the currently selected DOM element
func HandleGetSelectedElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		resp, fetchErr := client.client.Get(fmt.Sprintf("http://%s:%d/selected-element", client.host, client.port))
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// runAccessibilityAudit runs a WCAG-compliant accessibility audit
func HandleRunAccessibilityAudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	client := NewBrowserToolsClient()
	auditClient := http.DefaultClient
	
	result, apiErr := client.withServerConnection(func() (string, error) {
		req, reqErr := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/runAccessibilityAudit", client.host, client.port), nil)
		if reqErr != nil {
			return "", reqErr
		}
		
		resp, fetchErr := auditClient.Do(req)
		if fetchErr != nil {
			return "", fetchErr
		}
		defer resp.Body.Close()
		
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", readErr
		}
		
		var data interface{}
		if jsonErr := json.Unmarshal(body, &data); jsonErr != nil {
			return string(body), nil
		}
		
		formatted, formatErr := json.MarshalIndent(data, "", "  ")
		if formatErr != nil {
			return string(body), nil
		}
		return string(formatted), nil
	})
	
	if apiErr != nil {
		return err(apiErr.Error())
}

	return ok(result)
}

// runPerformanceAudit runs a performance audit
func HandleRunPerformanceAudit(ctx context.Context, args map[string]interface{})