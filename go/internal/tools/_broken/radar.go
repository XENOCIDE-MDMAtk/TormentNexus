package tools, then the imports without those two, then all the functions as before, right?

Wait let's make sure: the original code's functions are all there, right? Let's list them: HandleRadarBuild, HandleRadarTest, HandleRadarTypecheck, HandleRadarStart, HandleRadarStop, HandleRadarAPI, HandleRadarDiff, HandleRadarStatus, findRepoRoot, boolToYesNo. All present.

Wait let's check for any other possible errors. Oh, in HandleRadarStart, when we do cmd.Start(), that's correct. The health check part: if fetchErr is not nil, it returns ok, that's allowed? Wait the rule says return ok("text") if e != nil return err(e.Error()), but that's just a rule for when there's an error, right? Wait no, the rule 3 says "return ok("text"), if e != nil { return err(e.Error()) }" — wait, no, that's probably a guideline for when to use ok vs e, but in the original code, when the health check fails, it returns ok saying started but health pending, that's intentional, right? Because the process did start, just health check failed, so that's not an error, so ok is correct. That's not a compilation error.

Wait another thing: in HandleRadarStatus, the line _ = proc — that's to avoid unused variable, which is fine, no error there.

Wait let's check the imports again: after removing net/url and regexp, the import block is:

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

Yes, that's all stdlib, no unused ones now.

Wait is there any other compilation error? Let's see: all functions are declared correctly, variables are used, no missing imports. Let's check: getInt is used in HandleRadarStart, HandleRadarAPI, that's fine. getString used in HandleRadarAPI, fine. getBool? Wait wait, the pre-defined has getBool, but is it used here? Wait the original code doesn't use getBool, but that's okay, right? Because the pre-defined functions are available, we don't have to use all of them. The error was only about unused imports, not unused functions. Oh right, the errors listed by reviewers are only the two unused imports, so removing those fixes it.

Wait wait, let's confirm: the original errors are:
- imported and not used: "net/url"
- imported and not used: "regexp"
- imported package net/url is unused
- imported package regexp is unused

So those are the only compilation errors, right? So removing those two imports fixes all errors. Let's write the full codepackage tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	radarBinaryPath string
	radarPID        int
)

func HandleRadarBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "make", "build")
	cmd.Dir = findRepoRoot()
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("Build failed: %s\n%s", runErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Build succeeded\n%s", string(output)))
}

func HandleRadarTest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "make", "test")
	cmd.Dir = findRepoRoot()
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("Tests failed: %s\n%s", runErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Tests passed\n%s", string(output)))
}

func HandleRadarTypecheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "make", "tsc")
	cmd.Dir = findRepoRoot()
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("Type-check failed: %s\n%s", runErr.Error(), string(output)))
}

	return ok(fmt.Sprintf("Type-check passed\n%s", string(output)))
}

func HandleRadarStart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9300
	}
	repoRoot := findRepoRoot()
	binaryPath := filepath.Join(repoRoot, "radar")
	if _, statErr := os.Stat(binaryPath); os.IsNotExist(statErr) {
		return err("Radar binary not found. Run radar_build first.")
}

	cmd := exec.CommandContext(ctx, binaryPath, "--port", strconv.Itoa(port), "--no-open")
	cmd.Dir = repoRoot
	if startErr := cmd.Start(); startErr != nil {
		return err(fmt.Sprintf("Failed to start Radar: %s", startErr.Error()))
}

	radarBinaryPath = binaryPath
	radarPID = cmd.Process.Pid
	time.Sleep(2 * time.Second)
	healthURL := fmt.Sprintf("http://localhost:%d/api/cluster-info", port)
	client := http.DefaultClient
	resp, fetchErr := client.Get(healthURL)
	if fetchErr != nil {
		return ok(fmt.Sprintf("Radar started (PID: %d) on port %d. Health check pending.\nBinary: %s", radarPID, port, binaryPath))
}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return ok(fmt.Sprintf("Radar started successfully (PID: %d) on port %d\nHealth check: OK\nBinary: %s", radarPID, port, binaryPath))
}

	return ok(fmt.Sprintf("Radar started (PID: %d) on port %d (status: %d)\nBinary: %s", radarPID, port, resp.StatusCode, binaryPath))
}

