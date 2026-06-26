package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	unlaGatewayURL = os.Getenv("UNLA_GATEWAY_URL")
	unlaAPIURL     = os.Getenv("UNLA_API_URL")
)

func init() {
	if unlaGatewayURL == "" {
		unlaGatewayURL = "http://localhost:5235"
	}
	if unlaAPIURL == "" {
		unlaAPIURL = "http://localhost:8080"
	}
}

// HandleListMCPServers lists all configured MCP servers
func HandleListMCPServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	apiURL := fmt.Sprintf("%s/api/v1/mcp/servers?tenant_id=%s", unlaAPIURL, url.Values{"tenant_id": []string{tenantID}})

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleGetMCPServer retrieves details of a specific MCP server
func HandleGetMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverID, _ :=getString(args, "server_id")
	if serverID == "" {
		return err("server_id is required")
}

	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	apiURL := fmt.Sprintf("%s/api/v1/mcp/servers/%s?tenant_id=%s", unlaAPIURL, serverID, url.Values{"tenant_id": []string{tenantID}})

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleTestGatewayConfig tests and validates a gateway configuration
func HandleTestGatewayConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		configPath = "configs/mcp-gateway.yaml"
	}

	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return err(fmt.Sprintf("config file not found: %s", configPath))
}

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/mcp-gateway/main.go", "test", "-c", configPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("config test failed: %s\nOutput: %s", execErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Configuration test passed:\n%s", string(output)))
}

// HandleReloadGatewayConfig triggers hot-reload of gateway configuration
func HandleReloadGatewayConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config_path")
	if configPath == "" {
		configPath = "configs/mcp-gateway.yaml"
	}

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/mcp-gateway/main.go", "reload", "-c", configPath)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("reload failed: %s\nOutput: %s", execErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Configuration reloaded successfully:\n%s", string(output)))
}

// HandleGatewayHealth checks the health status of the MCP gateway
func HandleGatewayHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	healthURL := fmt.Sprintf("%s/health", unlaGatewayURL)

	client := http.DefaultClient
	resp, fetchErr := client.Get(healthURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("health check failed: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleGatewayStatus returns comprehensive gateway status including active sessions
func HandleGatewayStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	statusURL := fmt.Sprintf("%s/api/v1/status", unlaGatewayURL)

	client := http.DefaultClient
	resp, fetchErr := client.Get(statusURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("status check failed: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleListTools lists all available tools from MCP servers
func HandleListTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	toolsURL := fmt.Sprintf("%s/api/v1/mcp/tools?tenant_id=%s", unlaAPIURL, url.Values{"tenant_id": []string{tenantID}})

	client := http.DefaultClient
	resp, fetchErr := client.Get(toolsURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to list tools: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleCallTool calls a specific tool on an MCP server
func HandleCallTool(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	toolName, _ :=getString(args, "tool_name")
	if toolName == "" {
		return err("tool_name is required")
}

	serverID, _ :=getString(args, "server_id")
	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	toolArgs, hasArgs := args["arguments"]
	var arguments map[string]interface{}
	if hasArgs {
		if argsMap, found := toolArgs.(map[string]interface{}); found {
			arguments = argsMap
		}
	}
	if arguments == nil {
		arguments = make(map[string]interface{})

	payload := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
		"tenant_id": tenantID,
	}
	if serverID != "" {
		payload["server_id"] = serverID
	}

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	callURL := fmt.Sprintf("%s/api/v1/mcp/call", unlaAPIURL)
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", callURL, strings.NewReader(string(body)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonMarshalErr := json.MarshalIndent(result, "", "  ")
	if jsonMarshalErr != nil {
		return err(jsonMarshalErr.Error())
}

	return ok(string(data))
}

}

// HandleCreateMCPServer creates a new MCP server configuration
func HandleCreateMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	config, _ :=getString(args, "config")
	if config == "" {
		return err("config is required")
}

	payload := map[string]interface{}{
		"name":      name,
		"config":    config,
		"tenant_id": tenantID,
	}

	description, _ :=getString(args, "description")
	if description != "" {
		payload["description"] = description
	}

	enabled, _ :=getBool(args, "enabled")
	payload["enabled"] = enabled

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	createURL := fmt.Sprintf("%s/api/v1/mcp/servers", unlaAPIURL)
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "POST", createURL, strings.NewReader(string(body)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonMarshalErr := json.MarshalIndent(result, "", "  ")
	if jsonMarshalErr != nil {
		return err(jsonMarshalErr.Error())
}

	return ok(string(data))
}

// HandleUpdateMCPServer updates an existing MCP server configuration
func HandleUpdateMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverID, _ :=getString(args, "server_id")
	if serverID == "" {
		return err("server_id is required")
}

	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	payload := map[string]interface{}{
		"tenant_id": tenantID,
	}

	if name := getString(args, "name"); name != "" {
		payload["name"] = name
	}
	if config := getString(args, "config"); config != "" {
		payload["config"] = config
	}
	if description := getString(args, "description"); description != "" {
		payload["description"] = description
	}

	if _, hasEnabled := args["enabled"]; hasEnabled {
		payload["enabled"] = getBool(args, "enabled")

	body, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	updateURL := fmt.Sprintf("%s/api/v1/mcp/servers/%s", unlaAPIURL, serverID)
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "PUT", updateURL, strings.NewReader(string(body)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonMarshalErr := json.MarshalIndent(result, "", "  ")
	if jsonMarshalErr != nil {
		return err(jsonMarshalErr.Error())
}

	return ok(string(data))
}

}

// HandleDeleteMCPServer deletes an MCP server configuration
func HandleDeleteMCPServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverID, _ :=getString(args, "server_id")
	if serverID == "" {
		return err("server_id is required")
}

	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	deleteURL := fmt.Sprintf("%s/api/v1/mcp/servers/%s?tenant_id=%s", unlaAPIURL, serverID, url.Values{"tenant_id": []string{tenantID}})

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return err(fmt.Sprintf("delete failed with status %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("MCP server %s deleted successfully", serverID))
}

// HandleGetGatewayMetrics retrieves gateway metrics
func HandleGetGatewayMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	metricsURL := fmt.Sprintf("%s/api/v1/metrics", unlaGatewayURL)

	client := http.DefaultClient
	resp, fetchErr := client.Get(metricsURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to get metrics: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleListTenants lists all configured tenants
func HandleListTenants(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tenantsURL := fmt.Sprintf("%s/api/v1/tenants", unlaAPIURL)

	client := http.DefaultClient
	resp, fetchErr := client.Get(tenantsURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to list tenants: %s", fetchErr.Error()))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(decodeErr.Error())
}

	data, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(data))
}

// HandleGetSession retrieves session information
func HandleGetSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	if sessionID == "" {
		return err("session_id is required")
}

	tenantID, _ :=getString(args, "tenant_id")
	if tenantID == "" {
		tenantID = "default"
	}

	sessionURL := fmt.Sprintf("%s/api/v1/sessions/%s?tenant_id=%s", unlaGatewayURL, sessionID, url.Values{"tenant_id": []string{tenantID}})

}