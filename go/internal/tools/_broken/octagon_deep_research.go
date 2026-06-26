package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// HandleSearch performs a web search using DuckDuckGo
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	searchURL := fmt.Sprintf("https://duckduckgo.com/?q=%s&format=json", url.QueryEscape(query))

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ResearchBot/1.0)")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleScrape fetches and extracts content from a URL
func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url parameter is required")
}

	parsedURL, parseErr := url.Parse(targetURL)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		targetURL = parsedURL.String()

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ResearchBot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	content := string(body)
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.TrimSpace(content)

	if len(content) > 10000 {
		content = content[:10000] + "...[truncated]"
	}

	result := map[string]interface{}{
		"url":     targetURL,
		"content": content,
		"length":  len(content),
	}

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

}

// HandleAnalyze performs analysis on provided text content
func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)")
	phoneRegex := regexp.MustCompile(`\+?[\d\s\-\(\)]{10,}`)

	emails := emailRegex.FindAllString(text, -1)
	urls := urlRegex.FindAllString(text, -1)
	phones := phoneRegex.FindAllString(text, -1)

	wordCount := len(strings.Fields(text))
	charCount := len(text)

	result := map[string]interface{}{
		"word_count": wordCount,
		"char_count": charCount,
		"emails":     emails,
		"urls":       urls,
		"phones":     phones,
	}

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

// HandleQuery processes a research query and returns structured information
func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question parameter is required")
}

	contextText, _ :=getString(args, "context")

	result := map[string]interface{}{
		"question": question,
		"context":  contextText,
		"status":   "processed",
		"message":  "Query received and logged for research processing",
	}

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

// HandleSummarize generates a summary of the provided content
func HandleSummarize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	if content == "" {
		return err("content parameter is required")
}

	maxLength, _ :=getInt(args, "max_length")
	if maxLength == 0 {
		maxLength = 500
	}

	sentences := strings.Split(content, ".")
	var summarySentences []string
	currentLength := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		sentence = sentence + "."
		if currentLength+len(sentence) <= int(maxLength) {
			summarySentences = append(summarySentences, sentence)
			currentLength += len(sentence)

		if currentLength >= int(maxLength) {
			break
		}
	}

	summary := strings.Join(summarySentences, " ")
	if summary == "" {
		summary = content
		if len(summary) > int(maxLength) {
			summary = summary[:maxLength] + "..."
		}
	}

	result := map[string]interface{}{
		"original_length": len(content),
		"summary":         summary,
		"summary_length":  len(summary),
	}

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}
}