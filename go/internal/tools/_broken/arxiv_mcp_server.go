package tools

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// arXiv API response structures (simplified)
type feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []entry  `xml:"entry"`
}

type entry struct {
	Title       string   `xml:"title"`
	ID          string   `xml:"id"`
	Summary     string   `xml:"summary"`
	Published   string   `xml:"published"`
	Updated     string   `xml:"updated"`
	Authors     []author `xml:"author"`
	LinkPDF     link     `xml:"link[@title='pdf']"`
	LinkAltPDF  link     `xml:"link[@type='application/pdf']"`
	PrimaryCat  string   `xml:"arxiv:primary_category"`
	Comment     string   `xml:"arxiv:comment"`
	JournalRef  string   `xml:"arxiv:journal_ref"`
	Doi         string   `xml:"arxiv:doi"`
}

type author struct {
	Name string `xml:"name"`
}

type link struct {
	Href string `xml:"href,attr"`
}

// HTTP client with timeout
var http.DefaultClient = http.DefaultClient

// HandleSearch searches arXiv for papers matching a query.
// Expected args:
//   "query": string (required)
//   "max_results": int (optional, default 10)
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
	}
	max, _ :=getInt(args, "max_results")
	if max <= 0 {
		max = 10
	}

	// Build request URL
	values := url.Values{}
	values.Set("search_query", query)
	values.Set("start", "0")
	values.Set("max_results", strconv.Itoa(max))
	apiURL := fmt.Sprintf("http://export.arxiv.org/api/query?%s", values.Encode())

	req, apiErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("arXiv API returned status %d", resp.StatusCode))
	}
	body, apiErr := io.ReadAll(resp.Body)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	return ok(string(body))
}

// HandleGetPaperInfo retrieves metadata for a specific arXiv ID.
// Expected args:
//   "id": string (required) – e.g., "2301.00001"
func HandleGetPaperInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required argument: id")
	}
	// arXiv API for a single ID
	values := url.Values{}
	values.Set("id_list", id)
	apiURL := fmt.Sprintf("http://export.arxiv.org/api/query?%s", values.Encode())

	req, apiErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("arXiv API returned status %d", resp.StatusCode))
	}
	body, apiErr := io.ReadAll(resp.Body)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	// Parse XML to extract a readable summary
	var f feed
	if parseErr := xml.Unmarshal(body, &f); parseErr != nil {
		return err(parseErr.Error())
	}
	if len(f.Entries) == 0 {
		return err("no entry found for given ID")
	}
	e := f.Entries[0]
	authors := []string{}
	for _, a := range e.Authors {
		authors = append(authors, a.Name)

	summary := strings.TrimSpace(e.Summary)
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}
	info := fmt.Sprintf("Title: %s\nAuthors: %s\nPublished: %s\nSummary: %s",
		strings.TrimSpace(e.Title),
		strings.Join(authors, ", "),
		e.Published,
		summary,
	)
	return ok(info)
}

}

// HandleDownloadPDF downloads the PDF of a given arXiv ID to a temporary file.
// Expected args:
//   "id": string (required)
func HandleDownloadPDF(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required argument: id")
	}
	// Build PDF URL
	pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", id)

	req, apiErr := http.NewRequestWithContext(ctx, http.MethodGet, pdfURL, nil)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to download PDF, status %d", resp.StatusCode))
	}
	// Create temp file
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, fmt.Sprintf("%s.pdf", strings.ReplaceAll(id, "/", "_")))
	outFile, apiErr := os.Create(filePath)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	defer outFile.Close()
	_, apiErr = io.Copy(outFile, resp.Body)
	if apiErr != nil {
		return err(apiErr.Error())
	}
	return ok(fmt.Sprintf("PDF saved to %s", filePath))
}

// HandleListCategories returns a static list of popular arXiv categories.
func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	categories := []string{
		"cs.AI", "cs.LG", "cs.CV", "cs.CL", "cs.RO",
		"math.CO", "math.NT", "math.AG", "math.ST",
		"astro-ph", "cond-mat", "hep-ph", "quant-ph",
		"q-bio.BM", "q-bio.CB", "q-bio.GN",
	}
	return ok(strings.Join(categories, ", "))
}

// HandleCitationNetwork is a placeholder that returns a simple message.
// Expected args:
//   "id": string (required)
func HandleCitationNetwork(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing required argument: id")
	}
	// In a full implementation this would query Semantic Scholar.
	// Here we return a deterministic placeholder.
	msg := fmt.Sprintf("Citation network for %s would be generated here.", id)
	return ok(msg)
}