func HandleRadarStop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if radarPID == 0 {
		return err("No Radar process running (PID not tracked)")
}

	process, procErr := os.FindProcess(radarPID)
	if procErr != nil {
		return err(fmt.Sprintf("Could not find process: %s", procErr.Error()))
}

	if killErr := process.Kill(); killErr != nil {
		return err(fmt.Sprintf("Failed to kill process: %s", killErr.Error()))
}

	radarPID = 0
	return ok("Radar process stopped")
}

func HandleRadarAPI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 9300
	}
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		endpoint = "/api/cluster-info"
	}
	baseURL := fmt.Sprintf("http://localhost:%d%s", port, endpoint)
	client := http.DefaultClient
	resp, fetchErr := client.Get(baseURL)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to connect to %s: %s", baseURL, fetchErr.Error()))
}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read response: %s", readErr.Error()))
}

	var prettyJSON bytes.Buffer
	parseErr := json.Indent(&prettyJSON, body, "", " ")
	if parseErr != nil {
		return ok(fmt.Sprintf("Status: %d\nResponse:\n%s", resp.StatusCode, string(body)))
}

	return ok(fmt.Sprintf("Status: %d\nResponse:\n%s", resp.StatusCode, prettyJSON.String()))
}

func HandleRadarDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoRoot := findRepoRoot()
	cmd := exec.CommandContext(ctx, "git", "diff", "main..HEAD", "--name-only")
	cmd.Dir = repoRoot
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		if strings.Contains(string(output), "fatal") {
			return ok("Not on a feature branch or main branch not found. Showing unstaged changes.")
}

		return err(fmt.Sprintf("Git diff failed: %s", runErr.Error()))
}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return ok("No changes on this branch compared to main.")
}

	var uiFiles, goFiles, otherFiles []string
	for _, f := range files {
		if f == "" {
			continue
		}
		ext := filepath.Ext(f)
		if ext == ".tsx" || ext == ".ts" || ext == ".css" {
			uiFiles = append(uiFiles, f)
		} else if ext == ".go" {
			goFiles = append(goFiles, f)
		} else {
			otherFiles = append(otherFiles, f)

	}
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Changed files (%d total):\n\n", len(files)))
	if len(uiFiles) > 0 {
		result.WriteString("UI Files:\n")
		for _, f := range uiFiles {
			result.WriteString(fmt.Sprintf(" - %s\n", f))

		result.WriteString("\n")

	if len(goFiles) > 0 {
		result.WriteString("Go Files:\n")
		for _, f := range goFiles {
			result.WriteString(fmt.Sprintf(" - %s\n", f))

		result.WriteString("\n")

	if len(otherFiles) > 0 {
		result.WriteString("Other Files:\n")
		for _, f := range otherFiles {
			result.WriteString(fmt.Sprintf(" - %s\n", f))

	}
	needsVisualTest := len(uiFiles) > 0
	needsBuild := len(goFiles) > 0
	result.WriteString("\nRecommendations:\n")
	result.WriteString(fmt.Sprintf("- Run tests: yes\n"))
	result.WriteString(fmt.Sprintf("- Run type-check: yes\n"))
	result.WriteString(fmt.Sprintf("- Run build: %s\n", boolToYesNo(needsBuild)))
	result.WriteString(fmt.Sprintf("- Run visual-test: %s\n", boolToYesNo(needsVisualTest)))
	return ok(result.String())
}

}
}
}
}
}
}

func HandleRadarStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoRoot := findRepoRoot()
	var result strings.Builder
	if radarPID != 0 {
		if proc, procErr := os.FindProcess(radarPID); procErr == nil {
			result.WriteString(fmt.Sprintf("Radar running: PID %d\n", radarPID))
			_ = proc
		} else {
			result.WriteString("Radar: not running (stale PID)\n")
			radarPID = 0
		}
	} else {
		result.WriteString("Radar: not running\n")

	binaryPath := filepath.Join(repoRoot, "radar")
	if _, statErr := os.Stat(binaryPath); statErr == nil {
		result.WriteString(fmt.Sprintf("Binary: %s (exists)\n", binaryPath))
	} else {
		result.WriteString("Binary: not built\n")

	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = repoRoot
	if output, runErr := cmd.CombinedOutput(); runErr == nil {
		branch := strings.TrimSpace(string(output))
		result.WriteString(fmt.Sprintf("Branch: %s\n", branch))

	return ok(result.String())
}

}
}
}

func findRepoRoot() string {
	cwd, _ := os.Getwd()
	for {
		if _, statErr := os.Stat(filepath.Join(cwd, "CLAUDE.md")); statErr == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "."
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}