package workflow

import (
	"context"
	"errors"
	"testing"
)

func TestWorkflowRunPassesDependencyOutputs(t *testing.T) {
	wf := NewWorkflow("wf-1", "test", "dependency output flow", []*Step{
		{
			ID:   "step-a",
			Name: "Produce",
			Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
				return map[string]any{"value": "hello"}, nil
			},
		},
		{
			ID:        "step-b",
			Name:      "Consume",
			DependsOn: []string{"step-a"},
			Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
				if got := inputs["step-a.value"]; got != "hello" {
					return nil, errors.New("missing dependency output")
				}
				return map[string]any{"ok": true}, nil
			},
		},
	})

	if err := wf.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if wf.Status != StatusCompleted {
		t.Fatalf("expected workflow completed, got %s", wf.Status)
	}
	if wf.Steps[0].Status != StatusCompleted {
		t.Fatalf("expected step-a completed, got %s", wf.Steps[0].Status)
	}
	if wf.Steps[1].Status != StatusCompleted {
		t.Fatalf("expected step-b completed, got %s", wf.Steps[1].Status)
	}
	if got := wf.Steps[1].Output["ok"]; got != true {
		t.Fatalf("expected step-b output ok=true, got %#v", got)
	}
}

func TestWorkflowRunSkipsDependentStepOnFailure(t *testing.T) {
	wf := NewWorkflow("wf-2", "test", "skip on failure", []*Step{
		{
			ID:   "step-a",
			Name: "Fail",
			Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
				return nil, errors.New("boom")
			},
		},
		{
			ID:        "step-b",
			Name:      "Skipped",
			DependsOn: []string{"step-a"},
			Execute: func(ctx context.Context, inputs map[string]any) (map[string]any, error) {
				return map[string]any{"unexpected": true}, nil
			},
		},
	})

	err := wf.Run(context.Background())
	if err == nil {
		t.Fatal("expected Run to return error")
	}
	if wf.Status != StatusFailed {
		t.Fatalf("expected workflow failed, got %s", wf.Status)
	}
	if wf.Steps[0].Status != StatusFailed {
		t.Fatalf("expected step-a failed, got %s", wf.Steps[0].Status)
	}
	if wf.Steps[1].Status != StatusSkipped {
		t.Fatalf("expected step-b skipped, got %s", wf.Steps[1].Status)
	}
}
