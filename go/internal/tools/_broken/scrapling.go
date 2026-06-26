package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", parseErr))
}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return err("URL must start with http:// or https://")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "Scrapling/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	return ok(string(body))
}

func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	baseURL, _ :=getString(args, "base_url")
	if htmlContent == "" {
		return err("html parameter is required")
}

	linkRegex := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']*)["'][^>]*>`)")
	matches := linkRegex.FindAllStringSubmatch(htmlContent, -1)
	var links []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		rawLink := strings.TrimSpace(match[1])
		if rawLink == "" || strings.HasPrefix(rawLink, "javascript:") || strings.HasPrefix(rawLink, "mailto:") || strings.HasPrefix(rawLink, "tel:") {
			continue
		}
		if baseURL != "" {
			parsedBase, parseErr := url.Parse(baseURL)
			if parseErr == nil {
				parsedLink, parseErr := url.Parse(rawLink)
				if parseErr == nil {
					resolved := parsedBase.ResolveReference(parsedLink)
					rawLink = resolved.String()

			}
		}
		if !seen[rawLink] {
			seen[rawLink] = true
			links = append(links, rawLink)

	}
	sort.Strings(links)
	return ok(strings.Join(links, "\n"))
}

}
}

func HandleExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	if htmlContent == "" {
		return err("html parameter is required")
}

	reScript := regexp.MustCompile(`(?is)<(script|style).*?>.*?</\1>`)
	cleanHTML := reScript.ReplaceAllString(htmlContent, "")
	reTags := regexp.MustCompile(`<[^>]+>`)
	textContent := reTags.ReplaceAllString(cleanHTML, " ")
	reSpace := regexp.MustCompile(`\s+`)
	textContent = reSpace.ReplaceAllString(textContent, " ")
	textContent = strings.TrimSpace(textContent)
	return ok(textContent)
}

func HandleExtractMeta(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	metaName, _ :=getString(args, "name")
	property, _ :=getString(args, "property")
	if htmlContent == "" {
		return err("html parameter is required")
}

	var metaRegex *regexp.Regexp
	if metaName != "" {
		metaRegex = regexp.MustCompile(fmt.Sprintf(`<meta\s+[^>]*name=["']%s["'][^>]*content=["']([^"']*)["']`, regexp.QuoteMeta(metaName)))")
	} else if property != "" {
		metaRegex = regexp.MustCompile(fmt.Sprintf(`<meta\s+[^>]*property=["']%s["'][^>]*content=["']([^"']*)["']`, regexp.QuoteMeta(property)))")
	} else {
		return err("either name or property parameter must be specified")
}

	match := metaRegex.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return ok("")
}

	return ok(match[1])
}

func HandleCheckURLStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	client := http.DefaultClient,
	}
	req, reqErr := http.NewRequestWithContext(ctx, "HEAD", targetURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to check URL status: %v", fetchErr))
}

	defer resp.Body.Close()
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"final_url":   resp.Request.URL.String(),
	}
	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(jsonResult))
}