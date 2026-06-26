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

func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	parsedURL, urlErr := url.Parse(targetURL)
	if urlErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", urlErr))
}

	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
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

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	return ok(string(body))
}

func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	if htmlContent == "" {
		return err("html parameter is required")
}

	linkRegex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href=(["'])(.*?)\1`)")
	matches := linkRegex.FindAllStringSubmatch(htmlContent, -1)

	var links []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		link := match[2]
		if link != "" && !seen[link] {
			seen[link] = true
			links = append(links, link)

	}

	return ok(strings.Join(links, "\n"))
}

}

func HandleExtractEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	textContent, _ :=getString(args, "text")
	if textContent == "" {
		return err("text parameter is required")
}

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(textContent, -1)

	uniqueEmails := make(map[string]bool)
	for _, email := range emails {
		uniqueEmails[email] = true
	}

	var result []string
	for email := range uniqueEmails {
		result = append(result, email)

	return ok(strings.Join(result, "\n"))
}

}

func HandleExtractPhoneNumbers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	textContent, _ :=getString(args, "text")
	if textContent == "" {
		return err("text parameter is required")
}

	// Basic international phone number pattern
	phoneRegex := regexp.MustCompile(`(?:\+?(\d{1,3}))?[-. (]*(\d{3})[-. )]*(\d{3})[-. ]*(\d{4})`)
	matches := phoneRegex.FindAllStringSubmatch(textContent, -1)

	uniquePhones := make(map[string]bool)
	for _, match := range matches {
		phone := strings.Join(match[1:], "")
		if phone != "" {
			uniquePhones[phone] = true
		}
	}

	var result []string
	for phone := range uniquePhones {
		result = append(result, phone)

	return ok(strings.Join(result, "\n"))
}

}

func HandleCleanHTML(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	if htmlContent == "" {
		return err("html parameter is required")
}

	// Remove script tags
	scriptRegex := regexp.MustCompile(`<script\b[^>]*>(.*?)<\/script>`)
	cleaned := scriptRegex.ReplaceAllString(htmlContent, "")

	// Remove style tags
	styleRegex := regexp.MustCompile(`<style\b[^>]*>(.*?)<\/style>`)
	cleaned = styleRegex.ReplaceAllString(cleaned, "")

	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	cleaned = commentRegex.ReplaceAllString(cleaned, "")

	// Remove extra whitespace
	whitespaceRegex := regexp.MustCompile(`\s+`)
	cleaned = whitespaceRegex.ReplaceAllString(cleaned, " ")

	return ok(strings.TrimSpace(cleaned))
}

func HandleExtractMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	htmlContent, _ :=getString(args, "html")
	if htmlContent == "" {
		return err("html parameter is required")
}

	titleRegex := regexp.MustCompile(`<title>(.*?)<\/title>`)
	titleMatch := titleRegex.FindStringSubmatch(htmlContent)
	title := ""
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	metaRegex := regexp.MustCompile(`<meta\s+([^>]*?)>`)
	metaMatches := metaRegex.FindAllStringSubmatch(htmlContent, -1)

	metadata := make(map[string]string)
	metadata["title"] = title

	for _, match := range metaMatches {
		if len(match) < 2 {
			continue
		}
		attrs := match[1]
		nameRegex := regexp.MustCompile(`name=["'](.*?)["']`)
		propRegex := regexp.MustCompile(`property=["'](.*?)["']`)
		contentRegex := regexp.MustCompile(`content=["'](.*?)["']`)

		nameMatch := nameRegex.FindStringSubmatch(attrs)
		propMatch := propRegex.FindStringSubmatch(attrs)
		contentMatch := contentRegex.FindStringSubmatch(attrs)

		var key string
		if len(nameMatch) > 1 {
			key = nameMatch[1]
		} else if len(propMatch) > 1 {
			key = propMatch[1]
		} else {
			continue
		}

		if len(contentMatch) > 1 {
			metadata[key] = contentMatch[1]
		}
	}

	jsonData, jsonErr := json.Marshal(metadata)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal metadata: %v", jsonErr))
}

	return ok(string(jsonData))
}