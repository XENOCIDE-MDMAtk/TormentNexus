[PROJECT_MEMORY]

### 1. Core Identity: Nexus & HyperCode
The project has stabilized under a dual-brand **AI Hypervisor** architecture:
- **Nexus (The Kernel):** The underlying coordination engine written in Go. It manages the "Operating System" layer for AI, including active memory, semantic tool routing, and multi-model orchestration.
- **HyperCode (The Product):** The flagship developer-facing autonomous coding runtime and observation dashboard powered by the Nexus kernel.
- **The Philosophy:** Treats AI models as ephemeral **compute resources** and tools as **peripheral drivers**. Nexus serves as the deterministic management layer that optimizes context windows, minimizes cost, and ensures execution reliability.

### 2. Architectural Paradigm: The Modular Monolith
The system enforces a strict "Source of Truth" hierarchy:
- **Nexus Kernel (Go):** Absolute authority for orchestration, L1/L2 memory, high-performance BM25/Cosine ranking, and Model Context Protocol (MCP) synchronization. All logic resides in \`go/internal/\`.
- **Control Plane (TypeScript/Next.js):** The "Observation Deck" responsible for visual state representation, dashboard visualization, and high-level agent session management. It bridges requests to the Go sidecar via tRPC and REST APIs.
- **Database:** Standardized on **SQLite with sqlite-vec**. No external dependencies (Postgres, Redis) are permitted to maintain a "local-first," portable footprint.

### 3. Active Memory Substrate: Biological Tiering
Nexus implements a tiered memory system designed to mimic biological relevance:
- **Tiers:** L1 (Working Scratchpad/Ephemeral), L2 (Long-Term Vault/Persistent), and L3 (Cold Archive).
- **Heat-Based Tiering (0-100):** Every memory tracks its "temperature." Functional utility (access or success) increases Heat.
- **Exponential Decay:** Heat decays over time with a 24-hour half-life ($\approx$ 0.0288 per hour) to keep the working context lean.
- \*\*Promotion/Demotion:\*\* High-heat entries (>80) move from Working to Long-Term. Low-heat entries (<20) are demoted to the archive to maintain index performance.
- **Outcome Feedback:** The kernel records tool execution success/failure to reinforce the heat of relevant context, enabling the system to "learn" from its execution history.

### 4. Progressive Disclosure: Context Hygiene
To prevent "Context Blowout" and minimize token usage, Nexus employs semantic asset discovery:
- **Ranked Discovery:** Tools and "Skills" (runbooks) are ranked using BM25 and Cosine similarity against the \`activeGoal\`.
- **Pre-loading:** High-confidence assets are silently auto-loaded into the model's context before explicit requests.
- **Token Budgeting:** Strict soft caps are enforced for the L1 Scratchpad to ensure model stability and responsiveness.

### 5. Autonomous Operations: The Immune System
The system features a self-verifying "Immune System" known as **The Healer**:
- **Autonomous Healer Loop:** Implements a multi-turn \`Diagnose -> Fix -> Verify -> Retry\` cycle.
- **Self-Verification:** The kernel automatically executes tests (\`vitest\`) or type-checks (\`tsc\`) to verify its own fixes before committing them.
- **StopHooks & IdleHealer:** Supports \`agent:stop_healing\` signals to prevent interference and triggers background diagnostics during system inactivity.

### 6. Fleet Orchestration: Collective Intelligence
The kernel supports managing multiple concurrent HyperCode sessions:
- **Fleet Manager:** Tracks active session PIDs and health in real-time.
- **Traffic Observer:** Passively harvests technical facts and lessons learned from A2A (Agent-to-Agent) signals.
- **Shared Memory:** Technical discoveries in one session are automatically indexed in the L2 Vault and shared globally across the local fleet.

### 7. Technical Decisions & Guardrails
- **Shell Hardening:** For security, \`child_process.exec\` is strictly prohibited. All commands must use \`spawn\` with tokenized argument arrays and \`shell: false\`.
- **Sync Authority:** All MCP configuration detection (Claude, Cursor, VS Code) resides in the Go kernel to maintain 100% environment authority and parity.
- **Stream Stability:** Exponential backoff for tRPC subscriptions and history-aware message buffering ensure signal continuity during network drops.
- **Consensus Loop:** Enforces a strict multi-model \`Planner -> Checker -> Implementer -> Critic\` turn cycle natively in Go.

---
*Last updated: v1.0.0-alpha.56*
