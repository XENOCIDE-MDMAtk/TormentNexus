

"use client";

import { trpc } from "@/utils/trpc";
import { useState } from "react";
import { motion } from "framer-motion";
import { Loader2, RefreshCw, Download, Play, CheckCircle, AlertTriangle } from "lucide-react";
import { toast } from "sonner";

type ExpertTrpc = {
    expert: {
        research: { useMutation: () => any };
        code: { useMutation: () => any };
    };
};

export default function KnowledgeDashboard() {
    const trpcWithExpert = trpc as unknown as typeof trpc & ExpertTrpc;
    // Real submodule data
    const submodulesQuery = trpc.submodule.list.useQuery();
    const submodulesData = submodulesQuery.data || [];

    const resourcesQuery = trpc.knowledge.getResources.useQuery();

    const [ingestUrl, setIngestUrl] = useState("");
    const [ingestLog, setIngestLog] = useState("");
    const ingestMutation = trpc.knowledge.ingest.useMutation();

    // Research State
    const [researchQuery, setResearchQuery] = useState("");
    const [researchDepth, setResearchDepth] = useState(2);
    const researchMutation = trpcWithExpert.expert.research.useMutation();

    const handleResearch = () => {
        if (!researchQuery) return;
        researchMutation.mutate({ query: researchQuery, depth: researchDepth, breadth: 3 });
    };


    // Coder State
    const [coderTask, setCoderTask] = useState("");
    const coderMutation = trpcWithExpert.expert.code.useMutation({
        onSuccess: () => {
            toast.success("Coder task started");
            setCoderTask("");
        },
        onError: (err) => toast.error("Coder task failed: " + err.message)
    });

    const handleCode = () => {
        if (!coderTask) return;
        coderMutation.mutate({ task: coderTask });
    };

    // Submodule Mutations
    const utils = trpc.useUtils();
    const updateAllMutation = trpc.submodule.updateAll.useMutation({
        onSuccess: (res) => {
            if (res.success) toast.success("Submodules synced successfully");
            else toast.error("Sync failed: " + res.output);
            utils.submodule.list.invalidate();
        }
    });

    const installMutation = trpc.submodule.installDependencies.useMutation({
        onSuccess: (res) => {
            if (res.success) toast.success("Dependencies installed");
            else toast.error("Install failed: " + res.output);
            utils.submodule.list.invalidate();
        }
    });

    const buildMutation = trpc.submodule.build.useMutation({
        onSuccess: (res) => {
            if (res.success) toast.success("Build complete");
            else toast.error("Build failed: " + res.output);
            utils.submodule.list.invalidate();
        }
    });

    const enableMutation = trpc.submodule.enable.useMutation({
        onSuccess: (res) => {
            if (res.success) toast.success("Submodule enabled");
            else toast.error("Enable failed: " + res.output);
            utils.submodule.list.invalidate();
        }
    });

    const handleIngest = async () => {
        if (!ingestUrl) return;
        setIngestLog(`Ingesting: ${ingestUrl}...`);
        try {
            const result = await ingestMutation.mutateAsync({ url: ingestUrl });
            setIngestLog(`Success: ${result}`);
            setIngestUrl("");
        } catch (e: any) {
            setIngestLog(`Error: ${e.message}`);
        }
    };

    const handleUpdateAll = async () => {
        updateAllMutation.mutate();
    };

    const [actioning, setActioning] = useState<string | null>(null);

    const handleAction = async (action: 'install' | 'build' | 'enable', path: string) => {
        setActioning(`${action}-${path}`);
        try {
            if (action === 'install') await installMutation.mutateAsync({ path });
            if (action === 'build') await buildMutation.mutateAsync({ path });
            if (action === 'enable') await enableMutation.mutateAsync({ path });
        } catch (e) {
            toast.error(`Action failed: ${(e as Error).message}`);
        } finally {
            setActioning(null);
        }
    };

    const submodules = submodulesData;
    const resources = resourcesQuery.data || { categories: [] };

    return (
        <div className="p-8 bg-gray-900 min-h-screen text-gray-100 font-mono">
            <header className="mb-8 border-b border-gray-700 pb-4 flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-blue-400">KNOWLEDGE BASE</h1>
                    <p className="text-gray-400">External Intelligence & Submodules</p>
                </div>
                <div className="text-right">
                    <div className="text-sm text-gray-500">LAST SYNC</div>
                    <div className="text-blue-500 font-bold">{resources.lastUpdated ? new Date(resources.lastUpdated).toLocaleString() : 'Never'}</div>
                </div>
            </header>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Submodules Section */}
                <section className="bg-gray-800 rounded-lg border border-gray-700 p-6">
                    <div className="flex justify-between items-center mb-4">
                        <h2 className="text-xl font-bold text-yellow-400 flex items-center">
                            <span className="mr-2">📦</span> Submodules ({submodules.length})
                        </h2>
                        <button
                            onClick={handleUpdateAll}
                            disabled={updateAllMutation.isPending}
                            className={`flex items-center gap-2 px-3 py-1 bg-yellow-600 hover:bg-yellow-500 rounded text-sm font-bold transition-colors disabled:opacity-50 disabled:cursor-not-allowed`}
                        >
                            {updateAllMutation.isPending ? <Loader2 className="animate-spin h-3 w-3" /> : <RefreshCw className="h-3 w-3" />}
                            {updateAllMutation.isPending ? 'SYNCING...' : 'SYNC ALL'}
                        </button>
                    </div>

                    <div className="overflow-y-auto max-h-[600px] space-y-2 pr-2">
                        {submodulesQuery.isLoading ? (
                            <div className="flex justify-center p-8">
                                <Loader2 className="animate-spin text-gray-500" />
                            </div>
                        ) : submodules.length === 0 ? (
                            <div className="text-gray-500 italic text-center p-8 border border-dashed border-gray-700 rounded">
                                <p>No submodules detected.</p>
                                <p className="text-xs mt-2">Check .gitmodules or run Sync All.</p>
                            </div>
                        ) : (
                            submodules.map((sub: any) => (
                                <div key={sub.path} className="bg-gray-900 p-3 rounded border-l-2 border-gray-700 hover:border-yellow-500 transition-colors group">
                                    <div className="flex justify-between items-start">
                                        <div>
                                            <span className="font-bold text-blue-300 block">{sub.name}</span>
                                            <span className="text-gray-500 text-xs block truncate max-w-[300px]" title={sub.path}>{sub.path}</span>
                                        </div>
                                        <div className="flex flex-col items-end gap-1">
                                            <span className={`text-[10px] px-2 py-0.5 rounded uppercase font-bold tracking-wider ${sub.status === 'clean' ? 'bg-green-900/50 text-green-400 border border-green-800' : 'bg-red-900/50 text-red-400 border border-red-800'}`}>
                                                {sub.status}
                                            </span>
                                            <span className="text-[10px] font-mono text-gray-600">{sub.commit.substring(0, 7)}</span>
                                        </div>
                                    </div>

                                    <div className="mt-2 flex flex-wrap gap-2">
                                        {sub.capabilities?.map((cap: string) => (
                                            <span key={cap} className="text-[10px] bg-purple-900/30 text-purple-300 px-1.5 py-0.5 rounded border border-purple-800/50">
                                                {cap}
                                            </span>
                                        ))}
                                    </div>

                                    <div className="mt-3 flex gap-2 border-t border-gray-800 pt-2 opacity-60 group-hover:opacity-100 transition-opacity">
                                        <button
                                            onClick={() => handleAction('install', sub.path)}
                                            disabled={!!actioning || sub.isInstalled}
                                            className={`flex items-center gap-1 text-xs px-2 py-1 rounded border ${sub.isInstalled
                                                ? 'border-green-900 text-green-600 bg-green-900/10 cursor-default'
                                                : 'border-gray-600 text-gray-300 hover:bg-gray-800'}`}
                                        >
                                            {actioning === `install-${sub.path}` ? <Loader2 className="animate-spin h-3 w-3" /> : (sub.isInstalled ? <CheckCircle className="h-3 w-3" /> : <Download className="h-3 w-3" />)}
                                            {sub.isInstalled ? 'Installed' : 'Install'}
                                        </button>

                                        <button
                                            onClick={() => handleAction('build', sub.path)}
                                            disabled={!!actioning || !sub.isInstalled}
                                            className={`flex items-center gap-1 text-xs px-2 py-1 rounded border ${sub.isBuilt
                                                ? 'border-green-900 text-green-600 bg-green-900/10' // Can rebuild
                                                : 'border-gray-600 text-gray-300 hover:bg-gray-800 disabled:opacity-30'}`}
                                        >
                                            {actioning === `build-${sub.path}` ? <Loader2 className="animate-spin h-3 w-3" /> : <Play className="h-3 w-3" />}
                                            Build
                                        </button>

                                        {!sub.isInstalled && (
                                            <button
                                                onClick={() => handleAction('enable', sub.path)}
                                                disabled={!!actioning}
                                                className="ml-auto text-xs text-yellow-600 hover:text-yellow-500 flex items-center gap-1"
                                            >
                                                Enable
                                            </button>
                                        )}
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                    {/* Ingestion Section */}
                    <section className="bg-gray-800 rounded-lg border border-gray-700 p-6 flex flex-col gap-4 h-fit">
                        <h2 className="text-xl font-bold text-green-400 flex items-center">
                            <span className="mr-2">🧠</span> Ingest Knowledge
                        </h2>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                className="flex-1 bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:border-green-500 outline-none placeholder:text-gray-600"
                                placeholder="Enter URL to ingest (e.g., https://example.com/docs)"
                                value={ingestUrl}
                                onChange={(e) => setIngestUrl(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && handleIngest()}
                            />
                            <button
                                onClick={handleIngest}
                                disabled={ingestMutation.isPending}
                                className="bg-green-600 hover:bg-green-500 text-white px-4 py-2 rounded font-bold disabled:opacity-50 flex items-center gap-2"
                            >
                                {ingestMutation.isPending && <Loader2 className="animate-spin h-4 w-4" />}
                                INGEST
                            </button>
                        </div>
                        {ingestLog && (
                            <div className="bg-black p-3 rounded text-xs font-mono text-gray-300 break-all border border-gray-700 max-h-[200px] overflow-y-auto">
                                {ingestLog}
                            </div>
                        )}
                    </section>

                </section>

                {/* Deep Research Section */}
                <section className="bg-gray-800 rounded-lg border border-gray-700 p-6 flex flex-col gap-4 lg:col-span-2">
                    <h2 className="text-xl font-bold text-blue-400 flex items-center">
                        <span className="mr-2">🕵️</span> Deep Research Agent
                    </h2>
                    <div className="flex flex-col gap-4">
                        <div className="flex gap-2">
                            <input
                                type="text"
                                className="flex-1 bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:border-blue-500 outline-none placeholder:text-gray-600"
                                placeholder="Research Topic (e.g., 'Latest advancements in solid state batteries')"
                                value={researchQuery}
                                onChange={(e) => setResearchQuery(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && handleResearch()}
                            />
                            <div className="flex items-center gap-2 bg-gray-900 border border-gray-700 rounded px-3">
                                <span className="text-xs text-gray-500">DEPTH</span>
                                <input
                                    type="number"
                                    min="1" max="5"
                                    aria-label="Research depth"
                                    title="Research depth"
                                    value={researchDepth}
                                    onChange={(e) => setResearchDepth(parseInt(e.target.value))}
                                    className="bg-transparent w-10 text-center outline-none text-white"
                                />
                            </div>
                            <button
                                onClick={handleResearch}
                                disabled={researchMutation.isPending}
                                className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded font-bold disabled:opacity-50 flex items-center gap-2"
                            >
                                {researchMutation.isPending && <Loader2 className="animate-spin h-4 w-4" />}
                                RESEARCH
                            </button>
                        </div>

                        {researchMutation.data && (
                            <div className="bg-gray-900 p-4 rounded border border-gray-700 mt-2">
                                <h3 className="text-lg font-bold text-green-400 mb-2">Research Report</h3>
                                <div className="prose prose-invert max-w-none text-sm text-gray-300">
                                    <p className="whitespace-pre-wrap">{researchMutation.data.summary}</p>

                                    <h4 className="font-bold text-gray-400 mt-4 mb-2">Sources:</h4>
                                    <ul className="list-disc pl-5 space-y-1">
                                        {researchMutation.data.findings?.map((f: any, i: number) => (
                                            <li key={i}>
                                                <a href={f.content} target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:underline">
                                                    {f.source}
                                                </a>
                                            </li>
                                        ))}
                                    </ul>
                                </div>
                            </div>
                        )}

                        {researchMutation.error && (
                            <div className="p-4 bg-red-900/20 border border-red-800 text-red-300 rounded">
                                Error: {researchMutation.error.message}
                            </div>
                        )}
                    </div>
                </section>

                {/* Coder Agent Section */}
                <section className="bg-gray-800 rounded-lg border border-gray-700 p-6 flex flex-col gap-4 lg:col-span-2">
                    <h2 className="text-xl font-bold text-purple-400 flex items-center">
                        <span className="mr-2">👨‍💻</span> Coder Agent
                    </h2>
                    <div className="flex gap-2">
                        <input
                            type="text"
                            className="flex-1 bg-gray-900 border border-gray-700 rounded px-3 py-2 text-white focus:border-purple-500 outline-none placeholder:text-gray-600"
                            placeholder="Coding Task (e.g., 'Create a new utility file utils/math.ts')"
                            value={coderTask}
                            onChange={(e) => setCoderTask(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleCode()}
                        />
                        <button
                            onClick={handleCode}
                            disabled={coderMutation.isPending}
                            className="bg-purple-600 hover:bg-purple-500 text-white px-4 py-2 rounded font-bold disabled:opacity-50 flex items-center gap-2"
                        >
                            {coderMutation.isPending && <Loader2 className="animate-spin h-4 w-4" />}
                            CODE
                        </button>
                    </div>
                    {coderMutation.data && (
                        <div className="bg-gray-900 p-3 rounded text-sm border border-gray-700">
                            <h4 className="text-green-400 font-bold mb-1">Success!</h4>
                            <p className="text-gray-300">Files Changed: {coderMutation.data.filesChanged?.join(', ')}</p>
                            <p className="text-gray-500 text-xs mt-1">{coderMutation.data.reasoning}</p>
                        </div>
                    )}
                    {coderMutation.error && (
                        <div className="p-3 bg-red-900/20 border border-red-800 text-red-300 rounded text-sm">
                            Error: {coderMutation.error.message}
                        </div>
                    )}
                </section>
            </div>
        </div>
    );
}
