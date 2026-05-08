package mcp

import "testing"

func TestTokenizeFiltersShortTokensAndNormalizes(t *testing.T) {
	got := Tokenize("Go AI, Search-tools for MCP!!")
	want := []string{"search", "tools", "for", "mcp"}
	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected token[%d]=%q, got %q", i, want[i], got[i])
		}
	}
}

func TestRankToolsPrefersNameMatches(t *testing.T) {
	tools := []ToolEntry{
		{Name: "core__search_tools", Description: "General repository search", Server: "core"},
		{Name: "core__planner", Description: "Search through implementation plans", Server: "core"},
	}

	ranked := RankTools("search", tools, 10)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 ranked results, got %#v", ranked)
	}
	if ranked[0].Name != "core__search_tools" {
		t.Fatalf("expected name match to rank first, got %#v", ranked)
	}
	if ranked[0].Score <= ranked[1].Score {
		t.Fatalf("expected first result to score higher, got %#v", ranked)
	}
}

func TestRankToolsUsesTagsAndSemanticGroups(t *testing.T) {
	tools := []ToolEntry{
		{Name: "core__planner", Description: "Plan work", Server: "core", ToolTags: []string{"roadmap"}, SemanticGroup: "planning", SemanticGroupLabel: "Planning"},
		{Name: "core__search_tools", Description: "Search code", Server: "core", ToolTags: []string{"code"}, SemanticGroup: "discovery", SemanticGroupLabel: "Discovery"},
	}

	ranked := RankTools("planning roadmap", tools, 10)
	if len(ranked) == 0 {
		t.Fatalf("expected ranked results, got %#v", ranked)
	}
	if ranked[0].Name != "core__planner" {
		t.Fatalf("expected planning-tagged tool first, got %#v", ranked)
	}
	if ranked[0].ScoreBreakdown["tags"] == 0 && ranked[0].ScoreBreakdown["semantic_group"] == 0 {
		t.Fatalf("expected tag or semantic score contribution, got %#v", ranked[0].ScoreBreakdown)
	}
}

func TestRankToolsEmptyQueryReturnsDefaultListing(t *testing.T) {
	tools := []ToolEntry{{Name: "a"}, {Name: "b"}}
	ranked := RankTools("", tools, 1)
	if len(ranked) != 1 {
		t.Fatalf("expected 1 result with limit 1, got %#v", ranked)
	}
	if ranked[0].Rank != 1 || ranked[0].MatchReason != "Default listing" {
		t.Fatalf("unexpected default listing result: %#v", ranked[0])
	}
}
