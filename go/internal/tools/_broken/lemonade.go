package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com"
)

var (
	githubClient = http.DefaultClient
)

func HandleLemonadeInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platform, _ :=getString(args, "platform")
	version, _ :=getString(args, "version")
	if platform == "" || version == "" {
		return err("platform and version are required")
}

	var downloadURL string
	switch platform {
	case "windows":
		downloadURL = fmt.Sprintf("https://github.com/lemonade-sdk/lemonade/releases/download/v%s/lemonade-v%s-windows-x64.msi", version, version)
	case "macos":
		downloadURL = fmt.Sprintf("https://github.com/lemonade-sdk/lemonade/releases/download/v%s/lemonade-v%s-macos.dmg", version, version)
	case "linux":
		downloadURL = fmt.Sprintf("https://github.com/lemonade-sdk/lemonade/releases/download/v%s/lemonade-v%s-linux-x64.deb", version, version)
	default:
		return err("unsupported platform")
}

	resp, fetchErr := githubClient.Get(downloadURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to download installer: %s", resp.Status))
}

	tempFile, tempErr := os.CreateTemp("", "lemonade-installer-*.tmp")
	if tempErr != nil {
		return err(tempErr.Error())
}

	defer os.Remove(tempFile.Name())

	_, copyErr := io.Copy(tempFile, resp.Body)
	if copyErr != nil {
		return err(copyErr.Error())
}

	cmd := exec.Command("start", "/wait", tempFile.Name())
	runErr := cmd.Run()
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(fmt.Sprintf("Successfully installed Lemonade v%s for %s", version, platform))
}

func HandleLemonadeUninstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platform, _ :=getString(args, "platform")
	if platform == "" {
		return err("platform is required")
}

	var cmd *exec.Cmd
	switch platform {
	case "windows":
		cmd = exec.Command("msiexec", "/x", "{LEMONADE-GUID}")
	case "macos":
		cmd = exec.Command("osascript", "-e", "tell application \"Finder\" to eject (every disk whose name contains \"Lemonade\")")
	case "linux":
		cmd = exec.Command("sudo", "apt", "remove", "--purge", "lemonade")
	default:
		return err("unsupported platform")
}

	runErr := cmd.Run()
	if runErr != nil {
		return err(runErr.Error())
}

	return ok(fmt.Sprintf("Successfully uninstalled Lemonade for %s", platform))
}

func HandleLemonadeVersionCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, fetchErr := githubClient.Get(fmt.Sprintf("%s/repos/lemonade-sdk/lemonade/releases/latest", githubAPIURL))
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to fetch latest release: %s", resp.Status))
}

	var release map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&release)
	if parseErr != nil {
		return err(parseErr.Error())
}

	latestVersion, _ :=getString(release, "tag_name")
	if latestVersion == "" {
		return err("could not determine latest version")
}

	currentVersion, _ :=getString(args, "current_version")
	if currentVersion == "" {
		return ok(fmt.Sprintf("Latest version: %s", latestVersion))
}

	if latestVersion == currentVersion {
		return ok("Your version is up to date")
}

	return ok(fmt.Sprintf("Update available: %s (current: %s)", latestVersion, currentVersion))
}

func HandleLemonadeRunnerStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, fetchErr := githubClient.Get(fmt.Sprintf("%s/repos/lemonade-sdk/lemonade/actions/runners", githubAPIURL))
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to fetch runner status: %s", resp.Status))
}

	var runners struct {
		Runners []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"runners"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&runners)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var activeRunners []string
	for _, runner := range runners.Runners {
		if runner.Status == "online" {
			activeRunners = append(activeRunners, runner.Name)

	}

	sort.Strings(activeRunners)
	return ok(fmt.Sprintf("Active runners: %s", strings.Join(activeRunners, ", ")))
}

}

func HandleLemonadeLabelCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueNumber, _ :=getInt(args, "issue_number")
	if issueNumber == 0 {
		return err("issue_number is required")
}

	resp, fetchErr := githubClient.Get(fmt.Sprintf("%s/repos/lemonade-sdk/lemonade/issues/%d", githubAPIURL, issueNumber))
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to fetch issue details: %s", resp.Status))
}

	var issue struct {
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}

	parseErr := json.NewDecoder(resp.Body).Decode(&issue)
	if parseErr != nil {
		return err(parseErr.Error())
}

	var labels []string
	for _, label := range issue.Labels {
		labels = append(labels, label.Name)

	hasEngineLabel := false
	hasAreaLabel := false
	hasRuntimeLabel := false

	for _, label := range labels {
		if strings.HasPrefix(label, "engine::") {
			hasEngineLabel = true
		} else if strings.HasPrefix(label, "area::") {
			hasAreaLabel = true
		} else if strings.HasPrefix(label, "runtime::") {
			hasRuntimeLabel = true
		}
	}

	if hasEngineLabel && hasAreaLabel {
		return err("issue cannot have both engine and area labels")
}

	if hasEngineLabel && hasRuntimeLabel {
		return err("issue cannot have both engine and runtime labels")
}

	return ok(fmt.Sprintf("Issue #%d labels: %s", issueNumber, strings.Join(labels, ", ")))
}

}

func HandleLemonadeValidationCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Validation checks would be implemented here")
}