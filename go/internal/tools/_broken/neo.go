package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ToolResponse is defined in parity.go, assumed available.
// ok() and err("error") are defined in parity.go, assumed available.
// getString, getInt, getBool are defined in parity.go, assumed available.

// neoStatus checks the operational status of the Neo.mjs ecosystem.
func HandleNeoStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyFiles := []string{
		".agents/ANTIGRAVITY_RULES.md",
		".agents/skills/architecture-pre-flight/SKILL.md",
		".agents/skills/blocked-task-state/SKILL.md",
		".agents/skills/blog-post/SKILL.md",
	}

	missing := []string{}
	existing := []string{}

	for _, file := range keyFiles {
		if _, e := os.Stat(file); os.IsNotExist(e) {
			missing = append(missing, file)
		} else {
			existing = append(existing, file)

	}

	sort.Strings(missing)
	sort.Strings(existing)

	status := "healthy"
	if len(missing) > 0 {
		status = "degraded"
	}

	result := map[string]interface{}{
		"status":      status,
		"timestamp":   time.Now().Format(time.RFC3339),
		"files_found": len(existing),
		"files_missing": len(missing),
		"missing_files": missing,
		"verified_files": existing,
	}

	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonResult))
}

}

// neoRoute determines the appropriate architectural skill or workflow based on a query.
func HandleNeoRoute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	lowerQuery := strings.ToLower(query)

	selectedDiscipline := ""
	reason := ""
	blastRadius := ""

	if strings.Contains(lowerQuery, "blog") || strings.Contains(lowerQuery, "post") || strings.Contains(lowerQuery, "seo") {
		selectedDiscipline = "blog-post"
		reason = "Query explicitly mentions public-facing content or SEO."
		blastRadius = "Public Artifact"
	} else if strings.Contains(lowerQuery, "blocked") || strings.Contains(lowerQuery, "input") || strings.Contains(lowerQuery, "failed") {
		selectedDiscipline = "blocked-task-state"
		reason = "Query indicates a task state transition requiring A2A signaling."
		blastRadius = "Internal State"
	} else if strings.Contains(lowerQuery, "file") || strings.Contains(lowerQuery, "placement") || strings.Contains(lowerQuery, ".mjs") {
		selectedDiscipline = "structural-pre-flight"
		reason = "Query relates to file placement or structural changes."
		blastRadius = "Local Scope"
	} else if strings.Contains(lowerQuery, "skill") || strings.Contains(lowerQuery, "create") {
		selectedDiscipline = "create-skill"
		reason = "Query explicitly mentions skill creation."
		blastRadius = "Substrate Expansion"
	} else {
		selectedDiscipline = "architecture-pre-flight"
		reason = "Query spans multiple domains or lacks specific triggers, requiring high-level arbitration."
		blastRadius = "Cross-Substrate"
	}

	result := map[string]interface{}{
		"query":             query,
		"selected_discipline": selectedDiscipline,
		"reasoning":         reason,
		"blast_radius":      blastRadius,
		"next_action":       fmt.Sprintf("Invoke %s workflow", selectedDiscipline),
	}

	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonResult))
}

// neoFetch retrieves content from a URL, simulating the WebFetch capability mentioned in the blog guide.
func HandleNeoFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rawURL, _ :=getString(args, "url")
	if rawURL == "" {
		return err("url parameter is required")
}

	parsedURL, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return err("only http and https schemes are allowed")
}

	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Neo-MJS-Agent/1.0")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("received non-200 status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	maxLen := 10000
	content := string(body)
	if len(content) > maxLen {
		content = content[:maxLen] + "... [truncated]"
	}

	result := map[string]interface{}{
		"url":      rawURL,
		"status":   resp.Status,
		"content":  content,
		"length":   len(body),
	}

	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonResult))
}

// neoGitDiff checks for formatting noise as mandated by the Anti-Reformatting Protocol.
func HandleNeoGitDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "path")

	cmd := exec.CommandContext(ctx, "git", "diff", "--stat")
	if filePath != "" {
		cmd = exec.CommandContext(ctx, "git", "diff", "--stat", filePath)

	cmd.Dir = "."

	output, runErr := cmd.Output()
	if runErr != nil {
		if exitErr, found := runErr.(*exec.ExitError); found {
			return err(exitErr.Stderr)
}

		return err(runErr.Error())
}

	diffStat := string(output)

	isClean := strings.TrimSpace(diffStat) == ""

	result := map[string]interface{}{
		"command":      "git diff --stat",
		"output":       diffStat,
		"is_clean":     isClean,
		"recommendation": func() string {
			if isClean {
				return "No changes detected. Safe to proceed."
			}
			return "Changes detected. Verify against Anti-Reformatting Protocol."
		}(),
	}

	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonResult))
}

}

// neoValidateSource verifies external claims as per the Blog Post Authoring Guide.
func HandleNeoValidateSource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	claim, _ :=getString(args, "claim")
	sourceURL, _ :=getString(args, "source_url")

	if claim == "" || sourceURL == "" {
		return err("both 'claim' and 'source_url' are required")
}

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", sourceURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Neo-MJS-Agent/1.0")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("source returned status %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	content := string(body)

	claimWords := strings.Fields(claim)
	found := 0
	for _, word := range claimWords {
		if len(word) > 3 && strings.Contains(strings.ToLower(content), strings.ToLower(word)) {
			found++
		}
	}

	coverage := float64(found) / float64(len(claimWords))
	isVerified := coverage > 0.5

	result := map[string]interface{}{
		"claim":           claim,
		"source_url":      sourceURL,
		"content_preview": content[:min(200, len(content))],
		"match_coverage":  coverage,
		"is_verified":     isVerified,
		"recommendation": func() string {
			if isVerified {
				return "Claim appears supported by the source."
			}
			return "Claim not clearly supported. Verify manually or cut the claim."
		}(),
	}

	jsonResult, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonResult))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}