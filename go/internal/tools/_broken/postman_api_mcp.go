package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func HandleSendRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("url is required")
}

	headers, _ :=getString(args, "headers")
	body, _ :=getString(args, "body")
	
	client := http.DefaultClient
	
	req, reqErr := http.NewRequest(method, urlStr, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	// Parse headers
	if headers != "" {
		headerLines := strings.Split(headers, "\n")
		for _, line := range headerLines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))

		}
	}
	
	if body != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		req.Body = io.NopCloser(strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()
	
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
		"body":        string(respBody),
	}
	
	return ok(result)
}
}
}