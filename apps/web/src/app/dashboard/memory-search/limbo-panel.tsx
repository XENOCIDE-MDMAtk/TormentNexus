"use client";

import { useState, useCallback } from "react";
import { Skull, Search, RotateCcw, Loader2, Trash2 } from "lucide-react";

interface MemoryRecord {
  id: string;
  content: string;
  heat_score: number;
  importance: number;
  memory_kind?: string;
}

export function LimboPanel() {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<MemoryRecord[]>([]);
  const [stats, setStats] = useState<Record<string, number>>({});

  const fetchLimbo = useCallback(async (searchQuery = "") => {
    setLoading(true);
    try {
      const url = searchQuery.trim()
        ? `/api/go/api/memory/limbo/search?q=${encodeURIComponent(searchQuery)}`
        : "/api/go/api/memory/limbo/search";
      const res = await fetch(url);
      const d = await res.json();
      setResults(d.data ?? []);
      if (d.stats) setStats(d.stats);
    } catch {
      // Limbo fetch is best-effort
    }
    setLoading(false);
  }, []);

  const resurrect = async (id: string) => {
    try {
      await fetch("/api/go/api/memory/limbo/resurrect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      });
      setResults(prev => prev.filter(r => r.id !== id));
      fetchLimbo(query);
    } catch {
      // Resurrect is best-effort
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Skull className="w-4 h-4 text-zinc-500" />
          <span className="text-sm font-medium">L4 Limbo Vault</span>
        </div>
        <button
          onClick={() => fetchLimbo()}
          className="text-xs text-zinc-500 hover:text-zinc-300"
          title="Refresh limbo vault stats and entries"
        >
          <RotateCcw className="w-3 h-3 inline mr-1" />
          Refresh
        </button>
      </div>

      {/* Stats */}
      {Object.keys(stats).length > 0 && (
        <div className="flex gap-2 text-xs">
          {Object.entries(stats).map(([reason, count]) => (
            <span key={reason} className="px-2 py-1 bg-zinc-900 rounded border border-zinc-800 text-zinc-400">
              {reason}: <span className="text-red-400">{count}</span>
            </span>
          ))}
        </div>
      )}

      {/* Search */}
      <div className="flex gap-2">
        <input
          type="text"
          placeholder="Search limbo vault..."
          value={query}
          onChange={e => setQuery(e.target.value)}
          onKeyDown={e => e.key === "Enter" && fetchLimbo(query)}
          className="flex-1 px-3 py-1.5 bg-zinc-900 border border-zinc-700 rounded text-xs focus:outline-none focus:border-zinc-500"
          title="Search buried memories by keyword"
        />
        <button
          onClick={() => fetchLimbo(query)}
          disabled={loading}
          className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50"
          title="Search the L4 limbo vault for buried memories"
        >
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <Search className="w-3 h-3" />}
        </button>
      </div>

      {/* Results */}
      <div className="space-y-2 max-h-80 overflow-y-auto">
        {results.map(r => (
          <div key={r.id} className="flex items-start gap-2 p-2 bg-zinc-900/50 rounded border border-zinc-800 text-xs">
            <div className="flex-1 min-w-0">
              <div className="text-zinc-300 truncate">{r.content?.slice(0, 120) || r.id}</div>
              <div className="text-zinc-600 mt-0.5">
                Heat: {r.heat_score?.toFixed(0)} | Importance: {r.importance?.toFixed(2)} | {r.memory_kind || "fact"}
              </div>
            </div>
            <button
              onClick={() => resurrect(r.id)}
              className="shrink-0 p-1 rounded hover:bg-emerald-900/30 text-zinc-600 hover:text-emerald-400 transition-colors"
              title="Resurrect this memory from L4 limbo back to the L2 vault"
            >
              <RotateCcw className="w-3 h-3" />
            </button>
          </div>
        ))}
        {!loading && results.length === 0 && (
          <div className="text-center py-6 text-zinc-600 text-xs">
            <p>Limbo vault is empty</p>
            <p className="mt-1">Memories are sent here when discarded, lost, or decayed</p>
          </div>
        )}
      </div>
    </div>
  );
}
