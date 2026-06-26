package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	browserBaseAPI    = "https://www.browserbase.com/api"
	validSessionRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)
)

func HandleCreateSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	apiURL := fmt.Sprintf("%s/sessions", browserBaseAPI)
	payload := map[string]interface{}{
		"projectId": projectID,
	}

	if geo := getString(args, "geo"); geo != "" {
		payload["geo"] = geo
	}

	if record := getBool(args, "record"); record {
		payload["record"] = true
	}

	jsonData, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", marshalErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(jsonData)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result struct {
		ID string `json:"id"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	if !validSessionRegex.MatchString(result.ID) {
		return err(fmt.Sprintf("invalid session ID received: %s", result.ID))
}

	return ok(result.ID)
}

func HandleGetSession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	if sessionID == "" {
		return err("session_id is required")
}

	if !validSessionRegex.MatchString(sessionID) {
		return err("invalid session_id format")
}

	apiURL := fmt.Sprintf("%s/sessions/%s", browserBaseAPI, sessionID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result struct {
		ID        string `json:"id"`
		ProjectID string `json:"projectId"`
		Status    string `json:"status"`
		URL       string `json:"url"`
		CreatedAt string `json:"createdAt"`
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	response := fmt.Sprintf(
		"Session %s\nProject: %s\nStatus: %s\nURL: %s\nCreated: %s",
		result.ID,
		result.ProjectID,
		result.Status,
		result.URL,
		result.CreatedAt,
	)

	return ok(response)
}

func HandleListSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	apiURL := fmt.Sprintf("%s/sessions?projectId=%s", browserBaseAPI, url.QueryEscape(projectID))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var sessions []struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		CreatedAt string `json:"createdAt"`
	}

	if decodeErr := json.NewDecoder(resp.Body).Decode(&sessions); decodeErr != nil {
		return err(fmt.Sprintf("failed to decode response: %v", decodeErr))
}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Sessions for project %s:\n", projectID))
	for _, s := range sessions {
		builder.WriteString(fmt.Sprintf("- %s: %s (created %s)\n", s.ID, s.Status, s.CreatedAt))

	return ok(builder.String())
}

}

func HandleDestroySession(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	if sessionID == "" {
		return err("session_id is required")
}

	if !validSessionRegex.MatchString(sessionID) {
		return err("invalid session_id format")
}

	apiURL := fmt.Sprintf("%s/sessions/%s", browserBaseAPI, sessionID)

	req, reqErr := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	return ok(fmt.Sprintf("Session %s destroyed successfully", sessionID))
}

func HandleGetSessionScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sessionID, _ :=getString(args, "session_id")
	if sessionID == "" {
		return err("session_id is required")
}

	if !validSessionRegex.MatchString(sessionID) {
		return err("invalid session_id format")
}

	apiURL := fmt.Sprintf("%s/sessions/%s/screenshot", browserBaseAPI, sessionID)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "image/png")

	client := http.DefaultClient
	resp, apiErr := client.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	data, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read image data: %v", readErr))
}

	encodedData := base64.StdEncoding.EncodeToString(data)
	return ok(fmt.Sprintf("data:image/png;base64,%s", encodedData))
}

func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}