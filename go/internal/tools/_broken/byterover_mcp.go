package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const byteroverAPIBase = "https://api.byterover.com/v1"

func HandleSearchCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BYTEROVER_API_KEY")
	if apiKey == "" {
		return err("BYTEROVER_API_KEY environment variable is not set")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	repo, _ :=getString(args, "repo")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	http.DefaultClient := http.DefaultClient
	queryParams := url.Values{}
	queryParams.Set("q", query)
	if repo != "" {
		queryParams.Set("repo", repo)

	queryParams.Set("limit", strconv.Itoa(limit))
	reqURL := fmt.Sprintf("%s/code/search?%s", byteroverAPIBase, queryParams.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	res, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if res.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Code search failed with status %d: %s", res.StatusCode, string(body)))
}

	return ok(string(body))
}

}

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BYTEROVER_API_KEY")
	if apiKey == "" {
		return err("BYTEROVER_API_KEY environment variable is not set")
}

	owner, _ :=getString(args, "owner")
	if owner == "" {
		return err("owner parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 20
	}

	http.DefaultClient := http.DefaultClient
	queryParams := url.Values{}
	queryParams.Set("limit", strconv.Itoa(limit))
	reqURL := fmt.Sprintf("%s/repos/%s?%s", byteroverAPIBase, url.PathEscape(owner), queryParams.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	res, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if res.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("List repos failed with status %d: %s", res.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleGetRepoInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BYTEROVER_API_KEY")
	if apiKey == "" {
		return err("BYTEROVER_API_KEY environment variable is not set")
}

	owner, _ :=getString(args, "owner")
	if owner == "" {
		return err("owner parameter is required")
}

	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("repo parameter is required")
}

	http.DefaultClient := http.DefaultClient
	reqURL := fmt.Sprintf("%s/repos/%s/%s", byteroverAPIBase, url.PathEscape(owner), url.PathEscape(repo))

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	res, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if res.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Get repo info failed with status %d: %s", res.StatusCode, string(body)))
}

	return ok(string(body))
}

func HandleSearchIssues(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("BYTEROVER_API_KEY")
	if apiKey == "" {
		return err("BYTEROVER_API_KEY environment variable is not set")
}

	owner, _ :=getString(args, "owner")
	if owner == "" {
		return err("owner parameter is required")
}

	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("repo parameter is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}

	http.DefaultClient := http.DefaultClient
	queryParams := url.Values{}
	queryParams.Set("q", query)
	queryParams.Set("limit", strconv.Itoa(limit))
	reqURL := fmt.Sprintf("%s/repos/%s/%s/issues/search?%s", byteroverAPIBase, url.PathEscape(owner), url.PathEscape(repo), queryParams.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	res, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if res.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Issue search failed with status %d: %s", res.StatusCode, string(body)))
}

	return ok(string(body))
}