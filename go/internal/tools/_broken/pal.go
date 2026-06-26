package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	http.DefaultClient = http.DefaultClient
)

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func HandlePalindromeCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	cleaned := strings.ToLower(strings.ReplaceAll(text, " ", ""))
	re := regexp.MustCompile(`[^a-z0-9]`)
	cleaned = re.ReplaceAllString(cleaned, "")

	if cleaned == reverseString(cleaned) {
		return ok(fmt.Sprintf("'%s' is a palindrome", text))
}

	return ok(fmt.Sprintf("'%s' is not a palindrome", text))
}

func HandleReverseText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	reversed := reverseString(text)
	return ok(fmt.Sprintf("Reversed text: %s", reversed))
}

func HandleCountWords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	words := strings.Fields(text)
	count := len(words)
	return ok(fmt.Sprintf("Word count: %d", count))
}

func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("url parameter is required")
}

	parsedUrl, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		return err(fmt.Sprintf("invalid URL: %v", parseErr))
}

	resp, fetchErr := http.DefaultClient.Get(parsedUrl.String())
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch URL: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	var content struct {
		Text string `json:"text"`
	}
	readErr := json.NewDecoder(resp.Body).Decode(&content)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	return ok(fmt.Sprintf("Fetched content: %s", content.Text))
}

func HandleExtractEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(text, -1)

	if len(emails) == 0 {
		return ok("No emails found in the text")
}

	return ok(fmt.Sprintf("Found emails: %v", strings.Join(emails, ", ")))
}

func HandleTitleCase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text parameter is required")
}

	words := strings.Fields(text)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])

	}

	titleCase := strings.Join(words, " ")
	return ok(fmt.Sprintf("Title case: %s", titleCase))
}
}