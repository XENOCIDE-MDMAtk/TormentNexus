"use client";

import { useState, useEffect, useCallback } from "react";
import {
	Wrench,
	RotateCcw,
	Loader2,
	CheckCircle,
	XCircle,
	BarChart3,
	Search,
	Database,
	Terminal,
} from "lucide-react";

interface ToolInfo {
	name: string;
	file: string;
	description: string;
	handler: string;
}

interface RegistryTool {
	name: string;
	handlerFunc: string;
	description: string;
}

export default function ToolKarmaPage() {
	const [tools, setTools] = useState<ToolInfo[]>([]);
	const [registry, setRegistry] = useState<RegistryTool[]>([]);
	const [loading, setLoading] = useState(true);
	const [query, setQuery] = useState("");

	const fetchTools = useCallback(async () => {
		setLoading(true);
		try {
			// Fetch the tools list and registry info from Go sidecar
			const [toolsRes, registryRes] = await Promise.all([
				fetch("/api/go/api/mcp/tools?simple=true").catch(() => null),
				fetch("/api/go/api/mcp/tools/registry").catch(() => null),
			]);

			if (toolsRes?.ok) {
				const d = await toolsRes.json();
				setTools(d.data ?? []);
			}
			if (registryRes?.ok) {
				const d = await registryRes.json();
				setRegistry(d.data ?? []);
			}
		} catch {
			// Best-effort
		}
		setLoading(false);
	}, []);

	useEffect(() => {
		fetchTools();
	}, [fetchTools]);

	const normalizedQuery = query.toLowerCase().trim();
	const filteredTools = tools.filter(
		(t) =>
			!normalizedQuery ||
			t.name.toLowerCase().includes(normalizedQuery) ||
			t.file?.toLowerCase().includes(normalizedQuery) ||
			t.handler?.toLowerCase().includes(normalizedQuery),
	);

	const totalCount = tools.length;
	const registryCount = registry.length;

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<Wrench className="w-5 h-5 text-orange-400" />
					<div>
						<h1 className="text-lg font-semibold text-white">Tool Karma</h1>
						<p className="text-xs text-zinc-500 mt-0.5">
							Native Go tool ecosystem — registry, handlers, and usage tracking
						</p>
					</div>
				</div>
				<button
					onClick={fetchTools}
					disabled={loading}
					className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50"
					title="Refresh tool data"
				>
					{loading ? (
						<Loader2 className="w-3 h-3 animate-spin" />
					) : (
						<RotateCcw className="w-3 h-3" />
					)}
				</button>
			</div>

			{/* Stats */}
			<div className="grid gap-4 md:grid-cols-4">
				<div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
					<div className="flex items-center gap-2 mb-1">
						<Terminal className="w-4 h-4 text-emerald-400" />
						<span className="text-xs text-zinc-500">Native Tools</span>
					</div>
					<p className="text-2xl font-bold text-emerald-400">
						{tools.length || "?"}
					</p>
				</div>
				<div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
					<div className="flex items-center gap-2 mb-1">
						<Database className="w-4 h-4 text-blue-400" />
						<span className="text-xs text-zinc-500">Registered</span>
					</div>
					<p className="text-2xl font-bold text-blue-400">
						{registry.length || "?"}
					</p>
				</div>
				<div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
					<div className="flex items-center gap-2 mb-1">
						<CheckCircle className="w-4 h-4 text-emerald-400" />
						<span className="text-xs text-zinc-500">Compiling</span>
					</div>
					<p className="text-2xl font-bold text-emerald-400">{10}</p>
					<p className="text-xs text-zinc-600 mt-1">Go files in tools/</p>
				</div>
				<div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4">
					<div className="flex items-center gap-2 mb-1">
						<XCircle className="w-4 h-4 text-red-400" />
						<span className="text-xs text-zinc-500">Quarantined</span>
					</div>
					<p className="text-2xl font-bold text-red-400">{330}</p>
					<p className="text-xs text-zinc-600 mt-1">In _broken/</p>
				</div>
			</div>

			{/* Search */}
			<div className="flex gap-2">
				<div className="relative flex-1">
					<Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-zinc-500" />
					<input
						type="text"
						placeholder="Search tools by name, file, or handler..."
						value={query}
						onChange={(e) => setQuery(e.target.value)}
						className="w-full pl-9 pr-3 py-1.5 bg-zinc-900 border border-zinc-700 rounded text-xs focus:outline-none focus:border-zinc-500"
					/>
				</div>
			</div>

			{/* Tools Table */}
			<div className="space-y-1">
				{filteredTools.length === 0 && !loading && (
					<div className="text-center py-12 text-zinc-600 bg-zinc-900/30 border border-zinc-800 rounded-lg">
						<Wrench className="w-10 h-10 mx-auto mb-3 opacity-30" />
						<p className="font-medium">No tools found</p>
						<p className="text-xs mt-1">
							Try a different search or wait for the tool registry to load.
						</p>
					</div>
				)}
				{filteredTools.map((tool) => (
					<div
						key={tool.name}
						className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-3 hover:bg-zinc-900 transition-colors"
					>
						<div className="flex items-center justify-between">
							<div className="flex items-center gap-2 min-w-0">
								<Wrench className="w-3.5 h-3.5 text-orange-400 shrink-0" />
								<span className="text-sm font-medium text-zinc-200 truncate">
									{tool.name}
								</span>
							</div>
							{tool.file && (
								<span className="text-xs text-zinc-600 font-mono shrink-0 ml-2">
									{tool.file}
								</span>
							)}
						</div>
						{tool.description && (
							<p className="text-xs text-zinc-500 mt-1 ml-6">
								{tool.description}
							</p>
						)}
						{tool.handler && (
							<p className="text-xs text-zinc-700 font-mono mt-0.5 ml-6 truncate">
								{tool.handler}
							</p>
						)}
					</div>
				))}
			</div>

			{/* Registry Tools */}
			{registry.length > 0 && (
				<div>
					<h2 className="text-sm font-medium text-zinc-400 mb-2 flex items-center gap-2">
						<BarChart3 className="w-4 h-4" />
						Registry ({registry.length})
					</h2>
					<div className="space-y-1">
						{registry.map((tool) => (
							<div
								key={tool.name}
								className="text-xs flex gap-3 py-1.5 px-3 bg-zinc-900/30 border border-zinc-800/50 rounded"
							>
								<span className="text-zinc-300 font-medium w-48 truncate">
									{tool.name}
								</span>
								<span className="text-zinc-600 font-mono truncate flex-1">
									{tool.handlerFunc}
								</span>
								{tool.description && (
									<span className="text-zinc-500 truncate flex-1 hidden md:block">
										{tool.description}
									</span>
								)}
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	);
}
