# Handoff - v1.0.0-alpha.115

## Summary
Phase 113: Implemented predictive conversational tool injection. The system now auto-injects contextually relevant tools into the LLM's context window on every ListTools call, based on the direction of the ongoing conversation as judged by a fast local model (Gemma 12b via Ollama) with cloud fallback.

## Accomplishments

### Phase 113 — Predictive Conversational Tool Injection

**Core service**: `packages/core/src/mcp/ConversationalToolInjector.ts`
- Maintains a sliding window of the last 8 conversation turns
- Throttled predictions (3s minimum between LLM calls)
- **Prediction chain**: Go sidecar (`/api/mcp/tools/predict-conversational`) → Ollama Gemma 12b → cheapest cloud model via `LLMService`
- Structured JSON array output with robust extraction and catalog validation
- Only injects non-alwaysOn, non-already-loaded tools from the catalog

**NativeSessionMetaTools extensions**:
- `injectConversationalTools(names)` — loads predicted tools without displacing pinned always-on tools
- `getCatalogSnapshot()` — compact catalog view passed to LLM prediction prompts

**MCPServer wiring**:
- `ConversationalToolInjector` instantiated in constructor (shares `llmService` + `modelSelector`)
- `appendConversationTurn(role, text)` — public API
- `getConversationInjector()` — typed public getter for dashboard queries
- `getDirectModeTools()` runs prediction + injection on every ListTools request (before building the visible tool list)
- `CallToolRequestSchema` auto-captures tool name + string args as "tool" turns
- `CONVERSATION_TURN` WebSocket message type for explicit dashboard pushes
- `BROWSER_CHAT_SURFACE` feeds user input text to the window

**tRPC endpoints** in `mcpRouter.ts`:
- `mcp.appendConversationTurn` (mutation) — dashboard / Go kernel can push turns directly
- `mcp.getConversationWindow` (query) — returns window snapshot + token count for debug panel

### Environment Variables
- `TORMENTNEXUS_SIDECAR_URL` (default: `http://127.0.0.1:4300`)
- `TORMENTNEXUS_OLLAMA_URL` (default: `http://127.0.0.1:11434`)
- `TORMENTNEXUS_LOCAL_PREDICT_MODEL` (default: `gemma3:12b`)

### Verification
- TypeScript compile: ✅ 0 errors (verified 3×)
- Always-on tools: ✅ pinned, immune to eviction
- All injection failures: ✅ non-fatal, caught with console.warn

## Files Modified
- `packages/core/src/mcp/ConversationalToolInjector.ts` **[NEW]**
- `packages/core/src/mcp/NativeSessionMetaTools.ts`
- `packages/core/src/MCPServer.ts`
- `packages/core/src/routers/mcpRouter.ts`

## Next Steps
- Go kernel: implement `/api/mcp/tools/predict-conversational` endpoint to short-circuit the Ollama fallback with an embedded Gemma model
- Dashboard: add a "Predictive Injection" debug panel using `mcp.getConversationWindow`
- Wire `appendConversationTurn` calls from the tRPC bridge clients for assistant message turns (currently only tool-call turns are auto-captured)
