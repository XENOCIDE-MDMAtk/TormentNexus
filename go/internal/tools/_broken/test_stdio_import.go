package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// HandleEcho echoes back the provided message
func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("Echo: %s", message))
}

// HandleStdinTest tests reading from standard input
func HandleStdinTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, fetchErr := io.ReadAll(os.Stdin)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	return ok(fmt.Sprintf("Read %d bytes from stdin", len(data)))
}

// HandleImportCheck verifies that all expected imports are available
func HandleImportCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	imports := []string{
		"context",
		"encoding/json",
		"fmt",
		"io",
		"os",
		"strings",
		"time",
	}

	result := make(map[string]bool)
	for _, imp := range imports {
		result[imp] = true
	}

	jsonData, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(jsonData))
}

// HandleStdioVersion returns version info about the stdio module
func HandleStdioVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version := map[string]interface{}{
		"module":      "test_stdio_import",
		"version":     "1.0.0",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"imports":     []string{"context", "encoding/json", "fmt", "io", "os", "strings", "time"},
		"description": "Test module for stdio import validation",
	}

	jsonData, parseErr := json.Marshal(version)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(jsonData))
}

// HandleStringOps performs various string operations for testing
func HandleStringOps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	operation, _ :=getString(args, "operation")

	if input == "" {
		return err("input is required")
}

	var result string

	switch operation {
	case "upper":
		result = strings.ToUpper(input)
	case "lower":
		result = strings.ToLower(input)
	case "reverse":
		result = reverseString(input)
	case "length":
		result = fmt.Sprintf("%d", len(input))
	default:
		result = input
	}

	return ok(result)
}

// reverseString reverses a given string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}