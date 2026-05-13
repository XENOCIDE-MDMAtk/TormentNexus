# AGENTS — borg Contributor & Agent Guide

> **CRITICAL: ALL AGENTS MUST READ `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` BEFORE PROCEEDING.**

This file serves as a reference point for multi-agent workflows (Claude → Gemini → GPT) and human operators orchestrating autonomous sessions.

Outstanding! Magnificent! Insanely Great!!! Please continue to proceed as you recommend based on your ongoing analysis and assessments! Keep going! Keep on goin'! Don't stop! Don't ever stop baby! Don't ever quit! Don't stop the party!!!

## 1. Multi-Agent Workflows

### Handoff Protocol
- Agents communicate primarily through `HANDOFF.md`.
- When your turn finishes, document **exactly** what you did, what failed, and what the next agent must do.
- Include: files changed, tests run, build status, version bumped, any remaining issues.
- Update `MEMORY.md` with any new systemic observations or recurring bugs discovered.

### Model Specializations
| Model | Strengths |
|---|---|
| **Gemini** | Speed, recursive scripts, massive context processing, repo maintenance, bulk refactoring |
| **Claude** | Deep implementation, UI/UX perfection, documentation, styling, type safety |
| **GPT** | Architecture, systemic debugging, strict type enforcement, Go porting |

### Iteration Cycle
**Read → Strategize → Execute → Validate → Commit → Handoff. Never stop the party.**

   
Please summarize anything you have learned during this session that was not obvious at the start, or anything else you would like to inform yourself upon newer sessions. 

## 2. Session Protocol

### Session Start
1. Read `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`
2. Read `VERSION` file — verify it matches `package.json` and dashboard display
3. Read `HANDOFF.md` — pick up where previous agent left off
4. Read `MEMORY.md` — learn from accumulated observations
5. Run `git fetch --all && git status` — verify clean state on `main`
6. Understand repo structure before making changes

### During Execution
- Work autonomously unless action is destructive or genuinely ambiguous
- Prefer small, verifiable changes over broad rewrites
- Use parallel tool calls when safe
- Keep status labels and documentation honest
- After any `pnpm install`, run `pnpm rebuild better-sqlite3` on Node 24

### Session End
1. Update `HANDOFF.md` with complete session summary
2. Update `MEMORY.md` with new observations
3. Bump `VERSION` file and sync all `package.json` files
4. Update `CHANGELOG.md` with what changed
5. Commit with version number in message: `feat: description (v1.0.0-alpha.X)`
6. Push to both remotes: `origin` and `borg-upstream`
7. Update `TODO.md` and `ROADMAP.md` if priorities changed

Please include https://github.com/robertpelloni/jules-autopilot as a dashboard for integrating google jules cloud dev env, and also https://github.com/robertpelloni/opencode-autopilot, and please also make a dashboard with links to all convenient major AI tools, like a section for the billing for API key for gemini/vertex/aistudio/cloud, azure, openrouter, chatgpt/openai api key, claude api key, and then also links to the pro plan membership subscription management page, and then use the spreadsheet i linked with all the ai tools on another dashboard page to detect all installed tools and link to their homepage, and then detect whether they are connected using account oauth or api key, and show usage for each provider both account and api key and show budgets, copilot plus subscription, and then also make a dashboard page similar to the jules-app page for copilot cloud, claude chrome, claude cloud, openai codex cloud, and other providers like blocks, basically useful dashboard pages that organize everything into one place both locally and remote.

## 3. Version Management

### Single Source of Truth
- **`VERSION`** file contains the one true version string (e.g., `1.0.0-alpha.7`)
- All `package.json` files must be synced to this value
- Dashboard reads version from `VERSION` file at runtime
- Every meaningful commit should bump the version

### Version Bump Protocol
```
1. Update VERSION file
2. Run: node -e "const v=require('fs').readFileSync('VERSION','utf8').trim(); ['package.json',...require('fs').readdirSync('packages').map(p=>'packages/'+p+'/package.json')].forEach(f=>{try{const j=JSON.parse(require('fs').readFileSync(f,'utf8'));if(j.version&&j.version.startsWith('1.0.0')){j.version=v;require('fs').writeFileSync(f,JSON.stringify(j,null,4)+'\n')}}catch{}})"
3. Update CHANGELOG.md with version entry
4. git add -A && git commit -m "feat: description, bump VERSION to X.Y.Z"
5. git push origin main && git push borg-upstream main --force-with-lease
```

### Remotes
- `origin` → `https://github.com/robertpelloni/borg.git`
- `borg-upstream` → `https://github.com/robertpelloni/borg.git`

---

## 4. Build & Validation

### Required Checks
```bash
pnpm rebuild better-sqlite3   # Node 24 requires this after install
pnpm -C packages/core exec tsc --noEmit
pnpm -C packages/cli exec tsc --noEmit
cd go && go build -buildvcs=false ./cmd/borg
```

### Startup Verification
```bash
node scripts/build_startup.mjs --profile=go-primary
# Then: .\start.bat (Windows) or ./start.sh (Linux)
```

### Runtime Ports

