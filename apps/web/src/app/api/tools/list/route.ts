import { NextResponse } from "next/server";

interface ToolInfo {
	name: string;
	description: string;
	alwaysOn: boolean;
}

/**
 * Returns the full list of built-in TormentNexus MCP tools.
 *
 * When the Go sidecar is reachable it fetches the live tool list from
 * the sidecar's MCP tool registry.  When the sidecar is unreachable
 * (dev mode / offline) it falls back to a static snapshot.
 */
export async function GET() {
	const GO_SIDECAR =
		process.env.TORMENTNEXUS_GO_SIDECAR_URL || "http://127.0.0.1:7778";

	// Try the live sidecar first
	try {
		const res = await fetch(`${GO_SIDECAR}/api/mcp/tools`, {
			signal: AbortSignal.timeout(3000),
		});
		if (res.ok) {
			const data = await res.json();
			const toolsList = data && Array.isArray(data.data) ? data.data : (Array.isArray(data) ? data : []);
			return NextResponse.json({ tools: toolsList });
		}
	} catch {
		// fall through to static list
	}

	// Static fallback — matches the MCP server's built-in tools
	const tools: ToolInfo[] = [
		// Process Management (always-on)
		{
			name: "list_processes",
			description: "List active system processes on Windows",
			alwaysOn: true,
		},
		{
			name: "kill_process",
			description: "Kill a process by PID",
			alwaysOn: true,
		},
		// Input Simulation (always-on)
		{
			name: "simulate_input",
			description: "Send keyboard input via PowerShell SendKeys",
			alwaysOn: true,
		},
		// UI Inspection (always-on)
		{
			name: "detect_chat_surface",
			description: "Inspect active window and classify chat surface",
			alwaysOn: true,
		},
		{
			name: "inspect_window_ui",
			description: "List visible UI elements from the active window",
			alwaysOn: true,
		},
		{
			name: "detect_chat_state",
			description:
				"Detect whether chat is waiting for input or has action buttons",
			alwaysOn: true,
		},
		// Chat Automation (always-on)
		{
			name: "set_chat_input",
			description: "Set text in the active chat composer",
			alwaysOn: true,
		},
		{
			name: "submit_chat_input",
			description: "Submit the current chat input",
			alwaysOn: true,
		},
		{
			name: "click_action_buttons",
			description: "Click UI buttons by label text",
			alwaysOn: true,
		},
		{
			name: "click_chat_button",
			description: "Click a button on the active chat surface",
			alwaysOn: true,
		},
		{
			name: "advance_chat",
			description: "Single-step autopilot: click buttons or type bump text",
			alwaysOn: true,
		},
		// MCP Server Management (always-on)
		{
			name: "mcp_list_servers",
			description: "List configured MCP servers from the Go sidecar",
			alwaysOn: true,
		},
		{
			name: "mcp_list_tools",
			description: "List available MCP tools from the Go sidecar",
			alwaysOn: true,
		},
		{
			name: "mcp_call_tool",
			description: "Call an MCP tool through the Go sidecar",
			alwaysOn: true,
		},
		{
			name: "mcp_status",
			description: "Get MCP runtime status from the Go sidecar",
			alwaysOn: true,
		},
		{
			name: "mcp_server_test",
			description: "Test a downstream MCP server connection",
			alwaysOn: true,
		},
		// System (always-on)
		{
			name: "system_status",
			description: "Get overall system health status",
			alwaysOn: true,
		},
		{
			name: "billing_status",
			description: "Get billing and provider status",
			alwaysOn: true,
		},
		// Supervisor (always-on)
		{
			name: "list_surface_profiles",
			description: "List known supervisor surface profiles",
			alwaysOn: true,
		},
		{
			name: "get_supervisor_settings",
			description: "Get supervisor default settings",
			alwaysOn: true,
		},
		{
			name: "update_supervisor_settings",
			description: "Update supervisor default settings",
			alwaysOn: true,
		},
		// Accessory (always-on)
		{
			name: "list_accessory_tools",
			description:
				"List all built-in Go accessory tools from the root registry",
			alwaysOn: true,
		},
		// Go sidecar native tool stubs (optional — can be toggled on)
		{
			name: "echo",
			description: "Echo back the provided message",
			alwaysOn: false,
		},
		{
			name: "hello_world",
			description: "Return a greeting message",
			alwaysOn: false,
		},
		{
			name: "current_time",
			description: "Return the current system time",
			alwaysOn: false,
		},
		{
			name: "weather",
			description: "Get weather information for a location",
			alwaysOn: false,
		},
		{
			name: "calc",
			description: "Evaluate a mathematical expression",
			alwaysOn: false,
		},
		{
			name: "read_file",
			description: "Read the contents of a file",
			alwaysOn: false,
		},
		{
			name: "write_file",
			description: "Write content to a file",
			alwaysOn: false,
		},
		{
			name: "list_dir",
			description: "List directory contents",
			alwaysOn: false,
		},
		{
			name: "run_command",
			description: "Run a shell command",
			alwaysOn: false,
		},
		{
			name: "search_text",
			description: "Search for text across files using ripgrep",
			alwaysOn: false,
		},
		{
			name: "semantic_search",
			description: "Semantic search across the L2 memory vault",
			alwaysOn: false,
		},
		{
			name: "reinforce_memory",
			description: "Reinforce or decay a memory record by ID",
			alwaysOn: false,
		},
		{
			name: "catalog_search",
			description: "Search the MCP catalog for available servers",
			alwaysOn: false,
		},
		{
			name: "session_list",
			description: "List imported sessions",
			alwaysOn: false,
		},
		{
			name: "session_get",
			description: "Get a specific imported session with its memories",
			alwaysOn: false,
		},
		{
			name: "provider_status",
			description: "Get current model provider availability",
			alwaysOn: false,
		},
		{
			name: "provider_switch",
			description: "Switch the active model provider",
			alwaysOn: false,
		},
		{
			name: "code_execute",
			description: "Execute code in a sandboxed environment",
			alwaysOn: false,
		},
		{
			name: "git_status",
			description: "Get git working tree status",
			alwaysOn: false,
		},
		{ name: "git_commit", description: "Create a git commit", alwaysOn: false },
		{
			name: "skill_list",
			description: "List available skills from the skill registry",
			alwaysOn: false,
		},
		{
			name: "skill_get",
			description: "Get a specific skill by name",
			alwaysOn: false,
		},
		{
			name: "skill_search",
			description: "Search skills by keyword",
			alwaysOn: false,
		},
		{
			name: "memory_get_all",
			description: "Get all L2 memory records",
			alwaysOn: false,
		},
		{
			name: "memory_save",
			description: "Save a new memory record",
			alwaysOn: false,
		},
		{
			name: "workflow_list",
			description: "List configured workflows",
			alwaysOn: false,
		},
		{
			name: "workflow_trigger",
			description: "Trigger a workflow execution",
			alwaysOn: false,
		},
		{
			name: "mesh_status",
			description: "Get mesh network status",
			alwaysOn: false,
		},
		{
			name: "assimilation_status",
			description: "Get MCP assimilation pipeline status",
			alwaysOn: false,
		},
	];

	return NextResponse.json({ tools });
}
