package tools

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client = http.DefaultClient

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	response := fmt.Sprintf("Information for %s", name)
	return ok(response)
}

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	status, _ :=getString(args, "status")
	if status == "" {
		return err("status is required")
}

	response := fmt.Sprintf("Current status: %s", status)
	return ok(response)
}

func HandleSetConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	response := fmt.Sprintf("Configuration set: %s = %s", key, value)
	return ok(response)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}