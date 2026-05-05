# Borg (Super AI Plugin) - Comprehensive Project Summary

This document serves as the canonical record of the Borg project's architecture, design patterns, and strategic decisions, synthesized from deep codebase exploration and the implementation of critical performance optimizations.

## 🏗️ 1. Core Architecture: Local-First Modular Monolith
Borg is a high-performance, privacy-centric AI orchestration hub. It is structured as a **TypeScript monorepo** (managed by `pnpm`), with an active strategic pivot toward a **Go (Golang) native control plane** ("The Go Lane") for performance-critical logic. It adheres to a **local-first** philosophy, prioritizing local data ownership, privacy, and low-latency execution.

### Key Workspace Breakdown (`packages/` & `apps/`)
- **`packages/core/` (The Hub)**: The central backbone service. It manages database persistence, MCP (Model Context Protocol) server lifecycles, agent execution loops, and real-time state synchronization. Built using **Fastify/Hono** and **Socket.io**.
- **`apps/web/` (The Control Panel)**: A rich **Next.js 16 / React** dashboard for operator visibility, configuration, and monitoring. Uses **tRPC** for end-to-end type safety between the frontend and the Hub.
- **`packages/cli/` (`borg` / `super-ai`)**: A terminal interface built with **Commander** for direct system control, session management, and agent orchestration.
- **`go/` (The Go Lane)**: A high-performance sidecar/replacement layer implementing core logic (orchestration, memory, code execution) in Go for superior concurrency and resource efficiency.
- **Adapters & Extensions**: Specialized packages for LLM provider integration (`packages/adapters/`), VSCode integration (`packages/vscode/`), and browser automation (`packages/browser/`).

## 💾 2. Data Management & Persistence
- **Relational Storage**: Standardized on **Drizzle ORM** with the **SQLite** (`better-sqlite3`) driver. This supports a zero-config local setup while maintaining a strict, type-safe schema.
- **Repository Pattern**: Data access is organized into modular repositories (e.g., `ToolsRepository`, `ToolSetsRepository`).
- **N+1 Query Resolution Pattern**: A core performance standard is replacing sequential hydration loops (O(N) queries) with **batched fetching** using Drizzle's `inArray` operator (O(1) queries), followed by efficient in-memory grouping.
- **Vector Storage**: Uses **LanceDB** in the TypeScript stack and **sqlite-vec** in the Go lane for local semantic search, tool ranking, and Retrieval-Augmented Generation (RAG).
- **Dual-Tier Memory (L1/L2)**:
    - **L1 (Scratchpad)**: Ephemeral, fast memory tied to the active session (Chain of Thought, tool outputs).
    - **L2 (The Vault)**: Persistent, semantic storage for exact transcripts (`raw`) and LLM-compressed lessons learned (`heuristic`).
- **State Portability**: All sessions, handoffs, and configuration profiles are persisted as JSON/Markdown files to ensure the system state is portable, inspectable, and resilient to restarts.

## 🤖 3. AI Orchestration & MCP (Model Context Protocol)
- **Universal Aggregator**: Borg acts as a central proxy and aggregator for MCP servers, providing a unified tool surface to agents.
- **Progressive Disclosure Pattern**: To maintain context hygiene and reduce token bloat, the system implements a tiered discovery engine:
    - **Layer 1 (Search)**: Semantic/BM25-style ranking of the global tool/skill inventory.
    - **Layer 2 (The Router)**: Injection of only the top highly relevant tool schemas (top 5-10) into the model context.
    - **Layer 3 (Auto-Load)**: Silent loading of tools on high-confidence matches.
    - **Layer 4 (LRU Eviction)**: Graceful unloading of idle tools based on an idle-first policy.
- **Agent-to-Agent (A2A) Protocol**: A custom communication standard enabling multi-model swarms (Planner, Implementer, Tester, Critic) to debate, bid on tasks, and reach consensus.

## 🛠️ 4. Key Engineering Patterns & Decisions
- **System Doctor**: A proactive diagnostic service (`SystemDoctor`) that verifies the local environment (Node, Go, Git, Docker) before system startup.
- **Waterfall LLM Routing**: A resilient inference client that catches 429/5xx errors and automatically cascades the payload through a prioritized provider chain (Primary API → OpenRouter → Local).
- **Lazy Initialization**: Resource-heavy MCP servers and services are loaded on-demand to ensure fast startup.
- **Handoff Management**: The `HandoffManager` enables state persistence across restarts, allowing for long-running autonomous tasks.
- **Observability Hooks**: The system emits `pre_tool_call` and `post_tool_call` events, enabling real-time traffic inspection and passive memory extraction via the `TrafficObserver`.
- **Responsive UI Visualization**: Frontend canvas components (like the Knowledge Graph) utilize a shared `useResizeObserver` hook for dynamic, performant layout management.

## 🚀 5. Strategic Engineering Vision
Borg aspires to be the definitive local AI orchestration platform, achieving absolute feature parity with advanced coding tools while providing a superior, integrated UI experience. The roadmap focus remains on **high performance** (Go migration), **context hygiene** (progressive disclosure), and **autonomy** (Council of Supervisors).

---
*Status: The Assigned Performance Task (N+1 query optimization in `ToolSetsRepository.findAll`) has been successfully implemented and submitted.*

I have also attempted to address your request for submodule and repository synchronization. However, I encountered network connectivity issues when trying to reach GitHub (`github.com port 443: Connection timed out`). This prevents me from fetching upstream changes or pushing updates at this moment. I have documented the session history and will include it in the handoff for the next agent.