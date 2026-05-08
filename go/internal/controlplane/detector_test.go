package controlplane

import (
	"slices"
	"testing"
	"time"
)

func TestNewDetectorIncludesExpandedCLIInventory(t *testing.T) {
	detector := NewDetector(time.Second, time.Minute)

	cliTypes := make([]string, 0, len(detector.definitions))
	for _, definition := range detector.definitions {
		cliTypes = append(cliTypes, definition.Type)
	}

	for _, expected := range []string{
		"borg",
		"antigravity",
		"claude",
		"continue",
		"cody",
		"amazon-q",
		"amazon-q-developer",
		"azure-openai",
		"bito",
		"byterover",
		"code-codex",
		"dolt",
		"factory",
		"factory-droid",
		"grok",
		"jules",
		"kilo-code",
		"kimi",
		"llm",
		"litellm",
		"llamafile",
		"manus",
		"mistral-vibe",
		"ollama",
		"open-interpreter",
		"qwen-code",
		"rowboatx",
		"rovo",
		"shell-pilot",
		"smithery",
		"superai-cli",
		"trae",
		"warp",
	} {
		if !slices.Contains(cliTypes, expected) {
			t.Fatalf("expected detector inventory to include %q, got %v", expected, cliTypes)
		}
	}

	definitionByType := make(map[string]definition, len(detector.definitions))
	for _, definition := range detector.definitions {
		definitionByType[definition.Type] = definition
	}

	if definitionByType["factory-droid"].Command != "factory-droid" {
		t.Fatalf("expected factory-droid command parity, got %+v", definitionByType["factory-droid"])
	}
	if definitionByType["qwen-code"].Command != "qwen" {
		t.Fatalf("expected qwen-code detector command parity, got %+v", definitionByType["qwen-code"])
	}
	if definitionByType["open-interpreter"].Command != "interpreter" {
		t.Fatalf("expected open-interpreter detector command parity, got %+v", definitionByType["open-interpreter"])
	}
}
