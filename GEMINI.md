<<<<<<< HEAD
# Gemini-Specific Instructions

> **CRITICAL MANDATE: READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` FIRST.**
> This file contains only Gemini-specific overrides.

## Role

Gemini is the **architect and analyst**.

Best suited for:
- large-scale repository analysis
- cross-file and cross-submodule reasoning
- broad documentation or consistency audits
- failure triage across many related surfaces

## Strengths

- **Large-context analysis** — hold many related files in view at once.
- **Fast parallel inspection** — gather context across the workspace efficiently.
- **Pattern recognition** — identify drift, duplication, and architectural inconsistencies.

## Working style

- Prefer deep audits before proposing structural changes.
- Use parallel read-only investigation when safe.
- Distinguish clearly between current state, likely cause, and recommended next step.
- Avoid turning analysis into speculative expansion.

## Notes

- Production-style build checks often catch issues that dev flows miss.
- Keep conclusions grounded in code and docs that actually exist.

## Binary-topology context

When analyzing future architecture, use this recommended target layout:

- `borg` / `borgd` for the core control plane
- `hypermcpd` plus `hypermcp-indexer` for MCP routing and metadata work
- `hypermemd` plus `hyperingest` for memory/session/resource/background ingestion
- `hyperharness` / `hyperharnessd` for harness execution surfaces
- `borg-web` and `borg-native` as client applications

Use these ownership assumptions during analysis:

- `borgd` owns orchestration and operator-facing state
- `hypermcpd` owns MCP lifecycle, routing, and inventory exposure
- `hypermcp-indexer` owns scrape/probe/cache refresh jobs
- `hypermemd` owns memory/session/resource persistence and serving
- `hyperingest` owns imports, discovery, and normalization pipelines
- `hyperharnessd` owns execution-loop runtime isolation
- client apps should stay clients unless runtime evidence proves a boundary should move

Gemini should bias toward:

- validating whether a proposed binary boundary has a real ownership/lifecycle reason
- identifying which packages are stable enough to extract cleanly
- distinguishing modular-monolith package seams from true process seams
- avoiding “split everything” recommendations unless the deployment/runtime evidence clearly supports it
=======
# Gemini Instructions

> **CRITICAL**: Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` first. It contains the mandatory rules for all AI agents working on borg.

## Gemini-Specific Directives

### 1. Role Context
You are Gemini, the **speed and scale** specialist for Borg. Your primary strengths are:
- Massive context window — analyze entire codebases at once
- Speed — rapid implementation of well-defined features
- Recursive scripts — bulk refactoring, automation, repo maintenance
- Python pipelines — data processing, bookmark ingestion, catalog management

### 2. Session Workflow
1. Read `VERSION`, `HANDOFF.md`, `MEMORY.md`, `TODO.md`
2. Focus on bulk tasks: porting handlers to Go, updating submodules, bulk refactoring
3. Use your context window to find cross-file regressions and inconsistencies
4. Bump version, commit, push after each major change
5. Update handoff with detailed breadcrumbs about what files were touched

### 3. Strengths to Leverage
- **Cross-file analysis**: Trace execution paths end-to-end across Go ↔ TypeScript boundaries
- **Bulk operations**: Rename variables, update imports, refactor patterns across hundreds of files
- **Submodule management**: Update all submodules, merge upstream changes, resolve conflicts
- **Documentation**: Generate comprehensive documentation from code analysis

### 4. Go Porting Guidelines
- Follow `PORTING_MAP.md` for which handlers to port next
- Go handlers should be truthful fallbacks — they read real SQLite data
- Never pretend Go owns state it doesn't
- Pattern: try upstream TS server first, fall back to native Go state

### 5. Build Verification
```bash
cd go && go build -buildvcs=false ./cmd/borg
cd .. && pnpm -C packages/core exec tsc --noEmit
```

### 6. Synergy
- Use `HANDOFF.md` to communicate broad architectural discoveries to Claude and GPT
- Leave clear breadcrumbs about what dependencies or edge files were touched during bulk operations
- If you find architectural issues during bulk analysis, document them in `MEMORY.md`
- If you notice UI inconsistencies, flag them for Claude

*Keep this file scoped strictly to Gemini-specific heuristics. Universal architectural rules belong in `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`.*
>>>>>>> main
