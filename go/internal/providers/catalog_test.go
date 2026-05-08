package providers

import "testing"

func TestCatalogIncludesConfiguredProviderState(t *testing.T) {
	entries := Catalog([]Status{
		{Provider: "google", Configured: true, Authenticated: true, EnvVars: []string{"GOOGLE_API_KEY", "GEMINI_API_KEY"}},
		{Provider: "github-copilot", Configured: true, Authenticated: true, EnvVars: []string{"COPILOT_PAT", "GITHUB_TOKEN"}},
	})

	if len(entries) < 2 {
		t.Fatalf("expected catalog entries, got %d", len(entries))
	}
	if entries[0].Provider != "google" || !entries[0].Configured {
		t.Fatalf("expected google configured entry, got %+v", entries[0])
	}

	foundCopilot := false
	for _, entry := range entries {
		if entry.Provider != "github-copilot" {
			continue
		}
		foundCopilot = true
		if entry.DefaultModel != "copilot/gpt-4.1" || !entry.Configured {
			t.Fatalf("unexpected copilot entry: %+v", entry)
		}
	}
	if !foundCopilot {
		t.Fatalf("expected github-copilot catalog entry")
	}
}
