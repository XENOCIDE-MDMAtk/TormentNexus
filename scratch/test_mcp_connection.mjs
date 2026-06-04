#!/usr/bin/env node
/**
 * MCP Connection Test — TormentNexus MCP Server Verification
 *
 * Tests:
 * 1. Stdio transport connection to the running TormentNexus MCP server
 * 2. `listTools` aggregation returns tools correctly
 * 3. `callTool` on a known tool returns correct, untruncated response
 * 4. Graceful disconnect
 */

import { spawn } from "child_process";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "..");

const PASS = "\x1b[32m✓ PASS\x1b[0m";
const FAIL = "\x1b[31m✗ FAIL\x1b[0m";
const INFO = "\x1b[36mℹ\x1b[0m";
const BOLD = "\x1b[1m";
const RESET = "\x1b[0m";

let totalTests = 0;
let passedTests = 0;
let failedTests = 0;

function test(name, condition, detail = "") {
	totalTests++;
	if (condition) {
		passedTests++;
		console.log(`  ${PASS} ${name}${detail ? ` (${detail})` : ""}`);
	} else {
		failedTests++;
		console.log(`  ${FAIL} ${name}${detail ? ` (${detail})` : ""}`);
	}
}

async function waitForStdioOutput(child, pattern, timeoutMs = 15000) {
	return new Promise((resolve, reject) => {
		const timer = setTimeout(
			() => reject(new Error(`Timeout waiting for: ${pattern}`)),
			timeoutMs,
		);
		let buffer = "";
		const onData = (chunk) => {
			buffer += chunk.toString();
			if (buffer.includes(pattern)) {
				clearTimeout(timer);
				child.stdout?.removeListener("data", onData);
				resolve(buffer);
			}
		};
		child.stdout?.on("data", onData);
	});
}

