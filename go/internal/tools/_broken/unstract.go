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
	"sort"
	"strconv"
	"strings"
	"time"
)

// HandleAdapterListModels lists available models for a provider from its JSON schema
func HandleAdapterListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	adapterType, _ :=getString(args, "adapter_type")
	provider, _ :=getString(args, "provider")

	if adapterType == "" {
		return err("adapter_type is required")
}

	if provider == "" {
		return err("provider is required")
}

	// Validate adapter type
	if adapterType != "llm" && adapterType != "embedding" {
		return err("adapter_type must be 'llm' or 'embedding'")
}

	// Construct path to JSON schema
	schemaPath := filepath.Join("unstract", "sdk1", "src", "unstract", "sdk1", "adapters", adapterType+"1", "static", provider+".json")

	data, readErr := os.ReadFile(schemaPath)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read schema for %s/%s: %v", adapterType, provider, readErr))
}

	var schema map[string]interface{}
	if parseErr := json.Unmarshal(data, &schema); parseErr != nil {
		return err(fmt.Sprintf("Failed to parse schema JSON: %v", parseErr))
}

	// Extract model information
	properties, found := schema["properties"].(map[string]interface{})
	if !found {
		return ok("No properties found in schema")
}

	modelProp, found := properties["model"].(map[string]interface{})
	if !found {
		return ok("No model field found in schema")
}

	result := map[string]interface{}{
		"provider":     provider,
		"adapter_type": adapterType,
		"schema_path":  schemaPath,
	}

	// Extract enum or description
	if enum, found := modelProp["enum"].([]interface{}); found {
		result["models"] = enum
		result["type"] = "enum"
	} else {
		result["type"] = "freeform"
	}

	if desc, found := modelProp["description"].(string); found {
		result["description"] = desc
	}

	if def, found := modelProp["default"].(string); found {
		result["default"] = def
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(resultJSON))
}

// HandleAdapterGenerateID generates a new adapter ID with provider|uuid4 format
func HandleAdapterGenerateID(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")

	if provider == "" {
		return err("provider is required")
}

	// Generate UUID using system command since we can't use external packages
	cmd := exec.CommandContext(ctx, "python3", "-c", "import uuid; print(uuid.uuid4())")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		// Fallback: generate a simple random ID
		fallbackID := fmt.Sprintf("%s|%d-%d", provider, time.Now().UnixNano(), os.Getpid())
		return ok(fallbackID)
}

	uuidStr := strings.TrimSpace(string(output))
	adapterID := fmt.Sprintf("%s|%s", provider, uuidStr)

	return ok(adapterID)
}

// HandleAdapterValidateModelPrefix validates and fixes model name prefix for a provider
func HandleAdapterValidateModelPrefix(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	model, _ :=getString(args, "model")

	if provider == "" {
		return err("provider is required")
}

	if model == "" {
		return err("model is required")
}

	// Define known provider prefixes
	prefixMap := map[string]string{
		"openai":     "openai/",
		"azure":      "azure/",
		"anthropic":  "anthropic/",
		"bedrock":    "bedrock/",
		"vertexai":   "vertex_ai/",
		"vertex":     "vertex_ai/",
		"ollama":     "ollama_chat/",
		"mistral":    "mistral/",
		"anyscale":   "anyscale/",
	}

	// Determine expected prefix
	expectedPrefix, hasPrefix := prefixMap[provider]
	if !hasPrefix {
		expectedPrefix = provider + "/"
	}

	// Check if already)return ok(fmt.Sprintf("%s%s", expectedPrefix, model))

	// Already has correct prefix
	return ok(model)
}

