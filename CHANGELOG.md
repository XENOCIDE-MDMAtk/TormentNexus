# Changelog

## [1.0.0-alpha.62] - 2026-05-19

### Added
- **Deep Link Protocol Scheme (`hypercode://`) in Go**:
  - Built robust URI handling for `hypercode://attach?session=ID` and `hypercode://create?cliType=aider` commands.
  - Implemented single-instance CLI dispatcher. Clicking deep links routes actions through the active `hypercoded` daemon via HTTP REST.
- **SQLite L2 Vector Vault Visualizer**:
  - Implemented persistent database queries (`GetAllVaultRecords`) in Go fetching chronic vault memories ordered by importance and heat.
  - Wired the new tRPC `vaultRecords` query to the Next.js control plane to hook persistent SQLite vector records into the UI.
  - Re-designed the Healer dashboard in glassmorphic dark-mode, showing streaming active pathogens side-by-side with real persistent L2 Vault records.
- **Next.js Dashboard Routes**:
  - Added premium, highly interactive dashboard console cards for Blocks, Claude Chrome, Claude Cloud, Copilot, and OpenAI Codex.
- **LLM Instruction Unification**:
  - Resolved merge conflict markers and aligned role guidelines across `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md` under `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`.

### Changed
- Standardized documentation identity to Hypercode Kernel & HyperCode.
- Replaced git merge conflict markers across multiple internal Kotlin and Markdown files with unified content logic.

## [1.0.0-alpha.61] - 2026-05-17

### Added
- **Autonomous Healer Loop (The Immune System)**:
  - New `HealerService` in the Go kernel with a multi-turn `diagnose -> fix -> verify -> retry` loop.
  - Integration with `CodeExecutor` for native, sandboxed verification (tsc, vitest, go test).
  - L2 Vault persistence: All healing events and extracted facts are saved as long-term memory for fleet-wide intelligence sharing.

### Changed
- Standardized documentation identity to Hypercode Kernel & HyperCode.
- Updated `VERSION.md`, `ROADMAP.md`, and `TODO.md` to reflect Phase 5 active sprint goals.
- Unified `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` as the single source of truth for all AI agents.
- Resolved merge conflict markers and aligned role guidelines across `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md`.

## [1.0.0-alpha.60] - 2026-05-16

### Added
- Fully integrated Go-native `MemoryManager` into the core TS control plane.
- Wires up `sqlite-vec` storage backend, replacing the deprecated `@hypercode/hypercode` implementation.
- Dual-tier cache invalidation for the L1/L2 memory boundaries.

### Changed
- Shifted authority of MCP configuration sync entirely to the Go sidecar.
- Removed legacy TS synchronization scripts for VSCode and Cursor.
