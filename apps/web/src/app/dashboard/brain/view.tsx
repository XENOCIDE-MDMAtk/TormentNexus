"use client";

import type React from "react";
import { useState, useEffect, useMemo } from "react";
import {
	Card,
	CardHeader,
	CardTitle,
	CardContent,
	Button,
	Input,
	Tabs,
	TabsContent,
	TabsList,
	TabsTrigger,
	Badge,
	ScrollArea,
	KnowledgeGraph,
} from "@tormentnexus/ui";
import {
	Loader2,
	Brain,
	Search,
	Database,
	History,
	Zap,
	Filter,
	Plus,
	Save,
	Download,
	RefreshCw,
	ChevronRight,
	Globe,
	Sparkles,
	Bot,
	FileText,
	Map,
	Settings,
	Cpu,
	Info,
} from "lucide-react";
import { trpc } from "@/utils/trpc";
import { toast } from "sonner";
import {
	filterMemoryRecords,
	getMemoryBadgeLabel,
	getMemoryDetailSections,
	getMemoryModeHint,
	getMemoryPivotSections,
	getMemoryPreview,
	getMemoryProvenance,
	getMemoryRecordKey,
	getMemorySessionId,
	getRelatedMemoryRecords,
	getMemoryTimestamp,
	getMemoryTitle,
	groupMemoryWindowAroundAnchor,
	groupMemoryRecordsByDay,
	MEMORY_MODEL_PILLARS,
	MEMORY_SEARCH_MODES,
	sortMemoryRecordsByTimestamp,
	type MemoryRecord,
	type MemoryPivotAction,
	type RelatedMemoryRecord,
	type MemorySearchMode,
} from "../memory/memory-dashboard-utils";

type ExpertTrpc = {
	expert: {
		research: { useMutation: () => any };
		code: { useMutation: () => any };
	};
};

type MemoryInterchangeFormat =
	| "json"
	| "csv"
	| "jsonl"
	| "json-provider"
	| "tormentnexus-store";

type HydrationReport = {
	startedAt: string;
	completedAt: string;
	totalEntries: number;
	sections: string[];
	projectContextEntries: number;
	architectureEntries: number;
	agentInstructionsEntries: number;
	configEntries: number;
	repoGraphEntries: number;
	environmentEntries: number;
};

type HydrationStatus = {
	totalEntries: number;
	sections: string[];
	sectionCounts: Record<string, number>;
};

const InfoTooltip = ({ text }: { text: string }) => (
	<span
		className="inline-flex items-center justify-center ml-1 text-zinc-500 hover:text-zinc-300 cursor-help transition-colors select-none"
		title={text}
	>
		<Info className="w-3.5 h-3.5" />
	</span>
);

const SECTION_ICONS: Record<
	string,
	React.ComponentType<{ className?: string }>
> = {
	project_context: FileText,
	architecture: Map,
	agent_instructions: Brain,
	configuration: Settings,
	repo_graph: Database,
	environment: Cpu,
};

const SECTION_COLORS: Record<string, string> = {
	project_context: "text-blue-400",
	architecture: "text-purple-400",
	agent_instructions: "text-emerald-400",
	configuration: "text-amber-400",
	repo_graph: "text-cyan-400",
	environment: "text-rose-400",
};

const MEMORY_FORMAT_OPTIONS: Array<{
	value: MemoryInterchangeFormat;
	label: string;
}> = [
	{ value: "json", label: "Canonical JSON" },
	{ value: "csv", label: "Canonical CSV" },
	{ value: "jsonl", label: "Canonical JSONL" },
	{ value: "json-provider", label: "TormentNexus JSON Provider" },
	{ value: "tormentnexus-store", label: "tormentnexus Store" },
];

