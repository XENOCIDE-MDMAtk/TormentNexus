package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func HandleProcessList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "ps", "aux")
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("failed to execute ps command: %v", runErr))
}

	return ok(string(output))
}

func HandleServiceStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceName, _ :=getString(args, "service_name")
	if serviceName == "" {
		return err("service_name parameter is required")
}

	cmd := exec.CommandContext(ctx, "systemctl", "status", serviceName)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return ok(string(output))
}

	return ok(string(output))
}

func HandleServiceControl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serviceName, _ :=getString(args, "service_name")
	action, _ :=getString(args, "action")

	if serviceName == "" {
		return err("service_name parameter is required")
}

	if action == "" {
		return err("action parameter is required")
}

	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"reload":  true,
	}
	if !validActions[action] {
		return err(fmt.Sprintf("invalid action: %s. Valid actions are: start, stop, restart, reload", action))
}

	cmd := exec.CommandContext(ctx, "systemctl", action, serviceName)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("failed to %s service %s: %v", action, serviceName, runErr))
}

	return ok(fmt.Sprintf("Service %s %sed successfully", serviceName, action))
}

func HandleLogTail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	logPath, _ :=getString(args, "path")
	lines, _ :=getInt(args, "lines")

	if logPath == "" {
		return err("path parameter is required")
}

	if lines == 0 {
		lines = 20
	}

	absPath, pathErr := filepath.Abs(logPath)
	if pathErr != nil {
		return err(fmt.Sprintf("failed to resolve path: %v", pathErr))
}

	cmd := exec.CommandContext(ctx, "tail", "-n", strconv.Itoa(lines), absPath)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("failed to tail log file: %v", runErr))
}

	return ok(string(output))
}

func HandleNetworkCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		target = "8.8.8.8"
	}

	cmd := exec.CommandContext(ctx, "ping", "-c", "4", target)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("ping failed: %v", runErr))
}

	return ok(string(output))
}

func HandleSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	commands := map[string][]string{
		"hostname": {"hostname"},
		"uptime":   {"uptime"},
		"os":       {"uname", "-a"},
		"cpu":      {"lscpu"},
		"memory":   {"free", "-h"},
	}

	var result strings.Builder
	for name, cmdArgs := range commands {
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			result.WriteString(fmt.Sprintf("%s: error executing command: %v\n", name, runErr))
			continue
		}
		result.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", strings.ToUpper(name), string(output)))

	return ok(result.String())
}
}