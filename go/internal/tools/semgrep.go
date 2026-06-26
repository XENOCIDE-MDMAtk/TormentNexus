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
	"strings"
	"time"
)

func runSemgrep(ctx context.Context, args []string, workDir string) (string, int, error) {
	bin, lookupErr := exec.LookPath("semgrep")
	if lookupErr != nil {
		return "", -1, fmt.Errorf("semgrep not found in PATH: %v", lookupErr)
}

	cmd := exec.CommandContext(ctx, bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = os.Environ()
	out, runErr := cmd.CombinedOutput()
	exitCode := 0
	if runErr != nil {
		if exitErr, found := runErr.(*exec.ExitError); found {
			exitCode = exitErr.ExitCode()
		} else {
			return string(out), -1, runErr
		}
	}
	return string(out), exitCode, nil
}

func HandleSemgrepScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	config, _ :=getString(args, "config")
	language, _ :=getString(args, "language")
	rules, _ :=getString(args, "rules")
	severity, _ :=getString(args, "severity")
	jsonOutput, _ :=getBool(args, "json")
	autofix, _ :=getBool(args, "autofix")
	dryRun, _ :=getBool(args, "dry_run")

	cmdArgs := []string{"scan", "--no-git", path}
	if config != "" {
		cmdArgs = append(cmdArgs, "--config", config)

	if language != "" {
		cmdArgs = append(cmdArgs, "--lang", language)

	if rules != "" {
		cmdArgs = append(cmdArgs, "--rule", rules)

	if severity != "" {
		cmdArgs = append(cmdArgs, "--severity", severity)

	if jsonOutput {
		cmdArgs = append(cmdArgs, "--json")

	if autofix {
		cmdArgs = append(cmdArgs, "--autofix")

	if dryRun {
		cmdArgs = append(cmdArgs, "--dryrun")

	output, exitCode, runErr := runSemgrep(ctx, cmdArgs, "")
	if runErr != nil {
		return err(fmt.Sprintf("failed to run semgrep scan: %v", runErr))
}

	if jsonOutput && exitCode == 0 {
		var result map[string]interface{}
		parseErr := json.Unmarshal([]byte(output), &result)
		if parseErr != nil {
			return ok(fmt.Sprintf("Semgrep scan completed (raw output):\n%s", output))
}

		formatted, fmtErr := json.MarshalIndent(result, "", "  ")
		if fmtErr != nil {
			return ok(fmt.Sprintf("Semgrep scan completed:\n%s", output))
}

		return ok(fmt.Sprintf("Semgrep scan completed (JSON):\n%s", string(formatted)))
}

	summary := "Semgrep scan completed"
	if exitCode != 0 {
		summary = fmt.Sprintf("Semgrep scan completed with findings (exit code %d)", exitCode)

	return ok(fmt.Sprintf("%s\n\n%s", summary, output))
}

}
}
}
}
}
}
}
}

func HandleSemgrepVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	output, _, runErr := runSemgrep(ctx, []string{"--version"}, "")
	if runErr != nil {
		return err(fmt.Sprintf("failed to get semgrep version: %v", runErr))
}

	version := strings.TrimSpace(output)
	return ok(fmt.Sprintf("Semgrep version: %s", version))
}

func HandleSemgrepRegistrySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	language, _ :=getString(args, "language")
	category, _ :=getString(args, "category")

	params := url.Values{}
	params.Set("q", query)
	if language != "" {
		params.Set("lang", language)

	if category != "" {
		params.Set("category", category)

	apiURL := fmt.Sprintf("https://semgrep.dev/api/v1/registry?%s", params.Encode())

	client := http.DefaultClient
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to search registry: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("registry search failed (HTTP %d): %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return ok(fmt.Sprintf("Registry search raw response:\n%s", string(body)))
}

	formatted, fmtErr := json.MarshalIndent(result, "", "  ")
	if fmtErr != nil {
		return ok(fmt.Sprintf("Registry search completed:\n%s", string(body)))
}

	return ok(fmt.Sprintf("Semgrep registry search results:\n%s", string(formatted)))
}

}
}

func HandleSemgrepLint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	config, _ :=getString(args, "config")
	if config == "" {
		config = "auto"
	}
	strict, _ :=getBool(args, "strict")
	jsonOutput, _ :=getBool(args, "json")

	cmdArgs := []string{"scan", "--no-git", "--config", config, path}
	if strict {
		cmdArgs = append(cmdArgs, "--strict")

	if jsonOutput {
		cmdArgs = append(cmdArgs, "--json")

	output, exitCode, runErr := runSemgrep(ctx, cmdArgs, "")
	if runErr != nil {
		return err(fmt.Sprintf("failed to run semgrep lint: %v", runErr))
}

	if jsonOutput && exitCode == 0 {
		var result map[string]interface{}
		parseErr := json.Unmarshal([]byte(output), &result)
		if parseErr != nil {
			return ok(fmt.Sprintf("Semgrep lint completed (raw output):\n%s", output))
}

		formatted, fmtErr := json.MarshalIndent(result, "", "  ")
		if fmtErr != nil {
			return ok(fmt.Sprintf("Semgrep lint completed:\n%s", output))
}

		return ok(fmt.Sprintf("Semgrep lint completed (JSON):\n%s", string(formatted)))
}

	summary := "Semgrep lint completed — no findings"
	if exitCode != 0 {
		summary = fmt.Sprintf("Semgrep lint found issues (exit code %d)", exitCode)

	return ok(fmt.Sprintf("%s\n\n%s", summary, output))
}
}
}
}