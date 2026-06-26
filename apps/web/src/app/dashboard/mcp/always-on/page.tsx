"use client";

import { useEffect, useState } from "react";
import { Wrench, Power, PowerOff, Search, RefreshCw } from "lucide-react";

interface ToolInfo {
	name: string;
	description: string;
	alwaysOn: boolean;
}

export default function AlwaysOnToolsPage() {
	const [tools, setTools] = useState<ToolInfo[]>([]);
	const [search, setSearch] = useState("");
	const [loading, setLoading] = useState(true);

	const fetchTools = async () => {
		setLoading(true);
		try {
			const res = await fetch("/api/tools/list");
			const data = await res.json();
			setTools(data.tools ?? []);
		} catch {
			// Fallback: show built-in tools from MCP server
			setTools([
				{
					name: "list_processes",
					description: "List active system processes",
					alwaysOn: true,
				},
				{
					name: "kill_process",
					description: "Kill a process by PID",
					alwaysOn: true,
				},
				{
					name: "simulate_input",
					description: "Send keyboard input via SendKeys",
					alwaysOn: true,
				},
				{
					name: "detect_chat_surface",
					description: "Inspect active window chat surface",
					alwaysOn: true,
				},
				{
					name: "inspect_window_ui",
					description: "List visible UI elements",
					alwaysOn: true,
				},
				{
					name: "detect_chat_state",
					description: "Detect chat waiting state",
					alwaysOn: true,
				},
				{
					name: "set_chat_input",
					description: "Set text in chat composer",
					alwaysOn: true,
				},
				{
					name: "submit_chat_input",
					description: "Submit chat input",
					alwaysOn: true,
				},
				{
					name: "click_action_buttons",
					description: "Click UI buttons by label",
					alwaysOn: true,
				},
				{
					name: "advance_chat",
					description: "Single-step autopilot",
					alwaysOn: true,
				},
				{
					name: "mcp_list_servers",
					description: "List configured MCP servers",
					alwaysOn: true,
				},
				{
					name: "mcp_list_tools",
					description: "List available MCP tools",
					alwaysOn: true,
				},
				{
					name: "mcp_call_tool",
					description: "Call an MCP tool",
					alwaysOn: true,
				},
				{
					name: "mcp_status",
					description: "Get MCP runtime status",
					alwaysOn: true,
				},
				{
					name: "mcp_server_test",
					description: "Test a downstream MCP server",
					alwaysOn: true,
				},
				{
					name: "list_surface_profiles",
					description: "List supervisor surface profiles",
					alwaysOn: true,
				},
				{
					name: "get_supervisor_settings",
					description: "Get supervisor settings",
					alwaysOn: true,
				},
				{
					name: "update_supervisor_settings",
					description: "Update supervisor settings",
					alwaysOn: true,
				},
				{
					name: "list_accessory_tools",
					description: "List all built-in Go accessory tools",
					alwaysOn: true,
				},
				{
					name: "system_status",
					description: "Get overall system health",
					alwaysOn: true,
				},
				{
					name: "billing_status",
					description: "Get billing and provider status",
					alwaysOn: true,
				},
			]);
		}
		setLoading(false);
	};

	useEffect(() => {
		fetchTools();
	}, []);

	const toggleAlwaysOn = async (name: string) => {
		try {
			await fetch("/api/tools/always-on", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					name,
					alwaysOn: !tools.find((t) => t.name === name)?.alwaysOn,
				}),
			});
			setTools((prev) =>
				prev.map((t) =>
					t.name === name ? { ...t, alwaysOn: !t.alwaysOn } : t,
				),
			);
		} catch {
			// toggle locally if API unavailable
			setTools((prev) =>
				prev.map((t) =>
					t.name === name ? { ...t, alwaysOn: !t.alwaysOn } : t,
				),
			);
		}
	};

	const filtered = tools.filter(
		(t) =>
			t.name.toLowerCase().includes(search.toLowerCase()) ||
			t.description.toLowerCase().includes(search.toLowerCase()),
	);

	const alwaysOn = filtered.filter((t) => t.alwaysOn);
	const optional = filtered.filter((t) => !t.alwaysOn);

	return (
		<div className="p-6 space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-bold flex items-center gap-2">
						<Wrench className="w-6 h-6" />
						Always-On Tools
					</h1>
					<p className="text-zinc-400 text-sm mt-1">
						Configure which built-in TormentNexus tools are always available to
						the MCP client. Parity tools are always-on by default.
					</p>
				</div>
				<button
					onClick={fetchTools}
					className="flex items-center gap-2 px-3 py-2 bg-zinc-800 rounded-lg hover:bg-zinc-700 text-sm"
				>
					<RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
					Refresh
				</button>
			</div>

			<div className="relative">
				<Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
				<input
					type="text"
					placeholder="Search tools..."
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					className="w-full pl-10 pr-4 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-sm focus:outline-none focus:border-zinc-500"
				/>
			</div>

			{loading ? (
				<div className="text-center py-12 text-zinc-500">Loading tools...</div>
			) : (
				<div className="space-y-8">
					{alwaysOn.length > 0 && (
						<section>
							<h2 className="text-lg font-semibold text-emerald-400 mb-3 flex items-center gap-2">
								<Power className="w-4 h-4" />
								Always-On ({alwaysOn.length})
							</h2>
							<div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
								{alwaysOn.map((tool) => (
									<ToolCard
										key={tool.name}
										tool={tool}
										onToggle={toggleAlwaysOn}
									/>
								))}
							</div>
						</section>
					)}

					{optional.length > 0 && (
						<section>
							<h2 className="text-lg font-semibold text-zinc-400 mb-3 flex items-center gap-2">
								<PowerOff className="w-4 h-4" />
								Optional ({optional.length})
							</h2>
							<div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
								{optional.map((tool) => (
									<ToolCard
										key={tool.name}
										tool={tool}
										onToggle={toggleAlwaysOn}
									/>
								))}
							</div>
						</section>
					)}
				</div>
			)}
		</div>
	);
}

function ToolCard({
	tool,
	onToggle,
}: {
	tool: ToolInfo;
	onToggle: (name: string) => void;
}) {
	return (
		<div
			className={`border rounded-lg p-4 flex items-start justify-between gap-3 transition-colors ${
				tool.alwaysOn
					? "border-emerald-700/50 bg-emerald-950/20"
					: "border-zinc-800 bg-zinc-900/50"
			}`}
		>
			<div className="min-w-0 flex-1">
				<div className="font-mono text-sm font-medium truncate">
					{tool.name}
				</div>
				<div className="text-xs text-zinc-500 mt-1 line-clamp-2">
					{tool.description}
				</div>
			</div>
			<button
				onClick={() => onToggle(tool.name)}
				className={`shrink-0 p-2 rounded-lg transition-colors ${
					tool.alwaysOn
						? "bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/30"
						: "bg-zinc-800 text-zinc-500 hover:bg-zinc-700"
				}`}
				title={tool.alwaysOn ? "Disable always-on" : "Enable always-on"}
			>
				{tool.alwaysOn ? (
					<Power className="w-4 h-4" />
				) : (
					<PowerOff className="w-4 h-4" />
				)}
			</button>
		</div>
	);
}