export default function CognitiveBrainDashboard() {
	const trpcWithExpert = trpc as unknown as typeof trpc & ExpertTrpc;
	const utils = trpc.useUtils();

	// --- Tab Selection ---
	const [activeTab, setActiveTab] = useState("vault");

	// --- SPACED REPETITION STATE ---
	const [dueMemories, setDueMemories] = useState<any[]>([]);
	const [dueLoading, setDueLoading] = useState(false);
	const [currentReviewIndex, setCurrentReviewIndex] = useState(0);
	const [showReviewAnswer, setShowReviewAnswer] = useState(false);

	// --- COGNITIVE GRAPH STATE ---
	const graphQuery = trpc.graph.getSymbolsGraph.useQuery();
	const { nodes: rawNodes = [], links: rawLinks = [] } = graphQuery.data || {};

	const nodes = rawNodes.map((node: any) => ({
		id: node.id,
		label: node.name,
		type: "concept" as const,
		val: node.val || 1,
	}));

	const links = rawLinks.map((link: any) => ({
		source: link.source,
		target: link.target,
		value: 1,
	}));

	// --- URL INGESTION STATE ---
	const [ingestUrl, setIngestUrl] = useState("");
	const [ingestLog, setIngestLog] = useState("");
	const ingestMutation = trpc.knowledge.ingest.useMutation();
	const resourcesQuery = trpc.knowledge.getResources.useQuery();
	const resources = resourcesQuery.data || { categories: [] };

	// --- EXPERT AGENTS STATE ---
	const [researchQuery, setResearchQuery] = useState("");
	const [researchDepth, setResearchDepth] = useState(2);
	const researchMutation = trpcWithExpert.expert.research.useMutation();

	const [coderTask, setCoderTask] = useState("");
	const coderMutation = trpcWithExpert.expert.code.useMutation({
		onSuccess: () => {
			toast.success("Coder task started");
			setCoderTask("");
		},
		onError: (err) => toast.error("Coder task failed: " + err.message),
	});

	// --- MEMORY VAULT STATE ---
	const [searchQuery, setSearchQuery] = useState("");
	const [memoryType, setMemoryType] = useState<
		"session" | "working" | "long_term"
	>("working");
	const [searchMode, setSearchMode] = useState<MemorySearchMode>("all");
	const [newFact, setNewFact] = useState("");
	const [exportFormat, setExportFormat] =
		useState<MemoryInterchangeFormat>("json");
	const [convertToFormat, setConvertToFormat] =
		useState<MemoryInterchangeFormat>("tormentnexus-store");
	const [importing, setImporting] = useState(false);
	const [converting, setConverting] = useState(false);
	const [selectedRecordKey, setSelectedRecordKey] = useState<string | null>(
		null,
	);
	const [activePivot, setActivePivot] = useState<MemoryPivotAction | null>(
		null,
	);
	const [selectedSessionId, setSelectedSessionId] = useState<string | null>(
		null,
	);
	const [consolidating, setConsolidating] = useState(false);

	const handleConsolidate = async () => {
		setConsolidating(true);
		try {
			const res = await fetch("/api/go/api/memory/sleep-cycle", { method: "POST" });
			if (res.ok) {
				toast.success("Memory consolidation & sleep cycle executed successfully!");
				// Invalidate the cache to refresh data
				utils.memory.invalidate();
			} else {
				toast.error("Failed to consolidate memory vault");
			}
		} catch {
			toast.error("Error calling memory consolidation endpoint");
		}
		setConsolidating(false);
	};

	// --- CORE SCRATCHPAD STATE ---
	const [scratchpad, setScratchpad] = useState<Record<string, string>>({});
	const [scratchLoading, setScratchLoading] = useState(false);
	const [editingKey, setEditingKey] = useState<string | null>(null);
	const [editingValue, setEditingValue] = useState("");

	const fetchScratchpad = async () => {
		setScratchLoading(true);
		try {
			const res = await fetch("/api/go/api/memory/scratchpad/get");
			if (res.ok) {
				const d = await res.json();
				setScratchpad(d.scratchpad || {});
			}
		} catch {
			toast.error("Failed to load core memory scratchpad");
		}
		setScratchLoading(false);
	};

	const handleSaveScratchpad = async (key: string, value: string) => {
		try {
			const res = await fetch("/api/go/api/memory/scratchpad/set", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ key, value }),
			});
			if (res.ok) {
				toast.success(`Core memory '${key}' updated!`);
				setScratchpad((prev) => ({ ...prev, [key]: value }));
				setEditingKey(null);
			} else {
				toast.error("Failed to update core memory");
			}
		} catch {
			toast.error("Error saving core memory");
		}
	};

	// --- COMPREHENSIVE SCRAPER STATE ---
	const [resourceScraping, setResourceScraping] = useState(false);
	const [serverScraping, setServerScraping] = useState(false);
	const [scraperLog, setScraperLog] = useState("");

	const handleTriggerResourceScraper = async () => {
		setResourceScraping(true);
		setScraperLog("Initializing prompts & skills assimilation scanner...\n");
		try {
			const res = await fetch("/api/go/api/assimilation/trigger/resources", { method: "POST" });
			const d = await res.json();
			if (res.ok && d.success) {
				toast.success("Skill & prompt resources successfully assimilated!");
				setScraperLog((prev) => prev + d.output);
			} else {
				toast.error("Resource assimilation failed");
				setScraperLog((prev) => prev + `Error: ${d.error || "unknown"}\n${d.output || ""}`);
			}
		} catch (e: any) {
			toast.error("Failed to run resource scraper");
			setScraperLog((prev) => prev + `Exception: ${e.message}\n`);
		}
		setResourceScraping(false);
	};

	const handleTriggerServerScraper = async () => {
		setServerScraping(true);
		setScraperLog("Initializing MCP servers catalog assimilation scanner...\n");
		try {
			const res = await fetch("/api/go/api/assimilation/trigger/servers", { method: "POST" });
			const d = await res.json();
			if (res.ok && d.success) {
				toast.success("MCP servers successfully catalogued and synced!");
				setScraperLog((prev) => prev + d.output);
			} else {
				toast.error("MCP server catalog sync failed");
				setScraperLog((prev) => prev + `Error: ${d.error || "unknown"}\n${d.output || ""}`);
			}
		} catch (e: any) {
			toast.error("Failed to run server scraper");
			setScraperLog((prev) => prev + `Exception: ${e.message}\n`);
		}
		setServerScraping(false);
	};



	const trimmedSearchQuery = searchQuery.trim();
	const hasSearchQuery = trimmedSearchQuery.length > 0;

	const { data: stats } = trpc.memory.getAgentStats.useQuery(undefined, {
		refetchInterval: 10000,
	});

	const recentObservationsQuery = trpc.memory.getRecentObservations.useQuery(
		{
			limit: 6,
			namespace: "project",
		},
		{ refetchInterval: 10000 },
	);

	const recentPromptsQuery = trpc.memory.getRecentUserPrompts.useQuery(
		{
			limit: 5,
		},
		{ refetchInterval: 10000 },
	);

	const recentSessionSummariesQuery =
		trpc.memory.getRecentSessionSummaries.useQuery(
			{
				limit: 4,
			},
			{ refetchInterval: 10000 },
		);

	const genericSearchQuery = trpc.memory.searchAgentMemory.useQuery(
		{
			query: trimmedSearchQuery,
			type: memoryType,
			limit: 20,
		},
		{ enabled: searchMode === "all" || searchMode === "facts" },
	);

	const observationSearchQuery = trpc.memory.searchObservations.useQuery(
		{
			query: trimmedSearchQuery,
			limit: 20,
			namespace: "project",
		},
		{
			enabled:
				(searchMode === "observations" || searchMode === "all") &&
				hasSearchQuery,
		},
	);

	const promptSearchQuery = trpc.memory.searchUserPrompts.useQuery(
		{
			query: trimmedSearchQuery,
			limit: 20,
		},
		{
			enabled:
				(searchMode === "prompts" || searchMode === "all") && hasSearchQuery,
		},
	);

	const sessionSummarySearchQuery = trpc.memory.searchSessionSummaries.useQuery(
		{
			query: trimmedSearchQuery,
			limit: 20,
		},
		{
			enabled:
				(searchMode === "session_summaries" || searchMode === "all") &&
				hasSearchQuery,
		},
	);

	const pivotSearchQuery = trpc.memory.searchMemoryPivot.useQuery(
		{
			pivot: activePivot?.group ?? "session",
			value: activePivot?.query ?? "",
			limit: 20,
		},
		{ enabled: activePivot !== null },
	);

	// --- HYDRATION STATE ---
	const [hydrationStatus, setHydrationStatus] =
		useState<HydrationStatus | null>(null);
	const [hydrationReport, setHydrationReport] =
		useState<HydrationReport | null>(null);
	const [hydrationLoading, setHydrationLoading] = useState(false);
	const [hydrating, setHydrating] = useState(false);
	const [hydrationError, setHydrationError] = useState<string | null>(null);

	const fetchHydrationStatus = async () => {
		setHydrationLoading(true);
		try {
			const endpoints = [
				"/api/go/memory/hydration/status",
				"/api/go/api/memory/hydration/status",
			];
			for (const endpoint of endpoints) {
				try {
					const resp = await fetch(endpoint, {
						signal: AbortSignal.timeout(3000),
					});
					if (resp.ok) {
						const data = await resp.json();
						if (data.success) {
							setHydrationStatus(data.data as HydrationStatus);
							setHydrationError(null);
							return;
						}
					}
				} catch {
					// Ignore fetch errors, will show hydration error below
				}
			}
			setHydrationError("Could not reach hydration store");
		} finally {
			setHydrationLoading(false);
		}
	};

	const handleHydrate = async () => {
		setHydrating(true);
		setHydrationError(null);
		try {
			const endpoints = [
				"/api/go/memory/hydrate",
				"/api/go/api/memory/hydrate",
			];
			for (const endpoint of endpoints) {
				try {
					const resp = await fetch(endpoint, {
						method: "POST",
						signal: AbortSignal.timeout(30000),
					});
					if (resp.ok) {
						const data = await resp.json();
						if (data.success) {
							setHydrationReport(data.data as HydrationReport);
							await fetchHydrationStatus();
							toast.success("Memory hydrated successfully!");
							return;
						}
					}
				} catch {
					// Ignore fetch errors, will show hydration error below
				}
			}
			setHydrationError("Hydration failed — could not reach Go sidecar");
		} finally {
			setHydrating(false);
		}
	};

	const fetchDueMemories = async () => {
		setDueLoading(true);
		try {
			const resp = await fetch("/api/go/api/memory/spaced-repetition/due");
			if (resp.ok) {
				const data = await resp.json();
				if (data.success) {
					setDueMemories(data.due_records || []);
					setCurrentReviewIndex(0);
					setShowReviewAnswer(false);
				}
			}
		} catch (err) {
			console.error("Failed to fetch due memories", err);
		} finally {
			setDueLoading(false);
		}
	};

	const handleGradeMemory = async (memoryID: string, quality: number) => {
		try {
			const resp = await fetch("/api/go/api/memory/spaced-repetition/review", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ memory_id: memoryID, quality }),
			});
			if (resp.ok) {
				const data = await resp.json();
				if (data.success) {
					toast.success(`Memory graded ${quality}/5 successfully!`);
					setShowReviewAnswer(false);
					if (currentReviewIndex < dueMemories.length - 1) {
						setCurrentReviewIndex((prev) => prev + 1);
					} else {
						await fetchDueMemories();
					}
				} else {
					toast.error("Failed to grade memory: " + data.error);
				}
			} else {
				toast.error("Failed to grade memory: server returned error");
			}
		} catch (err: any) {
			toast.error("Failed to grade memory: " + err.message);
		}
	};

	useEffect(() => {
		if (activeTab === "sync") {
			fetchHydrationStatus();
		} else if (activeTab === "spaced-repetition") {
			fetchDueMemories();
		}
	}, [activeTab]);

	const mergeMemoryRecords = (
		groups: Array<MemoryRecord[] | undefined>,
	): MemoryRecord[] => {
		const records: Record<string, MemoryRecord> = {};
		for (const group of groups) {
			for (const memory of group ?? []) {
				records[getMemoryRecordKey(memory)] = memory;
			}
		}
		return Object.values(records);
	};

	const refreshMemoryViews = async () => {
		await Promise.all([
			utils.memory.getAgentStats.invalidate(),
			utils.memory.searchAgentMemory.invalidate(),
			utils.memory.searchObservations.invalidate(),
			utils.memory.searchUserPrompts.invalidate(),
			utils.memory.searchSessionSummaries.invalidate(),
			utils.memory.searchMemoryPivot.invalidate(),
			utils.memory.getMemoryTimelineWindow.invalidate(),
			utils.memory.getCrossSessionMemoryLinks.invalidate(),
			utils.memory.getRecentObservations.invalidate(),
			utils.memory.getRecentUserPrompts.invalidate(),
			utils.memory.getRecentSessionSummaries.invalidate(),
		]);
	};

	const activeResults = useMemo<MemoryRecord[]>(() => {
		if (activePivot) {
			return filterMemoryRecords(
				(pivotSearchQuery.data as MemoryRecord[] | undefined) ?? [],
				searchMode,
			);
		}

		if (searchMode === "all") {
			return mergeMemoryRecords(
				hasSearchQuery
					? [
							genericSearchQuery.data as MemoryRecord[] | undefined,
							observationSearchQuery.data as MemoryRecord[] | undefined,
							promptSearchQuery.data as MemoryRecord[] | undefined,
							sessionSummarySearchQuery.data as MemoryRecord[] | undefined,
						]
					: [
							genericSearchQuery.data as MemoryRecord[] | undefined,
							recentObservationsQuery.data as MemoryRecord[] | undefined,
							recentPromptsQuery.data as MemoryRecord[] | undefined,
							recentSessionSummariesQuery.data as MemoryRecord[] | undefined,
						],
			);
		}

		if (searchMode === "observations") {
			return hasSearchQuery
				? ((observationSearchQuery.data as MemoryRecord[] | undefined) ?? [])
				: ((recentObservationsQuery.data as MemoryRecord[] | undefined) ?? []);
		}

		if (searchMode === "prompts") {
			return hasSearchQuery
				? ((promptSearchQuery.data as MemoryRecord[] | undefined) ?? [])
				: ((recentPromptsQuery.data as MemoryRecord[] | undefined) ?? []);
		}

		if (searchMode === "session_summaries") {
			return hasSearchQuery
				? ((sessionSummarySearchQuery.data as MemoryRecord[] | undefined) ?? [])
				: ((recentSessionSummariesQuery.data as MemoryRecord[] | undefined) ??
						[]);
		}

		return filterMemoryRecords(
			(genericSearchQuery.data as MemoryRecord[] | undefined) ?? [],
			searchMode,
		);
	}, [
		genericSearchQuery.data,
		hasSearchQuery,
		observationSearchQuery.data,
		promptSearchQuery.data,
		recentObservationsQuery.data,
		recentPromptsQuery.data,
		recentSessionSummariesQuery.data,
		activePivot,
		pivotSearchQuery.data,
		searchMode,
		sessionSummarySearchQuery.data,
	]);

	const timelineRecords = useMemo(
		() => sortMemoryRecordsByTimestamp(activeResults),
		[activeResults],
	);
	const timelineGroups = useMemo(
		() => groupMemoryRecordsByDay(timelineRecords),
		[timelineRecords],
	);

	const selectedMemory = useMemo(() => {
		if (!timelineRecords.length) {
			return null;
		}
		return (
			timelineRecords.find(
				(memory) => getMemoryRecordKey(memory) === selectedRecordKey,
			) ?? timelineRecords[0]
		);
	}, [selectedRecordKey, timelineRecords]);

	const timelineWindowQuery = trpc.memory.getMemoryTimelineWindow.useQuery(
		{
			sessionId: selectedSessionId ?? "",
			anchorTimestamp: selectedMemory ? getMemoryTimestamp(selectedMemory) : 0,
			before: 3,
			after: 3,
		},
		{
			enabled: Boolean(selectedMemory && selectedSessionId),
		},
	);

	const crossSessionLinksQuery =
		trpc.memory.getCrossSessionMemoryLinks.useQuery(
			{
				memoryId: selectedMemory?.id ?? "",
				limit: 4,
			},
			{
				enabled: Boolean(selectedMemory?.id),
			},
		);

	const relatedRecords = useMemo(() => {
		if (!selectedMemory) {
			return [];
		}
		return getRelatedMemoryRecords(selectedMemory, timelineRecords);
	}, [selectedMemory, timelineRecords]);

	const pivotSections = useMemo(() => {
		if (!selectedMemory) {
			return [];
		}
		return getMemoryPivotSections(selectedMemory);
	}, [selectedMemory]);

	const sessionWindowRecords = useMemo(() => {
		if (!selectedMemory) {
			return [];
		}
		return (
			(timelineWindowQuery.data as MemoryRecord[] | undefined) ?? []
		).filter(
			(memory) =>
				getMemoryRecordKey(memory) !== getMemoryRecordKey(selectedMemory),
		);
	}, [selectedMemory, timelineWindowQuery.data]);

	const sessionWindowGroups = useMemo(() => {
		if (!selectedMemory) {
			return [];
		}
		return groupMemoryWindowAroundAnchor(selectedMemory, sessionWindowRecords);
	}, [selectedMemory, sessionWindowRecords]);

	const crossSessionLinks = useMemo(() => {
		return (
			(crossSessionLinksQuery.data as RelatedMemoryRecord[] | undefined) ?? []
		);
	}, [crossSessionLinksQuery.data]);

	const handlePivotAction = (action: MemoryPivotAction) => {
		setActivePivot(action);
		setSearchMode(action.mode);
		setSearchQuery(action.query);
		setSelectedRecordKey(null);
	};

	useEffect(() => {
		if (!timelineRecords.length) {
			if (selectedRecordKey !== null) {
				setSelectedRecordKey(null);
			}
			return;
		}

		const hasCurrentSelection = selectedRecordKey
			? timelineRecords.some(
					(memory) => getMemoryRecordKey(memory) === selectedRecordKey,
				)
			: false;

		if (!hasCurrentSelection) {
			setSelectedRecordKey(getMemoryRecordKey(timelineRecords[0]));
		}
	}, [selectedRecordKey, timelineRecords]);

	useEffect(() => {
		setSelectedSessionId(
			selectedMemory ? getMemorySessionId(selectedMemory) : null,
		);
	}, [selectedMemory]);

	useEffect(() => {
		if (activeTab === "core-memory") {
			fetchScratchpad();
		}
	}, [activeTab]);

	const activeLoading = activePivot
		? pivotSearchQuery.isLoading
		: searchMode === "all"
			? hasSearchQuery
				? genericSearchQuery.isLoading ||
					observationSearchQuery.isLoading ||
					promptSearchQuery.isLoading ||
					sessionSummarySearchQuery.isLoading
				: genericSearchQuery.isLoading ||
					recentObservationsQuery.isLoading ||
					recentPromptsQuery.isLoading ||
					recentSessionSummariesQuery.isLoading
			: searchMode === "observations"
				? hasSearchQuery
					? observationSearchQuery.isLoading
					: recentObservationsQuery.isLoading
				: searchMode === "prompts"
					? hasSearchQuery
						? promptSearchQuery.isLoading
						: recentPromptsQuery.isLoading
					: searchMode === "session_summaries"
						? hasSearchQuery
							? sessionSummarySearchQuery.isLoading
							: recentSessionSummariesQuery.isLoading
						: genericSearchQuery.isLoading;

	const activeEmptyMessage = activePivot
		? `No related memory records found for this ${activePivot.group} pivot yet.`
		: hasSearchQuery
			? `No matching ${searchMode === "all" ? "memory records" : searchMode.replace("_", " ")} found right now.`
			: searchMode === "all" || searchMode === "facts"
				? `No matching memories found in the ${memoryType} tier.`
				: `No ${searchMode.replace("_", " ")} have been captured yet.`;

	const addFactMutation = trpc.memory.addFact.useMutation({
		onSuccess: () => {
			toast.success("Fact added to memory");
			setNewFact("");
			void refreshMemoryViews();
		},
		onError: (err) => {
			toast.error(`Failed to add fact: ${err.message}`);
		},
	});

	const handleAddFact = (e: React.FormEvent) => {
		e.preventDefault();
		if (!newFact.trim()) return;
		addFactMutation.mutate({
			content: newFact,
			type: memoryType === "long_term" ? "long_term" : "working",
		});
	};

	const handleIngest = async () => {
		if (!ingestUrl) return;
		setIngestLog(`Ingesting: ${ingestUrl}...`);
		try {
			const result = await ingestMutation.mutateAsync({ url: ingestUrl });
			setIngestLog(`Success: ${result}`);
			setIngestUrl("");
			resourcesQuery.refetch();
		} catch (e: any) {
			setIngestLog(`Error: ${e.message}`);
		}
	};

	const handleResearch = () => {
		if (!researchQuery) return;
		researchMutation.mutate({
			query: researchQuery,
			depth: researchDepth,
			breadth: 3,
		});
	};

	const handleCode = () => {
		if (!coderTask) return;
		coderMutation.mutate({ task: coderTask });
	};

	return (
		<div className="h-full w-full p-6 flex flex-col space-y-6 overflow-hidden bg-black text-slate-100">
			{/* Header */}
			<header className="flex flex-wrap justify-between items-center border-b border-zinc-800 pb-4 shrink-0 gap-4">
				<div className="flex flex-col gap-1">
					<div className="flex items-center gap-4">
						<h1 className="text-3xl font-bold bg-gradient-to-r from-purple-500 via-pink-500 to-blue-500 bg-clip-text text-transparent flex items-center gap-2">
							<Brain className="h-8 w-8 text-pink-500 animate-pulse" />
							TormentNexus Cognitive Hub
						</h1>
						<button
							onClick={handleConsolidate}
							disabled={consolidating}
							className="px-3 py-1 bg-emerald-700/80 hover:bg-emerald-600 rounded text-xs font-semibold text-white flex items-center gap-1.5 transition-colors disabled:opacity-50"
							title="Run SM-2 Spaced Repetition consolidation, forgetting curve decay, and reflection loops manually"
						>
							{consolidating ? (
								<Loader2 className="w-3.5 h-3.5 animate-spin" />
							) : (
								<Sparkles className="w-3.5 h-3.5 text-emerald-300" />
							)}
							{consolidating ? "Consolidating..." : "Consolidate Vault"}
						</button>
					</div>
					<p className="text-zinc-400 text-sm">
						Unified control plane for Vector Memory, Knowledge Graphs, Web Ingestion, and Expert Systems.
					</p>
				</div>

				{/* Stats panel in header */}
				<div className="flex flex-wrap gap-2 text-xs">
					<div className="flex items-center" title="Transient cache of recent chat sessions and temporary data">
						<StatCard
							label="Session"
							value={(stats as any)?.sessionCount || 0}
							icon={<History className="h-3.5 w-3.5 text-zinc-400" />}
						/>
					</div>
					<div className="flex items-center" title="Active short-term workspace memory containing high-frequency variables">
						<StatCard
							label="Working"
							value={(stats as any)?.workingCount || 0}
							icon={<Zap className="h-3.5 w-3.5 text-yellow-500" />}
						/>
					</div>
					<div className="flex items-center" title="Persistent vector database memory that survives across runs">
						<StatCard
							label="Long Term"
							value={(stats as any)?.longTermCount || 0}
							icon={<Database className="h-3.5 w-3.5 text-pink-500" />}
						/>
					</div>
					<div className="flex items-center" title="Accumulated cognitive observations made during execution analysis">
						<StatCard
							label="Observations"
							value={(stats as any)?.observationCount || 0}
							icon={<RefreshCw className="h-3.5 w-3.5 text-green-500" />}
						/>
					</div>
				</div>
			</header>

			{/* Main Tabs System */}
			<Tabs
				value={activeTab}
				onValueChange={setActiveTab}
				className="w-full flex-1 flex flex-col min-h-0"
			>
				<TabsList className="flex flex-wrap gap-1 mb-4 bg-zinc-900 border border-zinc-800 p-1 rounded-lg shrink-0 w-fit">
					<TabsTrigger
						value="vault"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Memory Vault
					</TabsTrigger>
					<TabsTrigger
						value="core-memory"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all border border-emerald-500/20 text-emerald-400"
					>
						Core Memory (Letta)
					</TabsTrigger>
					<TabsTrigger
						value="spaced-repetition"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all border border-pink-500/20 text-pink-400"
					>
						Spaced Repetition
					</TabsTrigger>
					<TabsTrigger
						value="graph"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Cognitive Graph
					</TabsTrigger>
					<TabsTrigger
						value="ingest"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Web Ingestion
					</TabsTrigger>
					<TabsTrigger
						value="agents"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Expert Agents
					</TabsTrigger>
					<TabsTrigger
						value="observations"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Observations Log
					</TabsTrigger>
					<TabsTrigger
						value="sync"
						className="text-sm font-medium px-4 py-1.5 rounded-md transition-all"
					>
						Data Sync & Hydration
					</TabsTrigger>
				</TabsList>

				{/* CORE MEMORY SCRATCHPAD TAB */}
				<TabsContent
					value="core-memory"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="space-y-6 flex-1 overflow-y-auto">
						<Card className="bg-zinc-900/60 border-zinc-850">
							<CardHeader>
								<CardTitle className="text-lg font-semibold text-emerald-400 flex items-center gap-2">
									<Brain className="h-5 w-5" />
									Letta Core Memory Block (Active Ephemeral Instructions)
								</CardTitle>
								<p className="text-zinc-500 text-xs mt-1">
									Directly view and edit the active scratchpad blocks. Pinned instruction values are embedded inside the agent's short-term system instructions.
								</p>
							</CardHeader>
							<CardContent>
								{scratchLoading ? (
									<div className="flex justify-center py-12">
										<Loader2 className="h-8 w-8 animate-spin text-emerald-500" />
									</div>
								) : (
									<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
										{Object.entries(scratchpad).map(([key, val]) => (
											<div key={key} className="border border-zinc-800 bg-zinc-950 p-4 rounded-lg space-y-3">
												<div className="flex items-center justify-between border-b border-zinc-900 pb-2">
													<span className="font-mono text-sm font-semibold text-emerald-400 capitalize flex items-center">
														{key} Block
														<InfoTooltip text={key === "persona" ? "Defines the agent's style, format constraints, response demeanor, and core behavior directives." : "Contains information about the operator user, workspace preferences, environment parameters, and active task parameters."} />
													</span>
													{editingKey !== key ? (
														<Button
															size="sm"
															variant="outline"
															className="text-xs bg-zinc-900 hover:bg-zinc-800 border-zinc-800 text-zinc-300"
															onClick={() => {
																setEditingKey(key);
																setEditingValue(val);
															}}
														>
															Edit Key Block
														</Button>
													) : (
														<div className="flex gap-2">
															<Button
																size="sm"
																variant="outline"
																className="text-xs border-zinc-800 text-zinc-500"
																onClick={() => setEditingKey(null)}
															>
																Cancel
															</Button>
															<Button
																size="sm"
																className="text-xs bg-emerald-700 hover:bg-emerald-600 text-white"
																onClick={() => handleSaveScratchpad(key, editingValue)}
															>
																Save Block
															</Button>
														</div>
													)}
												</div>

												{editingKey === key ? (
													<textarea
														value={editingValue}
														onChange={(e) => setEditingValue(e.target.value)}
														className="w-full h-48 bg-zinc-950 border border-zinc-800 rounded p-2 text-xs font-mono text-zinc-300 focus:outline-none focus:border-zinc-700"
													/>
												) : (
													<pre className="text-xs font-mono text-zinc-400 bg-zinc-950 p-3 rounded max-h-48 overflow-y-auto whitespace-pre-wrap">
														{val || "No content configured."}
													</pre>
												)}
											</div>
										))}
									</div>
								)}
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				{/* MEMORY VAULT TAB */}
				<TabsContent
					value="vault"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="grid grid-cols-1 lg:grid-cols-4 gap-6 flex-1 min-h-0">
						{/* Sidebar: Filters & Manual Input */}
						<div className="lg:col-span-1 space-y-4 overflow-y-auto pr-1">
							<Card className="bg-zinc-900/60 border-zinc-850">
								<CardHeader className="pb-2">
									<CardTitle className="text-xs font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
										<Filter className="h-3.5 w-3.5" />
										Memory Filters
									</CardTitle>
								</CardHeader>
								<CardContent className="space-y-3">
									<div className="space-y-1">
										<label className="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">
											Storage Tier
										</label>
										<div className="flex flex-col gap-1">
											<TierButton
												active={memoryType === "session"}
												onClick={() => setMemoryType("session")}
												label="Session"
												description="Transient cache"
											/>
											<TierButton
												active={memoryType === "working"}
												onClick={() => setMemoryType("working")}
												label="Working"
												description="Active context"
											/>
											<TierButton
												active={memoryType === "long_term"}
												onClick={() => setMemoryType("long_term")}
												label="Long Term"
												description="Persistent database"
											/>
										</div>
										<p className="text-[11px] leading-relaxed text-zinc-500 mt-2">
											{getMemoryModeHint(searchMode, memoryType)}
										</p>
									</div>
								</CardContent>
							</Card>

							<Card className="bg-zinc-900/60 border-zinc-850">
								<CardHeader className="pb-2">
									<CardTitle className="text-xs font-bold text-zinc-400 uppercase tracking-widest flex items-center gap-2">
										<Plus className="h-3.5 w-3.5" />
										Manual Injection
									</CardTitle>
								</CardHeader>
								<CardContent>
									<form onSubmit={handleAddFact} className="space-y-2">
										<textarea
											value={newFact}
											onChange={(e) => setNewFact(e.target.value)}
											className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white h-20 focus:ring-1 focus:ring-pink-500 outline-none resize-none"
											placeholder="Manually seed a fact in current tier..."
										/>
										<Button
											type="submit"
											size="sm"
											disabled={addFactMutation.isPending || !newFact.trim()}
											className="w-full bg-pink-600 hover:bg-pink-500 text-white text-xs"
										>
											{addFactMutation.isPending ? (
												<Loader2 className="h-3.5 w-3.5 animate-spin" />
											) : (
												<Save className="h-3.5 w-3.5 mr-1.5" />
											)}
											Store Fact
										</Button>
									</form>
								</CardContent>
							</Card>

							<Card className="bg-zinc-900/60 border-zinc-850 border-l-2 border-l-amber-500">
								<CardContent className="p-3 text-[11px] text-zinc-400 space-y-2">
									<h4 className="font-bold text-white uppercase text-[10px] tracking-wider">
										Reinforcement Mechanics
									</h4>
									<p className="leading-relaxed">
										Memory relevance decays dynamically using biomimetic curves.
										Successive correct retrievals reinforce importance.
									</p>
								</CardContent>
							</Card>
						</div>

						{/* Main Explorer View */}
						<Card className="lg:col-span-3 bg-zinc-900/40 border-zinc-850 flex flex-col overflow-hidden">
							<CardHeader className="border-b border-zinc-800 bg-zinc-950/20 pb-3">
								<div className="space-y-3">
									<div className="flex flex-wrap gap-1.5">
										{MEMORY_SEARCH_MODES.map((mode) => (
											<button
												key={mode.value}
												type="button"
												onClick={() => setSearchMode(mode.value)}
												className={`rounded-full border px-3 py-1 text-[11px] font-semibold transition-colors ${
													searchMode === mode.value
														? "border-pink-500/60 bg-pink-500/10 text-pink-200"
														: "border-zinc-800 bg-zinc-950 text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200"
												}`}
												title={mode.description}
											>
												{mode.label}
											</button>
										))}
									</div>
									<div className="relative">
										<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-zinc-500" />
										<input
											type="text"
											value={searchQuery}
											onChange={(e) => {
												const nextValue = e.target.value;
												setSearchQuery(nextValue);
												if (
													activePivot &&
													nextValue.trim() !== activePivot.query
												) {
													setActivePivot(null);
												}
											}}
											placeholder={`Search ${searchMode === "all" ? `${memoryType} records` : searchMode.replace("_", " ")}...`}
											className="w-full bg-zinc-950 border border-zinc-800 rounded-lg pl-9 pr-4 py-2 text-xs text-white focus:ring-1 focus:ring-pink-500 outline-none transition-all"
										/>
									</div>
								</div>
							</CardHeader>
							<CardContent className="flex-1 p-0 overflow-hidden flex flex-col">
								<ScrollArea className="flex-1">
									{activeLoading ? (
										<div className="p-12 flex flex-col items-center justify-center text-zinc-500 gap-2">
											<Loader2 className="h-6 w-6 animate-spin" />
											<p className="text-xs font-mono uppercase tracking-widest">
												Searching synapses...
											</p>
										</div>
									) : !activeResults || activeResults.length === 0 ? (
										<div className="p-16 text-center text-zinc-500">
											<Brain className="h-10 w-10 mx-auto mb-3 opacity-20" />
											<p className="text-sm font-medium">Empty vault</p>
											<p className="text-xs mt-1">{activeEmptyMessage}</p>
										</div>
									) : (
										<div className="grid min-h-full lg:grid-cols-[minmax(0,1.2fr)_minmax(300px,0.8fr)]">
											<div className="border-r border-zinc-800">
												{timelineGroups.map((group) => (
													<div
														key={group.key}
														className="border-b border-zinc-850 last:border-b-0"
													>
														<div className="sticky top-0 z-10 border-b border-zinc-850 bg-zinc-950/90 px-3 py-1.5 text-[9px] font-bold uppercase tracking-wider text-zinc-400 backdrop-blur">
															{group.label}
														</div>
														<div className="divide-y divide-zinc-900">
															{group.items.map((memory) => {
																const recordKey = getMemoryRecordKey(memory);
																const isSelected =
																	recordKey === selectedRecordKey;

																return (
																	<button
																		key={recordKey}
																		type="button"
																		onClick={() =>
																			setSelectedRecordKey(recordKey)
																		}
																		className={`w-full text-left p-3.5 transition-colors flex flex-col gap-1.5 ${isSelected ? "bg-zinc-800/40" : "hover:bg-zinc-800/20"}`}
																	>
																		<div className="flex items-center justify-between gap-2">
																			<Badge
																				variant="outline"
																				className={`text-[9px] px-1.5 py-0.5 border-zinc-700 text-zinc-300 font-mono`}
																			>
																				{getMemoryBadgeLabel(memory)}
																			</Badge>
																			<span className="text-[10px] text-zinc-500 font-mono">
																				{new Date(
																					getMemoryTimestamp(memory),
																				).toLocaleTimeString()}
																			</span>
																		</div>
																		<p className="text-xs font-semibold text-white truncate">
																			{getMemoryTitle(memory)}
																		</p>
																		<p className="text-[11px] text-zinc-400 line-clamp-2 leading-relaxed">
																			{getMemoryPreview(memory)}
																		</p>
																	</button>
																);
															})}
														</div>
													</div>
												))}
											</div>

											{/* Details sidebar pane */}
											<div className="bg-zinc-950/30 flex flex-col">
												{selectedMemory ? (
													<div className="flex flex-col h-full">
														<div className="border-b border-zinc-850 p-4">
															<div className="flex items-center justify-between mb-2">
																<Badge
																	variant="outline"
																	className="border-zinc-700 text-[10px] text-zinc-300"
																>
																	{getMemoryBadgeLabel(selectedMemory)}
																</Badge>
																<span className="text-[10px] text-zinc-500">
																	{new Date(
																		getMemoryTimestamp(selectedMemory),
																	).toLocaleString()}
																</span>
															</div>
															<h2 className="text-sm font-bold text-white">
																{getMemoryTitle(selectedMemory)}
															</h2>
															<p className="mt-2 text-xs leading-relaxed text-zinc-300 whitespace-pre-wrap">
																{getMemoryPreview(selectedMemory)}
															</p>
														</div>

														<div className="flex-1 space-y-4 p-4 overflow-y-auto">
															{pivotSections.map((section) => (
																<div
																	key={section.title}
																	className="rounded-lg border border-zinc-800 bg-zinc-955/60 p-3"
																>
																	<h3 className="text-[10px] font-bold uppercase tracking-wider text-zinc-500 mb-2">
																		{section.title}
																	</h3>
																	<div className="flex flex-wrap gap-1.5">
																		{section.actions.map((action) => (
																			<button
																				key={action.key}
																				onClick={() =>
																					handlePivotAction(action)
																				}
																				className="rounded border border-zinc-700 bg-black/10 px-2 py-1 text-left text-[11px] text-zinc-300 hover:bg-zinc-800 transition-all font-mono"
																			>
																				{action.label}
																			</button>
																		))}
																	</div>
																</div>
															))}

															{getMemoryDetailSections(selectedMemory).map(
																(section) => (
																	<div
																		key={section.title}
																		className="rounded-lg border border-zinc-800 bg-zinc-955/60 p-3"
																	>
																		<h3 className="text-[10px] font-bold uppercase tracking-wider text-zinc-500 mb-1.5">
																			{section.title}
																		</h3>
																		{section.body && (
																			<p className="text-xs text-zinc-300 leading-relaxed">
																				{section.body}
																			</p>
																		)}
																		{section.items && (
																			<ul className="space-y-1 mt-1 text-xs">
																				{section.items.map((item, idx) => (
																					<li
																						key={idx}
																						className="bg-black/20 p-1.5 font-mono text-[10px] text-zinc-400 rounded"
																					>
																						{item}
																					</li>
																				))}
																			</ul>
																		)}
																	</div>
																),
															)}
														</div>
													</div>
												) : (
													<div className="p-8 text-center text-zinc-600 text-xs">
														Select a record to view detailed metrics.
													</div>
												)}
											</div>
										</div>
									)}
								</ScrollArea>
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				{/* SPACED REPETITION TAB */}
				<TabsContent
					value="spaced-repetition"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="flex-1 flex flex-col items-center justify-center max-w-3xl mx-auto w-full p-4">
						{dueLoading ? (
							<div className="flex flex-col items-center gap-3">
								<Loader2 className="h-8 w-8 animate-spin text-pink-500" />
								<p className="text-sm font-mono text-zinc-400">Loading synapses due for review...</p>
							</div>
						) : dueMemories.length === 0 ? (
							<Card className="bg-zinc-900 border border-zinc-800 p-8 text-center max-w-md w-full shadow-xl">
								<Sparkles className="h-12 w-12 text-pink-500 mx-auto mb-4 animate-bounce" />
								<h3 className="text-xl font-bold text-white mb-2">All Caught Up!</h3>
								<p className="text-sm text-zinc-400">
									No memories are due for spaced repetition review right now.
									Check back later or ingest more facts!
								</p>
							</Card>
						) : (
							<div className="w-full space-y-6">
								{/* Progress Header */}
								<div className="flex justify-between items-center text-xs text-zinc-400">
									<span>Reviewing memory {currentReviewIndex + 1} of {dueMemories.length}</span>
									<span className="bg-pink-500/10 text-pink-300 border border-pink-500/20 px-2.5 py-0.5 rounded-full font-mono font-bold">
										{dueMemories.length} due
									</span>
								</div>

								{/* Flashcard */}
								<Card className="bg-zinc-900/80 border border-zinc-800 p-8 shadow-2xl rounded-2xl flex flex-col gap-6 relative min-h-[250px] justify-between backdrop-blur-md">
									<div className="space-y-4">
										<div className="flex justify-between items-center">
											<Badge variant="outline" className="text-[10px] uppercase font-mono px-2 py-0.5 border-zinc-700 text-zinc-300">
												{dueMemories[currentReviewIndex]?.memory_kind || dueMemories[currentReviewIndex]?.memory_type || "memory"}
											</Badge>
											{dueMemories[currentReviewIndex]?.category && (
												<span className="text-[10px] text-zinc-500 font-bold uppercase tracking-wider">
													Category: {dueMemories[currentReviewIndex]?.category}
												</span>
											)}
										</div>

										<p className="text-base text-white font-medium leading-relaxed whitespace-pre-wrap font-mono p-4 bg-black/40 border border-zinc-850 rounded-lg">
											{dueMemories[currentReviewIndex]?.content}
										</p>
									</div>

									{/* Reveal / Grading Options */}
									<div className="pt-6 border-t border-zinc-850 flex flex-col items-center gap-4">
										{!showReviewAnswer ? (
											<Button
												onClick={() => setShowReviewAnswer(true)}
												className="bg-gradient-to-r from-pink-600 to-purple-600 hover:from-pink-500 hover:to-purple-500 text-white font-semibold px-8 py-2 rounded-lg shadow-lg shadow-pink-900/20 transition-all text-xs"
											>
												Reveal Details &amp; Grade
											</Button>
										) : (
											<div className="w-full space-y-4">
												{/* Additional info revealed */}
												<div className="grid grid-cols-2 gap-4 text-xs bg-black/40 p-3 rounded-lg border border-zinc-800 font-mono text-zinc-400">
													<div>
														<span className="text-zinc-500">Importance:</span> {(dueMemories[currentReviewIndex]?.importance * 100).toFixed(0)}%
													</div>
													<div>
														<span className="text-zinc-500">Heat Score:</span> {dueMemories[currentReviewIndex]?.heat_score?.toFixed(1) || "100.0"}
													</div>
													{dueMemories[currentReviewIndex]?.tags && (
														<div className="col-span-2">
															<span className="text-zinc-500">Tags:</span> {dueMemories[currentReviewIndex]?.tags}
														</div>
													)}
												</div>

												{/* Grade Buttons */}
												<div className="space-y-2">
													<p className="text-[11px] font-bold text-center text-zinc-500 uppercase tracking-widest">How well did you recall this memory?</p>
													<div className="grid grid-cols-6 gap-2">
														{[
															{ val: 0, label: "Forgot", color: "bg-red-900/50 hover:bg-red-800 text-red-200 border-red-800/40" },
															{ val: 1, label: "Wrong", color: "bg-orange-950 hover:bg-orange-900 text-orange-200 border-orange-900/40" },
															{ val: 2, label: "Hard", color: "bg-amber-950 hover:bg-amber-900 text-amber-200 border-amber-900/40" },
															{ val: 3, label: "Good", color: "bg-yellow-950/80 hover:bg-yellow-900 text-yellow-200 border-yellow-800/40" },
															{ val: 4, label: "Easy", color: "bg-emerald-950 hover:bg-emerald-900 text-emerald-200 border-emerald-900/40" },
															{ val: 5, label: "Perfect", color: "bg-green-900/50 hover:bg-green-800 text-green-200 border-green-800/40" },
														].map((grade) => (
															<button
																key={grade.val}
																onClick={() => handleGradeMemory(dueMemories[currentReviewIndex].id, grade.val)}
																className={`border rounded-lg p-2 flex flex-col items-center transition-all ${grade.color}`}
															>
																<span className="text-base font-bold">{grade.val}</span>
																<span className="text-[9px] font-medium tracking-tight mt-0.5">{grade.label}</span>
															</button>
														))}
													</div>
												</div>
											</div>
										)}
									</div>
								</Card>
							</div>
						)}
					</div>
				</TabsContent>

				{/* COGNITIVE GRAPH TAB */}
				<TabsContent
					value="graph"
					className="flex-1 flex flex-col min-h-0 outline-none relative"
				>
					<div className="absolute inset-0 bg-zinc-950 border border-zinc-850 rounded-2xl overflow-hidden shadow-inner flex flex-col">
						<div className="p-3 bg-zinc-900/50 border-b border-zinc-800 flex justify-between items-center shrink-0 z-10">
							<span className="text-xs text-zinc-400">
								Interactive visual map of all resolved concepts, databases,
								files, and their relational connections.
							</span>
							<Button
								size="sm"
								variant="outline"
								onClick={() => graphQuery.refetch()}
								className="border-zinc-700 hover:bg-zinc-800 h-7 text-xs"
							>
								<RefreshCw className="h-3 w-3 mr-1" /> Refetch Graph
							</Button>
						</div>
						<div className="flex-1 min-h-0 relative">
							<KnowledgeGraph
								nodes={nodes}
								links={links}
								loading={graphQuery.isLoading}
							/>
						</div>
					</div>
				</TabsContent>

				{/* WEB INGESTION TAB */}
				<TabsContent
					value="ingest"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0 overflow-y-auto">
						<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
							<CardHeader className="p-0">
								<CardTitle className="text-lg font-bold text-green-400 flex items-center gap-2">
									<Globe className="h-5 w-5" /> Ingest Knowledge Source
								</CardTitle>
							</CardHeader>
							<CardContent className="p-0 flex flex-col gap-4">
								<p className="text-sm text-zinc-400">
									Parse documentation, API references, or articles directly into
									the TormentNexus context engine.
								</p>
								<div className="flex gap-2">
									<Input
										type="text"
										className="flex-1 bg-black border-zinc-800 text-white outline-none placeholder:text-zinc-650"
										placeholder="Enter URL (e.g. https://mcp.dev/docs)"
										value={ingestUrl}
										onChange={(e) => setIngestUrl(e.target.value)}
										onKeyDown={(e) => e.key === "Enter" && handleIngest()}
									/>
									<Button
										onClick={handleIngest}
										disabled={ingestMutation.isPending}
										className="bg-green-600 hover:bg-green-500 text-white font-semibold"
									>
										{ingestMutation.isPending ? "Ingesting..." : "Ingest"}
									</Button>
								</div>
								{ingestLog && (
									<div className="bg-black p-3 rounded-lg text-xs font-mono text-zinc-300 break-all border border-zinc-800 max-h-[250px] overflow-y-auto leading-relaxed">
										{ingestLog}
									</div>
								)}
							</CardContent>
						</Card>

						<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
							<CardHeader className="p-0">
								<CardTitle className="text-lg font-bold text-violet-400 flex items-center gap-2">
									<RefreshCw className="h-5 w-5 text-violet-500 animate-pulse" /> Comprehensive Library Scrapers
								</CardTitle>
							</CardHeader>
							<CardContent className="p-0 flex flex-col gap-4">
								<p className="text-sm text-zinc-400">
									Scan known libraries, local registers, and remote indexes to discover, categorise, and ingest MCP server schemas, skill scripts, and prompt templates.
								</p>
								<div className="grid grid-cols-2 gap-4">
									<Button
										onClick={handleTriggerResourceScraper}
										disabled={resourceScraping}
										className="bg-violet-700 hover:bg-violet-600 text-white font-semibold flex items-center justify-center gap-1.5 transition-colors"
									>
										{resourceScraping && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
										{resourceScraping ? "Scraping..." : "Scrape Prompts & Skills"}
									</Button>
									<Button
										onClick={handleTriggerServerScraper}
										disabled={serverScraping}
										className="bg-indigo-700 hover:bg-indigo-600 text-white font-semibold flex items-center justify-center gap-1.5 transition-colors"
									>
										{serverScraping && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
										{serverScraping ? "Scraping..." : "Scrape MCP Servers"}
									</Button>
								</div>
								{scraperLog && (
									<pre className="bg-black p-3 rounded-lg text-xs font-mono text-zinc-400 break-all border border-zinc-800 max-h-[250px] overflow-y-auto leading-relaxed whitespace-pre-wrap">
										{scraperLog}
									</pre>
								)}
							</CardContent>
						</Card>

						<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
							<CardHeader className="p-0 flex justify-between flex-row items-center">
								<CardTitle className="text-lg font-bold text-zinc-300">
									Ingested Resource Index
								</CardTitle>
								<span className="text-xs text-zinc-500">
									Last Sync:{" "}
									{resources.lastUpdated
										? new Date(resources.lastUpdated).toLocaleTimeString()
										: "Never"}
								</span>
							</CardHeader>
							<CardContent className="p-0 flex-1 overflow-y-auto">
								{resourcesQuery.isLoading ? (
									<div className="flex justify-center p-8">
										<Loader2 className="animate-spin text-zinc-500" />
									</div>
								) : resources.categories?.length === 0 ? (
									<p className="text-zinc-500 italic text-center p-8">
										No external doc sources ingested yet.
									</p>
								) : (
									<div className="space-y-4">
										{resources.categories?.map((cat: any) => (
											<div
												key={cat.name}
												className="border-l-2 border-zinc-800 pl-4 py-1"
											>
												<h4 className="font-semibold text-zinc-200 text-sm">
													{cat.name}
												</h4>
												<ul className="mt-1 space-y-1 text-xs text-zinc-400">
													{cat.items?.map((item: any) => (
														<li key={item.url} className="truncate">
															<a
																href={item.url}
																target="_blank"
																rel="noopener noreferrer"
																className="hover:text-blue-400 hover:underline"
															>
																{item.title || item.url}
															</a>
														</li>
													))}
												</ul>
											</div>
										))}
									</div>
								)}
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				{/* EXPERT AGENTS TAB */}
				<TabsContent
					value="agents"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0 overflow-y-auto">
						<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
							<CardHeader className="p-0">
								<CardTitle className="text-lg font-bold text-blue-400 flex items-center gap-2">
									<Bot className="h-5 w-5" /> Deep Research Agent
								</CardTitle>
							</CardHeader>
							<CardContent className="p-0 flex flex-col gap-4 flex-1">
								<div className="flex gap-2">
									<Input
										type="text"
										className="flex-1 bg-black border-zinc-800 text-white outline-none placeholder:text-zinc-650"
										placeholder="Research query..."
										value={researchQuery}
										onChange={(e) => setResearchQuery(e.target.value)}
										onKeyDown={(e) => e.key === "Enter" && handleResearch()}
									/>
									<div className="flex items-center gap-2 bg-black border border-zinc-800 rounded-lg px-3">
										<span className="text-[10px] text-zinc-500 font-semibold tracking-wider">
											DEPTH
										</span>
										<input
											type="number"
											min="1"
											max="5"
											aria-label="Research depth"
											value={researchDepth}
											onChange={(e) =>
												setResearchDepth(parseInt(e.target.value))
											}
											className="bg-transparent w-8 text-center outline-none text-white font-semibold text-xs"
										/>
									</div>
									<Button
										onClick={handleResearch}
										disabled={researchMutation.isPending}
										className="bg-blue-600 hover:bg-blue-500 text-white font-semibold"
									>
										Research
									</Button>
								</div>
								<div className="flex-1 overflow-y-auto">
									{researchMutation.isPending && (
										<div className="flex justify-center p-8">
											<Loader2 className="animate-spin text-zinc-500" />
										</div>
									)}
									{researchMutation.data && (
										<div className="bg-black p-4 border border-zinc-800 rounded-lg text-sm text-zinc-300 whitespace-pre-wrap leading-relaxed">
											{researchMutation.data.summary}
										</div>
									)}
									{researchMutation.error && (
										<div className="p-4 bg-red-955/20 border border-red-800 text-red-300 rounded-lg text-xs font-mono">
											Error: {researchMutation.error.message}
										</div>
									)}
								</div>
							</CardContent>
						</Card>

						<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
							<CardHeader className="p-0">
								<CardTitle className="text-lg font-bold text-purple-400 flex items-center gap-2">
									<Sparkles className="h-5 w-5" /> Coder Agent
								</CardTitle>
							</CardHeader>
							<CardContent className="p-0 flex flex-col gap-4 flex-1">
								<div className="flex gap-2">
									<Input
										type="text"
										className="flex-1 bg-black border-zinc-800 text-white outline-none placeholder:text-zinc-650"
										placeholder="Coding task (e.g. 'Write a test file for helper/date.ts')..."
										value={coderTask}
										onChange={(e) => setCoderTask(e.target.value)}
										onKeyDown={(e) => e.key === "Enter" && handleCode()}
									/>
									<Button
										onClick={handleCode}
										disabled={coderMutation.isPending}
										className="bg-purple-600 hover:bg-purple-500 text-white font-semibold"
									>
										Code
									</Button>
								</div>
								<div className="flex-1 overflow-y-auto">
									{coderMutation.isPending && (
										<div className="flex justify-center p-8">
											<Loader2 className="animate-spin text-zinc-500" />
										</div>
									)}
									{coderMutation.data && (
										<div className="bg-black p-4 border border-zinc-800 rounded-lg text-sm text-zinc-300">
											<h4 className="text-green-400 font-bold mb-2">
												Task Complete!
											</h4>
											<p className="font-mono text-xs text-zinc-400">
												Files Changed:{" "}
												{coderMutation.data.filesChanged?.join(", ") || "None"}
											</p>
											<p className="mt-2 leading-relaxed text-zinc-300">
												{coderMutation.data.reasoning}
											</p>
										</div>
									)}
									{coderMutation.error && (
										<div className="p-4 bg-red-955/20 border border-red-800 text-red-300 rounded-lg text-xs font-mono">
											Error: {coderMutation.error.message}
										</div>
									)}
								</div>
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				{/* OBSERVATIONS LOG TAB */}
				<TabsContent
					value="observations"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0 overflow-y-auto">
						{/* Runtime Observations */}
						<Card className="bg-zinc-900 border-zinc-800 p-5 flex flex-col h-[550px]">
							<CardHeader className="p-0 pb-3 flex flex-row items-center justify-between border-b border-zinc-800">
								<CardTitle className="text-sm font-bold text-zinc-200 uppercase tracking-widest flex items-center gap-2">
									<RefreshCw className="h-4 w-4 text-emerald-500" />
									Runtime Observations
								</CardTitle>
								<Button
									size="sm"
									variant="ghost"
									onClick={() => recentObservationsQuery.refetch()}
									className="h-6 w-6 p-0 text-zinc-500 hover:text-zinc-200"
								>
									<RefreshCw className="h-3.5 w-3.5" />
								</Button>
							</CardHeader>
							<CardContent className="p-0 pt-3 flex-1 overflow-y-auto space-y-3">
								{recentObservationsQuery.isLoading ? (
									<div className="flex justify-center p-8">
										<Loader2 className="animate-spin text-zinc-500" />
									</div>
								) : !(recentObservationsQuery.data ?? []).length ? (
									<p className="text-zinc-550 italic text-center p-8 text-xs">
										No runtime observations captured yet.
									</p>
								) : (
									(recentObservationsQuery.data as any[]).map(
										(memory, index) => {
											const obs = memory.metadata?.structuredObservation;
											return (
												<div
													key={memory.id ?? index}
													className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3 text-xs"
												>
													<div className="mb-1.5 flex items-center justify-between gap-2">
														<Badge
															variant="outline"
															className="border-emerald-500/30 text-emerald-300 text-[9px] px-1 py-0"
														>
															{obs?.type ?? "observation"}
														</Badge>
														{obs?.toolName && (
															<span className="font-mono text-[9px] text-zinc-500">
																{obs.toolName}
															</span>
														)}
													</div>
													<p className="font-bold text-white text-[13px]">
														{obs?.title ?? "Untitled Observation"}
													</p>
													<p className="mt-1 text-zinc-400 leading-relaxed text-[11px] whitespace-pre-wrap">
														{obs?.narrative ?? memory.content}
													</p>
												</div>
											);
										},
									)
								)}
							</CardContent>
						</Card>

						{/* Session Summaries */}
						<Card className="bg-zinc-900 border-zinc-800 p-5 flex flex-col h-[550px]">
							<CardHeader className="p-0 pb-3 flex flex-row items-center justify-between border-b border-zinc-800">
								<CardTitle className="text-sm font-bold text-zinc-200 uppercase tracking-widest flex items-center gap-2">
									<History className="h-4 w-4 text-sky-500" />
									Session Summaries
								</CardTitle>
								<Button
									size="sm"
									variant="ghost"
									onClick={() => recentSessionSummariesQuery.refetch()}
									className="h-6 w-6 p-0 text-zinc-500 hover:text-zinc-200"
								>
									<RefreshCw className="h-3.5 w-3.5" />
								</Button>
							</CardHeader>
							<CardContent className="p-0 pt-3 flex-1 overflow-y-auto space-y-3">
								{recentSessionSummariesQuery.isLoading ? (
									<div className="flex justify-center p-8">
										<Loader2 className="animate-spin text-zinc-500" />
									</div>
								) : !(recentSessionSummariesQuery.data ?? []).length ? (
									<p className="text-zinc-550 italic text-center p-8 text-xs">
										No session summaries recorded yet.
									</p>
								) : (
									(recentSessionSummariesQuery.data as any[]).map(
										(memory, index) => {
											const summary = memory.metadata?.structuredSessionSummary;
											return (
												<div
													key={memory.id ?? index}
													className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3 text-xs"
												>
													<div className="mb-1.5 flex items-center justify-between gap-2">
														<Badge
															variant="outline"
															className="border-sky-500/30 text-sky-300 text-[9px] px-1 py-0"
														>
															{summary?.status ?? "summary"}
														</Badge>
														{summary?.cliType && (
															<span className="font-mono text-[9px] text-zinc-500">
																{summary.cliType}
															</span>
														)}
													</div>
													<p className="font-bold text-white text-[13px]">
														{summary?.name ??
															summary?.sessionId ??
															"Unnamed Session"}
													</p>
													<p className="mt-1 text-zinc-400 leading-relaxed text-[11px] whitespace-pre-wrap">
														{summary?.activeGoal ??
															summary?.lastObjective ??
															memory.content}
													</p>
												</div>
											);
										},
									)
								)}
							</CardContent>
						</Card>

						{/* Prompts & Directives */}
						<Card className="bg-zinc-900 border-zinc-800 p-5 flex flex-col h-[550px]">
							<CardHeader className="p-0 pb-3 flex flex-row items-center justify-between border-b border-zinc-800">
								<CardTitle className="text-sm font-bold text-zinc-200 uppercase tracking-widest flex items-center gap-2">
									<Brain className="h-4 w-4 text-violet-500" />
									Captured Prompts
								</CardTitle>
								<Button
									size="sm"
									variant="ghost"
									onClick={() => recentPromptsQuery.refetch()}
									className="h-6 w-6 p-0 text-zinc-500 hover:text-zinc-200"
								>
									<RefreshCw className="h-3.5 w-3.5" />
								</Button>
							</CardHeader>
							<CardContent className="p-0 pt-3 flex-1 overflow-y-auto space-y-3">
								{recentPromptsQuery.isLoading ? (
									<div className="flex justify-center p-8">
										<Loader2 className="animate-spin text-zinc-500" />
									</div>
								) : !(recentPromptsQuery.data ?? []).length ? (
									<p className="text-zinc-550 italic text-center p-8 text-xs">
										No captured prompts registered yet.
									</p>
								) : (
									(recentPromptsQuery.data as any[]).map((memory, index) => {
										const prompt = memory.metadata?.structuredUserPrompt;
										return (
											<div
												key={memory.id ?? index}
												className="rounded-lg border border-zinc-800 bg-zinc-950/70 p-3 text-xs"
											>
												<div className="mb-1.5 flex items-center justify-between gap-2">
													<Badge
														variant="outline"
														className="border-violet-500/30 text-violet-300 text-[9px] px-1 py-0"
													>
														{prompt?.role ?? "prompt"}
													</Badge>
													<span className="text-[9px] text-zinc-500">
														#{prompt?.promptNumber ?? "?"}
													</span>
												</div>
												<p className="text-zinc-300 font-mono text-[10px] leading-relaxed break-words whitespace-pre-wrap">
													{prompt?.content ?? memory.content}
												</p>
											</div>
										);
									})
								)}
							</CardContent>
						</Card>
					</div>
				</TabsContent>

				{/* DATA SYNC & HYDRATION TAB */}
				<TabsContent
					value="sync"
					className="flex-1 flex flex-col min-h-0 outline-none"
				>
					<div className="grid grid-cols-1 lg:grid-cols-2 gap-6 flex-1 min-h-0 overflow-y-auto">
						{/* Left Side: Backup & Convert */}
						<div className="space-y-6">
							<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
								<CardHeader className="p-0 border-b border-zinc-800 pb-3">
									<CardTitle className="text-lg font-bold text-cyan-400 flex items-center gap-2">
										<Download className="h-5 w-5" /> Memory Export &amp; Backup
									</CardTitle>
								</CardHeader>
								<CardContent className="p-0 flex flex-col gap-4 pt-2">
									<p className="text-sm text-zinc-400">
										Export stored RAG memory records from the active local
										database to other diagnostic schema formats.
									</p>
									<div className="space-y-2">
										<label className="text-xs font-bold text-zinc-400 uppercase">
											Export Format
										</label>
										<select
											value={exportFormat}
											onChange={(e) =>
												setExportFormat(
													e.target.value as MemoryInterchangeFormat,
												)
											}
											className="w-full bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white outline-none"
										>
											{MEMORY_FORMAT_OPTIONS.map((option) => (
												<option key={option.value} value={option.value}>
													{option.label}
												</option>
											))}
										</select>
									</div>
									<Button
										className="bg-cyan-700 hover:bg-cyan-600 text-white font-bold h-10"
										onClick={async () => {
											try {
												const res = await fetch(
													`/api/trpc/memory.exportMemories?input=${encodeURIComponent(JSON.stringify({ userId: "default", format: exportFormat }))}`,
												);
												const json = await res.json();
												const content = json?.result?.data?.data || "";
												const extension =
													exportFormat === "csv"
														? "csv"
														: exportFormat === "jsonl"
															? "jsonl"
															: "json";
												const blob = new Blob([content], {
													type:
														extension === "csv"
															? "text/csv"
															: "application/json",
												});
												const url = URL.createObjectURL(blob);
												const a = document.createElement("a");
												a.href = url;
												a.download = `tormentnexus-memories.${extension}`;
												a.click();
												URL.revokeObjectURL(url);
												toast.success(
													`Exported as ${MEMORY_FORMAT_OPTIONS.find((option) => option.value === exportFormat)?.label || exportFormat}`,
												);
											} catch (err: any) {
												toast.error(`Export failed: ${err.message}`);
											}
										}}
									>
										<Download className="h-4 w-4 mr-2" />
										Download Synapses
									</Button>
								</CardContent>
							</Card>

							<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
								<CardHeader className="p-0 border-b border-zinc-800 pb-3">
									<CardTitle className="text-lg font-bold text-purple-400 flex items-center gap-2">
										<Globe className="h-5 w-5" /> Import &amp; Convert Tiers
									</CardTitle>
								</CardHeader>
								<CardContent className="p-0 flex flex-col gap-4 pt-2">
									<p className="text-sm text-zinc-400">
										Restore memories from a JSON/CSV/JSONL backup file, or
										convert formats on the fly.
									</p>

									<div className="space-y-2">
										<label className="text-xs font-bold text-zinc-400 uppercase block">
											Import File
										</label>
										<input
											type="file"
											accept=".json,.csv,.jsonl"
											className="w-full text-xs text-zinc-400 file:mr-2 file:py-1 file:px-2.5 file:rounded file:border-0 file:text-xs file:bg-zinc-800 file:text-zinc-300 hover:file:bg-zinc-700"
											onChange={async (e) => {
												const file = e.target.files?.[0];
												if (!file) return;
												setImporting(true);
												try {
													const text = await file.text();
													const res = await fetch(
														"/api/trpc/memory.importMemories",
														{
															method: "POST",
															headers: { "Content-Type": "application/json" },
															body: JSON.stringify({
																userId: "default",
																format: exportFormat,
																data: text,
															}),
														},
													);
													const result = await res.json();
													toast.success(
														`Imported ${result?.result?.data?.imported || 0} memories`,
													);
													await refreshMemoryViews();
												} catch (err: any) {
													toast.error(`Import failed: ${err.message}`);
												} finally {
													setImporting(false);
												}
											}}
										/>
										{importing && (
											<div className="flex items-center gap-2 mt-2 text-xs text-cyan-400">
												<Loader2 className="h-3.5 w-3.5 animate-spin" />
												Ingesting backup...
											</div>
										)}
									</div>

									<div className="border-t border-zinc-800 pt-3 space-y-2">
										<label className="text-xs font-bold text-zinc-400 uppercase block">
											Convert Format
										</label>
										<div className="flex gap-2">
											<select
												value={convertToFormat}
												onChange={(e) =>
													setConvertToFormat(
														e.target.value as MemoryInterchangeFormat,
													)
												}
												className="flex-1 bg-zinc-950 border border-zinc-800 rounded-md p-2 text-xs text-white outline-none"
											>
												{MEMORY_FORMAT_OPTIONS.filter(
													(option) => option.value !== exportFormat,
												).map((option) => (
													<option key={option.value} value={option.value}>
														{option.label}
													</option>
												))}
											</select>
											<Button
												disabled={
													converting || convertToFormat === exportFormat
												}
												className="bg-zinc-800 hover:bg-zinc-750 text-cyan-300 font-bold border border-zinc-750 px-4"
												onClick={async () => {
													setConverting(true);
													try {
														const exportResponse = await fetch(
															`/api/trpc/memory.exportMemories?input=${encodeURIComponent(JSON.stringify({ userId: "default", format: exportFormat }))}`,
														);
														const exportJson = await exportResponse.json();
														const sourceData =
															exportJson?.result?.data?.data || "";

														const convertResponse = await fetch(
															"/api/trpc/memory.convertMemories",
															{
																method: "POST",
																headers: { "Content-Type": "application/json" },
																body: JSON.stringify({
																	userId: "default",
																	fromFormat: exportFormat,
																	toFormat: convertToFormat,
																	data: sourceData,
																}),
															},
														);
														const convertJson = await convertResponse.json();
														const convertedData =
															convertJson?.result?.data?.data || "";
														const extension =
															convertToFormat === "csv"
																? "csv"
																: convertToFormat === "jsonl"
																	? "jsonl"
																	: "json";
														const blob = new Blob([convertedData], {
															type:
																extension === "csv"
																	? "text/csv"
																	: "application/json",
														});
														const url = URL.createObjectURL(blob);
														const a = document.createElement("a");
														a.href = url;
														a.download = `tormentnexus-memory-converted.${extension}`;
														a.click();
														URL.revokeObjectURL(url);
														toast.success(
															`Converted ${exportFormat} → ${convertToFormat}`,
														);
													} catch (err: any) {
														toast.error(`Conversion failed: ${err.message}`);
													} finally {
														setConverting(false);
													}
												}}
											>
												{converting ? (
													<Loader2 className="h-4 w-4 animate-spin" />
												) : (
													<RefreshCw className="h-4 w-4 mr-2" />
												)}
												Convert
											</Button>
										</div>
									</div>
								</CardContent>
							</Card>
						</div>

						{/* Right Side: Hydration Control */}
						<div className="space-y-6">
							<Card className="bg-zinc-900 border-zinc-800 p-6 flex flex-col gap-4">
								<CardHeader className="p-0 border-b border-zinc-800 pb-3 flex flex-row items-center justify-between">
									<CardTitle className="text-lg font-bold text-emerald-400 flex items-center gap-2">
										<Zap className="h-5 w-5" /> Memory Hydration
									</CardTitle>
									<div className="flex gap-2">
										<Button
											variant="outline"
											size="sm"
											className="border-zinc-700 text-zinc-300 hover:bg-zinc-800 h-8"
											onClick={fetchHydrationStatus}
											disabled={hydrationLoading}
										>
											<RefreshCw
												className={`h-3.5 w-3.5 ${hydrationLoading ? "animate-spin" : ""}`}
											/>
										</Button>
										<Button
											size="sm"
											className="bg-emerald-600 hover:bg-emerald-500 text-white font-bold h-8"
											onClick={handleHydrate}
											disabled={hydrating}
										>
											{hydrating ? (
												<Loader2 className="h-3.5 w-3.5 animate-spin" />
											) : (
												<>
													<Zap className="h-3.5 w-3.5 mr-1" />
													Hydrate
												</>
											)}
										</Button>
									</div>
								</CardHeader>
								<CardContent className="p-0 pt-2 space-y-4">
									<p className="text-sm text-zinc-400">
										Bootstrap the Go sidecar context store with essential
										project knowledge for autonomous operation.
									</p>

									{/* Hydration Report */}
									{hydrationReport && (
										<div className="rounded-lg border border-emerald-500/20 bg-emerald-950/10 p-4 space-y-2">
											<div className="flex items-center gap-2 text-emerald-300 text-xs font-semibold">
												<Zap className="h-3.5 w-3.5" /> Hydration Successful
											</div>
											<div className="grid grid-cols-3 gap-2 text-center text-xs">
												{[
													{
														label: "Project",
														value: hydrationReport.projectContextEntries,
													},
													{
														label: "Architecture",
														value: hydrationReport.architectureEntries,
													},
													{
														label: "Instructions",
														value: hydrationReport.agentInstructionsEntries,
													},
													{
														label: "Config",
														value: hydrationReport.configEntries,
													},
													{
														label: "Repo Graph",
														value: hydrationReport.repoGraphEntries,
													},
													{
														label: "Env",
														value: hydrationReport.environmentEntries,
													},
												].map((item) => (
													<div
														key={item.label}
														className="rounded border border-zinc-800 bg-zinc-950 p-2"
													>
														<div className="text-[9px] uppercase text-zinc-500">
															{item.label}
														</div>
														<div className="mt-0.5 text-base font-semibold text-white">
															{item.value}
														</div>
													</div>
												))}
											</div>
										</div>
									)}

									{hydrationError && (
										<div className="p-3 bg-red-950/20 border border-red-800/30 text-red-300 text-xs rounded">
											{hydrationError}
										</div>
									)}

									{/* Status Info */}
									<div className="grid grid-cols-2 gap-4">
										<div className="rounded-lg border border-zinc-800 bg-zinc-950/40 p-4">
											<div className="text-[10px] uppercase tracking-wider text-zinc-500 font-bold">
												Total Hydrated
											</div>
											<div className="mt-1 text-2xl font-bold text-white">
												{hydrationStatus?.totalEntries ?? 0}
											</div>
										</div>
										<div className="rounded-lg border border-zinc-800 bg-zinc-950/40 p-4">
											<div className="text-[10px] uppercase tracking-wider text-zinc-500 font-bold">
												Store Status
											</div>
											<div className="mt-1 text-lg font-bold text-white">
												{(hydrationStatus?.totalEntries ?? 0) > 0
													? "Active (Hydrated)"
													: "Empty"}
											</div>
										</div>
									</div>

									{/* Section Counts */}
									{hydrationStatus?.sectionCounts && (
										<div className="space-y-1.5">
											<div className="text-xs font-bold text-zinc-500 uppercase tracking-wider">
												Breakdown by Section
											</div>
											<div className="space-y-1">
												{Object.entries(hydrationStatus.sectionCounts).map(
													([section, count]) => {
														const Icon = SECTION_ICONS[section] || Database;
														const color =
															SECTION_COLORS[section] || "text-zinc-400";
														return (
															<div
																key={section}
																className="flex items-center justify-between text-xs bg-zinc-950/40 border border-zinc-850 rounded px-3 py-1.5"
															>
																<div className="flex items-center gap-2 text-zinc-300">
																	<Icon className={`h-3.5 w-3.5 ${color}`} />
																	<span className="capitalize">
																		{section.replace("_", " ")}
																	</span>
																</div>
																<span className="font-mono text-zinc-400 font-bold">
																	{count}
																</span>
															</div>
														);
													},
												)}
											</div>
										</div>
									)}
								</CardContent>
							</Card>
						</div>
					</div>
				</TabsContent>
			</Tabs>
		</div>
	);
}

function StatCard({
	label,
	value,
	icon,
}: {
	label: string;
	value: number;
	icon: React.ReactNode;
}) {
	return (
		<div className="bg-zinc-900 border border-zinc-800 rounded-lg px-3 py-1 flex items-center gap-2">
			{icon}
			<div className="flex flex-col items-start leading-tight">
				<span className="text-[9px] text-zinc-500 uppercase font-bold tracking-wider">
					{label}
				</span>
				<span className="text-sm font-bold text-white tabular-nums">
					{value}
				</span>
			</div>
		</div>
	);
}

function TierButton({
	active,
	onClick,
	label,
	description,
}: {
	active: boolean;
	onClick: () => void;
	label: string;
	description: string;
}) {
	return (
		<button
			onClick={onClick}
			className={`w-full text-left p-2.5 rounded-lg border transition-all ${
				active
					? "bg-pink-500/10 border-pink-500/50 text-pink-400 shadow-[0_0_15px_rgba(236,72,153,0.05)]"
					: "bg-transparent border-transparent text-zinc-500 hover:bg-white/5"
			}`}
		>
			<div className="text-xs font-bold">{label}</div>
			<div
				className={`text-[10px] mt-0.5 ${active ? "text-pink-400/60" : "text-zinc-600"}`}
			>
				{description}
			</div>
		</button>
	);
}
