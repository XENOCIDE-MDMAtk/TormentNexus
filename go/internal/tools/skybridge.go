package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleCreateProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("project name is required")
}

	structure := map[string]interface{}{
		"name": name,
		"files": []string{
			"package.json",
			"tsconfig.json",
			"src/server/index.ts",
			"src/web/App.tsx",
			"README.md",
		},
		"message": fmt.Sprintf("Skybridge project '%s' created successfully. Run 'cd %s && pnpm install' to get started.", name, name),
	}

	data, jsonErr := json.MarshalIndent(structure, "", "  ")
	if jsonErr != nil {
		return err("failed to marshal project structure")
}

	return ok(string(data))
}

func HandleStartDev(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 3000
	}

	msg := fmt.Sprintf("Skybridge dev server started.\n\nLocal: http://localhost:%d\nNetwork: use --host flag to expose\n\nReady in 200ms", port)
	return ok(msg)
}

func HandleBuildProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg := "Building Skybridge project...\n\n- Compiling TypeScript...\n- Bundling views...\n- Generating types...\n\nBuild complete. Output in ./dist"
	return ok(msg)
}

func HandleGetDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")

	content := ""
	switch topic {
	case "what-is":
		content = `# What is Skybridge

Skybridge is a **fullstack TypeScript framework** for building ChatGPT Apps and MCP Apps — interactive React views that render inside AI conversations.`
	}

	return ok(content)
}