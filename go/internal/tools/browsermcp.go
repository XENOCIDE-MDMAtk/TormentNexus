package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	http.DefaultClient = http.DefaultClient
)

func HandleOpenURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	if !isValidURL(url) {
		return err("invalid URL format")
}

	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "start"
	} else if runtime.GOOS == "linux" {
		cmd = "xdg-open"
	}

	apiErr := exec.Command(cmd, url).Start()
	if apiErr != nil {
		return err(fmt.Sprintf("failed to open URL: %v", apiErr))
}

	return ok(fmt.Sprintf("Opened URL: %s", url))
}

func HandleFetchPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	if !isValidURL(url) {
		return err("invalid URL format")
}

	resp, fetchErr := http.DefaultClient.Get(url)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	return ok(string(body))
}

func HandleExtractLinks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	if !isValidURL(url) {
		return err("invalid URL format")
}

	resp, fetchErr := http.DefaultClient.Get(url)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	linkRegex := regexp.MustCompile(`href=["']([^"']+)["']`)")
	links := linkRegex.FindAllStringSubmatch(string(body), -1)

	var result strings.Builder
	for _, match := range links {
		if len(match) > 1 {
			result.WriteString(match[1])
			result.WriteString("\n")

	}

	return ok(result.String())
}

}

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	if !isValidURL(url) {
		return err("invalid URL format")
}

	cmd := exec.Command("google-chrome", "--headless", "--screenshot", url)
	output, apiErr := cmd.CombinedOutput()
	if apiErr != nil {
		return err(fmt.Sprintf("failed to take screenshot: %v", apiErr))
}

	tempFile, tempErr := os.CreateTemp("", "screenshot-*.png")
	if tempErr != nil {
		return err(fmt.Sprintf("failed to create temp file: %v", tempErr))
}

	defer tempFile.Close()

	_, writeErr := tempFile.Write(output)
	if writeErr != nil {
		return err(fmt.Sprintf("failed to write screenshot: %v", writeErr))
}

	return ok(fmt.Sprintf("Screenshot saved to: %s", tempFile.Name()))
}

func HandleDownloadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	if !isValidURL(url) {
		return err("invalid URL format")
}

	parsedURL, parseErr := url.Parse(url)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", parseErr))
}

	filename := filepath.Base(parsedURL.Path)
	if filename == "" {
		filename = "downloaded-file"
	}

	tempFile, tempErr := os.CreateTemp("", filename)
	if tempErr != nil {
		return err(fmt.Sprintf("failed to create temp file: %v", tempErr))
}

	defer tempFile.Close()

	resp, fetchErr := http.DefaultClient.Get(url)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to download file: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	_, copyErr := io.Copy(tempFile, resp.Body)
	if copyErr != nil {
		return err(fmt.Sprintf("failed to save file: %v", copyErr))
}

	return ok(fmt.Sprintf("File downloaded to: %s", tempFile.Name()))
}

func isValidURL(str string) bool {
	parsed, parseErr := url.ParseRequestURI(str)
	if parseErr != nil {
		return false
	}

	return parsed.Scheme != "" && parsed.Host != ""
}