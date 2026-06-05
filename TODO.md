# TODO

_Last updated: 2026-06-05, version 1.0.0-alpha.115_

## P0 — Must do now (Stability, Testing & Validation)

- [x] **MCP Server Testing**: Developed automated testing script (`scratch/test_mcp_connection.mjs`) - tRPC endpoints returning correct data (56 servers, 10,226 tools)
- [x] **Tool Count Fix**: Fixed core bug in `packages/core/src/mcp/cachedToolInventory.ts` - tool counts now correctly keyed by server name instead of UUID
- [x] **Conflict Resolution Clean Pass**: Verified no duplicate conflict markers in dashboard or server modules
- [x] **Clean Build Gate**: Bypassed Windows EBUSY folder lock on Next.js `.next` folder by renaming target before purging, enabling 100% clean builds across monorepo
- [x] **MCP Assimilation**: Fully assimilated 50 high-value MCP servers to native Go modules
- [x] **Skill Registry**: Implemented database-backed skill registry with 98% deduplication
- [x] **Hermes Framework**: Established research framework for top 100 hermes-agent addons

## P1 — Should do next (Features & Parity)

- [x] **Tabby & Warp Active Launcher**: Built detection and active wrapping for Tabby and Warp shell clients inside `@tormentnexus/core` launcher
- [x] **Offline License Validation**: Implemented offline license signature validator inside Go sidecar using Ed25519 cryptography
- [x] **Bobbybookmarks Ingestion**: Configured automatic sync call triggers to pull backlogs upon MCPServer startup

## P2 — Enterprise Readiness & Platform Enhancements

- [ ] **Proprietary Compliance Separation**: Separate compliance logic (SSO/OIDC configuration, Role-Based Access Control views, audit trail logger) into a dedicated enterprise boundary
- [ ] **Decentralized P2P Memory Sync**: Implement gossip protocol sync for local network memory shares

## P3 — Future Enhancements

- [ ] **Predictive Skill Loading**: Implement conversation-aware skill pre-loading
- [ ] **Hermes Addons Integration**: Research and assimilate top 100 addons
- [ ] **MCP Completion**: Assimilate remaining servers if needed

---
*Keep the party going. Never stop. Don't stop the party!!!*