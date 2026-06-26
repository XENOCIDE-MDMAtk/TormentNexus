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
	client = http.DefaultClient,
	}
	// Simple regex for stripping HTML tags
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	// Regex for finding href attributes
	hrefRegex = regexp.MustCompile(`href=["']([^"']+)["']`)")
)

// fetchPage performs the HTTP GET request and returns the body string
func fetchPage(targetURL string) (string, error) {
	if targetURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
}

	// Validate URL
	_, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return "", fmt.Errorf("invalid URL format: %v", parseErr)
}

	req, reqErr := http.NewRequest("GET", targetURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("failed to create request: %v", reqErr)
}

	// Set a standard User-Agent to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return "", fmt.Errorf("failed to fetch URL: %v", fetchErr)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %v", readErr)
}

	return string(bodyBytes), nil
}

// HandleFetchUrl fetches the raw HTML content of a given URL.
func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")

	html, fetchErr := fetchPage(targetURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	return ok(html)

// HandleExtractText fetches a URL and returns the visible text, stripping HTML tags.
func HandleExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")

	html, fetchErr := fetchPage(targetURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	// Remove script and style content to avoid reading JS/CSS
	scriptRegex := regexp.MustCompile(`(?i)<(script|style)[^>]*>.*?</\1>`)
	cleanHtml := scriptRegex.ReplaceAllString(html, "")

	// Remove remaining HTML tags
	text := htmlTagRegex.ReplaceAllString(cleanHtml, " ")

	// Normalize whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return ok(text)
}

// HandleGetLinks fetches a URL and extracts all href links found in anchor tags.
func HandleGetLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")

	html, fetchErr := fetchPage(targetURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	baseURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err(fmt.Sprintf("failed to parse base URL: %v", parseErr))
}

	matches := hrefRegex.FindAllStringSubmatch(html, -1)
	links := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			href := match[1]
			// Resolve relative URLs
			absURL, resolveErr := url.Parse(href)
			if resolveErr != nil {
				continue
			}
			resolved := baseURL.ResolveReference(absURL).String()

			// Filter fragments and basic duplicates
			if strings.Contains(resolved, "#") {
				resolved = strings.Split(resolved, "#")[0]
			}

			if resolved != "" && !seen[resolved] {
				seen[resolved] = true
				links = append(links, resolved)

		}
	}

	jsonLinks, marshalErr := json.MarshalIndent(links, "", "  ")
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal links: %v", marshalErr))
}

	return ok(string(jsonLinks))
}

}

// HandleSearch fetches a URL and checks if a specific query string exists in the text content.
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	query, _ :=getString(args, "query")

	if query == "" {
		return err("query string cannot be empty")
}

	html, fetchErr := fetchPage(targetURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	// Simple search in raw HTML (can be improved to text-only search)
	// For better accuracy, we strip tags first
	scriptRegex := regexp.MustCompile(`(?i)<(script|style)[^>]*>.*?</\1>`)
	cleanHtml := scriptRegex.ReplaceAllString(html, "")
	text := htmlTagRegex.ReplaceAllString(cleanHtml, " ")
	text = strings.ToLower(text)
	searchTerm := strings.ToLower(query)

	if strings.Contains(text, searchTerm) {
		return ok(fmt.Sprintf("Query '%s' found on the page.", query))
}

	return ok(fmt.Sprintf("Query '%s' not found on the page.", query))
}