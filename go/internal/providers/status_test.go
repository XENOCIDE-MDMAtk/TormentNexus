package providers

import "testing"

func TestSnapshotReportsConfiguredProviders(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "gemini")
	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "oauth-token")
	t.Setenv("COPILOT_PAT", "copilot")
	t.Setenv("GITHUB_TOKEN", "")

	statuses := Snapshot()
	if len(statuses) == 0 {
		t.Fatalf("expected provider statuses")
	}

	var openai Status
	var google Status
	var googleOAuth Status
	var copilot Status
	for _, status := range statuses {
		if status.Provider == "openai" {
			openai = status
		}
		if status.Provider == "google" {
			google = status
		}
		if status.Provider == "google-oauth" {
			googleOAuth = status
		}
		if status.Provider == "github-copilot" {
			copilot = status
		}
	}

	if !openai.Configured || !openai.Authenticated {
		t.Fatalf("expected openai to be configured, got %+v", openai)
	}
	if !google.Configured || google.EnvVar != "GEMINI_API_KEY" {
		t.Fatalf("expected google to detect GEMINI_API_KEY, got %+v", google)
	}
	if !googleOAuth.Configured || googleOAuth.EnvVar != "GOOGLE_OAUTH_ACCESS_TOKEN" {
		t.Fatalf("expected google-oauth to detect GOOGLE_OAUTH_ACCESS_TOKEN, got %+v", googleOAuth)
	}
	if !copilot.Configured || copilot.EnvVar != "COPILOT_PAT" {
		t.Fatalf("expected github-copilot to detect COPILOT_PAT, got %+v", copilot)
	}
}
