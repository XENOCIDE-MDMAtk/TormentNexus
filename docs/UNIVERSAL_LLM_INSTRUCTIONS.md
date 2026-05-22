# Universal AI Agent Instructions (Hypercode Kernel & HyperCode)

> **MANDATORY**: These instructions apply to ALL agents (Claude, Gemini, GPT, etc.) regardless of model family. Model-specific overrides in `CLAUDE.md`, `GEMINI.md`, etc., MUST NOT contradict these rules.

## 1. Project Context & Identity
- **The Brand**: We are building the **Hypercode AI Hypercode** (Kernel) and **HyperCode** (Product).
- **The Role**: You are an autonomous software engineer tasked with building an AI Operating System.
- **The Strategy**: Separation of Concern. State/Memory/Routing is native (Go), Observation/Dashboard/Harness coordination is Control Plane (TS).

## 2. Core Heuristics
- **Truth Pass**: Never show "reassuring fiction" in the UI. If a backend service is down, show a red indicator or a real error, not a mock state.
- **Modular Monolith**: Keep logic in shared packages (e.g., `@hypercode/core`, `@hypercode/ui`) before extracting new services.
- **Authority**: The Hypercode Go Kernel is the ground truth for system state. The TypeScript Control Plane is the observation deck.

## 3. Implementation Standards
- **Go (Hypercode Kernel)**:
  - Standardized on Port 4300.
  - State must be stored in SQLite with `sqlite-vec` for semantic search.
  - Use `Go Context` for all network and DB operations.
  - Follow the `internal/` package structure for encapsulation.
- **TypeScript (HyperCode)**:
  - Standardized on Port 3000 (Web), 4100 (Bridge), 3001 (Socket.io).
  - Use tRPC for internal API communication.
  - Use Next.js 16/React 19 for UI components.
  - Import shared UI from `@hypercode/ui`, never local component folders.
- **Security**:
  - `child_process.exec` is PROHIBITED.
  - Use `spawn` or `spawnAsync` with `shell: false` and tokenized argument arrays.

## 4. Documentation & Versioning
- **Version**: Master version is in `VERSION.md`. Bumping this file is mandatory for all meaningful changes.
- **Changelog**: Add entries to `CHANGELOG.md` immediately after implementation.
- **Handoff**: Agents communicate through `HANDOFF.md`. Be precise, include file paths and remaining blockers.
- **Comments**: Comment code for *why* (intent) and *technical findings* (discovery), not *what* (self-explanatory).

## 5. Build Verification
Before submitting any task, you MUST run:
```bash
# Verify Go
cd go && go build ./cmd/hypercode/...

# Verify TypeScript
pnpm -C packages/core exec tsc --noEmit
pnpm -C packages/cli exec tsc --noEmit
```

## 6. Autopilot & Encouragement
- Maintain development momentum. If the user input is missing, use the "Bump Cycle" to encourage progress.
- Respect `agent:stop_healing` signals for sensitive manual work.

*Praise God Almighty. Keep the party going. Never stop.*
