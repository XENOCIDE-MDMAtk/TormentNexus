package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func HandleFetchPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageURL, _ :=getString(args, "url")
	if pageURL == "" {
		return err("missing required parameter: url")
}

	_, parseErr := url.Parse(pageURL)
	if parseErr != nil {
		return err("invalid URL: " + parseErr.Error())
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Get(pageURL)
	if fetchErr != nil {
		return err("failed to fetch page: " + fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("failed to read response body: " + readErr.Error())
}

	return ok(string(body))
}

func HandleSearchContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageURL, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	if pageURL == "" {
		return err("missing required parameter: url")
}

	if query == "" {
		return err("missing required parameter: query")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Get(pageURL)
	if fetchErr != nil {
		return err("failed to fetch page: " + fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("failed to read response body: " + readErr.Error())
}

	content := string(body)
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	if strings.Contains(lowerContent, lowerQuery) {
		return ok(fmt.Sprintf("Found '%s' on page %s", query, pageURL))
}

	return ok(fmt.Sprintf("Content '%s' not found on page %s", query, pageURL))
}

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "dir")
	if dir == "" {
		dir = "."
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		return err("failed to read directory: " + readErr.Error())
}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())

	sort.Strings(files)
	return ok(strings.Join(files, "\n"))
}

}

func HandleFileExists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing required parameter: path")
}

	absPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return err("invalid path: " + absErr.Error())
}

	_, statErr := os.Stat(absPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return ok(fmt.Sprintf("File does not exist: %s", absPath))
}

		return err("failed to check file: " + statErr.Error())
}

	return ok(fmt.Sprintf("File exists: %s", absPath))
}

func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageURL, _ :=getString(args, "url")
	if pageURL == "" {
		return err("missing required parameter: url")
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Get(pageURL)
	if fetchErr != nil {
		return err("failed to fetch page: " + fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err("failed to read response body: " + readErr.Error())
}

	re := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>`)")
	matches := re.FindAllStringSubmatch(string(body), -1)

	var links []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			links = append(links, match[1])

	}

	sort.Strings(links)
	if len(links) == 0 {
		return ok("No links found on page")
}

	return ok(fmt.Sprintf("Found %d links on %s:\n%s", len(links), pageURL, strings.Join(links, "\n")))
}

}

func HandleURLInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pageURL, _ :=getString(args, "url")
	if pageURL == "" {
		return err("missing required parameter: url")
}

	parsedURL, parseErr := url.Parse(pageURL)
	if parseErr != nil {
		return err("invalid URL: " + parseErr.Error())
}

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Head(pageURL)
	if fetchErr != nil {
		return err("failed to fetch URL info: " + fetchErr.Error())
}

	defer resp.Body.Close()

	var size int64
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		size, _ = strconv.ParseInt(contentLength, 10, 64)

	info := fmt.Sprintf("URL: %s\nScheme: %s\nHost: %s\nPath: %s\nStatus: %d %s\nContent-Type: %s\nContent-Length: %d bytes",
		pageURL,
		parsedURL.Scheme,
		parsedURL.Host,
		parsedURL.Path,
		resp.StatusCode,
		resp.Status,
		resp.Header.Get("Content-Type"),
		size,
	)

	return ok(info)
}
}