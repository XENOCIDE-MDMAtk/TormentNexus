package vault

import (
	"fmt"
	"strings"
)

type Bridge struct {
	vault   *Vault
	embedFn EmbedFunc
}

func NewBridge(v *Vault, embedFn EmbedFunc) *Bridge {
	return &Bridge{vault: v, embedFn: embedFn}
}

func (b *Bridge) InjectContext(agentName, project, prompt string) (*Scratchpad, error) {
	sp, err := NewScratchpad()
	if err != nil { return nil, fmt.Errorf("create scratchpad: %w", err) }
	var systemCtx string
	if b.embedFn != nil {
		queryVec, err := b.embedFn(prompt)
		if err == nil && len(queryVec) > 0 {
			hits, err := b.vault.Search(queryVec, 5, 0.3, true)
			if err == nil && len(hits) > 0 { systemCtx = b.formatContext(hits) }
		}
	}
	if systemCtx == "" {
		entries, err := b.vault.RecentEntries(3, MemoryHeuristic)
		if err == nil && len(entries) > 0 {
			var parts []string
			for _, e := range entries { parts = append(parts, "- "+e.Content) }
			systemCtx = "[Recent Context]\n" + strings.Join(parts, "\n")
		}
	}
	if err := sp.InitSession(agentName, project, prompt, systemCtx); err != nil { sp.Close(); return nil, fmt.Errorf("init session: %w", err) }
	return sp, nil
}

func (b *Bridge) CommitSession(sp *Scratchpad, heuristicSummary string) error {
	session, entries, err := sp.Drain()
	if err != nil { return fmt.Errorf("drain: %w", err) }
	if err := b.vault.CommitSession(session, entries, heuristicSummary, b.embedFn); err != nil { return fmt.Errorf("commit: %w", err) }
	return nil
}

func (b *Bridge) formatContext(hits []ContextHit) string {
	var sb strings.Builder
	sb.WriteString("[Relevant Memory Context]\n")
	sb.WriteString("The following memories from past sessions may be relevant:\n\n")
	for _, hit := range hits {
		label := "Summary"
		if hit.Entry.MemoryType == MemoryRaw { label = "Transcript" }
		sb.WriteString(fmt.Sprintf("## %s (score=%.2f, session=%s)\n%s\n\n", label, hit.Score, shortID(hit.Entry.SessionID), hit.Entry.Content))
	}
	return sb.String()
}

func shortID(id string) string {
	if len(id) > 8 { return id[:8] }
	return id
}
