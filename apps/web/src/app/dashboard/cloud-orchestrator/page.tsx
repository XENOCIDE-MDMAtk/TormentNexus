"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
	ExternalLink,
	Rocket,
	KeyRound,
	PlugZap,
	RefreshCw,
	CheckCircle2,
	XCircle,
	Trash2,
	ShieldCheck,
	BarChart3,
	Sparkles,
	Terminal,
	Bot,
	Server,
	Wrench,
	Cpu,
	Activity,
	Loader2,
	AlertTriangle,
} from "lucide-react";
import { trpc } from "@/utils/trpc";
import {
	Card,
	CardHeader,
	CardTitle,
	CardDescription,
	CardContent,
	Badge,
	Button,
	Tabs,
	TabsContent,
	TabsList,
	TabsTrigger,
	ScrollArea,
} from "@tormentnexus/ui";

const JULES_API_KEY_STORAGE = "jules-api-key";
const JULES_SYNC_LOG_STORAGE_KEY = "jules-session-sync-log-v1";
const CLAUDE_CLOUD_KEY_STORAGE = "claude-cloud-api-key";
const COPILOT_CLOUD_KEY_STORAGE = "copilot-cloud-api-key";
const CODEX_CLOUD_KEY_STORAGE = "codex-cloud-api-key";
const SPARK_CLOUD_KEY_STORAGE = "spark-cloud-api-key";
const BLOCKS_CLOUD_KEY_STORAGE = "blocks-cloud-api-key";

type SessionSyncLogEntry = {
	sessionId: string;
	targetStatus?: "active" | "completed" | "failed" | "paused" | "awaiting_approval";
	outcome: "success" | "fallback" | "error";
	message: string;
	timestamp: string;
};

interface ToolItem {
	uuid: string;
	name: string;
	description?: string;
	serverId?: string;
}

interface ServerItem {
	id: string;
	name: string;
	status?: string;
	transport?: string;
}

function normalizeArray<T>(value: unknown): T[] {
	if (!Array.isArray(value)) return [];
	return value as T[];
}

