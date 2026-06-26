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
	"sort"
	"strings"
	"time"
)

func HandleGetImageInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("missing 'url' argument")
}

	parsedURL, apiErr := url.Parse(imageURL)
	if apiErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", apiErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "imagesorcery-mcp/1.0")
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", respErr))
}

	defer resp.Body.Close()
	const maxBody = 1024
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	info := map[string]interface{}{
		"url":            imageURL,
		"status_code":    resp.StatusCode,
		"content_type":   resp.Header.Get("Content-Type"),
		"content_length": resp.ContentLength,
		"body_preview":   string(body),
	}
	infoJSON, jsonErr := json.Marshal(info)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal info: %v", jsonErr))
}

	return ok(string(infoJSON))
}

func HandleListImages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "directory")
	if dirPath == "" {
		dirPath = "."
	}
	stat, statErr := os.Stat(dirPath)
	if statErr != nil {
		return err(fmt.Sprintf("directory does not exist: %v", statErr))
}

	if !stat.IsDir() {
		return err("path is not a directory")
}

	entries, readErr := os.ReadDir(dirPath)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read directory: %v", readErr))
}

	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}
	var images []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if imageExts[ext] {
			images = append(images, name)

	}
	sort.Strings(images)
	imagesJSON, jsonErr := json.Marshal(images)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal image list: %v", jsonErr))
}

	return ok(string(imagesJSON))
}

}

func HandleResizeImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("missing 'url' argument")
}

	width, _ :=getInt(args, "width")
	height, _ :=getInt(args, "height")
	if width <= 0 && height <= 0 {
		return err("at least one of 'width' or 'height' must be positive")
}

	parsedURL, apiErr := url.Parse(imageURL)
	if apiErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", apiErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "imagesorcery-mcp/1.0")
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", respErr))
}

	defer resp.Body.Close()
	const maxBody = 1024
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	result := map[string]interface{}{
		"url":         imageURL,
		"width":       width,
		"height":      height,
		"status_code": resp.StatusCode,
		"body_preview": string(bodyLEC),
	}
	resultJSON, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(resultJSON))
}

func HandleConvertImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("missing 'url' argument")
}

	format, _ :=getString(args, "format")
	if format == "" {
		return err("missing 'format' argument")
}

	parsedURL, apiErr := url.Parse(imageURL)
	if apiErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", apiErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "imagesorcery-mcp/1.0")
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", respErr))
}

	defer resp.Body.Close()
	const maxBody = 1024
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response body: %v", readErr))
}

	result := map[string]interface{}{
		"url":         imageURL,
		"format":      format,
		"status_code": resp.StatusCode,
		"body_preview": string(body),
	}
	resultJSON, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(resultJSON))
}

func HandleDownloadImage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("missing 'url' argument")
}

	outputPath, _ :=getString(args, "output_path")
	if outputPath == "" {
		return err("missing 'output_path' argument")
}

	parsedURL, apiErr := url.Parse(imageURL)
	if apiErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", apiErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "imagesorcery-mcp/1.0")
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", respErr))
}

	defer resp.Body.Close()
	file, fileErr := os.Create(outputPath)
	if fileErr != nil {
		return err(fmt.Sprintf("failed to create output file: %v", fileErr))
}

	defer file.Close()
	_, copyErr := io.Copy(file, resp.Body)
	if copyErr != nil {
		return err(fmt.Sprintf("failed to write file: %v", copyErr))
}

	result := map[string]interface{}{
		"url":         imageURL,
		"output_path": outputPath,
		"status":      "downloaded",
	}
	resultJSON, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(resultJSON))
}

func HandleValidateImageURL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imageURL, _ :=getString(args, "url")
	if imageURL == "" {
		return err("missing 'url' argument")
}

	parsedURL, apiErr := url.Parse(imageURL)
	if apiErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", apiErr))
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodHead, parsedURL.String(), nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "imagesorcery-mcp/1.0")
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", respErr))
}

	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	isImage := strings.HasPrefix(contentType, "image/")
	result := map[string]interface{}{
		"url":          imageURL,
		"valid":        resp.StatusCode == http.StatusOK,
		"is_image":     isImage,
		"content_type": contentType,
		"status_code":  resp.StatusCode,
	}
	resultJSON, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal result: %v", jsonErr))
}

	return ok(string(resultJSON))
}