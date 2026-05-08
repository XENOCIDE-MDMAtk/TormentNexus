package ai

import "testing"

func clearProviderEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"ANTHROPIC_API_KEY",
		"GOOGLE_API_KEY",
		"GEMINI_API_KEY",
		"OPENAI_API_KEY",
		"DEEPSEEK_API_KEY",
		"OPENROUTER_API_KEY",
	} {
		t.Setenv(key, "")
	}
}

func TestResolveProviderSelectionHonorsPriority(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")
	t.Setenv("GOOGLE_API_KEY", "google-key")

	selection, ok := resolveProviderSelection()
	if !ok {
		t.Fatal("expected provider selection")
	}
	if selection.ProviderName != "anthropic" {
		t.Fatalf("expected anthropic priority, got %#v", selection)
	}
	if selection.DefaultModel != "claude-sonnet-4-20250514" {
		t.Fatalf("unexpected default model: %#v", selection)
	}
}

func TestResolveProviderSelectionUsesGeminiAlias(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("GEMINI_API_KEY", "gemini-key")

	selection, ok := resolveProviderSelection()
	if !ok {
		t.Fatal("expected provider selection")
	}
	if selection.ProviderName != "google" || selection.EnvVar != "GEMINI_API_KEY" {
		t.Fatalf("expected google via GEMINI_API_KEY, got %#v", selection)
	}
}

func TestListConfiguredProvidersDeduplicatesAliases(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("GOOGLE_API_KEY", "google-key")
	t.Setenv("GEMINI_API_KEY", "gemini-key")
	t.Setenv("OPENAI_API_KEY", "openai-key")

	configured := ListConfiguredProviders()
	want := []string{"google", "openai"}
	if len(configured) != len(want) {
		t.Fatalf("expected %d configured providers, got %#v", len(want), configured)
	}
	for i := range want {
		if configured[i] != want[i] {
			t.Fatalf("expected configured[%d]=%q, got %q", i, want[i], configured[i])
		}
	}
}
