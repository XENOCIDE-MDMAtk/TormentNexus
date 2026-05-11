package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	Content string
	Usage   struct {
		InputTokens  int
		OutputTokens int
	}
	Provider string
	Model    string
}

type Provider interface {
	GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error)
}

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func (p *OpenAIProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://api.openai.com/v1/chat/completions"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": messages,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	return &LLMResponse{
		Content:  payload.Choices[0].Message.Content,
		Provider: "openai",
		Model:    model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  payload.Usage.PromptTokens,
			OutputTokens: payload.Usage.CompletionTokens,
		},
	}, nil
}

type AnthropicProvider struct {
	APIKey  string
	BaseURL string
}

func (p *AnthropicProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://api.anthropic.com/v1/messages"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"messages":   messages,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error: %s - %s", resp.Status, string(body))
	}

	var payload struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Content) == 0 {
		return nil, fmt.Errorf("no content returned from Anthropic")
	}

	return &LLMResponse{
		Content:  payload.Content[0].Text,
		Provider: "anthropic",
		Model:    model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  payload.Usage.InputTokens,
			OutputTokens: payload.Usage.OutputTokens,
		},
	}, nil
}

// GeminiProvider implements the Google Gemini (Generative AI) API
type GeminiProvider struct {
	APIKey  string
	BaseURL string
}

func (p *GeminiProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	// Convert messages to Gemini content format
	contents := make([]map[string]interface{}, 0, len(messages))
	var systemInstruction string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemInstruction = msg.Content
			continue
		}
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": msg.Content}},
		})
	}

	body := map[string]interface{}{
		"contents": contents,
	}
	if systemInstruction != "" {
		body["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}

	reqBody, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, model, p.APIKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(respBody))
	}

	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from Gemini")
	}

	// Concatenate all parts
	var sb strings.Builder
	for _, part := range payload.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	return &LLMResponse{
		Content:  sb.String(),
		Provider: "google",
		Model:    model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  payload.UsageMetadata.PromptTokenCount,
			OutputTokens: payload.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}

// DeepSeekProvider uses OpenAI-compatible API at api.deepseek.com
type DeepSeekProvider struct {
	APIKey string
}

func (p *DeepSeekProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	oai := &OpenAIProvider{
		APIKey:  p.APIKey,
		BaseURL: "https://api.deepseek.com/v1/chat/completions",
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "deepseek"
	return resp, nil
}

// OpenRouterProvider uses OpenAI-compatible API at openrouter.ai
type OpenRouterProvider struct {
	APIKey string
}

func (p *OpenRouterProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	oai := &OpenAIProvider{
		APIKey:  p.APIKey,
		BaseURL: "https://openrouter.ai/api/v1/chat/completions",
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "openrouter"
	return resp, nil
}

// LMStudioProvider uses local OpenAI-compatible API at localhost:1234
type LMStudioProvider struct {
	BaseURL string
}

func (p *LMStudioProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "http://localhost:1234/v1/chat/completions"
	}
	oai := &OpenAIProvider{
		APIKey:  "lm-studio",
		BaseURL: p.BaseURL,
	}
	resp, err := oai.GenerateText(ctx, model, messages)
	if err != nil {
		return nil, err
	}
	resp.Provider = "lmstudio"
	return resp, nil
}

// OllamaProvider uses local API at localhost:11434
type OllamaProvider struct {
	BaseURL string
}

func (p *OllamaProvider) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	if p.BaseURL == "" {
		p.BaseURL = "http://localhost:11434/api/chat"
	}

	body := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}
	reqBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		PromptEvalCount int `json:"prompt_eval_count"`
		EvalCount       int `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return &LLMResponse{
		Content:  payload.Message.Content,
		Provider: "ollama",
		Model:    model,
		Usage: struct {
			InputTokens  int
			OutputTokens int
		}{
			InputTokens:  payload.PromptEvalCount,
			OutputTokens: payload.EvalCount,
		},
	}, nil
}

// ProviderPriority defines the order for auto-routing
var ProviderPriority = []struct {
	EnvVar       string
	ProviderName string
	DefaultModel string
	Factory      func(apiKey string) Provider
}{
	{"ANTHROPIC_API_KEY", "anthropic", "claude-sonnet-4-20250514", func(k string) Provider { return &AnthropicProvider{APIKey: k} }},
	{"GOOGLE_AI_STUDIO_API_KEY", "google-ai-studio", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"GOOGLE_API_KEY", "google", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"GEMINI_API_KEY", "google", "gemini-2.5-flash", func(k string) Provider { return &GeminiProvider{APIKey: k} }},
	{"OPENAI_API_KEY", "openai", "gpt-4o", func(k string) Provider { return &OpenAIProvider{APIKey: k} }},
	{"DEEPSEEK_API_KEY", "deepseek", "deepseek-chat", func(k string) Provider { return &DeepSeekProvider{APIKey: k} }},
	{"OPENROUTER_API_KEY", "openrouter", "openrouter/free", func(k string) Provider { return &OpenRouterProvider{APIKey: k} }},
	{"", "lmstudio", "local-model", func(k string) Provider { return &LMStudioProvider{} }},
	{"", "ollama", "gemma:2b", func(k string) Provider { return &OllamaProvider{} }},
}

type providerSelection struct {
	EnvVar       string
	ProviderName string
	DefaultModel string
	Factory      func(apiKey string) Provider
	APIKey       string
}

func resolveProviderSelection() (providerSelection, bool) {
	for _, entry := range ProviderPriority {
		if key := os.Getenv(entry.EnvVar); key != "" {
			return providerSelection{
				EnvVar:       entry.EnvVar,
				ProviderName: entry.ProviderName,
				DefaultModel: entry.DefaultModel,
				Factory:      entry.Factory,
				APIKey:       key,
			}, true
		}
	}
	return providerSelection{}, false
}

// AutoRoute selects the best available provider based on environment variables.
// Priority: Anthropic > Gemini > OpenAI > DeepSeek > OpenRouter
// This acts as a lightweight fallback router when the main TypeScript Core is unavailable.
func AutoRoute(ctx context.Context, messages []Message) (*LLMResponse, error) {
	selection, ok := resolveProviderSelection()
	if !ok {
		return nil, fmt.Errorf("no LLM provider configured (set ANTHROPIC_API_KEY, GOOGLE_API_KEY, GEMINI_API_KEY, OPENAI_API_KEY, DEEPSEEK_API_KEY, or OPENROUTER_API_KEY)")
	}
	return selection.Factory(selection.APIKey).GenerateText(ctx, selection.DefaultModel, messages)
}

// AutoRouteWithModel selects the best provider and allows model override
func AutoRouteWithModel(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	selection, ok := resolveProviderSelection()
	if !ok {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	if model == "" {
		model = selection.DefaultModel
	}
	return selection.Factory(selection.APIKey).GenerateText(ctx, model, messages)
}

// ListConfiguredProviders returns which providers have API keys set
func ListConfiguredProviders() []string {
	var configured []string
	seen := map[string]struct{}{}
	for _, entry := range ProviderPriority {
		if os.Getenv(entry.EnvVar) == "" {
			continue
		}
		if _, ok := seen[entry.ProviderName]; ok {
			continue
		}
		seen[entry.ProviderName] = struct{}{}
		configured = append(configured, entry.ProviderName)
	}
	return configured
}
