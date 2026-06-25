package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HandleBuild executes the build script defined in package.json
func HandleBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "sh", "scripts/build.sh")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Build failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Build completed successfully:\n%s", string(output)))
}

// HandleStartHTTP starts the Next.js server in HTTP mode
func HandleStartHTTP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "next", "start")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("HTTP Start failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("HTTP Server started:\n%s", string(output)))
}

// HandleGenerateCert runs the certificate generation script
func HandleGenerateCert(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "sh", "scripts/generate-cert.sh")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Cert generation failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Certificate generated:\n%s", string(output)))
}

// HandleMigrate runs the migration script to SQLite
func HandleMigrate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "tsx", "scripts/migrate-to-sqlite.ts")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Migration failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Migration completed:\n%s", string(output)))
}

// HandleImportApps imports applications using the provided script
func HandleImportApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "tsx", "scripts/import-apps.ts")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Import failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Apps imported:\n%s", string(output)))
}

// HandleStatus checks the status of the project by attempting to read package.json
func HandleStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Read package.json to verify project structure
	data, readErr := os.ReadFile("package.json")
	if readErr != nil {
		return err(fmt.Sprintf("Could not read package.json: %v", readErr))
}

	var pkg struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	parseErr := json.Unmarshal(data, &pkg)
	if parseErr != nil {
		return err(fmt.Sprintf("Could not parse package.json: %v", parseErr))
}

	return ok(fmt.Sprintf("Project: %s v%s is ready.", pkg.Name, pkg.Version))
}

// HandleDeployToken runs the deployment token script
func HandleDeployToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "sh", "scripts/deploy-workload-base.sh")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Deploy token failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Deploy token script executed:\n%s", string(output)))
}

// HandleStop stops the PM2 process
func HandleStop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", "pm2 stop 0 & pm2 delete 0")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		// PM2 might return non-zero if process doesn't exist, which is often acceptable
		return ok(fmt.Sprintf("Stop command executed (may have ignored missing processes):\n%s", string(output)))
}

	return ok(fmt.Sprintf("Process stopped:\n%s", string(output)))
}

// HandleLint runs the linter
func HandleLint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "eslint", ".")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Linting failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("Linting passed:\n%s", string(output)))
}

// HandleDeleteAllApps runs the script to delete all apps
func HandleDeleteAllApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "tsx", "scripts/delete-all-apps.ts")
	cmd.Dir = "."

	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("Delete all apps failed: %v, output: %s", execErr, string(output)))
}

	return ok(fmt.Sprintf("All apps deleted:\n%s", string(output)))
}