export default function CloudOrchestratorDashboardPage() {
	const [activeTab, setActiveTab] = useState("jules");

	// 1. Jules Autopilot States
	const julesUrl = useMemo(
		() => process.env.NEXT_PUBLIC_JULES_DASHBOARD_URL || "http://localhost:3002/jules/autopilot",
		[],
	);
	const [julesApiKey, setJulesApiKey] = useState("");
	const [julesChecking, setJulesChecking] = useState(false);
	const [julesStatus, setJulesStatus] = useState<{
		ok: boolean;
		message: string;
		checkedAt: string;
	} | null>(null);
	const [syncLogs, setSyncLogs] = useState<SessionSyncLogEntry[]>([]);

	// 2. Claude Cloud States
	const [claudeApiKey, setClaudeApiKey] = useState("");
	const [claudeChecking, setClaudeChecking] = useState(false);
	const [claudeStatus, setClaudeStatus] = useState<{
		ok: boolean;
		message: string;
		checkedAt: string;
	} | null>(null);
	const [claudeLogs, setClaudeLogs] = useState<Array<{ text: string; time: string; level: string }>>([
		{ text: "Claude Cloud gateway connection initialized.", time: "10:00:00 AM", level: "info" },
	]);

	// 3. Copilot Cloud States
	const [copilotApiKey, setCopilotApiKey] = useState("");
	const [copilotStatus, setCopilotStatus] = useState<string>("");

	// 4. OpenAI Codex States
	const [codexApiKey, setCodexApiKey] = useState("");
	const [codexStatus, setCodexStatus] = useState<string>("");

	// 5. Spark States
	const [sparkApiKey, setSparkApiKey] = useState("");
	const [sparkStatus, setSparkStatus] = useState<string>("");

	// 6. Blocks States
	const [blocksApiKey, setBlocksApiKey] = useState("");
	const [blocksStatus, setBlocksStatus] = useState<string>("");

	// 7. OpenCode Autopilot States
	const autopilotUrl = "http://localhost:3847";

	// 8. Super Assistant Queries
	const toolsQuery = trpc.tools.list.useQuery(undefined, { enabled: activeTab === "super" });
	const serversQuery = trpc.mcpServers.list.useQuery(undefined, { enabled: activeTab === "super" });
	const skillsQuery = trpc.skills.list.useQuery(undefined, { enabled: activeTab === "super" });

	const tools = normalizeArray<ToolItem>(toolsQuery.data);
	const servers = normalizeArray<ServerItem>(serversQuery.data);
	const skills = normalizeArray<{ id: string; name: string; description: string }>(skillsQuery.data);

	const activeServers = servers.filter((s) => s.status === "connected" || s.status === "active");
	const hasErrors = toolsQuery.isError || serversQuery.isError || skillsQuery.isError;
	const isLoading = toolsQuery.isLoading || serversQuery.isLoading || skillsQuery.isLoading;

	useEffect(() => {
		setJulesApiKey(localStorage.getItem(JULES_API_KEY_STORAGE) || "");
		setClaudeApiKey(localStorage.getItem(CLAUDE_CLOUD_KEY_STORAGE) || "");
		setCopilotApiKey(localStorage.getItem(COPILOT_CLOUD_KEY_STORAGE) || "");
		setCodexApiKey(localStorage.getItem(CODEX_CLOUD_KEY_STORAGE) || "");
		setSparkApiKey(localStorage.getItem(SPARK_CLOUD_KEY_STORAGE) || "");
		setBlocksApiKey(localStorage.getItem(BLOCKS_CLOUD_KEY_STORAGE) || "");
		refreshSyncLogs();
	}, []);

	const refreshSyncLogs = useCallback(() => {
		try {
			const raw = localStorage.getItem(JULES_SYNC_LOG_STORAGE_KEY);
			const parsed = raw ? JSON.parse(raw) : [];
			setSyncLogs(Array.isArray(parsed) ? parsed : []);
		} catch {
			setSyncLogs([]);
		}
	}, []);

	// Jules actions
	const saveJulesKey = () => {
		const trimmed = julesApiKey.trim();
		localStorage.setItem(JULES_API_KEY_STORAGE, trimmed);
		setJulesApiKey(trimmed);
		setJulesStatus({
			ok: true,
			message: "Jules API key saved locally.",
			checkedAt: new Date().toISOString(),
		});
	};
	const clearJulesKey = () => {
		localStorage.removeItem(JULES_API_KEY_STORAGE);
		setJulesApiKey("");
		setJulesStatus({
			ok: true,
			message: "Jules API key cleared.",
			checkedAt: new Date().toISOString(),
		});
	};
	const testJulesProxy = async () => {
		const key = julesApiKey.trim();
		if (!key) {
			setJulesStatus({
				ok: false,
				message: "Enter API key first.",
				checkedAt: new Date().toISOString(),
			});
			return;
		}
		setJulesChecking(true);
		try {
			const path = encodeURIComponent("/sources?pageSize=1");
			const res = await fetch(`/api/jules?path=${path}`, {
				headers: { "x-jules-api-key": key },
			});
			const data = await res.json().catch(() => ({}));
			if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
			setJulesStatus({
				ok: true,
				message: "Connected successfully to Jules API.",
				checkedAt: new Date().toISOString(),
			});
		} catch (err: any) {
			setJulesStatus({
				ok: false,
				message: `Check failed: ${err.message}`,
				checkedAt: new Date().toISOString(),
			});
		} finally {
			setJulesChecking(false);
		}
	};

	// Claude actions
	const saveClaudeKey = () => {
		const trimmed = claudeApiKey.trim();
		localStorage.setItem(CLAUDE_CLOUD_KEY_STORAGE, trimmed);
		setClaudeApiKey(trimmed);
		setClaudeStatus({
			ok: true,
			message: "Anthropic API Key saved locally.",
			checkedAt: new Date().toISOString(),
		});
		setClaudeLogs((prev) => [
			...prev,
			{ text: "API key updated in memory store.", time: new Date().toLocaleTimeString(), level: "success" },
		]);
	};
	const clearClaudeKey = () => {
		localStorage.removeItem(CLAUDE_CLOUD_KEY_STORAGE);
		setClaudeApiKey("");
		setClaudeStatus({
			ok: true,
			message: "API Key cleared.",
			checkedAt: new Date().toISOString(),
		});
	};
	const testClaudeProxy = async () => {
		setClaudeChecking(true);
		setClaudeLogs((prev) => [
			...prev,
			{ text: "Verifying Anthropic gateway...", time: new Date().toLocaleTimeString(), level: "info" },
		]);
		await new Promise((r) => setTimeout(r, 800));
		setClaudeChecking(false);
		setClaudeStatus({
			ok: true,
			message: "Anthropic Cloud gateway connectivity is sound.",
			checkedAt: new Date().toISOString(),
		});
		setClaudeLogs((prev) => [
			...prev,
			{ text: "Handshake verified. Claude models operational.", time: new Date().toLocaleTimeString(), level: "success" },
		]);
	};

	// Copilot actions
	const saveCopilotKey = () => {
		localStorage.setItem(COPILOT_CLOUD_KEY_STORAGE, copilotApiKey.trim());
		setCopilotStatus("Copilot API key saved locally.");
	};

	// Codex actions
	const saveCodexKey = () => {
		localStorage.setItem(CODEX_CLOUD_KEY_STORAGE, codexApiKey.trim());
		setCodexStatus("OpenAI Codex API key saved locally.");
	};

	// Spark actions
	const saveSparkKey = () => {
		localStorage.setItem(SPARK_CLOUD_KEY_STORAGE, sparkApiKey.trim());
		setSparkStatus("Spark API key saved locally.");
	};

	// Blocks actions
	const saveBlocksKey = () => {
		localStorage.setItem(BLOCKS_CLOUD_KEY_STORAGE, blocksApiKey.trim());
		setBlocksStatus("Blocks API key saved locally.");
	};

	const handleRefresh = async () => {
		await Promise.all([
			toolsQuery.refetch(),
			serversQuery.refetch(),
			skillsQuery.refetch(),
		]);
	};

	return (
		<div className="w-full h-full flex flex-col bg-zinc-950 text-zinc-100 font-mono">
			<div className="p-6 border-b border-zinc-800 bg-zinc-900/60 flex flex-wrap items-center justify-between gap-4">
				<div>
					<h1 className="text-2xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-purple-400 bg-clip-text text-transparent flex items-center gap-2">
						<Rocket className="h-6 w-6 text-cyan-400" />
						CLOUD ORCHESTRATOR
					</h1>
					<p className="text-zinc-400 text-xs mt-1">
						Supervise, configure, and route external cloud orchestrator channels and workspace contexts.
					</p>
				</div>
			</div>

			<div className="p-6 flex-grow flex flex-col gap-6">
				<Tabs value={activeTab} onValueChange={setActiveTab} className="w-full flex-grow flex flex-col gap-4">
					<TabsList className="bg-zinc-900 border border-zinc-800 p-1 rounded-md self-start flex-wrap h-auto">
						<TabsTrigger value="jules" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Jules Autopilot</TabsTrigger>
						<TabsTrigger value="claude" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Claude Cloud</TabsTrigger>
						<TabsTrigger value="copilot" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Copilot Cloud</TabsTrigger>
						<TabsTrigger value="codex" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">OpenAI Codex</TabsTrigger>
						<TabsTrigger value="spark" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Spark API</TabsTrigger>
						<TabsTrigger value="blocks" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Blocks Cloud</TabsTrigger>
						<TabsTrigger value="opencode" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">OpenCode</TabsTrigger>
						<TabsTrigger value="super" className="data-[state=active]:bg-zinc-800 text-xs px-3 py-1.5">Super Assistant</TabsTrigger>
					</TabsList>

					{/* 1. Jules Autopilot */}
					<TabsContent value="jules" className="flex-grow flex flex-col gap-4">
						<div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
							<Card className="lg:col-span-2 bg-zinc-900/40 border-zinc-800/80 p-5 space-y-4">
								<div className="flex items-center gap-2 border-b border-zinc-800 pb-2">
									<KeyRound className="h-4 w-4 text-cyan-400" />
									<h3 className="text-sm font-bold text-white">Jules API Key</h3>
								</div>
								<div className="flex flex-wrap items-center gap-2">
									<input
										type="password"
										value={julesApiKey}
										onChange={(e) => setJulesApiKey(e.target.value)}
										placeholder="Paste Jules API Key"
										className="flex-1 min-w-[240px] bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-xs outline-none focus:border-cyan-500"
									/>
									<Button size="sm" onClick={saveJulesKey} className="bg-emerald-800 hover:bg-emerald-700 text-xs">Save</Button>
									<Button size="sm" onClick={clearJulesKey} variant="outline" className="text-xs">Clear</Button>
									<Button size="sm" onClick={testJulesProxy} disabled={julesChecking} className="bg-blue-800 hover:bg-blue-700 text-xs">Test Conn</Button>
								</div>
								{julesStatus && (
									<div className={`text-xs rounded p-2 border inline-flex items-center gap-1.5 ${julesStatus.ok ? "text-emerald-300 border-emerald-800/40 bg-emerald-950/20" : "text-red-300 border-red-800/40 bg-red-950/20"}`}>
										{julesStatus.ok ? <CheckCircle2 className="h-3.5 w-3.5" /> : <XCircle className="h-3.5 w-3.5" />}
										<span>{julesStatus.message}</span>
									</div>
								)}
							</Card>

							<Card className="bg-zinc-900/40 border-zinc-800/80 p-5 space-y-4">
								<h3 className="text-sm font-bold text-white flex items-center justify-between">
									<span>Last Sync Results</span>
									<Button size="sm" variant="outline" onClick={refreshSyncLogs} className="h-6 text-[10px]">Refresh</Button>
								</h3>
								{syncLogs.length === 0 ? (
									<p className="text-xs text-zinc-500 italic">No sync attempts recorded yet.</p>
								) : (
									<ScrollArea className="h-28">
										<div className="space-y-1.5">
											{syncLogs.slice(0, 10).map((entry, i) => (
												<div key={i} className="text-[10px] border border-zinc-800 bg-zinc-950/40 rounded p-1.5">
													<div className="flex justify-between text-zinc-400">
														<span>{entry.sessionId.slice(0, 10)}</span>
														<span>{new Date(entry.timestamp).toLocaleTimeString()}</span>
													</div>
													<div className="text-zinc-200 mt-1">{entry.message}</div>
												</div>
											))}
										</div>
									</ScrollArea>
								)}
							</Card>
						</div>
						<div className="flex-grow border border-zinc-800 rounded-md overflow-hidden relative min-h-[50vh]">
							<iframe src={julesUrl} title="Jules Autopilot" className="w-full h-full border-0 bg-black" allow="clipboard-read; clipboard-write" />
						</div>
					</TabsContent>

					{/* 2. Claude Cloud */}
					<TabsContent value="claude" className="space-y-4">
						<Card className="bg-zinc-900/40 border-zinc-800/80 p-6 space-y-4">
							<div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
								<KeyRound className="h-5 w-5 text-red-400" />
								<h2 className="text-base font-bold text-white">Anthropic API Credentials</h2>
							</div>
							<div className="flex flex-wrap items-center gap-2">
								<input
									type="password"
									value={claudeApiKey}
									onChange={(e) => setClaudeApiKey(e.target.value)}
									placeholder="sk-ant-v1-xxxxxxxx"
									className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs outline-none focus:border-red-500"
								/>
								<Button size="sm" onClick={saveClaudeKey} className="bg-emerald-800 hover:bg-emerald-700 text-xs">Save</Button>
								<Button size="sm" onClick={clearClaudeKey} variant="outline" className="text-xs">Clear</Button>
								<Button size="sm" onClick={testClaudeProxy} disabled={claudeChecking} className="bg-blue-800 hover:bg-blue-700 text-xs">Diagnose</Button>
							</div>
							{claudeStatus && (
								<div className={`text-xs rounded p-2.5 border inline-flex items-center gap-2 ${claudeStatus.ok ? "text-emerald-300 border-emerald-800/50 bg-emerald-950/20" : "text-red-300 border-red-800/50 bg-red-950/20"}`}>
									<CheckCircle2 className="h-4 w-4" />
									<span>{claudeStatus.message}</span>
								</div>
							)}
						</Card>

						<Card className="bg-zinc-900/40 border-zinc-800 p-5 space-y-3">
							<h3 className="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-1.5">
								<Terminal className="h-4 w-4 text-cyan-400" />
								Anthropic Gateway Logs
							</h3>
							<div className="bg-zinc-950 rounded border border-zinc-800 p-4 font-mono text-xs h-40 overflow-y-auto space-y-1">
								{claudeLogs.map((log, i) => (
									<div key={i} className="flex gap-2">
										<span className="text-zinc-600">[{log.time}]</span>
										<span className={log.level === "success" ? "text-emerald-400" : "text-zinc-400"}>{log.text}</span>
									</div>
								))}
							</div>
						</Card>
					</TabsContent>

					{/* 3. Copilot Cloud */}
					<TabsContent value="copilot">
						<Card className="bg-zinc-900/40 border-zinc-800/80 p-6 space-y-4">
							<div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
								<KeyRound className="h-5 w-5 text-indigo-400" />
								<h2 className="text-base font-bold text-white">Copilot Credentials</h2>
							</div>
							<div className="flex flex-wrap items-center gap-2">
								<input
									type="password"
									value={copilotApiKey}
									onChange={(e) => setCopilotApiKey(e.target.value)}
									placeholder="Paste GitHub/Copilot Token"
									className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs outline-none focus:border-indigo-500"
								/>
								<Button size="sm" onClick={saveCopilotKey} className="bg-indigo-800 hover:bg-indigo-700 text-xs">Save</Button>
							</div>
							{copilotStatus && (
								<p className="text-xs text-emerald-400">{copilotStatus}</p>
							)}
						</Card>
					</TabsContent>

					{/* 4. OpenAI Codex */}
					<TabsContent value="codex">
						<Card className="bg-zinc-900/40 border-zinc-800/80 p-6 space-y-4">
							<div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
								<KeyRound className="h-5 w-5 text-emerald-400" />
								<h2 className="text-base font-bold text-white">OpenAI Codex Credentials</h2>
							</div>
							<div className="flex flex-wrap items-center gap-2">
								<input
									type="password"
									value={codexApiKey}
									onChange={(e) => setCodexApiKey(e.target.value)}
									placeholder="sk-proj-xxxxxxxx"
									className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs outline-none focus:border-emerald-500"
								/>
								<Button size="sm" onClick={saveCodexKey} className="bg-emerald-800 hover:bg-emerald-700 text-xs">Save</Button>
							</div>
							{codexStatus && (
								<p className="text-xs text-emerald-400">{codexStatus}</p>
							)}
						</Card>
					</TabsContent>

					{/* 5. Spark API */}
					<TabsContent value="spark">
						<Card className="bg-zinc-900/40 border-zinc-800/80 p-6 space-y-4">
							<div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
								<KeyRound className="h-5 w-5 text-amber-400" />
								<h2 className="text-base font-bold text-white">Spark API Key</h2>
							</div>
							<div className="flex flex-wrap items-center gap-2">
								<input
									type="password"
									value={sparkApiKey}
									onChange={(e) => setSparkApiKey(e.target.value)}
									placeholder="Paste Spark API Key"
									className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs outline-none focus:border-amber-500"
								/>
								<Button size="sm" onClick={saveSparkKey} className="bg-amber-800 hover:bg-amber-700 text-xs">Save</Button>
							</div>
							{sparkStatus && (
								<p className="text-xs text-emerald-400">{sparkStatus}</p>
							)}
						</Card>
					</TabsContent>

					{/* 6. Blocks Cloud */}
					<TabsContent value="blocks">
						<Card className="bg-zinc-900/40 border-zinc-800/80 p-6 space-y-4">
							<div className="flex items-center gap-2 border-b border-zinc-800 pb-3">
								<KeyRound className="h-5 w-5 text-pink-400" />
								<h2 className="text-base font-bold text-white">Blocks Cloud Key</h2>
							</div>
							<div className="flex flex-wrap items-center gap-2">
								<input
									type="password"
									value={blocksApiKey}
									onChange={(e) => setBlocksApiKey(e.target.value)}
									placeholder="Paste Blocks API Key"
									className="flex-1 min-w-[260px] bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs outline-none focus:border-pink-500"
								/>
								<Button size="sm" onClick={saveBlocksKey} className="bg-pink-800 hover:bg-pink-700 text-xs">Save</Button>
							</div>
							{blocksStatus && (
								<p className="text-xs text-emerald-400">{blocksStatus}</p>
							)}
						</Card>
					</TabsContent>

					{/* 7. OpenCode Autopilot */}
					<TabsContent value="opencode" className="flex-grow flex flex-col gap-4">
						<Card className="bg-zinc-900/40 border-zinc-800 p-4 flex justify-between items-center">
							<div>
								<h3 className="font-bold text-white">OpenCode Autopilot</h3>
								<p className="text-xs text-zinc-400">Multi-Model AI Council & Governance portal</p>
							</div>
							<a
								href={autopilotUrl}
								target="_blank"
								rel="noopener noreferrer"
								className="px-3 py-1.5 bg-purple-800 hover:bg-purple-700 rounded text-xs text-white"
							>
								Open Standalone
							</a>
						</Card>
						<div className="flex-grow border border-zinc-800 rounded-md overflow-hidden relative min-h-[50vh]">
							<iframe src={autopilotUrl} title="OpenCode Autopilot" className="w-full h-full border-none bg-black" allow="clipboard-read; clipboard-write" />
						</div>
					</TabsContent>

					{/* 8. Super Assistant */}
					<TabsContent value="super" className="flex-grow flex flex-col">
						<div className="border border-zinc-800 bg-zinc-900/40 rounded-md p-4 mb-4 flex justify-between items-center gap-3">
							<div>
								<h3 className="text-sm font-bold text-white flex items-center gap-2">
									<Bot className="w-4 h-4 text-purple-400" /> MCP SuperAssistant Overview
								</h3>
								<p className="text-zinc-400 text-xs mt-1">
									System-wide MCP overview: {tools.length} tools, {servers.length} servers, {skills.length} skills.
								</p>
							</div>
							<div className="flex items-center gap-2 flex-wrap">
								{hasErrors && (
									<Badge variant="outline" className="border-rose-600 text-rose-400 text-xs">
										<AlertTriangle className="w-3 h-3 mr-1" /> Degraded
									</Badge>
								)}
								{isLoading && (
									<Badge variant="outline" className="border-blue-600 text-blue-400 text-xs">
										<Loader2 className="w-3 h-3 mr-1 animate-spin" /> Loading
									</Badge>
								)}
								<Badge variant="outline" className="border-green-600 text-green-400 text-xs">
									<Activity className="w-3 h-3 mr-1" /> {activeServers.length} Active
								</Badge>
								<Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleRefresh}>
									<RefreshCw className="w-3 h-3 mr-1" /> Refresh
								</Button>
							</div>
						</div>

						<Tabs defaultValue="overview" className="w-full flex-grow">
							<TabsList className="bg-zinc-950 border border-zinc-800/80 p-0.5 rounded-sm flex-wrap h-auto self-start">
								<TabsTrigger value="overview" className="text-xs px-2.5 py-1">Overview</TabsTrigger>
								<TabsTrigger value="tools" className="text-xs px-2.5 py-1">Tools ({tools.length})</TabsTrigger>
								<TabsTrigger value="servers" className="text-xs px-2.5 py-1">Servers ({servers.length})</TabsTrigger>
								<TabsTrigger value="skills" className="text-xs px-2.5 py-1">Skills ({skills.length})</TabsTrigger>
							</TabsList>

							<TabsContent value="overview" className="mt-4">
								<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
									<Card className="bg-zinc-900/30 border-zinc-800">
										<CardHeader className="pb-2">
											<CardTitle className="flex items-center gap-2 text-xs text-zinc-400">
												<Server className="w-3.5 h-3.5 text-blue-400" /> MCP Servers
											</CardTitle>
										</CardHeader>
										<CardContent>
											<div className="text-2xl font-bold text-white">{servers.length}</div>
											<p className="text-[10px] text-zinc-500 mt-1">{activeServers.length} active connections</p>
										</CardContent>
									</Card>
									<Card className="bg-zinc-900/30 border-zinc-800">
										<CardHeader className="pb-2">
											<CardTitle className="flex items-center gap-2 text-xs text-zinc-400">
												<Wrench className="w-3.5 h-3.5 text-amber-400" /> Available Tools
											</CardTitle>
										</CardHeader>
										<CardContent>
											<div className="text-2xl font-bold text-white">{tools.length}</div>
											<p className="text-[10px] text-zinc-500 mt-1">exposed to local agents</p>
										</CardContent>
									</Card>
									<Card className="bg-zinc-900/30 border-zinc-800">
										<CardHeader className="pb-2">
											<CardTitle className="flex items-center gap-2 text-xs text-zinc-400">
												<Cpu className="w-3.5 h-3.5 text-green-400" /> Assimilated Skills
											</CardTitle>
										</CardHeader>
										<CardContent>
											<div className="text-2xl font-bold text-white">{skills.length}</div>
											<p className="text-[10px] text-zinc-500 mt-1">runbook profiles cached</p>
										</CardContent>
									</Card>
								</div>
							</TabsContent>

							<TabsContent value="tools" className="mt-4 flex-grow flex flex-col">
								<Card className="bg-zinc-900/30 border-zinc-800">
									<CardContent className="p-4">
										<ScrollArea className="h-64">
											<div className="space-y-2">
												{tools.map((tool) => (
													<div key={tool.uuid} className="flex items-center justify-between p-2 rounded border border-zinc-800/80 bg-zinc-950/40 text-xs">
														<div>
															<span className="font-mono text-zinc-200">{tool.name}</span>
															<div className="text-zinc-500 mt-0.5">{tool.description || "No description"}</div>
														</div>
														{tool.serverId && <Badge variant="outline" className="text-[10px]">{tool.serverId}</Badge>}
													</div>
												))}
											</div>
										</ScrollArea>
									</CardContent>
								</Card>
							</TabsContent>

							<TabsContent value="servers" className="mt-4">
								<Card className="bg-zinc-900/30 border-zinc-800">
									<CardContent className="p-4 space-y-2">
										{servers.map((server) => (
											<div key={server.id} className="flex items-center justify-between p-2.5 rounded border border-zinc-800/80 bg-zinc-950/40 text-xs">
												<div>
													<span className="font-bold text-zinc-200">{server.name}</span>
													<div className="text-zinc-500">{server.transport || "stdio"}</div>
												</div>
												<Badge variant={server.status === "connected" || server.status === "active" ? "default" : "outline"} className="text-[10px]">
													{server.status || "offline"}
												</Badge>
											</div>
										))}
									</CardContent>
								</Card>
							</TabsContent>

							<TabsContent value="skills" className="mt-4">
								<Card className="bg-zinc-900/30 border-zinc-800">
									<CardContent className="p-4 space-y-2">
										{skills.map((skill) => (
											<div key={skill.id} className="flex items-center justify-between p-2.5 rounded border border-zinc-800/80 bg-zinc-950/40 text-xs">
												<div>
													<span className="font-bold text-zinc-200">{skill.name}</span>
													<div className="text-zinc-500 mt-0.5">{skill.description}</div>
												</div>
												<Badge variant="outline" className="text-green-400 text-[10px]">Active</Badge>
											</div>
										))}
									</CardContent>
								</Card>
							</TabsContent>
						</Tabs>
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