| Service | Port | URL |
|---|---|---|
| tRPC/REST API | 4100 | `http://0.0.0.0:4100/trpc` |
| MCP WebSocket | 3001 | `ws://localhost:3001` |
| Next.js Dashboard | 3000 | `http://localhost:3000/dashboard` |
| Go Sidecar | 4300 | `http://127.0.0.1:4300/api/index` |

> **Service Discovery**: All ports are configurable via environment variables (`BORG_GO_PORT`, `BORG_TRPC_UPSTREAM`, `BORG_BRIDGE_PORT`, `BORG_DASHBOARD_PORT`). See `go/internal/config/discovery.go`.
>
> **Dashboard Proxy**: The Next.js app proxies `/api/go/*` to the Go sidecar at port 4300. See `apps/web/src/app/api/go/[...path]/route.ts`.

---

## 5. Project Structure

```
borg/
├── VERSION                  # Single source of truth for version
├── packages/core/           # Main TypeScript control plane (593 .ts)
│   ├── src/MCPServer.ts     # Core MCP server (5000+ lines)
│   ├── src/orchestrator.ts  # Express/tRPC server + REST bridges
│   ├── src/providers/       # Provider registry, model selector
│   ├── src/routers/         # tRPC routers (savedScripts, etc.)
│   ├── src/daemons/         # HyperIngest workers (BobbyBookmarks, etc.)
│   └── src/config/          # council.json, llm config
├── packages/cli/            # CLI entrypoint (28 .ts)
├── packages/ai/             # AI SDK abstractions
├── packages/tools/          # Built-in tool definitions
├── packages/browser-extension/ # Chrome/Firefox extension (planned)
├── apps/web/                # Next.js 16 dashboard (311 .ts/.tsx, 91 pages)
├── apps/maestro/            # Electron/Wails visual orchestrator (submodule)
├── apps/cloud-orchestrator/ # Jules autopilot wrapper (submodule)
├── go/                      # Go-native server bridge (140+ .go)
│   ├── cmd/borg/       # Go binary entrypoint
│   ├── internal/httpapi/    # HTTP handlers (~45 route families)
│   ├── internal/mcp/        # MCP inventory, ranking, catalog, decision system
│   ├── internal/hsync/      # BobbyBookmarks, LinkCrawler, suggestions
│   └── internal/orchestration/ # Council, swarm, squad, pair orchestrator
├── submodules/
│   ├── hyperharness/        # LLM harness submodule
│   └── prism-mcp/           # Prism MCP reference
├── packages/claude-mem/     # Claude memory bridge (submodule)
├── mcp.jsonc                # MCP server definitions (34K+ lines)
├── docs/                    # All documentation
├── scripts/                 # Build and utility scripts
└── data/                    # Runtime data storage
```

### Submodules
| Path | Remote | Purpose |
|---|---|---|
| `apps/maestro` | github.com/robertpelloni/Maestro | Visual orchestrator (Electron/Wails) |
| `apps/cloud-orchestrator` | github.com/robertpelloni/jules-autopilot | Jules autopilot wrapper |
| `packages/claude-mem` | github.com/robertpelloni/claude-mem | Claude memory bridge |
| `submodules/hyperharness` | github.com/robertpelloni/hyperharness | LLM harness |
| `submodules/prism-mcp` | github.com/dcostenco/prism-mcp | Prism MCP reference |

---

## 6. Key Design Decisions

### "Borg → Borg" Rename
The project was originally called "Borg" and is being renamed to "Borg". Some internal references may still say "borg" (e.g., `borg-upstream` remote, `borg.config.json`). These are being migrated incrementally.

### Go Bridge Strategy
The Go server (`go/`) is a **bridge/fallback** — it provides native handlers for routes when the TS server is unavailable, but the TypeScript server is still the primary runtime. Do NOT split into separate binaries yet (per `UNIVERSAL_LLM_INSTRUCTIONS.md` modular-monolith-first rule).

### MCP Decision System
Borg's MCP layer is not "just an aggregator." It is a **decision system** with:
- Tiny permanent meta-tool surface (5-6 tools always visible)
- Ranked discovery, not raw search
- Silent high-confidence auto-load
- Deferred binary startup
- Small active working set with LRU eviction
- Profiles for common workflows
- Strong observability in dashboard inspector

### Memory Architecture
- SQLite (better-sqlite3) is the primary local store
- `.borg/mcp-cache.json` is the unified cache for the stdio loader
- `mcp.jsonc` is the manual config (never auto-deleted)
- DB tools and JSON tools are merged, never overwrite each other

---

## 7. Git Workflow

### Branch Strategy
- `main` is the only active branch
- No feature branches for this repo — commit directly to `main`
- For forked submodules under `robertpelloni`, merge any local feature branches into main

### Commit Message Format
```
type: description, bump VERSION to X.Y.Z

Types: feat, fix, docs, chore, refactor, test, perf
```

### Submodule Updates
```bash
git submodule update --remote --merge
# Then commit the updated submodule pointer
```

---

*For model-specific quirks, refer to `CLAUDE.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md`.*
>>>>>>> main
