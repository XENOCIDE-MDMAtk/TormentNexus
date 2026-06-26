package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"os/exec"
)

// HandleHTTPProbe performs an HTTP GET request to a given URL and returns status and response.
func HandleHTTPProbe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing required argument: url")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	start := time.Now()
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", fetchErr))
}

	defer resp.Body.Close()

	elapsed := time.Since(start)
	result := fmt.Sprintf("Status: %s\nTime: %v\n", resp.Status, elapsed)
	return ok(result)
}

// HandlePing executes the ping command against a host and returns its output.
func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("missing required argument: host")
}

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", host)
	output, pingErr := cmd.CombinedOutput()
	if pingErr != nil {
		// Include output even if command failed (e.g., host unreachable)
		return err(fmt.Sprintf("ping failed: %v\n%s", pingErr, string(output)))
}

	return ok(strings.TrimSpace(string(output)))
}