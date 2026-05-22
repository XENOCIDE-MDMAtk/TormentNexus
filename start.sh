#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

VER=$(cat VERSION 2>/dev/null || echo "dev")

echo "╔════════════════════════════════════════════════╗"
echo "║         Hypercode HYPERCODE v${VER}                   ║"
echo "║         The Neural Operating System            ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# ── 1. Build Go Sidecar ──────────────────────────
if command -v go &>/dev/null; then
	echo "[1/5] Building Go sidecar..."
	(cd go && go build -ldflags "-s -w -X github.com/hypercodehq/hypercode-go/internal/buildinfo.Version=${VER}" -buildvcs=false -o ../bin/hypercode ./cmd/hypercode 2>/dev/null) && echo "      ✓ bin/hypercode built" || echo "      [WARN] Go build failed"
else
	echo "[1/5] Go not found — skipping sidecar build."
fi

# ── 2. Install Dependencies ──────────────────────
echo "[2/5] Installing dependencies..."
HYPERCODE_SKIP_INSTALL="${HYPERCODE_SKIP_INSTALL:-0}"
if [ "$HYPERCODE_SKIP_INSTALL" = "1" ]; then
	echo "      Skipping (HYPERCODE_SKIP_INSTALL=1)"
else
	pnpm install --frozen-lockfile 2>/dev/null || pnpm install
fi

# ── 3. Rebuild native modules ────────────────────
echo "[3/5] Rebuilding native modules..."
HYPERCODE_SKIP_NATIVE="${HYPERCODE_SKIP_NATIVE:-0}"
if [ "$HYPERCODE_SKIP_NATIVE" != "1" ]; then
	pnpm rebuild better-sqlite3 2>/dev/null || true
fi

# ── 4. Build TypeScript ───────────────────────────
echo "[4/5] Building TypeScript..."
HYPERCODE_SKIP_BUILD="${HYPERCODE_SKIP_BUILD:-0}"
if [ "$HYPERCODE_SKIP_BUILD" != "1" ]; then
	pnpm run build:workspace 2>/dev/null || {
		pnpm -C packages/core exec tsc
		pnpm -C packages/cli exec tsc
	} || true
fi

# ── 5. Launch Services ────────────────────────────
echo "[5/5] Starting services..."

HYPERCODE_PORT="${HYPERCODE_PORT:-4000}"

# Start Go sidecar in background
if [ -x bin/hypercode ]; then
	echo "      Starting Go sidecar on port 4300..."
	bin/hypercode -port 4300 &>/dev/null &
	GO_PID=$!
fi

echo "      Starting TS control plane on port ${HYPERCODE_PORT}..."
echo ""
echo "  ✓ tRPC:     http://0.0.0.0:${HYPERCODE_PORT}/trpc"
echo "  ✓ REST:     http://0.0.0.0:${HYPERCODE_PORT}/api"
echo "  ✓ Go:       http://127.0.0.1:4300/api/index"
echo "  ✓ Dashboard: http://localhost:3000/dashboard"
echo ""
echo "  Press Ctrl+C to stop all services."
echo ""

cleanup() {
	echo ""
	echo "Shutting down..."
	[ -n "${GO_PID:-}" ] && kill "$GO_PID" 2>/dev/null
	exit 0
}
trap cleanup SIGINT SIGTERM

node packages/cli/dist/cli/src/index.js start --port "$HYPERCODE_PORT" "$@"
