"use client";

import { LimboPanel } from "./limbo-panel";
import { useState, useCallback } from "react";
import {
	Search,
	Database,
	Archive,
	Trash2,
	RotateCcw,
	Loader2,
} from "lucide-react";

interface MemoryRecord {
	id: string;
	session_id: string;
	memory_type: string;
	memory_kind: string;
	category: string;
	content: string;
	importance: number;
	heat_score: number;
}

interface FTSResult {
	record: MemoryRecord;
	score: number;
	tier: string;
}

const PAGE_SIZE = 20;

export default function MemorySearchPage() {
	const [query, setQuery] = useState("");
	const [offset, setOffset] = useState(0);
	const [total, setTotal] = useState(0);
	const [loading, setLoading] = useState(false);
	const [results, setResults] = useState<FTSResult[]>([]);
	const [coldCount, setColdCount] = useState<number | null>(null);
	const [limboStats, setLimboStats] = useState<Record<string, number>>({});
	const [error, setError] = useState("");

	const search = useCallback(
		async (newOffset = 0) => {
			if (!query.trim()) return;
			setLoading(true);
			setError("");
			setOffset(newOffset);
			try {
				const fts = await fetch(
					`/api/go/api/memory/fts-search?q=${encodeURIComponent(query)}&limit=${PAGE_SIZE}&offset=${newOffset}`,
				);
				if (!fts.ok) throw new Error(`FTS: ${fts.status}`);
				const ftsData = await fts.json();
				setResults(ftsData.data ?? []);
				setTotal(ftsData.total ?? 0);
			} catch (e) {
				setError(String(e));
			}
			setLoading(false);
		},
		[query],
	);

	const refreshStats = useCallback(async () => {
		try {
			const cold = await fetch("/api/go/api/memory/cold-archive");
			const coldData = await cold.json();
			setColdCount(coldData.count ?? 0);

			const limbo = await fetch("/api/go/api/memory/limbo/search");
			const limboData = await limbo.json();
			setLimboStats(limboData.stats ?? {});
		} catch {
			// Stats are best-effort, ignore failures
		}
	}, []);

	const buryMemory = async (id: string) => {
		try {
			await fetch("/api/go/api/memory/limbo/bury", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ id, reason: "discarded" }),
			});
			setResults((prev) => prev.filter((r) => r.record.id !== id));
			refreshStats();
		} catch {
			// Best-effort, UI stays consistent
		}
	};

	return (
		<div className="p-6 space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-2xl font-bold flex items-center gap-2">
						<Database className="w-6 h-6" />
						Memory Explorer
					</h1>
					<p className="text-zinc-400 text-sm mt-1">
						Full-text search across{" "}
						{coldCount !== null
							? `${coldCount + results.length + Object.values(limboStats).reduce((a, b) => a + b, 0)}+`
							: ""}{" "}
						L2 memories
					</p>
				</div>
				<button
					onClick={refreshStats}
					className="text-xs text-zinc-500 hover:text-zinc-300"
					title="Refresh cold archive and limbo statistics from the server"
				>
					Refresh stats
				</button>
			</div>

			{/* Stats bar */}
			<div className="flex gap-4 text-sm">
				<div className="px-3 py-1.5 bg-zinc-900 rounded-lg border border-zinc-800">
					<span className="text-zinc-500">Cold archive: </span>
					<span className="text-blue-400">{coldCount ?? "?"}</span>
				</div>
				{Object.entries(limboStats).map(([reason, count]) => (
					<div
						key={reason}
						className="px-3 py-1.5 bg-zinc-900 rounded-lg border border-zinc-800"
					>
						<span className="text-zinc-500">{reason}: </span>
						<span className="text-red-400">{count}</span>
					</div>
				))}
			</div>

			{/* Search bar */}
			<div className="flex gap-2">
				<div className="relative flex-1">
					<Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
					<input
						type="text"
						placeholder="Search memories... (FTS5 BM25 full-text search across all L2 vault records)"
						value={query}
						onChange={(e) => setQuery(e.target.value)}
						onKeyDown={(e) => e.key === "Enter" && search(offset)}
						className="w-full pl-10 pr-4 py-2 bg-zinc-900 border border-zinc-700 rounded-lg text-sm focus:outline-none focus:border-zinc-500"
					/>
				</div>
				<button
					onClick={() => search(0)}
					disabled={loading}
					className="px-4 py-2 bg-zinc-800 rounded-lg hover:bg-zinc-700 text-sm disabled:opacity-50"
					title="Execute BM25 full-text search across 86K+ indexed memories. Supports AND/OR, phrases, and prefix* matching."
				>
					{loading ? <Loader2 className="w-4 h-4 animate-spin" /> : "Search"}
				</button>
			</div>

			{error && <div className="text-red-400 text-sm">{error}</div>}

			{/* Results */}
			{results.length > 0 && (
				<p
					className="text-xs text-zinc-600"
					title={`Showing results ${offset + 1}-${offset + results.length}`}
				>
					Page {Math.floor(offset / PAGE_SIZE) + 1} ({offset + 1}–{offset + results.length}{total > 0 ? ` / ${total}` : ""})
				</p>
			)}

			<div className="space-y-3">
				{results.map(({ record }) => (
					<div
						key={record.id}
						className="border border-zinc-800 rounded-lg p-4 bg-zinc-900/50"
					>
						<div className="flex items-start justify-between gap-3">
							<div className="min-w-0 flex-1">
								<div className="text-xs text-zinc-500 font-mono mb-1 truncate">
									{record.id}
								</div>
								<div className="text-sm text-zinc-300 line-clamp-3 font-mono text-xs leading-relaxed">
									{record.content.slice(0, 500)}
								</div>
								<div className="flex gap-3 mt-2 text-xs text-zinc-600">
									<span>Heat: {record.heat_score?.toFixed(0)}</span>
									<span>Importance: {record.importance?.toFixed(2)}</span>
									<span>Kind: {record.memory_kind}</span>
									<span>Type: {record.memory_type}</span>
								</div>
							</div>
							<button
								onClick={() => buryMemory(record.id)}
								className="shrink-0 p-2 rounded-lg hover:bg-red-950/30 text-zinc-600 hover:text-red-400 transition-colors"
								title="Bury this memory in the L4 Limbo vault (discarded). It can be resurrected later from the limbo search."
							>
								<Trash2 className="w-4 h-4" />
							</button>
						</div>
					</div>
				))}
				{!loading && query && results.length === 0 && (
					<div className="text-center py-12 text-zinc-500">
						<Search className="w-8 h-8 mx-auto mb-3 opacity-30" />
						<p className="font-medium">No results found</p>
						<p className="text-xs text-zinc-600 mt-1">
							Try different keywords. FTS5 BM25 search supports AND/OR, phrases
							(&quot;in quotes&quot;), and prefix* matching.
						</p>
					</div>
				)}
				{/* Pagination */}
				{results.length > 0 && (
					<div className="flex justify-center gap-4 pt-4">
						<button
							onClick={() => search(Math.max(0, offset - PAGE_SIZE))}
							disabled={offset === 0 || loading}
							className="px-4 py-2 bg-zinc-800 rounded-lg hover:bg-zinc-700 text-sm disabled:opacity-30 disabled:cursor-not-allowed"
							title="Go to previous page of results"
						>
							Previous
						</button>
						<span className="px-4 py-2 text-sm text-zinc-500">
							Page {Math.floor(offset / PAGE_SIZE) + 1}
						</span>
						<button
							onClick={() => search(offset + PAGE_SIZE)}
							disabled={results.length < PAGE_SIZE || loading}
							className="px-4 py-2 bg-zinc-800 rounded-lg hover:bg-zinc-700 text-sm disabled:opacity-30 disabled:cursor-not-allowed"
							title="Go to next page of results"
						>
							Next
						</button>
					</div>
				)}

				{!loading && !query && results.length === 0 && (
					<div className="text-center py-16 text-zinc-600">
						<Database className="w-12 h-12 mx-auto mb-4 opacity-20" />
						<p className="font-medium">Explore your memory vault</p>
						<p className="text-xs text-zinc-600 mt-2 max-w-md mx-auto">
							Search across 86,000+ indexed L2 memories using BM25 full-text
							search. Results include heat scores, importance rankings, and
							memory kind classifications. Use the trash icon to bury irrelevant
							memories in the L4 Limbo vault.
						</p>
					</div>
				)}
			</div>

			{/* L4 Limbo Vault Section */}
			<details className="border border-zinc-800 rounded-lg p-4 bg-zinc-900/30">
				<summary className="cursor-pointer text-sm font-medium text-zinc-400 hover:text-zinc-200 select-none" title="The L4 Limbo vault stores memories that were discarded, lost, or decayed. Memories can be resurrected back to the L2 vault.">
					L4 Limbo Vault
				</summary>
				<div className="mt-4">
					<LimboPanel />
				</div>
			</details>
		</div>
	);
}
