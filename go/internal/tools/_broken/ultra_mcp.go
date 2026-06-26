package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// HandleSearchWeb searches the web for a given query using a simple web search API
func HandleSearchWeb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	// Use DuckDuckGo HTML scraping as a simple search mechanism
	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.0")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	// Extract results using regex
	resultPattern := regexp.MustCompile(`<a[^>]+class="result__a"[^>]*>([^<]+)`)
	matches := resultPattern.FindAllStringSubmatch(string(body), -1)

	var results []string
	for i, match := range matches {
		if i >= 5 {
			break
		}
		if len(match) > 1 {
			title := strings.TrimSpace(match[1])
			title = strings.ReplaceAll(title, "&amp;", "&")
			title = strings.ReplaceAll(title, "&lt;", "<")
			title = strings.ReplaceAll(title, "&gt;", ">")
			results = append(results, fmt.Sprintf("%d. %s", i+1, title))

	}

	if len(results) == 0 {
		return ok("No results found for query: " + query)
}

	return ok("Search results for '" + query + "':\n" + strings.Join(results, "\n"))
}

}

// HandleFetchURL fetches content from a given URL
func HandleFetchURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	// Validate URL
	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err("invalid URL: " + parseErr.Error())
}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return err("URL must use http or https scheme")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; UltraMCP/1.0)")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("HTTP error %d: %s", resp.StatusCode, resp.Status))
}

	// Limit reading to prevent memory issues
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if readErr != nil {
		return err(readErr.Error())
}

	content := string(body)
	if len(content) > 10000 {
		content = content[:10000] + "\n... (truncated)"
	}

	return ok(content)
}

// HandleRunCommand executes a shell command with safety restrictions
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command parameter is required")
}

	// Security: block dangerous commands
	dangerousPatterns := []string{
		"rm -rf /", "rm -rf /*", "mkfs", "dd if=", ">:", ">:/",
		"shutdown", "reboot", "halt", "poweroff",
	}
	lowerCmd := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			return err("command contains dangerous pattern: " + pattern)

	}

	// Determine shell to use
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)

	// Set timeout via context is handled by CommandContext
	output, runErr := cmd.CombinedOutput()

	result := string(output)
	if runErr != nil {
		result = fmt.Sprintf("Error: %s\nOutput: %s", runErr.Error(), result)

	if len(result) > 10000 {
		result = result[:10000] + "\n... (truncated)"
	}

	return ok(result)
}

}
}
}

// runtimeGOOS returns the current OS
func runtimeGOOS() string {
	return os.Getenv("GOOS")
}

// HandleReadFile reads content from a file path
func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "path")
	if filePath == "" {
		return err("path parameter is required")
}

	// Clean and validate path
	cleanPath := filepath.Clean(filePath)

	// Security: prevent directory traversal
	if strings.Contains(cleanPath, "..") {
		return err("path cannot contain parent directory references")
}

	// Check if file exists and is not a directory
	info, statErr := os.Stat(cleanPath)
	if statErr != nil {
		return err("cannot access file: " + statErr.Error())
}

	if info.IsDir() {
		return err("path is a directory, not a file")
}

	// Limit file size
	if info.Size() > 10*1024*1024 {
		return err("file too large (max 10MB)")
}

	content, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return err(readErr.Error())
}

	result := string(content)
	if len(result) > 10000 {
		result = result[:10000] + "\n... (truncated)"
	}

	return ok(result)
}

// HandleWriteFile writes content to a file path
func HandleWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "path")
	content, _ :=getString(args, "content")

	if filePath == "" {
		return err("path parameter is required")
}

	// Clean and validate path
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return err("path cannot contain parent directory references")
}

	// Create parent directories if needed
	dir := filepath.Dir(cleanPath)
	if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
		return err("cannot create directory: " + mkdirErr.Error())
}

	writeErr := os.WriteFile(cleanPath, []byte(content), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), cleanPath))
}

// HandleListDirectory lists files in a directory
func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	if dirPath == "" {
		dirPath = "."
	}

	// Clean and validate path
	cleanPath := filepath.Clean(dirPath)
	if strings.Contains(cleanPath, "..") {
		return err("path cannot contain parent directory references")
}

	info, statErr := os.Stat(cleanPath)
	if statErr != nil {
		return err("cannot access directory: " + statErr.Error())
}

	if !info.IsDir() {
		return err("path is not a directory")
}

	entries, readErr := os.ReadDir(cleanPath)
	if readErr != nil {
		return err(readErr.Error())
}

	var files []string
	for _, entry := range entries {
		prefix := "  "
		if entry.IsDir() {
			prefix = "[D]"
		} else {
			prefix = "[F]"
		}
		files = append(files, fmt.Sprintf("%s %s", prefix, entry.Name()))

	// Sort for consistent output
	sort.Strings(files)

	if len(files) == 0 {
		return ok("Directory is empty: " + cleanPath)
}

	result := fmt.Sprintf("Contents of %s (%d items):\n", cleanPath, len(files))
	result += strings.Join(files, "\n")

	return ok(result)
}

}

// HandleGetWeather gets weather information for a location
func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location parameter is required")
}

	// Use wttr.in as a free weather API
	encodedLocation := url.QueryEscape(location)
	weatherURL := fmt.Sprintf("https://wttr.in/%s?format=j1", encodedLocation)

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "curl/7.64.1")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	// Parse JSON response
	var weatherData map[string]interface{}
	if jsonErr := json.Unmarshal(body, &weatherData); jsonErr != nil {
		return err("failed to parse weather data: " + jsonErr.Error())
}

	// Extract current condition
	currentArray, ok1 := weatherData["current_condition"].([]interface{})
	if !ok1 || len(currentArray) == 0 {
		return err("unexpected weather data format")
}

	current, ok2 := currentArray[0].(map[string]interface{})
	if !ok2 {
		return err("unexpected weather data format")
}

	// Build result
	var parts []string
	parts = append(parts, fmt.Sprintf("Weather for: %s", location))

	if tempC, found := current["temp_C"]; found {
		parts = append(parts, fmt.Sprintf("Temperature: %s°C", tempC))

	if tempF, found := current["temp_F"]; found {
		parts = append(parts, fmt.Sprintf("(%s°F)", tempF))

	if desc, found := current["weatherDesc"].([]interface{}); ok && len(desc) > 0 {
		if descMap, found := desc[0].(map[string]interface{}); found {
			if value, found := descMap["value"].(string); found {
				parts = append(parts, fmt.Sprintf("Conditions: %s", value))

		}
	}
	if humidity, found := current["humidity"]; found {
		parts = append(parts, fmt.Sprintf("Humidity: %s%%", humidity))

	if wind, found := current["windspeedKmph"]; found {
		parts = append(parts, fmt.Sprintf("Wind: %s km/h", wind))

	return ok(strings.Join(parts, "\n"))
}
}
}
}
}
}