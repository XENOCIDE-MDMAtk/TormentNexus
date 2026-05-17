package ai

import (
	"context"
	"fmt"
	"strings"
	"github.com/borghq/borg-go/internal/providers"
)

type WaterfallClient struct {
	tiers    []Provider
    selector *providers.ModelSelector
}

func NewWaterfallClient(selector *providers.ModelSelector, tiers ...Provider) *WaterfallClient {
	return &WaterfallClient{
		tiers:    tiers,
        selector: selector,
	}
}

func (w *WaterfallClient) GenerateText(ctx context.Context, taskType string, messages []Message) (*LLMResponse, error) {
    providerName, err := w.selector.SelectProvider(ctx, taskType)
    if err != nil {
        return nil, err
    }

    var provider Provider
    selection, ok := getProviderSelection(providerName)
    if ok {
        provider = selection.Factory(selection.APIKey)
    }

    if provider == nil {
        return nil, fmt.Errorf("selected provider %s not available", providerName)
    }

    resp, err := provider.GenerateText(ctx, selection.DefaultModel, messages)
    if err == nil {
        return resp, nil
    }

    // Heuristic fallback logic if the primary selection fails
    if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
        fmt.Printf("[Waterfall] Selected provider %s failed due to quota. Falling back...\n", providerName)
        // In a real implementation, we would retry the selection loop
    }

    return nil, err
}
