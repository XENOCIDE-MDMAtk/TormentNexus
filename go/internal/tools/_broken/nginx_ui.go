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

 // nginxUIClient handles communication with nginx-ui API
 type nginxUIClient struct {
 	baseURL    string
 	http.DefaultClient *http.Client
 	token      string
 }

 // newNginxUIClient creates a new nginx-ui API client
 func newNginxUIClient() (*nginxUIClient, error) {
 	baseURL := os.Getenv("NGINX_UI_BASE_URL")
 	if baseURL == "" {
 		return nil, fmt.Errorf("NGINX_UI_BASE_URL environment variable is required")
}

 	token := os.Getenv("NGINX_UI_TOKEN")

 	return &nginxUIClient{
}
 		baseURL: strings.TrimSuffix(baseURL, "/"),
 		http.DefaultClient: &http.Client{
 			Timeout: 30 * time.Second,
 		},
 		token: token,
 	}, nil
 }

 // doRequest performs an HTTP request to nginx-ui API
 func (c *nginxUIClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
 	var reqBody io.Reader
 	if body != nil {
 		jsonData, e := json.Marshal(body)
 		if e != nil {
 			return nil, fmt.Errorf("failed to marshal request body: %w", e)
}

 		reqBody = strings.NewReader(string(jsonData))

 	reqURL := c.baseURL + path
 	req, e := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
 	if e != nil {
 		return nil, fmt.Errorf("failed to create request: %w", e)
}

 	req.Header.Set("Content-Type", "application/json")
 	if c.token != "" {
 		req.Header.Set("Authorization", "Bearer "+c.token)

 	resp, e := c.http.DefaultClient.Do(req)
 	if e != nil {
 		return nil, fmt.Errorf("failed to send request: %w", e)
}

 	defer resp.Body.Close()

 	respBody, e := io.ReadAll(resp.Body)
 	if e != nil {
 		return nil, fmt.Errorf("failed to read response body: %w", e)
}

 	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
 		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
}

 	return respBody, nil
 }

}

 // HandleListConfigs lists all nginx configurations
 func HandleListConfigs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
 	client, apiErr := newNginxUIClient()
 	if apiErr != nil {
 		return err("Failed to create nginx-ui client: " + apiErr.Error())
}

 	respBody, fetchErr := client.doRequest(ctx, http.MethodGet, "/api/configs", nil)
 	if fetchErr != nil {
 		return err("Failed to list configs: " + fetchErr.Error())
}

 	var result struct {
 		Data []struct {
 			ID   string `json:"id"`
 			Name string `json:"name"`
 			Path string `json:"path"`
 		} `json:"data"`
 	}

 	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
 		return err("Failed to parse response: " + parseErr.Error())
}

 	var output strings.Builder
 	output.WriteString("Nginx Configurations:\n")
 	for _, cfg := range result.Data {
 		output.WriteString(fmt.Sprintf("  ID: %s, Name: %s, Path: %s\n", cfg.ID, cfg.Name, cfg.Path))

 	return ok(output.String())
}

}

 // HandleGetConfig retrieves a specific nginx configuration
 func HandleGetConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
 	client, apiErr := newNginxUIClient()
 	if apiErr != nil {
 		return err("Failed to create nginx-ui client: " + apiErr.Error())
}

 	configID, _ :=getString(args, "id")
 	if configID == "" {
 		return err("Configuration ID is required")
}

 	path := fmt.Sprintf("/api/configs/%s", url.PathEscape(configID))
 	respBody, fetchErr := client.doRequest(ctx, http.MethodGet, path, nil)
 	if fetchErr != nil {
 		return err("Failed to get config: " + fetchErr.Error())
}

 	var result struct {
 		Data struct {
 			ID     string `json:"id"`
 			Name   string `json:"name"`
 			Path   string `json:"path"`
 			Content string `json:"content"`
 		} `json:"data"`
 	}

 	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
 		return err("Failed to parse response: " + parseErr.Error())
}

 	output := fmt.Sprintf("Configuration: %s\nID: %s\nPath: %s\n\nContent:\n%s",
 		result.Data.Name, result.Data.ID, result.Data.Path, result.Data.Content)

 	return ok(output)
}

 // HandleUpdateConfig updates an existing nginx configuration
 func HandleUpdateConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
 	client, apiErr := newNginxUIClient()
 	if apiErr != nil {
 		return err("Failed to create nginx-ui client: " + apiErr.Error())
}

 	configID, _ :=getString(args, "id")
 	if configID == "" {
 		return err("Configuration ID is required")
}

 	content, _ :=getString(args, "content")
 	if content == "" {
 		return err("Configuration content is required")
}

 	body := map[string]string{
 		"content": content,
 	}

 	path := fmt.Sprintf("/api/configs/%s", url.PathEscape(configID))
 	respBody, fetchErr := client.doRequest(ctx, http.MethodPut, path, body)
 	if fetchErr != nil {
 		return err("Failed to update config: " + fetchErr.Error())
}

 	var result struct {
 		Message string `json:"message"`
 	}

 	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
 		return err("Failed to parse response: " + parseErr.Error())
}

 	return ok("Configuration updated successfully: " + result.Message)
}

 // HandleReloadNginx reloads nginx configuration
 func HandleReloadNginx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
 	client, apiErr := newNginxUIClient()
 	if apiErr != nil {
 		return err("Failed to create nginx-ui client: " + apiErr.Error())
}

 	respBody, fetchErr := client.doRequest(ctx, http.MethodPost, "/api/nginx/reload", nil)
 	if fetchErr != nil {
 		return err("Failed to reload nginx: " + fetchErr.Error())
}

 	var result struct {
 		Message string `json:"message"`
 	}

 	if parseErr := json.Unmarshal(respBody, &result); parseErr != nil {
 		return err("Failed to parse response: " + parseErr.Error())
}

 	return ok("Nginx reloaded successfully: " + result.Message)
 }