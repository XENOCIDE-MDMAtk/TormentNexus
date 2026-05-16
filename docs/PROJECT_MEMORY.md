[PROJECT_MEMORY]

# AI Hypervisor (Nexus) - Comprehensive Architectural Memory

This document summarizes the foundational architecture, established patterns, and strategic decisions of the project as of version **1.0.0-alpha.56**.

## 1. Strategic Identity: Nexus & HyperCode
The project has successfully pivoted from "Borg" to a dual-brand infrastructure model:
*   **Nexus (The Kernel/Hypervisor):** The underlying coordination runtime and "AI Hypervisor." It treats LLMs as "guest operating systems" and manages the low-level memory, routing, and execution buses.
*   **HyperCode (The Product):** The user-facing, local-first autonomous coding environment powered by the Nexus kernel.

## 2. Active Tiered Memory Substrate (Implemented Phase 1)
*   **Heat Scoring (0-100):** Every memory entry tracks utility. Heat increases on access and decays exponentially (24-hour half-life).
*   **Tiered Hierarchy:** L1 (Working Memory, heat > 80) is promoted to context; L2 (Vault) is for semantic recall.
*   **Tool-Outcome Feedback:** `MemoryManager.recordToolOutcome()` boosts the heat of successful patterns and demotes failures.

## 3. "Kernel / Control Plane" Topology
*   **/kernel**: Deterministic brain (runtime, memory, router).
*   **/control-plane**: Observer layer (UI, Telemetry).

## 4. State Authority & The Sidecar Pattern
*   **Go Sidecar (Port 4300):** State authority and BM25 ranking.
*   **TS Bridge (Port 4100):** Primary control-plane bridge and tRPC host.

## 5. Intelligence Management: Progressive Disclosure
*   **Decision System:** Ranked discovery and LRU eviction ensures only 3-5 tools/skills are in the active working set.

## 6. Hardened Execution & Security
*   **Standard:** Tokenized argument arrays with `shell: false` for all command executions.
*   **Parity Principle:** 1:1 behavioral and schema parity for tools expected by proprietary models (e.g., Claude Code).

---
### Meta-Protocol for Future Sessions
1.  **Truth over Fiction:** Dashboards must reflect real state.
2.  **Autonomous Momentum:** Proceed through Phase 2 (Autonomy/Self-Healing).
3.  **Documentation Sync:** Every version bump syncs all meta files and manifests.
