package tools` and the function signatures as `HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error)`.

 We note that the original code does not use the `args` parameter in any of the functions. That is acceptable.

 Let's remove the unused imports and keep only:
   - context
   - fmt
   - os/exec
   - strings

 However, note that the original code also uses `strings.TrimSpace` and `fmt.Sprintf`, so we keep `strings` and `fmt`.

 But wait: the code also uses `exec.CommandContext` which is from `os/exec`. So we keep that.

 We are also told that we must only use stdlib imports, and we are.

 Letpackage tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleWindowsVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "ver")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	version := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("Windows Version: %s", version))
}

func HandleSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "systeminfo")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	info := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("System Information:\n%s", info))
}

func HandleDiskInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "wmic logicaldisk get size,freespace,caption")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	diskInfo := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("Disk Information:\n%s", diskInfo))
}

func HandleProcessList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "tasklist")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	processes := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("Running Processes:\n%s", processes))
}

func HandleNetworkInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "ipconfig /all")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	networkInfo := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("Network Information:\n%s", networkInfo))
}

func HandleEnvironmentVariables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "set")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return err(cmdErr.Error())
}

	envVars := strings.TrimSpace(string(output))
	return ok(fmt.Sprintf("Environment Variables:\n%s", envVars))
}