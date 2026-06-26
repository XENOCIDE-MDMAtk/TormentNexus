package tools

import (
	"context"
	"fmt"
	"time"
)

func HandleNavigate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	return ok(fmt.Sprintf("Navigated to %s", url))
}

func HandleGoBack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Navigated back in browser history")
}

func HandleGoForward(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Navigated forward in browser history")
}

func HandleWait(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	seconds, _ :=getInt(args, "time")
	if seconds <= 0 {
		seconds = 1
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	return ok(fmt.Sprintf("Waited for %d seconds", seconds))
}

func HandlePressKey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key parameter is required")
}

	return ok(fmt.Sprintf("Pressed key: %s", key))
}

func HandleGetConsoleLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Console logs retrieved successfully")
}

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Screenshot captured")
}

func HandleSnapshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Page snapshot captured")
}

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	element, _ :=getString(args, "element")
	if element == "" {
		return err("element parameter is required")
}

	return ok(fmt.Sprintf("Clicked on element: %s", element))
}

func HandleHover(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	element, _ :=getString(args, "element")
	if element == "" {
		return err("element parameter is required")
}

	return ok(fmt.Sprintf("Hovered over element: %s", element))
}

func HandleType(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	element, _ :=getString(args, "element")
	text, _ :=getString(args, "text")
	if element == "" {
		return err("element parameter is required")
}

	if text == "" {
		return err("text parameter is required")
}

	return ok(fmt.Sprintf("Typed '%s' into element: %s", text, element))
}

func HandleSelectOption(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	element, _ :=getString(args, "element")
	if element == "" {
		return err("element parameter is required")
}

	return ok(fmt.Sprintf("Selected option in element: %s", element))
}

func HandleDrag(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	startElement, _ :=getString(args, "startElement")
	endElement, _ :=getString(args, "endElement")
	if startElement == "" || endElement == "" {
		return err("startElement and endElement parameters are required")
}

	return ok(fmt.Sprintf("Dragged from %s to %s", startElement, endElement))
}