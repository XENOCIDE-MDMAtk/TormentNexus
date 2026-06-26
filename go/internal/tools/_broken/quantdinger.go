package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func ok(text string) ToolResponse {
	return ToolResponse{TextContent: text, Ok: true, Error: nil}
}

func err(e error) ToolResponse {
	return ToolResponse{TextContent: "", Ok: false, Error: e}
}

func getString(args map[string]interface{}, key string) string {
	if value, found := args[key]; found {
		if str, found := value.(string); found {
			return str
		}
	}
	return ""
}

func getInt(args map[string]interface{}, key string) int {
	if value, found := args[key]; found {
		if intVal, found := value.(int); found {
			return intVal
		}
	}
	return 0
}

func getBool(args map[string]interface{}, key string) bool {
	if value, found := args[key]; found {
		if boolVal, found := value.(bool); found {
			return boolVal
		}
	}
	return false
}

func HandleGetAgentWorkflowInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	detailLevel, _ :=getString(args, "detail_level")
	if detailLevel == "" {
		detailLevel = "summary"
	}
	var info string
	if strings.EqualFold(detailLevel, "full") {
		info = `QuantDinger Agent Workflow (Full)
===============================
Applies to: Changes to backend_api_python/, strategies, docker-compose.yml, scripts/, env.example, docs/agent/
Key docs to read first:
1. docs/agent/AGENT_ENVIRONMENT_DESIGN.md - SSOT for 3 layers: documentation contract -> command contract -> optional HTTP/MCP
2. docs/agent/AI_INTEGRATION_DESIGN.md - How external AI agents consume QuantDinger (Agent Gateway, scopes, MCP, trading safety)
3. docs/agent/AGENT_QUICKSTART.md - Operator/integrator walkthrough
4. docs/agent/agent-openapi.json - Machine-readable contract for /api/agent/v1
5. docs/agent/README.md - Index of agent-facing docs

Implemented Surface:
- Agent Gateway mounted at /api/agent/v1
- Auth: app/utils/agent_auth.py, tokens hashed at rest in qd_agent_tokens
- Async jobs: app/utils/agent_jobs.py, writes to qd_agent_jobs, SSE stream at GET /jobs/{id}/stream
- Audit: all calls (success/denial) appended to qd_agent_audit
- Trading: quick_trade.py enforces paper-only by default, live requires paper_only=false on token AND AGENT_LIVE_TRADING_ENABLED=true
- MCP: mcp_server/ thin wrapper over R+W+B endpoints, transports: stdio (default), sse, streamable-http
- Admin UI: Profile -> My Agent Token for all users, admins have /agent-tokens for audit`
	} else {
		info = `QuantDinger Agent Workflow (Summary)
==================================
Use this workflow when editing:
- backend_api_python/ (Flask API, services, routes)
- Strategy/backtest/trading logic
- docker-compose.yml, scripts/, env.example
- docs/agent/ (English only)

Key rules:
1. Read AGENT_ENVIRONMENT_DESIGN.md and AI_INTEGRATION_DESIGN.md before adding endpoints/tools
2. Update agent-openapi.json when changing /api/agent/v1 routes
3. Never commit secrets, production .env, or API keys
4. Do not add live trading automation that bypasses human review unless explicitly requested
5. All new agent-facing prose must be English`
	}
	return ok(info)
}