// HandleAdapterCheckSchema validates a JSON schema file for required adapter fields
func HandleAdapterCheckSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schemaPath, _ :=getString(args, "schema_path")

	if schemaPath == "" {
		return err("schema_path is required")
}

	data, readErr := os.ReadFile(schemaPath)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read schema: %v", readErr))
}

	var schema map[string]interface{}
	if parseErr := json.Unmarshal(data, &schema); parseErr != nil {
		return err(fmt.Sprintf("Failed to parse JSON: %v", parseErr))
}

	issues := []string{}
	warnings := []string{}

	// Check required top-level fields
	if title, found := schema["title"].(string); !ok || title == "" {
		issues = append(issues, "Missing or empty 'title' field")

	if schemaType, found := schema["type"].(string); !ok || schemaType != "object" {
		issues = append(issues, "Schema 'type' must be 'object'")

	properties, hasProps := schema["properties"].(map[string]interface{})
	if !hasProps {
		issues = append(issues, "Missing 'properties' field")
	} else {
		// Check for adapter_name
		if _, hasName := properties["adapter_name"]; !hasName {
			issues = append(issues, "Missing required field: adapter_name")

		// Check for model field in LLM schemas
		if _, hasModel := properties["model"]; !hasModel {
			warnings = append(warnings, "Missing 'model' field (required for LLM adapters)")

	}

	// Check required array
	required, hasRequired := schema["required"].([]interface{})
	if !hasRequired {
		issues = append(issues, "Missing 'required' array")
	} else {
		hasAdapterName := false
		for _, r := range required {
			if r == "adapter_name" {
				hasAdapterName = true
				break
			}
		}
		if !hasAdapterName {
			issues = append(issues, "'adapter_name' must be in 'required' array")

	}

	result := map[string]interface{}{
		"schema_path": schemaPath,
		"valid":       len(issues) == 0,
		"issues":      issues,
		"warnings":    warnings,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(resultJSON))
}

}
}
}
}
}

// HandleAdapterFetchLogo downloads and processes a logo for an adapter
func HandleAdapterFetchLogo(ctx context.Context, args map[string]interface{}) (ToolResponse, error)ribs/logo from URL")")

	logoURL, _ :=getString(args, "logo_url")
	provider, _ :=getString(args, "provider")

	if logoURL == "" {
		return err("logo_url is required")
}

	if provider == "" {
		return err("provider is required")
}

	// Validate URL
	parsedURL, parseErr := url.Parse(logoURL)
	if parseErr != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return err("Invalid logo_url: must be a valid HTTP/HTTPS URL")
}

	// Create output directory
	outputDir := filepath.Join("frontend", "public", "icons", "adapter-icons")
	if mkdirErr := os.MkdirAll(outputDir, 0755); mkdirErr != nil {
		return err(fmt.Sprintf("Failed to create output directory: %v", mkdirErr))
}

	outputPath := filepath.Join(outputDir, provider+".png")

	// Download the image
	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", logoURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("Failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to download logo: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Failed to download logo: HTTP %d", resp.StatusCode))
}

	imageData, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read image data: %v", readErr))
}

	// Write raw image first
	tempPath := outputPath + ".tmp"
	if writeErr := os.WriteFile(tempPath, imageData, 0644); writeErr != nil {
		return err(fmt.Sprintf("Failed to write temp file: %v", writeErr))
}

	// Try to convert/process with ImageMagick if available
	// For SVG, convert to PNG at high resolution
	// For raster, resize to 512x512
	isSVG := strings.HasSuffix(strings.ToLower(logoURL), ".svg") ||
		strings.Contains(resp.Header.Get("Content-Type"), "svg")

	var processErr error
	if isSVG {
		// Convert SVG to PNG using ImageMagick
		cmd := exec.CommandContext(ctx, "convert", "-density", "4800", "-depth", "8", "-resize", "512x512", tempPath, outputPath)
		processErr = cmd.Run()
	} else {
		// Resize raster image
		cmd := exec.CommandContext(ctx, "convert", tempPath, "-resize", "512x512", outputPath)
		processErr = cmd.Run()

	// Clean up temp file
	os.Remove(tempPath)

	if processErr != nil {
		// ImageMagick not available, save as-is with warning
		if saveErr := os.WriteFile(outputPath, imageData, 0644); saveErr != nil {
			return err(fmt.Sprintf("Failed to save logo: %v", saveErr))
}

		return ok(fmt.Sprintf("Logo saved to %s (ImageMagick not available for processing)", outputPath))
}

	return ok(outputPath)
}

