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

func makeSentryRequest(method, path, token string, params url.Values) (map[string]interface{}, error) {
	baseURL := os.Getenv("SENTRY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sentry.io/api/0"
	}
	fullURL := baseURL + path
	if params != nil && len(params) > 0 {
		fullURL = fullURL + "?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(context.Background(), method, fullURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Sentry API error %d: %s", resp.StatusCode, string(body))
}

	var result map[string]interface{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return result, nil
}

}

func makeSentryListRequest(method, path, token string, params url.Values) ([]interface{}, error) {
	baseURL := os.Getenv("SENTRY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sentry.io/api/0"
	}
	fullURL := baseURL + path
	if params != nil && len(params) > 0 {
		fullURL = fullURL + "?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(context.Background(), method, fullURL, nil)
	if reqErr != nil {
		return nil, reqErr
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, doErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Sentry API error %d: %s", resp.StatusCode, string(body))
}

	var result []interface{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return result, nil
}

}

// HandleListProjects lists all Sentry projects for the authenticated user
func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("SENTRY_AUTH_TOKEN environment variable is not set")
}

	result, apiErr := makeSentryListRequest("GET", "/projects/", token, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

// HandleListIssues lists issues for a Sentry project, optionally filtered by query and status
func HandleListIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("SENTRY_AUTH_TOKEN environment variable is not set")
}

	org, _ :=getString(args, "organization_slug")
	project, _ :=getString(args, "project_slug")
	if org == "" || project == "" {
		return err("organization_slug and project_slug are required")
}

	params := url.Values{}
	query, _ :=getString(args, "query")
	if query != "" {
		params.Set("query", query)

	status, _ :=getString(args, "status")
	if status != "" {
		params.Set("query", status)

	path := fmt.Sprintf("/projects/%s/%s/issues/", url.PathEscape(org), url.PathEscape(project))

	result, apiErr := makeSentryListRequest("GET", path, token, params)
	if apiErr != nil {
		return err(apiErr.Error())
}

	jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

}
}

// HandleGetIssueDetails retrieves details for a specific Sentry issue
func HandleGetIssueDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("SENTRY_AUTH_TOKEN environment variable is not set")
}

	issueID, _ :=getString(args, "issue_id")
	if issueID == "" {
		return err("issue_id is required")
}

	path := fmt.Sprintf("/issues/%s/", url.PathEscape(issueID))

	result, apiErr := makeSentryRequest("GET", path, token, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

// HandleListEvents lists events for a specific Sentry issue
func HandleListEvents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("SENTRY_AUTH_TOKEN environment variable is not set")
}

	issueID, _ :=getString(args, "issue_id")
	if issueID == "" {
		return err("issue_id is required")
}

	path := fmt.Sprintf("/issues/%s/events/", url.PathEscape(issueID))

	result, apiErr := makeSentryListRequest("GET", path, token, nil)
	if apiErr != nil {
		return err(apiErr.Error())
}

	jsonBytes, jsonErr := json.MarshalIndent(result, "", "  ")
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

// HandleUpdateIssue updates a Sentry issue (e.g., change status to resolved)
func HandleUpdateIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SENTRY_AUTH_TOKEN")
	if token == "" {
		return err("SENTRY_AUTH_TOKEN environment variable is not set")
}

	issueID, _ :=getString(args, "issue_id")
	if issueID == "" {
		return err("issue_id is required")
}

	status, _ :=getString(args, "status")
	assignedTo, _ :=getString(args, "assignedTo")

	if status == "" && assignedTo == "" {
		return err("at least one of status or assignedTo must be provided")
}

	updatePayload := map[string]interface{}{}
	if status != "" {
		validStatuses := map[string]bool{"resolved": true, "unresolved": true, "ignored": true}
		if !validStatuses[status] {
			return err("status must be one of: resolved, unresolved, ignored")
}

		updatePayload["status"] = status
	}
	if assignedTo != "" {
		updatePayload["assignedTo"] = assignedTo
	}

	bodyBytes, jsonErr := json.Marshal(updatePayload)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	baseURL := os.Getenv("SENTRY_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sentry.io/api/0"
	}
	fullURL := baseURL + fmt.Sprintf("/issues/%s/", url.PathEscape(issueID))

	req, reqErr := http.NewRequestWithContext(ctx, "PUT", fullURL, strings.NewReader(string(bodyBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, doErr := client.Do(req)
	if doErr != nil {
		return err(doErr.Error())
}

	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("Sentry API error %d: %s", resp.StatusCode, string(respBody)))
}

	return ok(string(respBody))
}