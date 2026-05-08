package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTiers(statuses ...int) ([]TierConfig, func()) {
	configs := make([]TierConfig, len(statuses))
	servers := make([]*httptest.Server, len(statuses))
	for i, code := range statuses {
		statusCode := code
		idx := i
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if statusCode != http.StatusOK {
				http.Error(w, `{"error":"mock error"}`, statusCode)
				return
			}
			resp := ChatResponse{
				ID: "resp-" + itoa(idx), Model: "test-model", ProviderID: configs[idx].ID,
				Choices: []Choice{{Index: 0, Message: Message{Role: RoleAssistant, Content: "hello from tier " + itoa(idx)}, FinishReason: "stop"}},
				Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		servers[i] = srv
		configs[i] = TierConfig{ID: "tier-" + itoa(i), Name: "Tier " + itoa(i), BaseURL: srv.URL, DefaultModel: "test-model", Timeout: 5 * time.Second, Priority: i}
	}
	cleanup := func() { for _, s := range servers { s.Close() } }
	return configs, cleanup
}

func TestWaterfall_SuccessOnFirstTier(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusOK, http.StatusOK, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	resp, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil { t.Fatalf("expected success, got: %v", err) }
	if resp.ProviderID != "tier-0" { t.Errorf("expected tier-0, got %s", resp.ProviderID) }
	if resp.Attempts != 1 { t.Errorf("expected 1 attempt, got %d", resp.Attempts) }
	req, fb, fail := router.Stats()
	if req != 1 || fb != 0 || fail != 0 { t.Errorf("stats: req=%d fb=%d fail=%d, want 1,0,0", req, fb, fail) }
}

func TestWaterfall_CascadesOn429(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusTooManyRequests, http.StatusOK, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	resp, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil { t.Fatalf("expected success after cascade, got: %v", err) }
	if resp.ProviderID != "tier-1" { t.Errorf("expected tier-1, got %s", resp.ProviderID) }
	if resp.Attempts != 2 { t.Errorf("expected 2 attempts, got %d", resp.Attempts) }
	_, fb, _ := router.Stats()
	if fb != 1 { t.Errorf("expected 1 fallback, got %d", fb) }
}

func TestWaterfall_CascadesOn503(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	resp, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil { t.Fatalf("expected success on tier-2, got: %v", err) }
	if resp.ProviderID != "tier-2" { t.Errorf("expected tier-2, got %s", resp.ProviderID) }
}

func TestWaterfall_FailsOnAllTiers(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusTooManyRequests, http.StatusServiceUnavailable, http.StatusBadGateway)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	_, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err == nil { t.Fatal("expected error when all tiers fail") }
	_, _, fail := router.Stats()
	if fail != 1 { t.Errorf("expected 1 failure, got %d", fail) }
}

func TestWaterfall_StopsOn4xx(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusBadRequest, http.StatusOK, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	_, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err == nil { t.Fatal("expected error on 400") }
	if pe, ok := err.(*ProviderError); !ok || pe.StatusCode != 400 { t.Errorf("expected ProviderError 400, got %v", err) }
}

func TestWaterfall_ExcludeSkipsTiers(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusOK, http.StatusOK, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	resp, err := router.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}, Exclude: []string{"tier-0", "tier-1"}})
	if err != nil { t.Fatalf("expected success on tier-2, got: %v", err) }
	if resp.ProviderID != "tier-2" { t.Errorf("expected tier-2, got %s", resp.ProviderID) }
}

func TestWaterfall_StreamFallback(t *testing.T) {
	configs, cleanup := setupTiers(http.StatusOK, http.StatusOK, http.StatusOK)
	defer cleanup()
	router := NewWaterfallRouter(configs)
	ch, err := router.ChatStream(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil { t.Fatalf("ChatStream: %v", err) }
	var chunks []StreamChunk
	for chunk := range ch { chunks = append(chunks, chunk) }
	if len(chunks) < 2 { t.Fatalf("expected at least 2 chunks, got %d", len(chunks)) }
	if !chunks[len(chunks)-1].Done { t.Error("last chunk should have Done=true") }
}

func TestDefaultTierConfigs(t *testing.T) {
	configs := DefaultTierConfigs()
	if len(configs) != 3 { t.Fatalf("expected 3 default tiers, got %d", len(configs)) }
	if configs[0].ID != "nvidia-nim" { t.Errorf("tier 0: got %s", configs[0].ID) }
	if configs[1].ID != "openrouter" { t.Errorf("tier 1: got %s", configs[1].ID) }
	if configs[2].ID != "lmstudio" { t.Errorf("tier 2: got %s", configs[2].ID) }
}
