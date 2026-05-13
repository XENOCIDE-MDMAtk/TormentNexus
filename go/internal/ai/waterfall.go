package ai

import (
	"context"
	"fmt"
	"strings"
)

// LLMClient represents a standard interface for text generation
type LLMClient interface {
	GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error)
}

// WaterfallClient wraps multiple providers and cascades upon failures (429, 5xx)
type WaterfallClient struct {
	tiers []LLMClient
}

func NewWaterfallClient(tiers ...LLMClient) *WaterfallClient {
	return &WaterfallClient{
		tiers: tiers,
	}
}

func (w *WaterfallClient) GenerateText(ctx context.Context, model string, messages []Message) (*LLMResponse, error) {
	var lastErr error
	for i, tier := range w.tiers {
		resp, err := tier.GenerateText(ctx, model, messages)
		if err == nil {
			if i > 0 {
				fmt.Printf("[Waterfall] Fallback successful at tier %d\n", i+1)
			}
			return resp, nil
		}

		lastErr = err

		// Heuristic to decide if we should cascade (e.g. rate limit, server error)
		errStr := err.Error()
		if strings.Contains(errStr, "429") || strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503") || strings.Contains(errStr, "timeout") {
			fmt.Printf("[Waterfall] Tier %d failed (%v). Cascading...\n", i+1, err)
			continue
		}

		// If it's a hard error (like bad request), fail immediately
		if strings.Contains(errStr, "400") || strings.Contains(errStr, "401") || strings.Contains(errStr, "403") {
			return nil, fmt.Errorf("hard error from tier %d: %w", i+1, err)
		}

		// Otherwise, log and continue cascading to be safe
		fmt.Printf("[Waterfall] Tier %d failed with unexpected error (%v). Cascading...\n", i+1, err)
	}
	return nil, fmt.Errorf("all waterfall tiers failed. Last error: %w", lastErr)
}
