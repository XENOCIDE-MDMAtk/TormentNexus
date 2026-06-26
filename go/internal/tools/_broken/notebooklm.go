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
)

var (
	notebookLMAPIBase = "https://notebooklm.google.com/api"
)

func HandleCreateNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")

	if title == "" {
		return err("title is required")
}

	payload := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	jsonData, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", jsonErr))
}

	req, reqErr := http.NewRequestWithContext(
		ctx,
		"POST",
		notebookLMAPIBase+"/notebooks",
		strings.NewReader(string(jsonData)),
	)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	notebookID, found := result["id"].(string)
	if !found {
		return err("invalid response format: missing notebook ID")
}

	return ok(fmt.Sprintf("Created notebook with ID: %s", notebookID))
}

func HandleAddSource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	notebookID, _ :=getString(args, "notebook_id")
	sourceURL, _ :=getString(args, "source_url")
	sourceType, _ :=getString(args, "source_type")

	if notebookID == "" {
		return err("notebook_id is required")
}

	if sourceURL == "" {
		return err("source_url is required")
}

	if sourceType == "" {
		sourceType = "webpage"
	}

	validTypes := map[string]bool{
		"webpage": true,
		"document": true,
		"text": true,
	}
	if !validTypes[sourceType] {
		return err(fmt.Sprintf("invalid source_type: %s", sourceType))
}

	payload := map[string]interface{}{
		"url":         sourceURL,
		"source_type": sourceType,
	}

	jsonData, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", jsonErr))
}

	req, reqErr := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/notebooks/%s/sources", notebookLMAPIBase, notebookID),
		strings.NewReader(string(jsonData)),
	)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	sourceID, found := result["id"].(string)
	if !found {
		return err("invalid response format: missing source ID")
}

	return ok(fmt.Sprintf("Added source with ID: %s to notebook %s", sourceID, notebookID))
}

func HandleAskQuestion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	notebookID, _ :=getString(args, "notebook_id")
	question, _ :=getString(args, "question")

	if notebookID == "" {
		return err("notebook_id is required")
}

	if question == "" {
		return err("question is required")
}

	payload := map[string]interface{}{
		"question": question,
	}

	jsonData, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal payload: %v", jsonErr))
}

	req, reqErr := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/notebooks/%s/questions", notebookLMAPIBase, notebookID),
		strings.NewReader(string(jsonData)),
	)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	answer, found := result["answer"].(string)
	if !found {
		return err("invalid response format: missing answer")
}

	return ok(answer)
}

func HandleListNotebooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, reqErr := http.NewRequestWithContext(
		ctx,
		"GET",
		notebookLMAPIBase+"/notebooks",
		nil,
	)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "application/json")

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(fmt.Sprintf("API request failed: %v", apiErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("API error: %s - %s", resp.Status, string(body)))
}

	var result struct {
		Notebooks []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			CreatedAt   string `json:"created_at"`
		} `json:"notebooks"`
	}

	if parseErr := json.NewDecoder(resp.Body).Decode(&result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	var sb strings.Builder
	for _, nb := range result.Notebooks {
		sb.WriteString(fmt.Sprintf("ID: %s\nTitle: %s\nDescription: %s\nCreated: %s\n\n",
			nb.ID, nb.Title, nb.Description, nb.CreatedAt))
	}

	return ok(sb.String())
}

func HandleValidateSourceURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sourceURL, _ :=getString(args, "source_url")

	if sourceURL == "" {
		return err("source_url is required")
}

	parsedURL, urlErr := url.ParseRequestURI(sourceURL)
	if urlErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", urlErr))
}

	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
	}
	if !validSchemes[parsedURL.Scheme] {
		return err("URL must use http or https scheme")
}

	// Simple check for common document extensions
	docExtensions := regexp.MustCompile(`\.(pdf|docx?|txt|md|html?)$`)
	if !docExtensions.MatchString(strings.ToLower(parsedURL.Path)) {
		return ok("URL appears valid but may not be a supported document type")
}

	return ok("URL is valid and appears to be a supported document type")
}