func HandleGetAgentEnvDesign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	layer, _ :=getString(args, "layer")
	if layer == "" {
		layer = "all"
	}
	var designInfo string
	switch strings.ToLower(layer) {
	case "documentation":
		designInfo = `Agent Environment Design - Documentation Layer
================================================
This is the topmost contract layer. All agent-facing documentation lives in docs/agent/ and is the source of truth for:
- API surface definitions (agent-openapi.json)
- Workflow rules (this SKILL.md)
- Integration guides (AI_INTEGRATION_DESIGN.md)
- Quickstart for operators (AGENT_QUICKSTART.md)
All changes to the agent surface must be reflected here first before implementation.`
	case "command":
		designInfo = `Agent Environment Design - Command Layer
=========================================
The command layer is the CLI/script interface for agents and operators, located in scripts/ and backend_api_python/app/services/.
Key rules:
- All commands must be idempotent where possible
- Commands that modify state require explicit confirmation flags
- No command may bypass the auth/audit layers defined in the documentation contract
- All command outputs must be parseable by AI agents (structured text, JSON where appropriate)`
	case "http":
		designInfo = `Agent Environment Design - HTTP/MCP Layer
=============================================
The HTTP layer is the /api/agent/v1 REST surface, with optional MCP wrapper.
Key rules:
- All endpoints require agent auth with appropriate scopes (app/utils/agent_auth.py)
- Async long-running operations use the job system with SSE streaming
- All requests and responses are audited to qd_agent_audit
- MCP tools are thin wrappers over existing REST endpoints, no new logic in MCP layer
- Trading endpoints enforce paper-only by default, live execution requires explicit token scope and env var`
	default:
		designInfo = `Agent Environment Design - All Layers
======================================
1. Documentation Layer (docs/agent/): Source of truth for all agent-facing contracts, workflows, and guides. Must be updated before any implementation changes.
2. Command Layer (scripts/, app/services/): CLI/script interface for agents, idempotent, audited, no auth bypass.
3. HTTP/MCP Layer (/api/agent/v1, mcp_server/): REST API with auth, async jobs, audit, and optional MCP wrapper. Trading endpoints enforce strict safety boundaries.

All three layers must stay in sync. Changes to one layer require corresponding updates to the others as defined in the documentation contract.`
	}
	return ok(designInfo)
}

func HandleGetAgentQuickstart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	step, _ :=getInt(args, "step")
	steps := []string{
		"1. Clone the repository and navigate to the project root",
		"2. Copy backend_api_python/env.example to backend_api_python/.env and set SECRET_KEY to a random 64-character hex value",
		"3. Launch the full stack with: docker compose up -d --build",
		"4. Access the UI at http://localhost:8888, backend API at http://localhost:5000",
		"5. Generate an agent token via Profile -> My Agent Token in the UI, or POST to /api/agent/v1/me/tokens",
		"6. Use the token with required scopes to call agent endpoints, or configure the MCP server (mcp_server/) with the token for IDE integration",
		"7. For local backend development without Docker: cd backend_api_python, create venv, pip install -r requirements.txt, python run.py",
	}
	if step > 0 && step <= len(steps) {
		return ok(steps[step-1])
}

	return ok(strings.Join(steps, "\n"))
}

func HandleCheckSafetyBoundaries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		category = "all"
	}
	var boundaries string
	switch strings.ToLower(category) {
	case "trading":
		boundaries = `Trading Safety Boundaries
========================
1. Paper-only by default: All trading operations run in paper mode unless explicitly overridden
2. Live trading requirements:
   - Agent token must have paper_only=false scope
   - Environment variable AGENT_LIVE_TRADING_ENABLED=true must be set
   - Both conditions must be met simultaneously; no exceptions
3. No automated order placement that bypasses human review unless explicitly requested and scoped
4. All trades are logged to qd_agent_audit for compliance`
	case "red_lines":
		boundaries = `Red Lines (Never Violate)
========================
1. Never commit real secrets, production .env files, API keys, or DB passwords to version control
2. Never add live trading or order placement automation that bypasses human review without explicit, scoped request
3. Never log or persist raw agent tokens (only hashed values are stored in qd_agent_tokens)
4. Never duplicate long strategy guide text in agent docs (link to official guides instead)
5. All new agent-facing content must be English`
	default:
		boundaries = `QuantDinger Safety Boundaries
=============================
Trading Safety:
- Paper-only by default for all trading operations
- Live trading requires both token scope (paper_only=false) AND env var AGENT_LIVE_TRADING_ENABLED=true
- No automated order placement bypassing human review unless explicitly requested

Red Lines (Non-Negotiable):
1. No secrets, production .env, API keys, or DB passwords in version control
2. No live trading automation without explicit, scoped approval
3. No raw token logging/persistence (only hashed tokens stored)
4. No duplicated strategy guide content in agent docs (link to official guides instead)
5. All new agent-facing content must be English`
	}
	return ok(boundaries)
}