/**
 * @file mcp-config-discovery.service.ts
 * @module packages/core/src/services/mcp-config-discovery.service
 *
 * Borg-native entry point for loading MCP server definitions from config.
 * The implementation still lives in the compatibility-named module for now,
 * but active imports can use this generic Borg-owned path immediately.
 */

export { getMcpServers } from "./fetch-metamcp.service.js";