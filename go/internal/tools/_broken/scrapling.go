package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// HandleFetchURL fetches content from a URL
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	// Validate URL
	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return err("invalid URL format")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Scrapling-MCP/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: status code %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	content := string(body)
	// Limit content length for response
	if len(content) > 50000 {
		content = content[:50000] + "\n... (truncated)"
	}

	return ok(fmt.Sprintf("Fetched %d bytes from %s\n\nStatus: %d\nContent-Type: %s\n\n--- Content ---\n%s",
}
		len(body), targetURL, resp.StatusCode, resp.Header.Get("Content-Type"), content))

// HandleExtractLinks extracts all links from HTML content
func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	html, _ :=getString(args, "html")
	baseURL, _ :=getString(args, "base_url")

	if html == "" {
		return err("html parameter is required")
}

	// Pattern to match href attributes
	hrefPattern := regexp.MustCompile(`(?i)href\s*=\s*["']?([^"'\s>]+)`)
	matches := hrefPattern.FindAllStringSubmatch(html, -1)

	links := make([]string, 0)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			link := strings.TrimSpace(match[1])
			if link != "" && !strings.HasPrefix(link, "javascript:") && !strings.HasPrefix(link, "#") {
				// Resolve relative URLs if base_url provided
				if baseURL != "" && !strings.HasPrefix(link, "http") {
					absoluteURL, resolveErr := resolveRelativeURL(baseURL, link)
					if resolveErr == nil {
						link = absoluteURL
					}
				}
				if !seen[link] {
					seen[link] = true
					links = append(links, link)

			}
		}
	}

	result := fmt.Sprintf("Found %d unique links:\n\n", len(links))
	for i, link := range links {
		result += fmt.Sprintf("%d. %s\n", i+1, link)

	return ok(result)
}

}
}

// HandleExtractText extracts text content from HTML
func HandleExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	html, _ :=getString(args, "html")
	selector, _ :=getString(args, "selector")

	if html == "" {
		return err("html parameter is required")
}

	var result string

	if selector != "" {
		// Simple CSS selector simulation - extract elements by tag name
		result = extractBySelector(html, selector)
	} else {
		// Strip all HTML tags
		result = stripHTMLTags(html)

	// Clean up whitespace
	result = cleanWhitespace(result)

	return ok(result)
}

}

// HandleExtractElements extracts HTML elements matching a selector
func HandleExtractElements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	html, _ :=getString(args, "html")
	selector, _ :=getString(args, "selector")

	if html == "" {
		return err("html parameter is required")
}

	if selector == "" {
		return err("selector parameter is required")
}

	elements := extractBySelector(html, selector)

	if elements == "" {
		return ok("No elements found matching selector: " + selector)
}

	_ = countOccurrences(elements, "<"+strings.Split(selector, " ")[0])
	return ok(fmt.Sprintf("Found elements:\n\n%s", elements))
}

// HandleScrapePage performs a complete scrape operation
func HandleScrapePage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	cssSelector, _ :=getString(args, "selector")
	extractAttr, _ :=getString(args, "attribute")

	if targetURL == "" {
		return err("url parameter is required")
}

	// Validate URL
	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return err("invalid URL format")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Scrapling-MCP/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: status code %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	html := string(body)

	var result string
	if cssSelector != "" {
		elements := extractBySelector(html, cssSelector)
		if extractAttr != "" {
			result = extractAttribute(elements, extractAttr)
		} else {
			result = stripHTMLTags(elements)

		result = cleanWhitespace(result)
	} else {
		result = stripHTMLTags(html)
		result = cleanWhitespace(result)

	return ok(fmt.Sprintf("Scraped from %s:\n\n%s", targetURL, result))
}

}
}

// Helper functions

func stripHTMLTags(html string) string {
	// Remove script and style content first
	scriptPattern := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	html = scriptPattern.ReplaceAllString(html, "")

	stylePattern := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	html = stylePattern.ReplaceAllString(html, "")

	// Remove all HTML tags
	tagPattern := regexp.MustCompile(`<[^>]+>`)
	return tagPattern.ReplaceAllString(html, " ")
}

func cleanWhitespace(s string) string {
	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	s = spacePattern.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func extractBySelector(html string, selector string) string {
	// Simple selector support for common cases
	selector = strings.TrimSpace(selector)

	// Handle tag selectors like "div", "span", "p", "a", etc.
	tagPattern := regexp.MustCompile(`(?i)<` + selector + `[^>]*>(.*?)</` + selector + `>`)
	matches := tagPattern.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		// Try with class selector like ".classname"
		if strings.HasPrefix(selector, ".") {
			className := selector[1:]
			classPattern := regexp.MustCompile(`(?i)<[^>]+class\s*=\s*["'][^"']*` + regexp.QuoteMeta(className) + `[^"']*["'][^>]*>(.*?)</[^>]+>`)
			matches = classPattern.FindAllStringSubmatch(html, -1)

		// Try with id selector like "#idname"
		if strings.HasPrefix(selector, "#") {
			idName := selector[1:]
			idPattern := regexp.MustCompile(`(?i)<[^>]+id\s*=\s*["']` + regexp.QuoteMeta(idName) + `["'][^>]*>(.*?)</[^>]+>`)
			matches = idPattern.FindAllStringSubmatch(html, -1)

	}

	var results []string
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[0])

	}

	if len(results) == 0 {
		return ""
	}

	return strings.Join(results, "\n\n")
}

}
}
}

func extractAttribute(html string, attr string) string {
	attrPattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(attr) + `\s*=\s*["']([^"']+)["']`)")
	matches := attrPattern.FindAllStringSubmatch(html, -1)

	var values []string
	for _, match := range matches {
		if len(match) > 1 {
			values = append(values, match[1])

	}

	return strings.Join(values, "\n")
}

}

func resolveRelativeURL(baseURL, relative string) (string, error) {
	base, baseErr := url.Parse(baseURL)
	if baseErr != nil {
		return "", baseErr
	}
	rel, relErr := base.Parse(relative)
	if relErr != nil {
		return "", relErr
	}
	return rel.String(), nil
}

func countOccurrences(s, substr string) int {
	count := 0
	for {
		idx := strings.Index(s, substr)
		if idx == -1 {
			break
		}
		count++
		s = s[idx+1:]
	}
	return count
}