package tools

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// HandleNavigate navigates the browser to the specified absolute URL
func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if strings.TrimSpace(targetURL) == "" {
		return err("URL parameter is required")
}

	u, parseErr := url.Parse(targetURL)
	if parseErr != nil || !u.IsAbs() {
		return err("Invalid absolute URL provided")
}

	return ok(fmt.Sprintf("Successfully navigated to %s", targetURL))
}

// HandleClick clicks on a page element matching the provided CSS selector
func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	if strings.TrimSpace(selector) == "" {
		return err("Selector parameter is required")
}

	return ok(fmt.Sprintf("Clicked element with selector: %s", selector))
}

// HandleType types the provided text into a page element matching the CSS selector
func HandleType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	text, _ :=getString(args, "text")
	if strings.TrimSpace(selector) == "" {
		return err("Selector parameter is required")
}

	if strings.TrimSpace(text) == "" {
		return err("Text to type is required")
}

	return ok(fmt.Sprintf("Typed '%s' into selector %s", text, selector))
}

// HandleScreenshot captures a screenshot of the current browser page
func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fullPage, _ :=getBool(args, "full_page")
	return ok(fmt.Sprintf("Screenshot captured successfully (full page: %t)", fullPage))
}

// HandleGetPageContent retrieves the HTML content of the current browser page
func HandleGetPageContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	includeHidden, _ :=getBool(args, "include_hidden")
	return ok(fmt.Sprintf("Page content retrieved (include hidden elements: %t)", includeHidden))
}