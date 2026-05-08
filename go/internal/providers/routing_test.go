package providers

import "testing"

func TestCatalogMarksConfiguredProvidersFromStatus(t *testing.T) {
	entries := Catalog([]Status{
		{Provider: "google", Configured: true, Authenticated: true, EnvVars: []string{"GOOGLE_API_KEY"}},
		{Provider: "openai", Configured: true, Authenticated: false, EnvVars: []string{"OPENAI_API_KEY"}},
	})

	foundGoogle := false
	foundOpenAI := false
	for _, entry := range entries {
		switch entry.Provider {
		case "google":
			foundGoogle = true
			if !entry.Configured || !entry.Authenticated {
				t.Fatalf("expected google configured/authenticated, got %#v", entry)
			}
		case "openai":
			foundOpenAI = true
			if !entry.Configured || entry.Authenticated {
				t.Fatalf("expected openai configured but not authenticated, got %#v", entry)
			}
		}
	}
	if !foundGoogle || !foundOpenAI {
		t.Fatalf("expected google and openai catalog entries, got %#v", entries)
	}
}

func TestBuildRoutingSummaryUsesConfiguredStatuses(t *testing.T) {
	summary := BuildRoutingSummary([]Status{
		{Provider: "anthropic", Configured: true, Authenticated: true},
		{Provider: "google", Configured: true, Authenticated: true},
		{Provider: "openai", Configured: false, Authenticated: false},
	})

	if summary.DefaultStrategy != "best" {
		t.Fatalf("expected default strategy best, got %#v", summary)
	}
	if len(summary.Tasks) != 6 {
		t.Fatalf("expected 6 task summaries, got %#v", summary.Tasks)
	}

	var coding *RoutingTaskSummary
	for i := range summary.Tasks {
		if summary.Tasks[i].TaskType == "coding" {
			coding = &summary.Tasks[i]
			break
		}
	}
	if coding == nil {
		t.Fatal("expected coding task summary")
	}
	if coding.Candidates[0].Provider != "google" || !coding.Candidates[0].Configured {
		t.Fatalf("expected google first/configured for coding, got %#v", coding.Candidates[0])
	}

	var planning *RoutingTaskSummary
	for i := range summary.Tasks {
		if summary.Tasks[i].TaskType == "planning" {
			planning = &summary.Tasks[i]
			break
		}
	}
	if planning == nil {
		t.Fatal("expected planning task summary")
	}
	if planning.Candidates[0].Provider != "anthropic" || !planning.Candidates[0].Configured {
		t.Fatalf("expected anthropic first/configured for planning, got %#v", planning.Candidates[0])
	}
}
