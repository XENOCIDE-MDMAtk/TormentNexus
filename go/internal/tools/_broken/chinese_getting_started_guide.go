package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func HandleWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
	}

	apiKey := os.Getenv("ZHIPU_API_KEY")
	if apiKey == "" {
		return err("ZHIPU_API_KEY environment variable is not set")
	}

	requestBody := map[string]interface{}{
		"tool": "web-search-pro",
		"messages": []map[string]string{
			{"role": "user", "content": query},
		},
		"stream": false,
	}

	jsonBody, jsonErr := json.Marshal(requestBody)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", jsonErr))
	}

	req, reqErr := http.NewRequestWithContext(ctx, "POST",
		"https://open.bigmodel.cn/api/paas/v4/tools",
		bytes.NewBuffer(jsonBody))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if respErr != nil {
		return err(fmt.Sprintf("request failed: %v", respErr))
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
	}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
	}

	choices, found := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ok("No search results found")
	}

	var searchResults []string
	for _, choice := range choices {
		choiceMap, found := choice.(map[string]interface{})
		if !found {
			continue
		}
		message, found := choiceMap["message"].(map[string]interface{})
		if !found {
			continue
		}
		toolCalls, found := message["tool_calls"].([]interface{})
		if !found {
			continue
		}
		for _, toolCall := range toolCalls {
			toolCallMap, found := toolCall.(map[string]interface{})
			if !found {
				continue
			}
			searchResult, found := toolCallMap["search_result"].([]interface{})
			if !found {
				continue
			}
			for _, item := range searchResult {
				itemMap, found := item.(map[string]interface{})
				if !found {
					continue
				}
				content, found := itemMap["content"].(string)
				if found {
					searchResults = append(searchResults, content)

			}
		}
	}

	if len(searchResults) == 0 {
		return ok("No search results found")
	}

	resultText := strings.Join(searchResults, "\n\n\n")
	return ok(resultText)
}

}

func HandleGetMCPInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	info := `Model Context Protocol (MCP) Programming Quick Start Guide

MCP is an innovative open-source protocol that redefines how large language models (LLM) interact with the external world. MCP provides a standardized method for any LLM to easily connect to various data sources and tools, enabling seamless information access and processing.

Core Features of MCP:
- Resources: Structured data that can be queried by clients
- Prompts: Pre-defined prompt templates
- Tools: Executable functions that models can call
- Sampling: Allows servers to request sampling from clients
- Roots: Directories the server can access
- Transports: Communication protocols (stdio and SSE)

This server implements the following tools:
1. web_search - Search the internet for information using Zhipu AI's web search API

Usage:
- Set ZHIPU_API_KEY environment variable with your Zhipu AI API key
- Call web_search with a query parameter to search the internet
- Results include summarized content from search results`

	return ok(info)
}