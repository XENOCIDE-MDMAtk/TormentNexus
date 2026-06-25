package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (t *ToolResponse) ok() error { return nil }
func (t *ToolResponse) err("error") error { return nil }
func (t *ToolResponse) getString() string { return t.TextContent }
func (t *ToolResponse) getInt() int { return 0 }
func (t *ToolResponse) getBool() bool { return false }
func ok(s string) ToolResponse { return ToolResponse{TextContent: s} }
func err(e error) ToolResponse { return ToolResponse{TextContent: e.Error() + " (error)"} }

func HandleXxx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	type ollamaRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}
	type ollamaResponse struct {
		Response string `json:"response"`
	}
	reqBody, jsonErr := json.Marshal(ollamaRequest{Model: model, Prompt: prompt})
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	client := http.DefaultClient
	ollamaURL := "http://localhost:11434/api/generate"
	httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, ollamaURL, strings.NewReader(string(reqBody)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, httpErr := client.Do(httpReq)
	if httpErr != nil {
		return err(httpErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("ollama returned status %d", resp.StatusCode))
}

	var ollResp ollamaResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&ollResp)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	return ok(ollResp.Response)
}

func HandleCreateNPC(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	directive, _ :=getString(args, "primary_directive")
	if directive == "" {
		return err("primary_directive is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "unknown"
	}
	npcInfo := map[string]string{
		"name":              name,
		"primary_directive": directive,
		"model":             model,
		"provider":          provider,
	}
	b, marshalErr := json.Marshal(npcInfo)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(b))
}

func HandleRunAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	model, _ :=getString(args, "model")
	if model == "" {
		return err("model is required")
}

	type ollamaRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}
	type ollamaResponse struct {
		Response string `json:"response"`
	}
	reqBody, jsonErr := json.Marshal(ollamaRequest{Model: model, Prompt: prompt})
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	client := http.DefaultClient
	ollamaURL := "http://localhost:11434/api/generate"
	httpReq, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, ollamaURL, strings.NewReader(string(reqBody)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, httpErr := client.Do(httpReq)
	if httpErr != nil {
		return err(httpErr.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("ollama returned status %d", resp.StatusCode))
}

	var ollResp ollamaResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&ollResp)
	if decodeErr != nil {
		return err(decodeErr.Error())
}

	result := fmt.Sprintf("Agent %s responded: %s", name, ollResp.Response)
	return ok(result)
}

func HandleListLargePythonFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	if repoPath == "" {
		return err("repo_path is required")
}

	minLines, _ :=getInt(args, "min_lines")
	if minLines == 0 {
		minLines = 500
	}
	var matches []string
	walkErr := filepath.Walk(repoPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".py") {
			f, openErr := os.Open(path)
			if openErr != nil {
				return nil
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			lineCount := 0
			for scanner.Scan() {
				lineCount++
			}
			if lineCount >= minLines {
				rel, relErr := filepath.Rel(repoPath, path)
				if relErr != nil {
					rel = path
				}
				matches = append(matches, fmt.Sprintf("%s (%d lines)", rel, lineCount))

		}
		return nil
	})
	if walkErr != nil {
		return err(walkErr.Error())
}

	if len(matches) == 0 {
		return ok("No Python files exceed the line threshold.")
}

	result := "Python files exceeding line count:\n" + strings.Join(matches, "\n")
	return ok(result)
}

}

func HandleFetchImageDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dataset, _ :=getString(args, "dataset_name")
	if dataset == "" {
		return err("dataset_name is required")
}

	maxImages, _ :=getInt(args, "max_images")
	if maxImages == 0 {
		maxImages = 10
	}
	placeholder := fmt.Sprintf("Fetched %d images from dataset %s (placeholder paths).", maxImages, dataset)
	return ok(placeholder)
}