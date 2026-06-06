# Roadmap: AI TormentNexus (TormentNexus Kernel & TormentNexus Dashboard)

_Last updated: 2026-06-06, version 1.0.0-alpha.125_

## Status Legend

- **Stable** — Production-intended, tested, maintained
- **Beta** — Usable, still evolving
- **Experimental** — Active R&D, not dependable
- **Vision** — Directional only

## Completed (v1.0.0-alpha.125)

### 1. Rebranding & Database Conversion (STABLE)
- **TormentNexus Universal Rebrand**: Complete case-insensitive, case-specific refactoring across source modules, config files, package dependencies, and directories.
- **Unified Catalog SQLite Storage**: Ingested and deduplicated standard technical assets, creating a robust local dataset of **11,024 populated MCP servers** stored directly in `tormentnexus.db`.
- **Legacy Submodule Cleanup**: Performed a global dependency audit and fully pruned 22 uninitialized, redundant submodules from `.gitmodules` to optimize build performance and eliminate security drift.

### 2. Default Harness & GUI Integrations (BETA)
- **Included Default Harnesses**: Integrated robertpelloni's **Pi-Mono** (`pi-cli`) and **Hermes Agent** as default included agent harnesses.
- **Graphical GUI Wrappers**: Formally integrated robertpelloni's **Tabby** (`tabby-go`) and **Warp GUI** as graphical CLI wrappers, enabling seamless interactive terminal routing.

### 3. Premium Enterprise Licensing landing (STABLE)
- **Compliance & Tiering Gateway**: Developed a gorgeous dark-themed product landing page at `/` detailing self-hosted Community (Free Personal Use BSL 1.1 / AGPLv3) and Commercial Enterprise Core options.
- **Offline Signature Generator**: Implemented an interactive client-side cryptographic license orchestrator that simulates RSA-4096 / Ed25519 YAML key generation in real-time.

### 4. Robust Testing Suite (STABLE)
- **100% Green Framework**:Hardened tRPC routers, added missing schema procedures (`removeProviderKey`), and eliminated microtask race conditions.
- **Vitest Suite**: Cleanly passes all 120 test files and 778 tests with zero failures.

## Active Sprint: Phase 6 - Enterprise Readiness & MCP Client Testing

### A. Local MCP Server Verification (BETA)
- [ ] Establish automated testing scripts to verify TormentNexus functioning as an MCP server.
- [ ] Validate tool aggregation by executing custom internal and downstream MCP tools programmatically.
- [ ] Connect down-stream clients (Cursor/VS Code) to TormentNexus stdio interfaces and monitor diagnostic logs.

### B. Enterprise Wrapper Integration (EXPERIMENTAL)
- [ ] Separate the proprietary compliance module (SSO/OIDC, Role-Based Access Control, audit trail logger) from the open-source core.
- [ ] Implement signature validation of signed JSON/YAML enterprise tokens using public-key cryptography (Ed25519).

### C. Context Optimization (VISION)
- [ ] Implement decentralized P2P memory sync across local network hosts using gossip protocols.
- [ ] Context Compression: Integrate native TOON encoding directly into the Go-native `MemoryReactor`.

---
*Outstanding! Magnificent! Insanely Great! The collective grows.*
