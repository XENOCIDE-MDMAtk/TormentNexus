package tools

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func HandleCodex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	sessionId, _ :=getString(args, "sessionId")
	resetSession, _ :=getBool(args, "resetSession")
	model, _ :=getString(args, "model")
	reasoningEffort, _ :=getString(args, "reasoningEffort")
	sandbox, _ :=getString(args, "sandbox")
	fullAuto, _ :=getBool(args, "fullAuto")
	workingDirectory, _ :=getString(args, "workingDirectory")
	callbackUri, _ :=getString(args, "callbackUri")

	defaultModel := os.Getenv("CODEX_DEFAULT_MODEL")
	if defaultModel == "" {
		defaultModel = "gpt-5.3-codex"
	}
	if model == "" {
		model = defaultModel
	}

	cmdArgs := []string{"exec"}

	if model != "" {
		cmdArgs = append(cmdArgs, "--model", model)

	if sandbox != "" {
		cmdArgs = append(cmdArgs, "--sandbox", sandbox)

	if fullAuto {
		cmdArgs = append(cmdArgs, "--full-auto")

	if reasoningEffort != "" {
		cmdArgs = append(cmdArgs, "-c", "model_reasoning_effort="+reasoningEffort)

	if workingDirectory != "" {
		cmdArgs = append(cmdArgs, "-c", "working_directory="+workingDirectory)

	if callbackUri != "" {
		cmdArgs = append(cmdArgs, "-c", "callback_uri="+callbackUri)

	cmdArgs = append(cmdArgs, "--skip-git-repo-check")

	if sessionId != "" {
		if resetSession {
			cmdArgs = append(cmdArgs, prompt)
		} else {
			cmdArgs = append(cmdArgs, "resume", sessionId, prompt)

	} else {
		cmdArgs = append(cmdArgs, prompt)

	cmd := exec.CommandContext(ctx, "codex", cmdArgs...)

	if workingDirectory != "" {
		cmd.Dir = workingDirectory
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return err("command timed out")
}

		output := stdout.String() + stderr.String()
		if output != "" {
			return ok(output)
}

		return err(e.Error())
}

	output := stdout.String() + stderr.String()

	if len(output) > 10*1024*1024 {
		output = output[:10*1024*1024] + "... (output truncated at 10MB)"
	}

	return ok(output)
}

}
}
}
}
}
}
}
}

func HandleReview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	uncommitted, _ :=getBool(args, "uncommitted")
	commit, _ :=getString(args, "commit")

	cmdArgs := []string{"review"}

	if uncommitted {
		cmdArgs = append(cmdArgs, "--uncommitted")
	} else if commit != "" {
		cmdArgs = append(cmdArgs, "--commit", commit)
	} else if base != "" {
		cmdArgs = append(cmdArgs, "--base", base)

	cmd := exec.CommandContext(ctx, "codex", cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		output := stdout.String() + stderr.String()
		if output != "" {
			return ok(output)
}

		return err(e.Error())
}

	output := stdout.String() + stderr.String()
	if len(output) > 10*1024*1024 {
		output = output[:10*1024*1024] + "... (output truncated at 10MB)"
	}

	return ok(output)
}

}

func HandleWebsearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	numResults, _ :=getInt(args, "numResults")
	searchDepth, _ :=getString(args, "searchDepth")

	cmdArgs := []string{"search", "--query", query}

	if numResults > 0 {
		cmdArgs = append(cmdArgs, "--num-results", strconv.Itoa(numResults))

	if searchDepth != "" {
		cmdArgs = append(cmdArgs, "--search-depth", searchDepth)

	cmd := exec.CommandContext(ctx, "codex", cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		output := stdout.String() + stderr.String()
		if output != "" {
			return ok(output)
}

		return err(e.Error())
}

	output := stdout.String() + stderr.String()
	if len(output) > 10*1024*1024 {
		output = output[:10*1024*1024] + "... (output truncated at 10MB)"
	}

	return ok(output)
}

}
}

func HandleListSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Session listing is available through the session management system")
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "codex", "--version")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		return err("Failed to ping codex server")
}

	output := stdout.String() + stderr.String()
	return ok("Codex server is running: " + strings.TrimSpace(output))
}

func HandleHelp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "codex", "--help")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := cmd.Run()
	if e != nil {
		output := stdout.String() + stderr.String()
		if output != "" {
			return ok(output)
}

		return err(e.Error())
}

	output := stdout.String() + stderr.String()
	if len(output) > 10*1024*1024 {
		output = output[:10*1024*1024] + "... (output truncated at 10MB)"
	}

	return ok(output)
}