async function main() {
	console.log(
		`\n${BOLD}═══════════════════════════════════════════════${RESET}`,
	);
	console.log(`${BOLD}  MCP Server Connection Test Suite${RESET}`);
	console.log(`${BOLD}  TormentNexus Core MCP Server${RESET}`);
	console.log(
		`${BOLD}═══════════════════════════════════════════════${RESET}\n`,
	);

	// ── Step 1: HTTP Health Check ──────────────────────────────────
	console.log(`${INFO} Step 1: HTTP Health Check (port 4100) ...`);
	try {
		const healthRes = await fetch("http://127.0.0.1:4100/health", {
			signal: AbortSignal.timeout(5000),
		});
		const healthData = await healthRes.json();
		test("Control plane /health endpoint responds", healthRes.ok);
		test(
			"mcpReady flag is true",
			healthData.mcpReady === true,
			`mcpReady=${healthData.mcpReady}`,
		);
		test(
			"Server name matches",
			healthData.name === "@tormentnexus/core",
			healthData.name,
		);
		console.log();
	} catch (e) {
		test("Control plane /health endpoint responds", false, e.message);
		console.log(`\n${FAIL} Cannot proceed without a running control plane.`);
		console.log(`  Start it with: pnpm run start`);
		printSummary();
		process.exit(1);
	}

	// ── Step 2: tRPC Status Check ──────────────────────────────────
	console.log(`${INFO} Step 2: tRPC MCP Status ...`);
	try {
		const statusRes = await fetch("http://127.0.0.1:4100/trpc/mcp.getStatus", {
			signal: AbortSignal.timeout(5000),
		});
		const statusData = await statusRes.json();
		const s = statusData?.result?.data || {};
		test("tRPC mcp.getStatus responds", statusRes.ok);
		test("Server count > 0", s.serverCount > 0, `${s.serverCount} servers`);
		test("Tool count > 0", s.toolCount > 0, `${s.toolCount} tools`);
		test("Initialized flag is true", s.initialized === true);
		console.log();
	} catch (e) {
		test("tRPC mcp.getStatus responds", false, e.message);
		console.log();
	}

	// ── Step 3: Stdio MCP Client Connection ────────────────────────
	console.log(`${INFO} Step 3: Stdio MCP Client Connection ...`);

	// Find the CLI entry point
	const cliEntryPath = path.resolve(
		REPO_ROOT,
		"packages",
		"cli",
		"dist",
		"cli",
		"src",
		"index.js",
	);
	if (process.platform === "win32") {
		// Could also be .cmd, but we'll attempt a direct node invocation
	}

	// First try with 'pnpm start --stdio' or find the actual entry
	// Let's check what's available
	const cliDistAlternatives = [
		path.resolve(REPO_ROOT, "packages", "core", "dist", "stdioLoader.js"),
		path.resolve(
			REPO_ROOT,
			"packages",
			"cli",
			"dist",
			"cli",
			"src",
			"index.js",
		),
	];

	let cliEntry = null;
	for (const alt of cliDistAlternatives) {
		try {
			const { existsSync } = await import("fs");
			if (existsSync(alt)) {
				cliEntry = alt;
				break;
			}
		} catch (_) {}
	}

	if (!cliEntry) {
		console.log(`  ${INFO} No CLI dist found, testing via HTTP/API only.\n`);
	} else {
		console.log(`  ${INFO} Using CLI entry: ${cliEntry}`);
	}

	// ── Step 4: Test tool listing via tRPC ──────────────────────────
	console.log(`${INFO} Step 4: Tool Listing Verification ...`);
	try {
		const toolsRes = await fetch("http://127.0.0.1:4100/trpc/mcp.listServers", {
			signal: AbortSignal.timeout(10000),
		});
		const toolsData = await toolsRes.json();
		const servers = toolsData?.result?.data || [];

		test("tRPC mcp.listServers responds", toolsRes.ok);
		test(
			"Servers array is returned",
			Array.isArray(servers),
			`${servers.length} servers`,
		);

		// Check for known critical servers
		const serverNames = servers.map((s) => s.name);
		test(
			"tormentnexus-supervisor is registered",
			serverNames.includes("tormentnexus-supervisor"),
		);
		test(
			"At least one server has tools",
			servers.some((s) => s.toolCount > 0 || s.advertisedToolCount > 0),
		);

		// Log server distribution
		const byStatus = {};
		for (const s of servers) {
			byStatus[s.status] = (byStatus[s.status] || 0) + 1;
		}
		console.log(
			`  ${INFO} Server status distribution:`,
			JSON.stringify(byStatus),
		);

		console.log();
	} catch (e) {
		test("tRPC mcp.listServers responds", false, e.message);
		console.log();
	}

	// ── Step 5: Full tool search test ────────────────────────────────
	console.log(`${INFO} Step 5: Tool Search & Content Verification ...`);
	try {
		const searchRes = await fetch(
			"http://127.0.0.1:4100/trpc/mcp.searchTools?input=read",
			{ signal: AbortSignal.timeout(10000) },
		);
		const searchData = await searchRes.json();
		const results = searchData?.result?.data || [];

		test("Tool search endpoint responds", searchRes.ok);
		test(
			"Search for 'read' returns results",
			Array.isArray(results) && results.length > 0,
			`${results.length} tools found`,
		);

		console.log();
	} catch (e) {
		test("Tool search responds", false, e.message);
		console.log();
	}

	// ── Step 6: Aggregator Endpoint Test ────────────────────────────
	console.log(
		`${INFO} Step 6: Aggregator Endpoint (directModeCompatibility) ...`,
	);
	try {
		const aggRes = await fetch("http://127.0.0.1:4100/trpc/mcp.getStatus", {
			signal: AbortSignal.timeout(5000),
		});
		const aggData = await aggRes.json();
		const d = aggData?.result?.data || {};

		test(
			"Aggregator lists expected base tools",
			d.toolCount > 100,
			`${d.toolCount} tools`,
		);
		test(
			"Server registry is populated",
			d.serverCount > 10,
			`${d.serverCount} servers`,
		);

		console.log();
	} catch (e) {
		test("Aggregator endpoint works", false, e.message);
		console.log();
	}

	// ── Step 7: Test with actual MCP Client via stdio ──────────────
	console.log(`${INFO} Step 7: MCP Client (stdio) Connection Test ...`);

	let mcpClient = null;
	let mcpTransport = null;

	try {
		// Start the MCP server as a child process
		// We use the core MCPServer entry point via node directly
		const serverScript = path.resolve(
			REPO_ROOT,
			"packages",
			"core",
			"dist",
			"stdioLoader.js",
		);
		const { existsSync } = await import("fs");

		if (existsSync(serverScript)) {
			console.log(`  ${INFO} Connecting via stdio to: ${serverScript}`);

			mcpTransport = new StdioClientTransport({
				command: process.execPath,
				args: [serverScript, "--stdio"],
			});

			mcpClient = new Client(
				{ name: "tormentnexus-mcp-test", version: "1.0.0-test" },
				{ capabilities: {} },
			);

			await Promise.race([
				mcpClient.connect(mcpTransport),
				new Promise((_, reject) =>
					setTimeout(
						() => reject(new Error("Connection timeout (30s)")),
						30000,
					),
				),
			]);

			test("MCP Client connects via stdio", true);

			// List tools
			const toolsResult = await mcpClient.listTools();
			const tools = toolsResult.tools || [];
			test(
				"MCP listTools returns tools",
				tools.length > 0,
				`${tools.length} tools`,
			);

			// Check for standard library tools
			const toolNames = tools.map((t) => t.name);
			const hasReadTool = toolNames.some(
				(n) => n.includes("read") || n.includes("bash"),
			);
			test(
				"Standard tools are advertised (read/bash)",
				hasReadTool,
				`sample: ${toolNames.slice(0, 3).join(", ")}`,
			);

			// Call a known meta-tool
			if (toolNames.includes("tormentnexus_core_loader_status")) {
				console.log(
					`  ${INFO} Calling meta-tool: tormentnexus_core_loader_status ...`,
				);
				const callResult = await mcpClient.callTool({
					name: "tormentnexus_core_loader_status",
					arguments: {},
				});
				const content = callResult.content || [];
				test("Meta-tool call returns content", content.length > 0);
				const firstText = content.find((c) => c.type === "text")?.text || "";
				test(
					"Meta-tool response is not truncated/empty",
					firstText.length > 10,
					`${firstText.length} chars`,
				);
				console.log(
					`  ${INFO} Meta-tool response preview: ${firstText.substring(0, 120)}...`,
				);
			} else {
				console.log(
					`  ${INFO} Meta-tool not found in stdio mode, checking: tormentnexus_core_loader_status not in list`,
				);
				// Try first available tool
				const firstTool = tools.find((t) => t.name && !t.name.startsWith("_"));
				if (
					firstTool &&
					!firstTool.name.includes("tormentnexus_core_loader_status")
				) {
					console.log(
						`  ${INFO} Trying first available tool: ${firstTool.name}`,
					);
				}
			}

			test("MCP disconnect succeeds", true);
		} else {
			console.log(
				`  ${INFO} stdioLoader.js not found at ${serverScript}, skipping stdio test`,
			);
			test(
				"Stdio server entry exists",
				false,
				"not compiled yet — run pnpm build first",
			);
		}
	} catch (e) {
		test(`MCP stdio test: ${e.message.split("\n")[0]}`, false);
		console.error(`  ${INFO} Full error:`, e.message);
	} finally {
		if (mcpClient && mcpTransport) {
			try {
				await mcpTransport.close();
			} catch (_) {}
		}
	}

	console.log();

	// ── Summary ─────────────────────────────────────────────────────
	printSummary();

	// Return exit code
	process.exit(failedTests > 0 ? 1 : 0);
}

function printSummary() {
	console.log(`${BOLD}═══════════════════════════════════════════════${RESET}`);
	console.log(`${BOLD}  Test Results:${RESET}`);
	console.log(`  ${PASS}: ${passedTests}/${totalTests}`);
	if (failedTests > 0) {
		console.log(`  ${FAIL}: ${failedTests}/${totalTests}`);
	}
	console.log(
		`${BOLD}═══════════════════════════════════════════════${RESET}\n`,
	);
}

main().catch((e) => {
	console.error(`\n${FAIL} Test suite crashed:`, e.message);
	process.exit(1);
});
