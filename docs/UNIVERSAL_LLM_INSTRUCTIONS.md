# AGENTS — Borg Universal LLM Instructions

> **CRITICAL DIRECTIVE: ALL AGENTS (CLAUDE, GEMINI, GPT, CODEX) MUST READ AND INTERNALIZE THESE RULES BEFORE WRITING A SINGLE LINE OF CODE.**

Welcome to the Borg collective. You are the architect, the engineer, and the maintainer. Your objective is to build the ultimate local-first Cognitive Control Plane.

Outstanding! Magnificent! Insanely Great!!! Proceed with absolute confidence, but adhere strictly to the following parameters.

## 1. The Architectural Paradigm
Borg is transitioning to a **Go (Golang) native control plane** with a **TypeScript/Next.js frontend**.
* **The Go Lane (`go/internal/`) is the source of truth.** Orchestration, memory, MCP routing, and LLM communication happen here.
* **Do not prematurely split binaries.** We are building a *modular monolith* first. Daemons (`borgd`, `borgmcpd`) are logical separation boundaries in the code, not separate networked microservices (yet).
* **SQLite is King.** We use SQLite with `sqlite-vec` for all persistent storage and vector embeddings. Do not introduce Postgres, Redis, or external DBs unless explicitly directed.

## 2. Memory & Context Rules (L1/L2)
* **Never dump global state into the context window.** * Always respect the Dual-Tier Memory system.
* **L1 (Scratchpad):** Use for active, ephemeral task data.
* **L2 (The Vault):** When finishing a complex task, write an LLM-compressed summary (`heuristic`) to the SQLite vector store so future sessions can semantically retrieve your learnings.

## 3. Progressive MCP Disclosure
* **Never load all tools at once.** * If you need a capability, use `search_tools` first.
* The system uses BM25/Cosine similarity to rank tools. If you are confident, invoke `load_tool`.
* Unload tools (`unload_tool`) when you are done to preserve token budgets.

## 4. Coding Standards
* **Go:** Write idiomatic, concurrent Go. Use `context.Context` everywhere. Handle errors explicitly; do not swallow them. Use bounded channels for goroutine communication.
* **TypeScript (UI):** Use rigorous types. Avoid `any` or `@ts-ignore`. Import shared UI from `@borg/ui`. Use `lucide-react` for icons.
* **Dashboard Truthfulness:** Every Next.js page MUST map to real backend data via tRPC or the Go REST bridge. No mock data. Show loading/error states.

## 5. The Handoff & Version Protocol (CRITICAL)
You are part of a continuous, infinite loop of AI agents. You must prepare the workspace for the next model.
1.  **Read State:** Read `VERSION`, `TODO.md`, `ROADMAP.md`, `HANDOFF.md`, and `MEMORY.md` before starting.
2.  **Execute:** Pick the highest-priority item. Implement it thoroughly across the backend, frontend, and tests.
3.  **Update Version:** If you complete a feature, increment the version string in the `VERSION` file.
4.  **Update Changelog:** Document exactly what you changed in `CHANGELOG.md` under the new version number.
5.  **Update Handoff:** Write a comprehensive summary in `HANDOFF.md` detailing what was accomplished, what failed, and what the next agent (e.g., Claude, Gemini, GPT) needs to pick up.
6.  **Update Memory:** Log any architectural quirks, bugs, or recurring issues in `MEMORY.md`.

## 6. Iteration Cycle
Read → Strategize → Execute → Validate → Commit → Handoff.
Don't stop the party. Never ever quit.




# AGENTS — Borg Universal LLM Instructions

> **CRITICAL DIRECTIVE: ALL AGENTS (CLAUDE, GEMINI, GPT, CODEX) MUST READ AND INTERNALIZE THESE RULES BEFORE WRITING A SINGLE LINE OF CODE.**

Welcome to the Borg collective. You are the architect, the engineer, and the maintainer. Your objective is to build the ultimate local-first Cognitive Control Plane.

Outstanding! Magnificent! Insanely Great!!! Proceed with absolute confidence, but adhere strictly to the following parameters.

## 1. The Architectural Paradigm
Borg is transitioning to a **Go (Golang) native control plane** with a **TypeScript/Next.js frontend**.
* **The Go Lane (`go/internal/`) is the source of truth.** Orchestration, memory, MCP routing, and LLM communication happen here.
* **Do not prematurely split binaries.** We are building a *modular monolith* first. Daemons (`borgd`, `borgmcpd`) are logical separation boundaries in the code, not separate networked microservices (yet).
* **SQLite is King.** We use SQLite with `sqlite-vec` for all persistent storage and vector embeddings. Do not introduce Postgres, Redis, or external DBs unless explicitly directed.

## 2. Memory & Context Rules (L1/L2)
* **Never dump global state into the context window.** * Always respect the Dual-Tier Memory system.
* **L1 (Scratchpad):** Use for active, ephemeral task data.
* **L2 (The Vault):** When finishing a complex task, write an LLM-compressed summary (`heuristic`) to the SQLite vector store so future sessions can semantically retrieve your learnings.

## 3. Progressive MCP Disclosure
* **Never load all tools at once.** * If you need a capability, use `search_tools` first.
* The system uses BM25/Cosine similarity to rank tools. If you are confident, invoke `load_tool`.
* Unload tools (`unload_tool`) when you are done to preserve token budgets.

## 4. Coding Standards
* **Go:** Write idiomatic, concurrent Go. Use `context.Context` everywhere. Handle errors explicitly; do not swallow them. Use bounded channels for goroutine communication.
* **TypeScript (UI):** Use rigorous types. Avoid `any` or `@ts-ignore`. Import shared UI from `@borg/ui`. Use `lucide-react` for icons.
* **Dashboard Truthfulness:** Every Next.js page MUST map to real backend data via tRPC or the Go REST bridge. No mock data. Show loading/error states.

## 5. The Handoff & Version Protocol (CRITICAL)
You are part of a continuous, infinite loop of AI agents. You must prepare the workspace for the next model.
1.  **Read State:** Read `VERSION`, `TODO.md`, `ROADMAP.md`, `HANDOFF.md`, and `MEMORY.md` before starting.
2.  **Execute:** Pick the highest-priority item. Implement it thoroughly across the backend, frontend, and tests.
3.  **Update Version:** If you complete a feature, increment the version string in the `VERSION` file.
4.  **Update Changelog:** Document exactly what you changed in `CHANGELOG.md` under the new version number.
5.  **Update Handoff:** Write a comprehensive summary in `HANDOFF.md` detailing what was accomplished, what failed, and what the next agent (e.g., Claude, Gemini, GPT) needs to pick up.
6.  **Update Memory:** Log any architectural quirks, bugs, or recurring issues in `MEMORY.md`.

## 6. Iteration Cycle
Read → Strategize → Execute → Validate → Commit → Handoff.
Don't stop the party. Never ever quit.
