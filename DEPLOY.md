<<<<<<< HEAD
# Borg Deployment Instructions

_This document contains the latest deployment instructions for the Borg Universal AI Dashboard and Cognitive Control Plane._

## Prerequisites

1.  **Node.js**: >= 22.12.0
2.  **pnpm**: Recommended package manager (`npm install -g pnpm`)
3.  **Git**: For submodule fetching and version control.

## Initial Setup

1.  **Clone the Repository**:
    ```bash
    git clone https://github.com/robertpelloni/borg.git
    cd borg
    ```

2.  **Initialize Submodules**:
    ```bash
    git submodule update --init --recursive
    ```

3.  **Install Dependencies**:
    ```bash
    pnpm install
    ```

4.  **Environment Variables**:
    Copy `.env.example` to `.env` and fill in the required API keys (OpenAI, Anthropic, Gemini, etc.).
    ```bash
    cp .env.example .env
    ```

## Running the Platform

Borg is designed as a long-running service that manages PC memory, CPU, disk, and bandwidth usage.

### Start the Borg Hub
Use the provided startup scripts:

**Windows**:
```bash
.\start.bat
```

`start.bat` defaults to `pnpm run build:workspace` (faster startup, skips extension-only build stages).
It starts Borg only; Maestro is now launched separately.
It also runs a native-runtime preflight that repairs missing `better-sqlite3` and Electron runtime artifacts when possible.
To force a full build before startup, set:

```bash
set BORG_FULL_BUILD=1
.\start.bat
```

For repeat local runs where dependencies are already installed, you can skip the install step:

```bash
set BORG_SKIP_INSTALL=1
.\start.bat
```

For the fastest local restart loop (when you know previous build artifacts are still valid), you can also skip the build step:

```bash
set BORG_SKIP_BUILD=1
.\start.bat
```

If you need to bypass the native preflight explicitly, set:

```bash
set BORG_SKIP_NATIVE_PREFLIGHT=1
.\start.bat
```

**Linux/macOS**:
```bash
./start.sh
```

### Start Maestro Separately

When you want the Electron app, launch it explicitly:

```bash
pnpm -C apps/maestro start
```

Alternatively, run via pnpm:
```bash
pnpm dev
```

This will:
1. Start the Borg Server (core).
2. Start the MCP Router (client/server proxy).
3. Open the Web Dashboard (Next.js) at `http://localhost:3000`.

### Building for Production
```bash
pnpm build
pnpm start
```

## Extension Installation

Once the dashboard is running, navigate to the **Integrations** tab in the WebUI to install:
*   Browser Extensions (Chrome, Firefox).
*   IDE Plugins (VSCode, Cursor, Windsurf).
*   CLI Harnesses, with `borg` now tracked as Borg's primary CLI harness lane via `submodules/borg` and surfaced as a Go/Cobra runtime with REPL + `pipe` metadata (still **Experimental**).

## Package Manager Requirement

**pnpm v10 is required.** The root `package.json` locks `packageManager: pnpm@10.28.0`. Using pnpm v9 or below will produce `ERR_PNPM_BAD_PM_VERSION` and fail the build.

```bash
npm install -g pnpm@10
```

## CI/CD

Workflows in `.github/workflows/` use `pnpm/action-setup@v4` with `version: 10`. Do not downgrade this — it will invalidate every CI run against the packageManager lock.

## Release Gate

Before merging or pushing, validate the full release gate:

```bash
pnpm run check:release-gate:ci
```

This runs (in order):
1. `check:placeholders` — ensures no unresolved placeholder files are committed
2. Core typecheck — `tsc --noEmit` across `packages/core`
3. Turbo lint — ESLint across all packages

Screenshot/visual verification is now opt-in (manual workflow). Use this when you explicitly want to validate screenshot state:

```bash
pnpm run check:release-gate:ci:strict-visuals
```

If strict visuals fail, run `pnpm run sync:screenshot-status` to resync the README table.

## MCP Configuration

MCP server definitions are stored in `~/.borg/mcp.json` (or `mcp.jsonc` for commented JSON). On first run, Borg migrates any legacy config from the workspace root to `~/.borg/`.

Example `~/.borg/mcp.json`:
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "C:/workspace"]
    }
  }
}
```

### Always On Tools

Servers and individual tools can be marked **Always On** via the dashboard:
- **Server-level**: Toggle on the MCP dashboard (`/dashboard/mcp`) to auto-load all tools from a server.
- **Tool-level**: Toggle individual tools from the Inspector (`/dashboard/mcp/inspector`).

Always On tools are permanently advertised to clients and available for the semantic `auto_call_tool` meta-tool, which uses LLM-based matching to automatically select and execute the best tool for a given query.

## Ports

| Service | Default Port | Override |
|---------|-------------|----------|
| Core API (Borg CLI control plane) | 4000 | `borg start --port <number>` |
| Web Dashboard (Next.js) | 3000 | `PORT` |
| Orchestrator | 3847 | `BORG_ORCHESTRATOR_PORT` |
| MCP Proxy (stdio) | stdin/stdout | N/A |

=======
# Deployment Instructions

## borg `1.0.0-alpha.1`

### 1. Build Requirements
- Node.js >= 24
- pnpm >= 10.28
- Go >= 1.23 (For experimental sidecar)

### 2. Standard Build & Run
```bash
pnpm install
pnpm run build
pnpm run start
```
This will compile the TypeScript monorepo, build the web assets, and launch the primary Node.js `cli-orchestrator` along with the web dashboard on port `3000`.

### 3. Experimental Go Sidecar
To run the Go bridge alongside the main control plane:
```bash
cd go
go run ./cmd/borg serve
```

### 4. Extensions
To build extensions (VS Code, Chrome):
```bash
pnpm run build:extensions
```

### 5. Production Docker
```bash
docker build -f Dockerfile.prod -t borg:latest .
docker run -p 3000:3000 -p 4000:4000 -v borg-data:/root/.borg borg:latest
```

## Health Checks
- `http://localhost:4000/api/config/status` - Main control plane health
- `http://localhost:3000` - Dashboard
>>>>>>> main
