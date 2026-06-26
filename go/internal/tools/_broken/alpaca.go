package tools

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HandleAlpacaFact returns a random fact about alpacas
func HandleAlpacaFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	facts := []string{
		"Alpacas are social animals that live in herds.",
		"Alpacas have three stomachs to digest their food.",
		"Alpacas communicate through body language and humming sounds.",
		"Alpaca fiber is softer than cashmere and warmer than wool.",
		"Alpacas are environmentally friendly grazers that don't pull grass up by the roots.",
		"Alpacas can live up to 20 years.",
		"Alpacas are related to camels and llamas.",
		"Alpacas come in 22 natural colors recognized in the US.",
	}
	// Simple deterministic selection based on current time
	index := int(time.Now().UnixNano()) % len(facts)
	return ok(facts[index])
}

// HandleAlpacaCount counts occurrences of "alpaca" (case-insensitive) in text
func HandleAlpacaCount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return ok("0")
	}
	// Case-insensitive regex for "alpaca"
	re := regexp.MustCompile(`(?i)alpaca`)
	count := len(re.FindAllString(text, -1))
	return ok(strconv.Itoa(count))
}

// HandleAlpacaSearch searches for alpaca-related terms in a query
func HandleAlpacaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	// List of alpaca-related terms
	terms := []string{"alpaca", "llama", "camelid", "vicuna", "guanaco", "fiber", "wool", "herd", "humming"}
	queryLower := strings.ToLower(query)
	matches := []string{}
	for _, term := range terms {
		if strings.Contains(queryLower, term) {
			matches = append(matches, term)

	}
	if len(matches) == 0 {
		return ok("No alpaca-related terms found")
	}
	result := "Found " + strconv.Itoa(len(matches)) + " alpaca-related term(s): " + strings.Join(matches, ", ")
	return ok(result)
}

}

// HandleAlpacaUrl encodes a string for use in an alpaca-related URL path
func HandleAlpacaUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	segment, _ :=getString(args, "segment")
	if segment == "" {
		return err("segment parameter is required")
}

	// Clean and encode the segment for URL path use
	segment = strings.TrimSpace(segment)
	segment = strings.ToLower(segment)
	// Replace spaces with hyphens, remove non-alphanumeric (except hyphens)
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	clean := re.ReplaceAllString(segment, "")
	clean = strings.Trim(clean, "-")
	if clean == "" {
		return err("segment contains no valid characters after cleaning")
}

	// URL encode the clean segment
	u := url.Values{}
	u.Set("path", clean)
	encoded := u.Encode()
	// Remove the "path=" prefix for just the encoded value
	encoded = strings.TrimPrefix(encoded, "path=")
	result := "/alpaca/" + encoded
	return ok(result)
}

// HandleAlpacaValidate checks if a string is a valid alpaca name
func HandleAlpacaValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	// Alpaca names should be 2-30 characters, letters, spaces, hyphens
	re := regexp.MustCompile(`^[a-zA-Z\s\-]{2,30}$`)
	if !re.MatchString(name) {
		return ok("false")
	}
	// Check for reserved words
	reserved := []string{"admin", "root", "system", "null"}
	nameLower := strings.ToLower(name)
	for _, r := range reserved {
		if nameLower == r {
			return ok("false")
		}
	}
	return ok("true")
}