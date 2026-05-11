# Memory

> **CRITICAL**: Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` first.

## Ongoing Observations & Codebase Knowledge

### 1. The Great Config Split (Fixed in 1.0.0-alpha.1)
**Observation**: Historically, `mcp.jsonc` in the user directory was intended to act as the sole source of truth. However, as SQLite was introduced, a destructive cycle emerged where `McpConfigService.syncWithDatabase()` would wipe out DB tools (and their `always_on` status) if `mcp.jsonc` lacked `_meta.tools`.
**Resolution**: We completely decoupled the manual config from the database. The system now exports a unified `.borg/mcp-cache.json` which the lightweight `stdioLoader` reads. DB tools are no longer destroyed by an empty JSON configuration.

### 2. The `always_on` Advertising Filter
**Observation**: `getDirectModeTools()` in `MCPServer.ts` enforces an "ULTRA-STREAMLINED ADVERTISING" filter. It ONLY returns tools marked `always_on`. If no tools have this flag, it defaults to returning *only* the internal meta-tools (`search_tools`, `load_tool`, etc.).
**Implication**: This is intended behavior to keep LLM context clean. Models are expected to use `search_tools` and `load_tool` to dynamically fetch what they need.

### 3. Config Directory Resolution
**Observation**: `getBorgConfigDir()` historically hardcoded `os.homedir() + '/.borg'`. 
**Resolution**: It now dynamically respects `process.env.BORG_CONFIG_DIR`, and falls back to checking `process.cwd()/mcp.jsonc` before defaulting to the user's home directory. This allows local repository configurations to be authoritative during development.
**Observation**: `getBorgConfigDir()` historically hardcoded `os.homedir() + '/.borg'`. 
**Resolution**: It now dynamically respects `process.env.BORG_CONFIG_DIR`, and falls back to checking `process.cwd()/mcp.jsonc` before defaulting to the user's home directory. This allows local repository configurations to be authoritative during development.

### 4. Binary Extraction Strategy
**Observation**: The project has aggressive plans to split into distinct daemons (`borgd`, `borgmcpd`, etc.).
**Implication**: DO NOT split these prematurely. Follow the modular-monolith-first rule defined in `UNIVERSAL_LLM_INSTRUCTIONS.md`. Treat the Go workspaces as experimental bridges for now.

### 5. better-sqlite3 and Node 24 (Fixed 2026-04-08)
**Observation**: `better-sqlite3@12.6.2` requires native `.node` bindings compiled for the exact Node version. Node v24.10.0 broke compatibility.
**Resolution**: `pnpm rebuild better-sqlite3` works (uses prebuild-install). `node-gyp rebuild` does NOT work on Node 24.
**Implication**: After any `pnpm install`, you MUST run `pnpm rebuild better-sqlite3` on Node 24. Add this to startup checks.

### 6. Gemini Model Names Change Frequently (Updated 2026-04-08)
**Observation**: `gemini-2.0-flash` was deprecated and returns 404. The current free-tier model is `gemini-2.5-flash`.
**Implication**: When adding Gemini models, verify current availability at https://ai.google.dev/gemini-api/docs/models. The ProviderRegistry should be updated whenever Google renames models.

### 7. Dashboard Polling Creates Noise
**Observation**: The Next.js dashboard polls multiple endpoints every 5 seconds, generating a constant stream of HTTP requests. If any endpoint returns 404 (like `/api/scripts`), it creates log spam and wasted cycles.
**Resolution**: Added REST API bridge routes in `orchestrator.ts` that serve the same data as the tRPC router, so the dashboard's native-control-plane fetch path works cleanly.

### 8. Worktree Complexity
**Observation**: The project uses git worktrees with the submodule structure at `.git/modules/borg`. The actual working directory (`borg-push`) can become detached from `main`.
**Resolution**: Manually update the worktree HEAD file to point to `refs/heads/main`. Don't try to use `git checkout main` across worktrees.

### 9. Go Sidecar Version Injection (Added 2026-04-08)
**Observation**: The Go binary had a hardcoded version `0.0.1-experimental` that was completely out of sync with the VERSION file.
**Resolution**: `go/internal/buildinfo/buildinfo.go` now uses `var Version` (not `const`) injected at build time via `-ldflags "-X internal/buildinfo.Version=$VER"`. The build script `scripts/build-go.sh` reads from the VERSION file automatically.
**Implication**: Always use `scripts/build-go.sh` or the ldflags pattern to build the Go binary. Never manually edit buildinfo.go.

### 10. Submodule Stash Pop Conflicts (Added 2026-04-08)
**Observation**: When updating submodules with `git merge origin/main` followed by `git stash pop`, merge conflicts appear in stash-applied files. Using `git checkout --ours` resolves to the merged main version, which is typically correct for our local changes.
**Implication**: When submodules have both upstream updates and local stashed changes, merge upstream first, then pop stash, then resolve conflicts keeping HEAD (the merged result).

### 11. Meta-Tool Decision System is Already Implemented (Added 2026-04-08)
**Observation**: The MCP meta-tool decision system was listed as TODO but is actually fully implemented in `packages/core/src/mcp/`. It includes:
- `NativeSessionMetaTools.ts` — 6 permanent meta-tools (search_tools, load_tool, get_tool_schema, list_loaded_tools, unload_tool, list_all_tools) plus auto_call_tool, search_published_catalog, install_published_server
- `toolSearchRanking.ts` — Multi-signal scoring (exact name, prefix, token, semantic group, tags, description, server name) with profile boosting
- `SessionToolWorkingSet.ts` — LRU + idle-first eviction with configurable caps (default 16 loaded, 8 hydrated)
- Full tRPC endpoints: getWorkingSet, getWorkingSetEvictionHistory, getToolSelectionTelemetry, searchTools
**Implication**: Do NOT re-implement. Focus on improving ranking quality, adding more profiles, and verifying the dashboard inspector shows all this data.

### 12. Package.json Sync Script Pattern (Added 2026-04-08)
**Observation**: There are 57 package.json files across the monorepo that all need version syncing. The inline Node.js script pattern (`node -e "const fs=..."`) is reliable for this.
**Implication**: Every version bump should sync all 57 files. The script excludes node_modules, .git, archive, and submodules directories.

### 13. Multi-Model Pair Programming Pattern (Added 2026-04-08)
**Observation**: For complex tasks, a single model often hallucinates or misses edge cases. A multi-model squad with rotating roles (Planner, Implementer, Tester) provides much higher reliability.
**Resolution**: Implemented `PairOrchestrator` which coordinates Claude, GPT, and Gemini in a shared chat history. It rotates roles every turn to ensure diverse perspective on the implementation.
**Implication**: Use `run_pair_session` for high-complexity structural changes.

### 14. Preemptive Tool Advertisement (Added 2026-04-08)
**Observation**: Models waste tokens and latency searching for tools. If we know they are likely to need a tool based on the topic, we should "advertise" it in the initial turn.
**Resolution**: Implemented `ToolPredictor` which uses a fast LLM turn to predict needed capabilities and preloads them into the working set before the main agent turn.
**Implication**: Preloading makes tools visible in `list_tools` without explicit model search.

### 15. Go Native Tool Execution Pattern (Added 2026-04-08)
**Observation**: Relying on the Node control plane for every tool call creates a single point of failure and higher latency.
**Resolution**: Ported core standard library and parity tools (read, write, bash, edit) to native Go in `go/internal/tools/`. The Go sidecar now implements a "Native First, Bridge Second" strategy in `handleAgentRunTool`.
**Implication**: The Go sidecar can fulfill many critical developer tasks autonomously even if the TypeScript server is restarting or unreachable.

### 16. Package Build Dependency in Monorepo (Added 2026-04-08)
**Observation**: Adding new files and exports to sub-packages (like `@borg/agents`) requires an explicit build of those packages before the main control plane (`@borg/core`) or CLI can see the changes, especially if they depend on built artifacts or have strict type checking.
**Resolution**: Run `pnpm build` in the affected sub-packages before building the consumer.
**Implication**: Automated build scripts should handle package topological sorting or ensure all dependencies are built.

### 17. Multi-Model Swarm Pattern (Added 2026-04-08)
**Observation**: For complex, multi-stage projects, a fixed implementation loop (PairOrchestrator) is sometimes too rigid. A swarm of models with specialized roles (Planner, Implementer, Tester, Critic) and a shared neural transcript provides more flexibility.
**Resolution**: Implemented `SwarmController` which manages a team of models and evaluates progress using a "Critic" model turn. The transcript acts as the shared working memory for the entire swarm.
**Implication**: Use `swarm_start_session` for goals that require multiple iterative cycles of planning and verification.

### 18. A2A Communication Broker (Added 2026-04-08)
**Observation**: Decentralized agents need a structured way to hand off tasks and share state without creating tight coupling between classes.
**Resolution**: Implemented `A2ABroker` (TS and Go) that routes typed messages (TASK_REQUEST, STATE_UPDATE, etc.) based on agent IDs. The broker includes history tracking for dashboard visibility.
**Implication**: All new autonomous agents should register with the `A2ABroker` to enable full orchestration and observability.

### 19. Session Archiving Strategy (Added 2026-04-08)
**Observation**: Long-term retention of raw JSON sessions is noisy and inefficient.
**Resolution**: Implemented `MemoryArchiver` to convert sessions to clean plaintext, extract key facts via LLM, and store them in a compressed ZIP archive.
**Implication**: The system now has a clear boundary between "hot" active session state and "cold" historical archive state, while preserving semantic knowledge.

### 20. A2A WebSocket Integration (Added 2026-04-08)
**Observation**: The `A2ABroker` needs to reach beyond the Node process to dashboard components and remote agents.
**Resolution**: Implemented a WebSocket listener for `A2A_SIGNAL` types. Sending this type to the main WebSocket server will now bridge it directly into the local agent broker.
**Implication**: Remote UI components can now directly trigger and observe agent-to-agent coordination.

### 21. A2A Liveness and Pruning (Added 2026-04-08)
**Observation**: Decentralized agents can crash or become unreachable, leaving stale entries in the broker pool.
**Resolution**: Implemented a heartbeat mechanism in `A2ABroker`. Agents now signal liveness every 15s, and the broker prunes any agent silent for more than 30s.
**Implication**: The dashboard and swarm controller now have a truthful view of currently available coordination capacity.

### 22. Deep Link Insight (Added 2026-04-08)
**Observation**: Basic webpage crawling is insufficient for automated control-plane growth. We need to know if a site *contains* tools we can use.
**Resolution**: Enhanced the Go Link Crawler with an LLM turn that explicitly looks for MCP servers, skills, and APIs.
**Implication**: Discovered links are now semantically enriched with their technical "capabilities," paving the way for auto-installing discovered MCP servers.

### 23. A2A Request-Response Pattern (Added 2026-04-08)
**Observation**: Simple broadcast signals are insufficient for negotiation or state lookups between agents.
**Resolution**: Implemented the `query` pattern in `A2ABroker`. By using `replyTo` matching and local promise maps, agents can now perform asynchronous "RPC-over-A2A" calls with built-in timeouts.
**Implication**: Agents can now ask other agents to perform sub-tasks or report their status without manual correlation logic in every agent class.

### 24. A2A Auditability (Added 2026-04-08)
**Observation**: Decentralized signaling makes it hard to understand why a swarm decision was made.
**Resolution**: Implemented `A2ALogger` in TS and Go. Every message routed through the broker is now appended to a persistent JSONL log. These logs are also bundled into session archives by `MemoryArchiver`.
**Implication**: Operators can now perform post-mortem analysis on agent coordination by inspecting the signal traffic logs.

### 25. Standard Tool Visibility Fix (Added 2026-04-08)
**Observation**: The "Ultra-Streamlined Advertising" filter was too aggressive, hiding basic standard library tools (bash, filesystem) from models. This forced a manual discovery turn that most agents (like pi) weren't prepared for.
**Resolution**: Modified `getDirectModeTools` to treat standard library and tool parity aliases as `alwaysOn` by default.
**Implication**: Borg is now immediately useful as an MCP server for any host agent, as basic coding capabilities are advertised upfront.

### 26. Directory Clutter Reduction (Added 2026-04-08)
**Observation**: The nested hash-based directory structure for session archives was creating thousands of nearly-empty subdirectories, making the `.borg` folder difficult to manage.
**Resolution**: Flattened the archive structure in `ImportedSessionStore` to store all session files in a single `sessions/` directory.
**Implication**: Improved filesystem performance and much cleaner project directory structure.

### 27. Multi-Turn Agent Coordination (Added 2026-04-08)
**Observation**: Simple broadcast signaling is not enough for complex handoffs. Agents need a way to ask questions and wait for specific answers.
**Resolution**: Implemented the `query` pattern in `A2ABroker` (TS and Go). Agents can now perform request-response cycles with built-in correlation and timeouts.
**Implication**: This enables more complex autonomous workflows where a planner can query a coder's status or a critic can ask for clarification on a test failure.

### 28. Native Go Auditing (Added 2026-04-08)
**Observation**: Without a native logger in Go, signal traffic through the sidecar was invisible.
**Resolution**: Ported `A2ALogger` to Go. It now captures every signal routed through the native broker in a JSONL format.
**Implication**: Full observability of agent coordination is now maintained even when the TypeScript control plane is inactive.

### 29. Go Native Skill Management (Added 2026-04-08)
**Observation**: The Go sidecar previously relied on the Node server to list and save skills, creating a dependency for "Total Autonomy".
**Resolution**: Implemented `SkillStore` in Go. It natively reads and writes `.md` runbooks with frontmatter metadata in the `.borg/skills` directory.
**Implication**: The Go sidecar can now independently manage the system's operational knowledge base.

### 30. Monorepo Build Sequencing (Added 2026-04-08)
**Observation**: Changes to foundational packages (like `adk` or `agents`) often fail to propagate to consumers (like `core`) due to missing builds or type definition staleness.
**Resolution**: Ensure a full topological build (ADK -> Agents -> Core -> CLI) when introducing new protocol members or cross-package exports.
**Implication**: Developers should use a workspace-wide build command or script to maintain environment sanity after structural changes.

### 31. Swarm Transcript Observability (Added 2026-04-08)
**Observation**: High-level swarm coordination was previously a "black box" in the dashboard, with no way to see the models' internal dialogue.
**Resolution**: Implemented the `getSwarmTranscript` tRPC endpoint and a corresponding "Neural Transcript" tab in the dashboard.
**Implication**: Operators can now watch model team collaboration in real-time, significantly improving trust and debugging for multi-agent workflows.

### 32. Native Go Configuration Management (Added 2026-04-08)
**Observation**: The Go sidecar's dependency on the Node server for reading `mcp.jsonc` limited its utility as a standalone fallback.
**Resolution**: Implemented `ConfigManager` in Go. It natively handles reading and writing the `mcpServers` object in `mcp.jsonc`.
**Implication**: The Go sidecar can now independently manage the MCP ecosystem, completing another key requirement for "Total Autonomy".

### 33. Specialized Swarm Personas (Added 2026-04-08)
**Observation**: Generic system prompts for swarm participants lead to role confusion and overlapping suggestions.
**Resolution**: Implemented specialized system prompts for `Planner`, `Implementer`, `Tester`, and `Critic` roles in `packages/ai/src/prompts/SystemPrompts.ts`.
**Implication**: Model collaboration is now much more structured, with each participant focusing on their specific domain (architecture, execution, verification, or evaluation).

### 34. Free-Tier Fallback Resilience (Added 2026-04-08)
**Observation**: Frequent quota exhaustion on frontier models makes the system unreliable for background tasks.
**Resolution**: Expanded the fallback chain to include multiple new free-tier options, including `google/gemini-2.0-flash-lite` and `openrouter/free`.
**Implication**: Utility calls now have a nearly guaranteed path to execution even when multiple paid API quotas are hit simultaneously.

### 35. Automated Browser Memory Ingestion (Added 2026-04-08)
**Observation**: Manually clicking "Sync to Memory" in the extension is a friction point that leads to missing context.
**Resolution**: Implemented a `MutationObserver` in the extension's `MemoryCaptureService`. It now watches for new DOM nodes matching AI message patterns and automatically schedules a capture.
**Implication**: Conversations from web chat interfaces now flow into the Borg memory bank in near real-time without user intervention.

### 36. Persistent A2A Audit Logs (Added 2026-04-08)
**Observation**: Live A2A traffic in the dashboard is transient and lost on page refresh, making long-term debugging difficult.
**Resolution**: Created a dedicated dashboard page that reads the `a2a_traffic.jsonl` log file from disk.
**Implication**: Operators can now audit historical agent coordination signals to understand why specific swarm decisions were made.

### 37. Runtime Dependency Hygiene (Added 2026-04-08)
**Observation**: Adding TypeScript dev-dependencies (like @types/adm-zip) is insufficient for runtime features; the actual implementation library (adm-zip) must be in the `dependencies` block of the consumer package.
**Resolution**: Added `adm-zip` to the dependencies of `packages/core`.
**Implication**: Always verify that new services have their core runtime dependencies properly declared in `package.json` to prevent production/startup crashes.

### 38. Cross-Runtime Handshake Parity (Added 2026-04-08)
**Observation**: Agents in Go sidecar were previously unable to participate in task-bidding, creating an orchestration gap between TS and Go agents.
**Resolution**: Implemented the native Go Handshake logic, matching the TS protocol. The `CoderAgent` in Go now correctly responds to `TASK_NEGOTIATION` with a `CAPABILITY_REPORT`.
**Implication**: The swarm controller can now transparently pick the best agent regardless of whether they are implemented in TypeScript or Go.

### 39. Go Namespace Conflict Management (Added 2026-04-08)
**Observation**: Go's package-based shadowing rules can cause built-in packages (like `encoding/json`) to become "undefined" if a local folder or variable with the same name is in scope.
**Resolution**: Added an explicit import alias (`mcp_pkg`) when importing the internal MCP package in files that also use the `json` library.
**Implication**: Prefer descriptive, aliased imports for internal Go packages to avoid collision with standard library names.

### 40. Unified Knowledge Base Persistence (Added 2026-04-08)
**Observation**: The system previously had a "split brain" knowledge base where only the Node server could persist discovered skills.
**Resolution**: Fully integrated the Go native `SkillStore` with the `HighValueIngestor`. Discovered technical skills are now written directly to disk by the sidecar.
**Implication**: Operational runbooks are now truly cross-runtime and persistent across control-plane restarts.

### 41. Bidding Process Visibility (Added 2026-04-08)
**Observation**: Task negotiation between agents was an invisible internal state, making it impossible for operators to know why a specific agent was selected.
**Resolution**: Implemented negotiation tracking in `A2ABroker` and created a dedicated "A2A Handshake Manager" dashboard page.
**Implication**: Every step of the bidding process—from task proposal to individual capability reports—is now visible and auditable in real-time.

### 42. Truthful Sidecar Status (Added 2026-04-08)
**Observation**: The Go sidecar's expert status endpoint returned hardcoded "offline" values, misleading the dashboard even when native agents were ready.
**Resolution**: Updated the Go server to check the actual state of `s.coderAgent` and other native components before reporting status.
**Implication**: The dashboard now reflects the true availability of Go-native agents, improving operator trust in the sidecar's fallback capabilities.

### 43. Automated Tool Integration (Added 2026-04-08)
**Observation**: Discovered tools and skills from Link Crawler were previously stagnant in the backlog, requiring manual operator intervention to become active.
**Resolution**: Updated the `HighValueIngestor` to automatically promote analyzed technical resources. MCP servers are added to the active configuration, and skills are saved to the native store.
**Implication**: The system's capabilities now grow autonomously as the background crawler discovers high-quality technical resources.

### 44. Native Go Context Harvesting (Added 2026-04-08)
**Observation**: The Go sidecar's lack of a context harvester meant its local memory view was always out of sync with recent file changes.
**Resolution**: Implemented the `MemoryReactor` in Go. It performs semantic chunking and incremental summarization of workspace files autonomously.
**Implication**: The Go sidecar now maintains its own live semantic index of the project, further reducing its dependency on the TypeScript control plane.

*Update this file whenever a major systemic pattern, recurring bug, or deep architectural quirk is discovered.*

### 9. tRPC Type Recursion Limit (Discovered 2026-04-29)
**Observation**: The appRouter has 67+ procedures. When `@trpc/react-query` tries to infer `AppRouter` types for the React client, TypeScript hits its internal type recursion depth limit and collapses to error strings like `"The property 'useContext' in your router collides with a built-in method"`.
**Resolution**: Use `createTRPCReact<any>()` instead of `createTRPCReact<AppRouter>()`. Type safety comes from the server-side router definition. A dynamic `require('@trpc/react-query')` is needed because pnpm strict isolation means `@trpc/react-query` isn't accessible from `packages/ui/node_modules`.

### 10. Turbo v2 Requires `extends` in All Package turbo.json (Discovered 2026-04-29)
**Observation**: Turbo v2.8.x requires `"extends": ["//"]` in every non-root `turbo.json`. Without it, the build fails with "add extends key here".
**Resolution**: Add `"extends": ["//"]` to all package-level turbo.json files.

### 11. Duplicate VS Code Extension Package (Discovered 2026-04-29)
**Observation**: `apps/vscode` and `packages/vscode` had the same package name (`borg-vscode-extension`), causing Turbo workspace collision.
**Resolution**: Removed `packages/vscode` (stale, alpha.34). `apps/vscode` is canonical (alpha.36+).

### 12. CLI Version Read Path (Discovered 2026-04-29)
**Observation**: The CLI `getVersion()` tried to read `VERSION.md` from `../../../..VERSION.md` relative to `__dirname` in compiled output. The compiled output is at `packages/cli/dist/cli/src/index.js`, so the path resolved to `packages/cli/VERSION` (wrong).
**Resolution**: Changed to `../../../../../VERSION` (5 levels up) and filename to `VERSION` (not `VERSION.md`).

### 13. @borg/* Stub Package Architecture (Discovered 2026-04-29)
**Observation**: The core package imports from 8 `@borg/*` packages that don't have real implementations. TypeScript compilation and Node runtime both fail without them.
**Resolution**: Created stub packages in `packages/` with three layers:
  - `src/index.ts` — TS type stubs (for `tsc --noEmit` via `exports.types` condition)
  - `dist/index.js` — ESM runtime stubs (for Node via `exports.import` condition)
  - `dist/index.d.ts` — Type declarations (fallback resolution)
  Key insight: `@borg/types` must export real zod schemas (tRPC `.input()` calls `'~standard' in schema`). `@borg/ai` must export real classes for `extends` (NormalizedQuotaService extends QuotaService). All others can be `undefined` stubs.
  The `exports` field in package.json must have both `import` and `types` conditions pointing to the correct files. Without `exports`, tsc resolves incorrectly (230 errors).
**Implication**: As real implementations are built, each stub can be replaced incrementally. The zod schemas in `@borg/types` are already functional and used by tRPC routers.

### 14. Dashboard Utility Stubs Must Match Usage Patterns (Discovered 2026-04-29)
**Observation**: Auto-generated stubs returning `(() => ({})) as any` crash Next.js build when pages call `.map()` on the result or access properties like `.topTools`.
**Resolution**: Utility functions must return the correct base type: arrays for anything used with `.map()`, objects for property access, strings for display, numbers for timestamps. The heuristic: function names containing "rows/records/sections/items/list/groups" → `[]`, "label/title/hint/message" → `''`, everything else → `{}`.
**Implication**: When replacing stubs with real implementations, the return type contracts are already implicitly defined by page usage.

### 15. Next.js useSearchParams Requires Suspense Boundary (Discovered 2026-04-29)
**Observation**: Pages using `useSearchParams()` from `next/navigation` fail during static prerendering with "useSearchParams() should be wrapped in a suspense boundary".
**Resolution**: Wrap the default export in a `<Suspense>` boundary. Pattern: `export default function PageWrapper() { return <Suspense fallback={...}><Page /></Suspense>; }` then rename the original to `function Page()`.

### 16. tRPC Service Dependency Chain (Discovered 2026-04-29)
**Observation**: 8 tRPC routers crash with `Cannot read properties of undefined` when MCP services aren't initialized. Root cause: `getMcpServer()` returns `{}` when server isn't ready, then getters like `getSuggestionService()` return `undefined`, and calling methods on `undefined` crashes.
**Resolution**: Wrap service calls in try/catch with safe defaults: `try { return getService()?.method() ?? []; } catch { return []; }`
**Implication**: When running with `--no-mcp`, all routers should return empty/safe defaults instead of 500 errors.

### 17. Server Startup Takes ~45 Seconds (Discovered 2026-04-29)
**Observation**: `borg start --no-mcp` takes ~45 seconds to fully bind port 4000. The MCPServer imports 53+ phases before Express starts listening.
**Implication**: CLI health checks and tests must wait at least 50 seconds before testing endpoints. The `borg status` command uses `AbortSignal.timeout(3000)` which may not be enough during startup.

### 18. Commander.js Global --json Doesn't Merge to Subcommands (Discovered 2026-04-29)
**Observation**: When `--json` is registered as a global option on the program AND as a subcommand option, the subcommand `opts` object is empty `{}`. The global option consumes the flag.
**Resolution**: Use `cmd.optsWithGlobals()` in the action handler: `async (opts, cmd) => { const allOpts = cmd ? cmd.optsWithGlobals() : opts; const isJson = allOpts.json === true; }`

### 19. Turbopack Cache Corruption (Discovered 2026-04-29)
**Observation**: Next.js 16 Turbopack crashes with `range start index 1607525553 out of range for slice of length 468465` in static_sorted_file.rs. The `.next/` cache gets corrupted.
**Resolution**: Delete `apps/web/.next/` and rebuild. Set `turbopack.root` in `next.config.js` to the monorepo root.
**Implication**: If the dashboard fails to start, first try clearing `.next/`.

### 20. Dashboard tRPC Proxy Port Mismatch (Discovered 2026-04-30)
**Observation**: The Next.js dashboard's tRPC API route (`apps/web/src/app/api/trpc/[trpc]/route.ts`) defaulted to `http://127.0.0.1:3001/trpc` (MCP WebSocket port) instead of `http://127.0.0.1:4000/trpc` (tRPC server). All dashboard data queries returned 502 errors.
**Resolution**: Changed `DEFAULT_UPSTREAM_TRPC_URL` from port 3001 to 4000. Can also be overridden via `BORG_TRPC_UPSTREAM` env var.
**Implication**: This was the root cause of empty dashboard data. The proxy is the only way the dashboard reaches the TS server.

### 21. Lightweight MCP Init Without Full Discovery (Discovered 2026-04-30)
**Observation**: When `--no-mcp` is used, `global.mcpServerInstance` is never set, causing ALL tRPC routers to return empty objects from `getMcpServer()`. Session supervisor, memory manager, and other services are unavailable.
**Resolution**: Added a lightweight init path that creates MCPServer with `skipAutoDrive: true, skipStdio: true` and sets `global.mcpServerInstance`. This gives routers their dependencies without the 45-second discovery phase.
**Implication**: Next server restart with `--no-mcp` will have functional tRPC routers (sessions, memory, agents) even without MCP tool discovery.

### 22. MCP Inventory Known When Persisted Data Exists (Discovered 2026-04-30)
**Observation**: `startupStatus` showed `inventoryReady: false` even though the database had 135 persisted servers and 1302 tools. The `inventoryKnown` check required `aggregatorStatus.initialized` which is only set during full MCP startup.
**Resolution**: Added `|| persistedServerCount > 0` to the `inventoryKnown` condition in `startupStatus.ts`.
**Implication**: `borg health` will now correctly report inventory as ready when data exists in the database.

### 23. CLI Provider Auto-Detection (Discovered 2026-04-30)
**Observation**: 8 API keys were available in environment variables (OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY, GEMINI_API_KEY, XAI_API_KEY, DEEPSEEK_API_KEY, MISTRAL_API_KEY, OPENROUTER_API_KEY) but `borg provider list` showed "No providers configured".
**Resolution**: Added environment variable scanning as fallback when no providers are explicitly configured. Shows provider name, "● Available" status, and the env var source.
**Implication**: Users see their available providers immediately without explicit configuration.
