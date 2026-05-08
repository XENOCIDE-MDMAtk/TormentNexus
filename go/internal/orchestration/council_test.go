package orchestration

import (
	"context"
	"errors"
	"testing"

	"github.com/borghq/borg-go/internal/ai"
)

func TestRunDebateApprovedFlow(t *testing.T) {
	original := autoRoute
	defer func() { autoRoute = original }()

	responses := []string{
		"Architect plan: add tests and verify routes.",
		"APPROVE - no blocking concerns.",
		"Final implementation plan: ship the tested fallback safely.",
	}
	call := 0
	autoRoute = func(ctx context.Context, messages []ai.Message) (*ai.LLMResponse, error) {
		resp := &ai.LLMResponse{Content: responses[call]}
		call++
		return resp, nil
	}

	result, err := RunDebate(context.Background(), "Ship feature", "ctx")
	if err != nil {
		t.Fatalf("RunDebate returned error: %v", err)
	}
	if !result.Approved {
		t.Fatalf("expected approved result, got %#v", result)
	}
	if result.Consensus != 1.0 {
		t.Fatalf("expected full consensus, got %#v", result)
	}
	if len(result.Contributions) != 3 {
		t.Fatalf("expected 3 contributions, got %#v", result)
	}
	if result.Contributions[0].Role != "Architect" || result.Contributions[1].Role != "Security Reviewer" || result.Contributions[2].Role != "Lead Engineer" {
		t.Fatalf("unexpected roles: %#v", result.Contributions)
	}
}

func TestRunDebateRejectedFlow(t *testing.T) {
	original := autoRoute
	defer func() { autoRoute = original }()

	responses := []string{
		"Architect plan: do risky thing.",
		"This has severe security issues.",
		"REJECTED: unsafe plan.",
	}
	call := 0
	autoRoute = func(ctx context.Context, messages []ai.Message) (*ai.LLMResponse, error) {
		resp := &ai.LLMResponse{Content: responses[call]}
		call++
		return resp, nil
	}

	result, err := RunDebate(context.Background(), "Ship risky feature", "ctx")
	if err != nil {
		t.Fatalf("RunDebate returned error: %v", err)
	}
	if result.Approved {
		t.Fatalf("expected rejected result, got %#v", result)
	}
	if result.Consensus != 0.0 {
		t.Fatalf("expected zero consensus, got %#v", result)
	}
}

func TestRunDebatePropagatesArchitectError(t *testing.T) {
	original := autoRoute
	defer func() { autoRoute = original }()

	autoRoute = func(ctx context.Context, messages []ai.Message) (*ai.LLMResponse, error) {
		return nil, errors.New("provider unavailable")
	}

	_, err := RunDebate(context.Background(), "Ship feature", "ctx")
	if err == nil {
		t.Fatal("expected RunDebate to return error")
	}
	if err.Error() != "architect failed: provider unavailable" {
		t.Fatalf("unexpected error: %v", err)
	}
}
