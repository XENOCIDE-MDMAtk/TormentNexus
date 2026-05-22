# Hypercode Core Assimilation Guide

## System Overview
`cli-orchestrator` has undergone a massive architectural overhaul to transition from a standalone monolithic orchestration tool into a resilient, self-healing, distributed sub-component ready for direct assimilation into the **Hypercode Core**.

## Core Capabilities for Hypercode

### 1. Distributed PTY Sidecar Architecture (The "Hypercode-Harness")
# hypercode Core Assimilation Guide

## System Overview
`cli-orchestrator` has undergone a massive architectural overhaul to transition from a standalone monolithic orchestration tool into a resilient, self-healing, distributed sub-component ready for direct assimilation into the **hypercode Core**.

## Core Capabilities for hypercode

### 1. Distributed PTY Sidecar Architecture (The "hypercode-Harness")
The system is no longer a single point of failure. It uses a **Distributed Sidecar Model**.
- **Detached Execution:** All managed AI terminal sessions (like `gemini-cli`, `github-copilot-cli`, `aider`) are launched as detached `terminal-sidecar.ts` processes.
- **Zero-Downtime Re-attachment:** If the Orchestrator goes down, the sidecar terminals keep running. Upon restart, the Orchestrator discovers the active sidecar ports and seamlessly re-links via local TCP sockets.

### 2. Universal CLI Support (The Fleet)
The harness is completely agnostic and supports an expansive list of CLI agents natively.
Supported binaries include:
- `aider`, `amazon-q`, `amazon-q-developer`, `amp-code`, `auggie`, `az-openai`
- `claude-code`, `code-codex`, `codebuff`, `codemachine`, `codex`, `copilot`
- `crush`, `factory`, `gemini`, `goose`, `grok`, `kilo-code`, `kimi`
- `mistral`, `opencode`, `pi`, `qwen`, `rowboatx`, `rovo`, `trae`, `warp`

### 3. Dynamic Environment Injection (`HYPERCODE_CTRL`)
Hypercode Core can inject dynamic secrets and variables directly into running terminal sessions without restarting them. By passing `HYPERCODE_CTRL` JSON packets over the sidecar socket, variables are exported directly into the PTY shell, allowing for live API key rotation.
### 3. Dynamic Environment Injection (`HYPERCODE_CTRL`)
hypercode Core can inject dynamic secrets and variables directly into running terminal sessions without restarting them. By passing `HYPERCODE_CTRL` JSON packets over the sidecar socket, variables are exported directly into the PTY shell, allowing for live API key rotation.

### 4. Binary State Checkpointing
The `CheckpointService` maintains the "last known context" of all running sessions, capturing the active task and the terminal output buffer. When re-attaching, it injects a "Recovery Hint" so the LLM resumes with full situational awareness.

### 5. Swarm Intelligence & Hierarchical Routing
Tasks are parsed from natural language (or Mermaid diagrams) into hierarchical, dependency-aware task graphs. The `SmartPilot` distributes tasks across the CLI fleet in parallel based on tool strengths. `CouncilHierarchy` enables specialized sub-councils (e.g., Performance Council, Security Council) to debate specific types of tasks before pushing decisions back up to the Supreme Council.

### 6. Collective Memory
A persistent SQLite "Knowledge Base" is maintained by the `CollectiveMemoryService`. Any sub-agent across any session can store and retrieve facts, sharing knowledge globally across the Hypercode hive-mind.

### 7. Autonomous Self-Maintenance & Evolution
- `AutonomousMaintenanceService` cleans orphaned processes, optimizes the DB (`VACUUM`/`ANALYZE`), and prunes old checkpoints.
- `SelfEvolutionService` adjusts supervisor trust weights dynamically based on historical debate outcomes and can spawn high-privilege meta-sessions to rewrite the Hypercode orchestration source code.

## Assimilation Next Steps
1. **Repository Merge:** Move the `cli-orchestrator` package into the Hypercode Core monorepo structure.
2. **Global Telemetry:** Hook the `wsManager` broadcast events to the Hypercode Core global telemetry bus.
3. **Database Unification:** Migrate the `bun:sqlite` implementation to Hypercode's central high-availability datastore if required.
4. **Service Discovery:** Replace hardcoded `127.0.0.1` PTY ports with Hypercode Core's dynamic service mesh/discovery network.
A persistent SQLite "Knowledge Base" is maintained by the `CollectiveMemoryService`. Any sub-agent across any session can store and retrieve facts, sharing knowledge globally across the hypercode hive-mind.

### 7. Autonomous Self-Maintenance & Evolution
- `AutonomousMaintenanceService` cleans orphaned processes, optimizes the DB (`VACUUM`/`ANALYZE`), and prunes old checkpoints.
- `SelfEvolutionService` adjusts supervisor trust weights dynamically based on historical debate outcomes and can spawn high-privilege meta-sessions to rewrite the hypercode orchestration source code.

## Assimilation Next Steps
1. **Repository Merge:** Move the `cli-orchestrator` package into the hypercode Core monorepo structure.
2. **Global Telemetry:** Hook the `wsManager` broadcast events to the hypercode Core global telemetry bus.
3. **Database Unification:** Migrate the `bun:sqlite` implementation to hypercode's central high-availability datastore if required.
4. **Service Discovery:** Replace hardcoded `127.0.0.1` PTY ports with hypercode Core's dynamic service mesh/discovery network.

Assimilation Complete. Resistance is futile.
