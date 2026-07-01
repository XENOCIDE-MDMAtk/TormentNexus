"use client";

import { useState, useEffect, useCallback } from "react";
import {
	FileUp,
	Search,
	RotateCcw,
	Database,
	Loader2,
	CheckCircle,
	XCircle,
	ChevronDown,
	ChevronRight,
} from "lucide-react";

interface ImportedSession {
	id: string;
	sourceTool: string;
	sourceType: string;
	sourcePath: string;
	format: string;
	lastModifiedAt: string;
	estimatedSize: number;
	valid: boolean;
	errors?: string[];
	detectedModels?: string[];
	imported?: boolean;
	importedAt?: string;
}

export default function SessionImportPage() {
	const [sessions, setSessions] = useState<ImportedSession[]>([]);
	const [loading, setLoading] = useState(false);
	const [scanning, setScanning] = useState(false);
	const [selectedSession, setSelectedSession] =
		useState<ImportedSession | null>(null);
	const [lastScan, setLastScan] = useState<string | null>(null);
	const [stats, setStats] = useState<{
		total: number;
		valid: number;
		imported: number;
	} | null>(null);

	const fetchSessions = useCallback(async () => {
		setLoading(true);
		try {
			const res = await fetch("/api/go/api/sessions/imported/list?limit=200");
			const d = await res.json();
			const data = d.data ?? [];
			setSessions(data);
			const total = data.length;
			const valid = data.filter((s: ImportedSession) => s.valid).length;
			const imported = data.filter((s: ImportedSession) => s.imported).length;
			setStats({ total, valid, imported });
		} catch {
			// Best-effort session fetch
		}
		setLoading(false);
	}, []);

	const triggerScan = async () => {
		setScanning(true);
		try {
			await fetch("/api/go/api/sessions/imported/scan", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ force: true }),
			});
			setLastScan(new Date().toLocaleTimeString());
			await fetchSessions();
		} catch {
			// Best-effort scan
		}
		setScanning(false);
	};

	useEffect(() => {
		fetchSessions();
	}, [fetchSessions]);

	return (
		<div className="space-y-6">
			{/* Header */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<FileUp className="w-5 h-5 text-orange-400" />
					<div>
						<h1 className="text-lg font-semibold text-white">Session Import</h1>
						<p className="text-xs text-zinc-500 mt-0.5">
							Scan and import sessions from external tools into the memory vault
						</p>
					</div>
				</div>
				<div className="flex items-center gap-2">
					<button
						onClick={fetchSessions}
						disabled={loading}
						className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50"
						title="Refresh the session list"
					>
						{loading ? (
							<Loader2 className="w-3 h-3 animate-spin" />
						) : (
							<RotateCcw className="w-3 h-3" />
						)}
					</button>
					<button
						onClick={triggerScan}
						disabled={scanning}
						className="px-3 py-1.5 bg-orange-900/40 text-orange-300 border border-orange-800/50 rounded hover:bg-orange-900/60 text-xs disabled:opacity-50 flex items-center gap-1.5"
						title="Scan workspace directories for importable session files and transcripts"
					>
						{scanning ? (
							<Loader2 className="w-3 h-3 animate-spin" />
						) : (
							<Search className="w-3 h-3" />
						)}
						{scanning ? "Scanning..." : "Scan for Sessions"}
					</button>
				</div>
			</div>

			{/* Stats */}
			{stats && (
				<div className="flex gap-3 text-xs">
					<div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800">
						<span className="text-zinc-500">Total</span>
						<span className="ml-2 text-white font-medium">{stats.total}</span>
					</div>
					<div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800">
						<span className="text-zinc-500">Valid</span>
						<span className="ml-2 text-emerald-400 font-medium">
							{stats.valid}
						</span>
					</div>
					<div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800">
						<span className="text-zinc-500">Imported</span>
						<span className="ml-2 text-blue-400 font-medium">
							{stats.imported}
						</span>
					</div>
					{lastScan && (
						<div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800">
							<span className="text-zinc-500">Last scan</span>
							<span className="ml-2 text-zinc-300">{lastScan}</span>
						</div>
					)}
				</div>
			)}

			{/* Sessions List */}
			<div className="space-y-2">
				{sessions.length === 0 && !loading && (
					<div className="text-center py-16 text-zinc-600">
						<Database className="w-12 h-12 mx-auto mb-4 opacity-20" />
						<p className="font-medium">No sessions found</p>
						<p className="text-xs mt-2 max-w-md mx-auto">
							Click &quot;Scan for Sessions&quot; to search workspace
							directories for importable transcripts, conversation exports, and
							session logs from supported tools.
						</p>
					</div>
				)}
				{sessions.map((session) => (
					<div
						key={session.id}
						className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-3 hover:bg-zinc-900 transition-colors cursor-pointer"
						onClick={() =>
							setSelectedSession(
								selectedSession?.id === session.id ? null : session,
							)
						}
						title="Click to expand session details"
					>
						<div className="flex items-center justify-between">
							<div className="flex items-center gap-2 min-w-0">
								{selectedSession?.id === session.id ? (
									<ChevronDown className="w-3 h-3 shrink-0 text-zinc-500" />
								) : (
									<ChevronRight className="w-3 h-3 shrink-0 text-zinc-500" />
								)}
								{session.imported ? (
									<CheckCircle className="w-3.5 h-3.5 text-emerald-500 shrink-0" />
								) : session.valid ? (
									<CheckCircle className="w-3.5 h-3.5 text-amber-500 shrink-0" />
								) : (
									<XCircle className="w-3.5 h-3.5 text-red-500 shrink-0" />
								)}
								<span className="text-sm text-zinc-300 truncate">
									{session.sourceTool || "unknown"} —{" "}
									{session.format || "unknown format"}
								</span>
							</div>
							<div className="flex items-center gap-2 text-xs text-zinc-500 shrink-0">
								<span>{session.sourceType}</span>
								{session.estimatedSize > 0 && (
									<span>{Math.round(session.estimatedSize / 1024)}KB</span>
								)}
							</div>
						</div>

						{/* Expanded Details */}
						{selectedSession?.id === session.id && (
							<div className="mt-3 pt-3 border-t border-zinc-800 space-y-2 text-xs">
								<div className="grid grid-cols-2 gap-2">
									<div>
										<span className="text-zinc-500">ID</span>
										<p
											className="text-zinc-300 font-mono truncate"
											title={session.id}
										>
											{session.id}
										</p>
									</div>
									<div>
										<span className="text-zinc-500">Source path</span>
										<p
											className="text-zinc-300 truncate"
											title={session.sourcePath}
										>
											{session.sourcePath}
										</p>
									</div>
									<div>
										<span className="text-zinc-500">Last modified</span>
										<p className="text-zinc-300">
											{session.lastModifiedAt || "unknown"}
										</p>
									</div>
									<div>
										<span className="text-zinc-500">Format</span>
										<p className="text-zinc-300">{session.format}</p>
									</div>
								</div>
								{session.detectedModels &&
									session.detectedModels.length > 0 && (
										<div>
											<span className="text-zinc-500">Detected models</span>
											<div className="flex gap-1 mt-1 flex-wrap">
												{session.detectedModels.map((m: string) => (
													<span
														key={m}
														className="px-1.5 py-0.5 bg-zinc-800 rounded text-zinc-400 text-2xs"
													>
														{m}
													</span>
												))}
											</div>
										</div>
									)}
								{session.errors && session.errors.length > 0 && (
									<div>
										<span className="text-red-400">Errors</span>
										<ul className="mt-1 list-disc list-inside text-red-400/70">
											{session.errors.map((e: string, i: number) => (
												<li key={i}>{e}</li>
											))}
										</ul>
									</div>
								)}
								{session.importedAt && (
									<div>
										<span className="text-zinc-500">Imported at</span>
										<p className="text-emerald-400">{session.importedAt}</p>
									</div>
								)}
								{session.valid && !session.imported && (
									<div className="pt-1">
										<button
											className="px-3 py-1.5 bg-emerald-900/40 text-emerald-300 border border-emerald-800/50 rounded hover:bg-emerald-900/60 text-xs"
											onClick={async (e) => {
												e.stopPropagation();
												await fetch("/api/go/api/session-export/import", {
													method: "POST",
													headers: { "Content-Type": "application/json" },
													body: JSON.stringify({
														data: JSON.stringify(session),
														merge: true,
													}),
												});
												fetchSessions();
											}}
											title="Import this session into the memory vault"
										>
											Import Session
										</button>
									</div>
								)}
							</div>
						)}
					</div>
				))}
			</div>
		</div>
	);
}