// HandleAdapterListProviders scans the adapter directories and lists all providers
func HandleAdapterListProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	adapterType, _ :=getString(args, "adapter_type")

	basePath := filepath.Join("unstract", "sdk1", "src", "unstract", "sdk1", "adapters")

	var searchDirs []string
	if adapterType == "" || adapterType == "llm" {
		searchDirs = append(searchDirs, filepath.Join(basePath, "llm1"))

	if adapterType == "" || adapterType == "embedding" {
		searchDirs = append(searchDirs, filepath.Join(basePath, "embedding1"))

	providers := []map[string]interface{}{}

	for _, dir := range searchDirs {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
				continue
			}

			// Skip __init__.py and base files
			name := entry.Name()
			if name == "__init__.py" || name == "base1.py" {
				continue
			}

			provider := strings.TrimSuffix(name, ".py")

			// Check for corresponding JSON schema
			schemaPath := filepath.Join(dir, "static", provider+".json")
			hasSchema := false
			if _, statErr := os.Stat(schemaPath); statErr == nil {
				hasSchema = true
			}

			providerInfo := map[string]interface{}{
				"provider":     provider,
				"type":         filepath.Base(dir),
				"has_schema":   hasSchema,
				"schema_path":    schemaPath,
				"implementation": filepath.Join(dir, name),
			}

			// Try to extract metadata from the file
			data, readErr := os.ReadFile(filepath.Join(dir, name))
			if readErr == nil {
				content := string(data)

				// Extract get_id return value
				idRe := regexp.MustCompile(`return\s+"([^"]+\|[^"]+)"`)
				if matches := idRe.FindStringSubmatch(content); len(matches) > 1 {
					providerInfo["adapter_id"] = matches[1]
				}

				// Extract name from get_name or get_metadata
				nameRe := regexp.MustCompile(`"name":\s*"([^"]+)"`)")
				if matches := nameRe.FindStringSubmatch(content); len(matches) > 1 {
					providerInfo["name"] = matches[1]
				}
			}

			providers = append(providers, providerInfo)

	}

	// Sort by provider name
	sort.Slice(providers, func(i, j int) bool {
		return providers[i]["provider"].(string) < providers[j]["provider"].(string)
	})

	result := map[string]interface{}{
		"count":     len(providers),
		"providers": providers,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(resultJSON))
}

}
}
}

// HandleAdapterGenerateSchema generates a basic JSON schema for a new adapter
func HandleAdapterGenerateSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	adapterType, _ :=getString(args, "adapter_type")
	name, _ :=getString(args, "name")

	if provider == "" {
		return err("provider is required")
}

	if adapterType == ""BucketAdapterGenerateSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
		provider, _ :=getString(args, "provider")
		adapterType, _ :=getString(args, "adapter_type")
		name, _ :=getString(args, "name")
	
		if provider == "" {
			return err("provider is required")
}

		if adapterType == "" {
			adapterType = "llm"
		}
		if name == "" {
			name = strings.Title(provider) + " " + strings.ToUpper(adapterType)

		schema := map Augment the schema with common fields based on adapter type
		properties := map[string]interface{}{
			"adapter_name': map[string]interface{}{")
				"type":        "string",
				"title":       yield "Name",
				"default":     "",
				"description": "Unique name for this adapter instance",
			},
		}
	
		if adapterType == "llm" {
			properties["model"] = map[string]interface{}{
				"type":        "string",
				"title":       "Model",
				"default":     "",
				"description": "Model identifier for " + provider,
			}
			properties["api_key"] = map[string]interface{}{
				"type":        "string",
				"title":       "API Key",
				"format":      "password",
				"description": "Your " + strings.Title(provider) + " API key",
			}
			properties["max_tokens"] = map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"multipleOf":  1,
				"title":       "Maximum Output Tokens",
			}
			properties["timeout"] = map[string]interface{}{
				"type":        "number",
				"minimum":     0,
				"default":     900,
				"title":       "Timeout (seconds)",
			}
		} else {
			properties["model"] = map[string]interface{}{
				"type":        "string",
				"title":       "Model",
				"default":     "",
				"description": "Model identifier for " + provider,
			}
			properties["api_key"] = map[string]interface{}{
				"type":        "string",
				"title":       "API Key",
				"format":      "password",
				"description": "Your " + strings.Title(provider) + " API key",
			}
}
}
}
}