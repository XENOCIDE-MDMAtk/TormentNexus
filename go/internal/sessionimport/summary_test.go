package sessionimport

import "testing"

func TestBuildSummaryAggregatesValidatedCandidates(t *testing.T) {
	summary := BuildSummary([]ValidationResult{
		{
			SourceTool:     "claude-code",
			SourceType:     "session",
			Format:         "jsonl",
			EstimatedSize:  10,
			Valid:          true,
			DetectedModels: []string{"claude", "gpt-5"},
		},
		{
			SourceTool:     "openai",
			SourceType:     "conversation-export",
			Format:         "json",
			EstimatedSize:  20,
			Valid:          false,
			DetectedModels: []string{"gpt-5"},
			Errors:         []string{"invalid JSON document"},
		},
		{
			SourceTool:     "claude-code",
			SourceType:     "session",
			Format:         "jsonl",
			EstimatedSize:  30,
			Valid:          true,
			DetectedModels: []string{"claude"},
		},
	})

	if summary.Count != 3 {
		t.Fatalf("expected 3 candidates, got %d", summary.Count)
	}
	if summary.ValidCount != 2 || summary.InvalidCount != 1 {
		t.Fatalf("expected valid/invalid counts 2/1, got %d/%d", summary.ValidCount, summary.InvalidCount)
	}
	if summary.TotalEstimatedSize != 60 {
		t.Fatalf("expected total estimated size 60, got %d", summary.TotalEstimatedSize)
	}
	if len(summary.BySourceTool) == 0 || summary.BySourceTool[0].Key != "claude-code" || summary.BySourceTool[0].Count != 2 {
		t.Fatalf("expected claude-code bucket first, got %+v", summary.BySourceTool)
	}
	if len(summary.ByModelHint) == 0 || summary.ByModelHint[0].Key != "claude" || summary.ByModelHint[0].Count != 2 {
		t.Fatalf("expected claude model bucket first, got %+v", summary.ByModelHint)
	}
	if len(summary.ByError) == 0 || summary.ByError[0].Key != "invalid JSON document" || summary.ByError[0].Count != 1 {
		t.Fatalf("expected invalid JSON bucket first, got %+v", summary.ByError)
	}
}
