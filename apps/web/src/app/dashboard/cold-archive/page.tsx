"use client";

import { useState, useEffect, useCallback } from "react";
import { Snowflake, Search, RotateCcw, Database, Loader2, ArrowUp } from "lucide-react";

interface ColdArchiveEntry {
  id: string;
  content: string;
  memory_kind: string;
  category: string;
  importance: number;
  heat_score: number;
  archived_at: string;
  created_at: string;
}

export default function ColdArchivePage() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<ColdArchiveEntry[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [promoting, setPromoting] = useState<string | null>(null);

  const search = useCallback(async (searchQuery = "") => {
    setLoading(true);
    try {
      const url = searchQuery.trim()
        ? `/api/go/api/memory/cold-archive/search?q=${encodeURIComponent(searchQuery)}&limit=50`
        : "/api/go/api/memory/cold-archive/search?limit=50";
      const res = await fetch(url);
      const d = await res.json();
      setResults(d.data ?? []);
      if (d.total !== undefined) setCount(d.total);
    } catch {
      // Best-effort
    }
    setLoading(false);
  }, []);

  const fetchCount = useCallback(async () => {
    try {
      const res = await fetch("/api/go/api/memory/cold-archive/count");
      const d = await res.json();
      if (d.count !== undefined) setCount(d.count);
      else if (d.data !== undefined && d.data.count !== undefined) setCount(d.data.count);
    } catch {
      // Best-effort
    }
  }, []);

  const promote = async (id: string) => {
    setPromoting(id);
    try {
      await fetch("/api/go/api/memory/cold-archive/promote", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      });
      setResults(prev => prev.filter(r => r.id !== id));
      fetchCount();
    } catch {
      // Best-effort
    }
    setPromoting(null);
  };

  useEffect(() => {
    search();
    fetchCount();
  }, [search, fetchCount]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Snowflake className="w-5 h-5 text-blue-400" />
          <div>
            <h1 className="text-lg font-semibold text-white">L3 Cold Archive</h1>
            <p className="text-xs text-zinc-500 mt-0.5">
              Long-term compressed memory tier for low-heat memories evicted from L2
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => { search(query); fetchCount(); }}
            disabled={loading}
            className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50"
            title="Refresh cold archive status"
          >
            <RotateCcw className="w-3 h-3" />
          </button>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="flex gap-3 text-xs">
        <div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800 flex items-center gap-2">
          <Database className="w-3.5 h-3.5 text-blue-400" />
          <span className="text-zinc-500">Archived memories</span>
          <span className="text-white font-medium">{count}</span>
        </div>
        <div className="px-3 py-2 bg-zinc-900 rounded border border-zinc-800 flex items-center gap-2">
          <Search className="w-3.5 h-3.5 text-zinc-500" />
          <span className="text-zinc-500">Results</span>
          <span className="text-white font-medium">{results.length}</span>
        </div>
      </div>

      {/* Search */}
      <div className="flex gap-2">
        <input
          type="text"
          placeholder="Search cold archive contents..."
          value={query}
          onChange={e => setQuery(e.target.value)}
          onKeyDown={e => e.key === "Enter" && search(query)}
          className="flex-1 px-3 py-1.5 bg-zinc-900 border border-zinc-700 rounded text-xs focus:outline-none focus:border-zinc-500"
          title="Search archived memories by content keyword"
        />
        <button
          onClick={() => search(query)}
          disabled={loading}
          className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50"
          title="Search cold archive entries"
        >
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <Search className="w-3 h-3" />}
        </button>
      </div>

      {/* Results */}
      <div className="space-y-2">
        {results.length === 0 && !loading && (
          <div className="text-center py-16 text-zinc-600 bg-zinc-900/30 border border-zinc-800 rounded-lg">
            <Snowflake className="w-12 h-12 mx-auto mb-4 opacity-20" />
            <p className="font-medium">Empty archive</p>
            <p className="text-xs mt-2 max-w-md mx-auto">
              Low-heat memories are automatically evicted from L2 to L3 cold storage when their
              heat score drops below 10.0. Search above to find archived entries.
            </p>
          </div>
        )}
        {results.map((entry) => (
          <div key={entry.id} className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-3 hover:bg-zinc-900 transition-colors">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <p className="text-sm text-zinc-300 truncate" title={entry.content}>
                  {entry.content?.slice(0, 200) || entry.id}
                </p>
                <div className="flex gap-3 mt-1.5 text-xs text-zinc-600">
                  <span>Kind: {entry.memory_kind || "fact"}</span>
                  <span>Category: {entry.category || "general"}</span>
                  <span>Importance: {entry.importance?.toFixed(2)}</span>
                  <span>Heat: {entry.heat_score?.toFixed(0)}</span>
                </div>
                <div className="flex gap-3 mt-0.5 text-2xs text-zinc-700">
                  <span>Archived: {entry.archived_at?.slice(0, 10) || "?"}</span>
                  <span>Created: {entry.created_at?.slice(0, 10) || "?"}</span>
                </div>
              </div>
              <button
                onClick={() => promote(entry.id)}
                disabled={promoting === entry.id}
                className="shrink-0 p-1.5 rounded hover:bg-amber-900/30 text-zinc-600 hover:text-amber-400 transition-colors disabled:opacity-50"
                title="Promote this memory back to the L2 vault (removes from cold archive)"
              >
                {promoting === entry.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <ArrowUp className="w-3.5 h-3.5" />}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
