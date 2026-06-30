"use client";

import { useEffect, useState } from "react";
import { Loader2, Activity } from "lucide-react";

const CATEGORY_DESC: Record<string, string> = {
  browser: "Control browser automation: manage pages, capture screenshots, scrape content, and inspect console logs across remote browser instances.",
  "browser-extension": "Bridge to the TormentNexus browser extension: save web memories, parse DOM content, and manage browser-extension stored data.",
  "cli-harnesses": "Detect and manage CLI harnesses installed on the system: view versions, capabilities, and install surfaces for AI coding tools.",
  "cloud-dev": "Manage cloud development sessions: create, monitor, and communicate with remote development environments across providers.",
  deerflow: "DeerFlow integration bridge: query available models, skills, and memory status through the DeerFlow service.",
  healer: "Self-healing diagnostics and auto-repair system: analyze errors, attempt automated fixes, and review repair history.",
  imports: "Import external sessions from Claude, Gemini, Aider, and other AI coding tools into the TormentNexus memory system.",
  "logs-metrics": "System logs and performance metrics: view provider breakdowns, routing history, monitoring stats, and system snapshots.",
  mesh: "P2P memory synchronization mesh: discover peers, query capabilities, broadcast messages, and sync memory across machines.",
  observability: "Real-time system observability: pulse events, provider status monitoring, and service health tracking.",
  runtime: "Runtime status overview: service health checks, lock file status, startup readiness, and imported instructions.",
};

export default function GenericDashboardPage() {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const path = typeof window !== "undefined" ? window.location.pathname.replace("/dashboard/", "/api/") : "";
  const pageName = path.split("/").pop()?.replace(/-/g, " ") || "Page";
  const desc = CATEGORY_DESC[path.split("/").pop() || ""] || `Browse ${pageName} data from the TormentNexus API.`;

  useEffect(() => {
    if (!path) return;
    fetch(path.startsWith("/api/") ? `/api/go${path}` : `/api/go/api/${path}`)
      .then(r => r.json().catch(() => ({ raw: true })))
      .then(d => { setData(d); setLoading(false); })
      .catch(e => { setError(String(e)); setLoading(false); });
  }, [path]);

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2 capitalize">
          <Activity className="w-6 h-6" />
          {pageName}
        </h1>
        <p className="text-zinc-400 text-sm mt-1" title={desc}>{desc}</p>
      </div>

      {loading && <div className="flex items-center gap-2 text-zinc-500"><Loader2 className="w-4 h-4 animate-spin" /> Loading...</div>}
      {error && <div className="text-red-400 bg-red-950/20 rounded-lg p-4 border border-red-900/50">{error}</div>}
      {data && (
        <div className="space-y-3">
          {Array.isArray(data) ? (
            data.length === 0 ? (
              <div className="text-center py-12 text-zinc-600">
                <p className="font-medium">No data available</p>
                <p className="text-xs mt-1">The API returned an empty list. Check back once data has been populated.</p>
              </div>
            ) : (
              data.slice(0, 50).map((item: any, i: number) => (
                <div key={i} className="border border-zinc-800 rounded-lg p-4 bg-zinc-900/50 hover:border-zinc-700 transition-colors">
                  {Object.entries(item).slice(0, 6).map(([k, v]) => (
                    <div key={k} className="flex gap-2 text-sm">
                      <span className="text-zinc-500 font-mono min-w-[120px]">{k}:</span>
                      <span className="text-zinc-300 truncate">{typeof v === 'object' ? JSON.stringify(v).slice(0, 100) : String(v).slice(0, 100)}</span>
                    </div>
                  ))}
                  {Object.keys(item).length > 6 && <div className="text-xs text-zinc-600 mt-1">... and {Object.keys(item).length - 6} more fields</div>}
                </div>
              ))
            )
          ) : (
            <div className="border border-zinc-800 rounded-lg p-4 bg-zinc-900/50">
              {Object.entries(data).slice(0, 20).map(([k, v]) => (
                <div key={k} className="flex gap-2 text-sm py-1">
                  <span className="text-zinc-500 font-mono min-w-[160px]">{k}:</span>
                  <span className="text-zinc-300">{typeof v === 'object' ? JSON.stringify(v).slice(0, 200) : String(v).slice(0, 200)}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
