package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func newClient() *http.Client {
	return &http.Client{
}
		Timeout: 30 * time.Second,
	}
}

// callAPI performs a GET request to the given URL and returns the body as bytes.
// The variable is named apiErr to avoid shadowing the err("error") function.
func callAPI(reqURL string) ([]byte, error) {
	client := newClient()
	resp, apiErr := client.Get(reqURL)
	if apiErr != nil {
		return nil, fmt.Errorf("http request failed: %w", apiErr)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
}

	return body, nil
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// HandleSearchPapers searches for academic papers using the Semantic Scholar API.
// Argument: query (string, required), limit (int, optional, default 10).
// Returns a JSON array of paper summaries.
func HandleSearchPapers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument 'query'")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	baseURL := "https://api.semanticscholar.org/graph/v1/paper/search"
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("fields", "title,url,abstract,authors,year,venue,externalIds")
	fullURL := baseURL + "?" + params.Encode()

	body, fetchErr := callAPI(fullURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("search failed: %v", fetchErr))
}

	// Parse the response to extract a clean summary
	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse search response: %v", parseErr))
}

	papers, found := result["data"].([]interface{})
	if !ok || len(papers) == 0 {
		return ok("No results found.")
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d papers:\n\n", len(papers)))
	for i, p := range papers {
		paper, found := p.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := paper["title"].(string)
		year, _ := paper["year"].(float64)
		venue, _ := paper["venue"].(string)
		externalIDs, _ := paper["externalIds"].(map[string]interface{})
		arxivID, _ := externalIDs["ArXiv"].(string)
		sb.WriteString(fmt.Sprintf("%d. %s", i+1, title))
		if arxivID != "" {
			sb.WriteString(fmt.Sprintf(" (arXiv:%s)", arxivID))

		if year > 0 {
			sb.WriteString(fmt.Sprintf(" (%d)", int(year)))

		if venue != "" {
			sb.WriteString(fmt.Sprintf(" - %s", venue))

		sb.WriteString("\n")

	return ok(sb.String())
}

}
}
}

// HandleGetPaper retrieves detailed information about a specific paper.
// Arguments: paper_id (string, required) – a Semantic Scholar ID (e.g. "649def34f8be52c8b66281af98ae884c09aef38b")
// or an ArXiv ID (e.g. "2106.12345").
func HandleGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperID, _ :=getString(args, "paper_id")
	if paperID == "" {
		return err("missing required argument 'paper_id'")
}

	// Determine whether the ID is an ArXiv ID (contains a dot) or a Semantic Scholar ID.
	// For ArXiv IDs we use the /hash/arxiv:XXXX endpoint, for S2 IDs we use the /hash/XXXX endpoint.
	var idParam string
	if strings.Contains(paperID, ".") {
		idParam = "arxiv:" + paperID
	} else if strings.HasPrefix(paperID, "arxiv:") {
		idParam = paperID
	} else {
		idParam = paperID
	}

	baseURL := "https://api.semanticscholar.org/graph/v1/paper/"
	fullURL := baseURL + strings.TrimPrefix(idParam, "arxiv:")
	fullURL += "?fields=title,url,abstract,authors,year,venue,citations,references,externalIds"

	body, fetchErr := callAPI(fullURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("fetch paper failed: %v", fetchErr))
}

	var paper map[string]interface{}
	if parseErr := json.Unmarshal(body, &paper); parseErr != nil {
		return err(fmt.Sprintf("failed to parse paper response: %v", parseErr))
}

	title, _ := paper["title"].(string)
	if title == "" {
		return ok("Paper not found.")
}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Title: %s\n", title))
	if year, found := paper["year"].(float64); ok && year > 0 {
		sb.WriteString(fmt.Sprintf("Year: %d\n", int(year)))

	if venue, found := paper["venue"].(string); ok && venue != "" {
		sb.WriteString(fmt.Sprintf("Venue: %s\n", venue))

	if abstract, found := paper["abstract"].(string); ok && abstract != "" {
		sb.WriteString(fmt.Sprintf("Abstract: %s\n", abstract[:min(len(abstract), 500)]))

	if authors, found := paper["authors"].([]interface{}); found {
		var names []string
		for _, a := range authors {
			if m, found := a.(map[string]interface{}); found {
				if name, found := m["name"].(string); found {
					names = append(names, name)

			}
		}
		if len(names) > 0 {
			sb.WriteString(fmt.Sprintf("Authors: %s\n", strings.Join(names, ", ")))

	}
	if citations, found := paper["citations"].([]interface{}); found {
		sb.WriteString(fmt.Sprintf("Citation count: %d\n", len(citations)))

	if references, found := paper["references"].([]interface{}); found {
		sb.WriteString(fmt.Sprintf("Reference count: %d\n", len(references)))

	if externalIDs, found := paper["externalIds"].(map[string]interface{}); found {
		if ai, found := externalIDs["ArXiv"].(string); found {
			sb.WriteString(fmt.Sprintf("ArXiv: %s\n", ai))

	}
	url, _ := paper["url"].(string)
	if url != "" {
		sb.WriteString(fmt.Sprintf("URL: %s\n", url))

	return ok(sb.String())
}

}
}
}
}
}
}
}
}
}

// min helper for strings (Go 1.20+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}