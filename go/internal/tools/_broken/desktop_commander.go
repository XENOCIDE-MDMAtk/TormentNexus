package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HandleRunCommand executes a shell command and returns its combined output
func HandleRunCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmdStr, _ :=getString(args, "command")
	cmdArgs := []string{}
	if rawArgs, found := args["args"]; found {
		if argSlice, found := rawArgs.([]interface{}); found {
			for _, a := range argSlice {
				if s, found := a.(string); found {
					cmdArgs = append(cmdArgs, s)

			}
		}
	}

	cmd := exec.Command(cmdStr, cmdArgs...)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return err(fmt.Sprintf("Command execution failed: %s\nOutput: %s", runErr.Error(), string(output)))
}

	return ok(string(output))
}

}

// HandleGetSystemInfo retrieves basic system information
func HandleGetSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, hostErr := os.Hostname()
	if hostErr != nil {
		return err(fmt.Sprintf("Failed to retrieve hostname: %s", hostErr.Error()))
}

	var osInfo string
	// Try Unix-like uname first
	unameCmd := exec.Command("uname", "-a")
	if unameOut, unameErr := unameCmd.CombinedOutput(); unameErr == nil {
		osInfo = string(unameOut)
	} else {
		// Fallback to Windows ver command
		verCmd := exec.Command("cmd", "/c", "ver")
		if verOut, verErr := verCmd.CombinedOutput(); verErr == nil {
			osInfo = string(verOut)
		} else {
			osInfo = "Unknown operating system"
		}
	}

	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		cwd = "Unknown"
	}

	info := fmt.Sprintf("Hostname: %s\nOS Information: %s\nCurrent Working Directory: %s",
		hostname, strings.TrimSpace(osInfo), cwd)
	return ok(info)
}

// HandleListDirectory lists contents of a directory with metadata
func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dirPath, _ :=getString(args, "path")
	if dirPath == "" {
		var cwdErr error
		dirPath, cwdErr = os.Getwd()
		if cwdErr != nil {
			return err(fmt.Sprintf("Failed to get current working directory: %s", cwdErr.Error()))

	}

	entries, readErr := os.ReadDir(dirPath)
	if readErr != nil {
		return err(fmt.Sprintf("Failed to read directory: %s", readErr.Error()))
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of directory: %s\n", dirPath))
	result.WriteString("Name\tType\tSize\n")

	for _, entry := range entries {
		info, statErr := entry.Info()
		if statErr != nil {
			continue
		}

		var size string
		if info.IsDir() {
			size = "<dir>"
		} else {
			size = fmt.Sprintf("%d", info.Size())

		result.WriteString(fmt.Sprintf("%s\t%s\t%s\n",
			entry.Name(),
			info.Mode().String(),
			size,
		))

	return ok(result.String())
}

}
}
}

// HandleCreateFile creates a new file with given content
func HandleCreateFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "path")
	if filePath == "" {
		return err("File path cannot be empty")
}

	content, _ :=getString(args, "content")
	if content == "" {
		return err("Content cannot be empty")
}

	file, createErr := os.Create(filePath)
	if createErr != nil {
		return err(fmt.Sprintf("Failed to create file: %s", createErr.Error()))
}

	defer file.Close()

	_, writeErr := file.WriteString(content)
	if writeErr != nil {
		return err(fmt.Sprintf("Failed to write to file: %s", writeErr.Error()))
}

	return ok(fmt.Sprintf("File created successfully: %s", filePath))
}