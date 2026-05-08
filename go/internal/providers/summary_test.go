package providers

import "testing"

func TestBuildSummaryAggregatesCatalogState(t *testing.T) {
	summary := BuildSummary([]Status{
		{Provider: "google", Configured: true, Authenticated: true},
		{Provider: "openai", Configured: true, Authenticated: true},
		{Provider: "github-copilot", Configured: true, Authenticated: true},
	})

	if summary.ProviderCount < 3 {
		t.Fatalf("expected catalog-backed provider count, got %d", summary.ProviderCount)
	}
	if summary.ConfiguredCount != 3 {
		t.Fatalf("expected 3 configured providers, got %d", summary.ConfiguredCount)
	}
	if len(summary.ByAuthMethod) == 0 {
		t.Fatalf("expected auth-method buckets")
	}
	if len(summary.ByPreferredTask) == 0 {
		t.Fatalf("expected preferred-task buckets")
	}
}
