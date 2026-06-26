package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Note: ToolResponse, ok, e, getString, getInt, getBool, TextContent are defined in parity.go

func HandleGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	name, _ :=getString(args, "name")

	if owner == "" || name == "" {
		return err("owner and name are required")
}

	client := http.DefaultClient
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("GitHub API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var data map[string]interface{}
	parseErr := json.Unmarshal(body, &data)
	if parseErr != nil {
		return err(parseErr.Error())
}

	description := getStringFromMap(data, "description")
	if description == "" {
		description = "No description available"
	}

	return ok(fmt.Sprintf("Repository %s/%s found. Description: %s", owner, name, description))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	client := http.DefaultClient
	params := url.Values{}
	params.Set("q", query)
	params.Set("per_page", "5")

	reqURL := "https://api.github.com/search/issues?" + params.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("GitHub API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var result struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Title   string `json:"title"`
			URL     string `json:"html_url"`
			State   string `json:"state"`
			RepoURL string `json:"repository_url"`
		} `json:"items"`
	}

	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if result.TotalCount == 0 {
		return ok("No issues found matching the query.")
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues:\n", result.TotalCount))
	for i, item := range result.Items {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n   URL: %s\n", i+1, item.State, item.Title, item.URL))

	return ok(sb.String())
}

}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	client := http.DefaultClient
	reqURL := fmt.Sprintf("https://api.github.com/users/%s", username)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return err("User not found")
}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("GitHub API error: %d - %s", resp.StatusCode, string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var data map[string]interface{}
	parseErr := json.Unmarshal(body, &data)
	if parseErr != nil {
		return err(parseErr.Error())
}

	name := getStringFromMap(data, "name")
	bio := getStringFromMap(data, "bio")
	publicRepos := getIntFromMap(data, "public_repos")

	return ok(fmt.Sprintf("User: %s (Name: %s, Bio: %s, Public Repos: %d)", username, name, bio, publicRepos))
}

// Helper functions to extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if v, found := m[key]; found {
		if s, found := v.(string); found {
			return s
		}
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if v, found := m[key]; found {
		switch val := v.(type) {
		case float64:
			return int(val)
}
		case int:
			return val
}
		case string:
			i, _ := strconv.Atoi(val)
			return i
		}
	}
	return 0
}