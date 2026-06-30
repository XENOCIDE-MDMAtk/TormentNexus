"use client";

import { useState, useEffect, useCallback } from "react";
import { Terminal, Play, RotateCcw, Loader2, Search, Wrench, FileJson, Clock, CheckCircle, XCircle } from "lucide-react";

interface ToolDef {
  name: string;
  description?: string;
  handler?: string;
  file?: string;
  parameters?: Record<string, unknown>;
  inputSchema?: { properties?: Record<string, { type?: string; description?: string }> };
}

interface ExecRecord {
  id: string;
  toolName: string;
  args: string;
  result: string;
  success: boolean;
  timestamp: string;
  duration: string;
}

export default function ToolConsolePage() {
  const [tools, setTools] = useState<ToolDef[]>([]);
  const [selectedTool, setSelectedTool] = useState<ToolDef | null>(null);
  const [args, setArgs] = useState("{}");
  const [result, setResult] = useState<string | null>(null);
  const [executing, setExecuting] = useState(false);
  const [history, setHistory] = useState<ExecRecord[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"execute" | "history" | "schema">("execute");
  const [error, setError] = useState("");

  const fetchTools = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/go/api/mcp/tools?simple=true");
      if (res.ok) {
        const d = await res.json();
        const allTools = d.data ?? d ?? [];
        setTools(Array.isArray(allTools) ? allTools : []);
      }
    } catch {
      // Best-effort
    }
    setLoading(false);
  }, []);

  const loadSchema = useCallback(async (toolName: string) => {
    try {
      const res = await fetch(`/api/go/api/mcp/tools/schema?tool=${toolName}`);
      if (res.ok) {
        const d = await res.json();
        const schema = d.data?.inputSchema ?? d.inputSchema ?? d.schema ?? {};
        setSelectedTool(prev => prev?.name === toolName ? { ...prev, inputSchema: schema } : prev);
        // Generate default args from schema
        if (schema.properties) {
          const defaults: Record<string, string | number | boolean> = {};
          for (const [key, val] of Object.entries(schema.properties as Record<string, { type?: string }>)) {
            if (val.type === "string") defaults[key] = "";
            else if (val.type === "number" || val.type === "integer") defaults[key] = 0;
            else if (val.type === "boolean") defaults[key] = false;
          }
          if (Object.keys(defaults).length > 0) {
            setArgs(JSON.stringify(defaults, null, 2));
          }
        }
      }
    } catch {
      // Best-effort
    }
  }, []);

  const selectTool = async (tool: ToolDef) => {
    setSelectedTool(tool);
    setResult(null);
    setError("");
    setArgs("{}");
    setActiveTab("execute");
    if (tool.name) await loadSchema(tool.name);
  };

  const execute = async () => {
    if (!selectedTool) return;
    setExecuting(true);
    setResult(null);
    setError("");
    const start = performance.now();

    try {
      let parsed: Record<string, unknown> = {};
      try { parsed = JSON.parse(args); } catch { setError("Invalid JSON in arguments"); setExecuting(false); return; }

      const res = await fetch("/api/go/api/mcp/tools/call", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          toolName: selectedTool.name,
          name: selectedTool.name,
          arguments: parsed,
          args: parsed,
        }),
      });

      const duration = ((performance.now() - start) / 1000).toFixed(2);
      const data = await res.json();
      const resultStr = JSON.stringify(data.data ?? data, null, 2);
      const success = data.success !== false && res.ok;

      setResult(resultStr);
      setError(success ? "" : data.error || data.detail || "Tool execution failed");

      setHistory(prev => [{
        id: Date.now().toString(36),
        toolName: selectedTool.name,
        args: JSON.stringify(parsed),
        result: resultStr.slice(0, 500),
        success,
        timestamp: new Date().toLocaleTimeString(),
        duration: `${duration}s`,
      }, ...prev].slice(0, 50));
    } catch (e) {
      setError(String(e));
      setResult(null);
    }
    setExecuting(false);
  };

  useEffect(() => { fetchTools(); }, [fetchTools]);

  const normalizedSearch = search.toLowerCase().trim();
  const filteredTools = tools.filter(t =>
    !normalizedSearch ||
    t.name?.toLowerCase().includes(normalizedSearch) ||
    t.description?.toLowerCase().includes(normalizedSearch)
  );

  const schemaProps = selectedTool?.inputSchema?.properties ?? selectedTool?.parameters ?? {};

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Terminal className="w-5 h-5 text-emerald-400" />
          <div>
            <h1 className="text-lg font-semibold text-white">Tool Execution Console</h1>
            <p className="text-xs text-zinc-500 mt-0.5">
              Browse, inspect, and execute native Go tools against the live system
            </p>
          </div>
        </div>
        <button onClick={fetchTools} disabled={loading} className="px-3 py-1.5 bg-zinc-800 rounded hover:bg-zinc-700 text-xs disabled:opacity-50">
          {loading ? <Loader2 className="w-3 h-3 animate-spin" /> : <RotateCcw className="w-3 h-3" />}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Tool Browser */}
        <div className="md:col-span-1 bg-zinc-900/50 border border-zinc-800 rounded-lg">
          <div className="p-3 border-b border-zinc-800">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3 h-3 text-zinc-500" />
              <input
                value={search}
                onChange={e => setSearch(e.target.value)}
                placeholder="Search tools..."
                className="w-full pl-9 pr-3 py-1.5 bg-zinc-900 border border-zinc-700 rounded text-xs focus:outline-none focus:border-zinc-500"
              />
            </div>
          </div>
          <div className="overflow-y-auto max-h-[60vh]">
            {loading && <div className="p-4 text-xs text-zinc-600 text-center"><Loader2 className="w-3 h-3 animate-spin inline mr-1" />Loading...</div>}
            {!loading && filteredTools.length === 0 && (
              <div className="p-4 text-xs text-zinc-600 text-center">No tools found</div>
            )}
            {filteredTools.map(tool => (
              <button
                key={tool.name}
                onClick={() => selectTool(tool)}
                className={`w-full text-left px-3 py-2 text-xs border-b border-zinc-800/50 hover:bg-zinc-800 transition-colors ${
                  selectedTool?.name === tool.name ? "bg-zinc-800 text-white" : "text-zinc-400"
                }`}
              >
                <div className="flex items-center gap-2">
                  <Wrench className="w-3 h-3 shrink-0 text-orange-400" />
                  <span className="font-medium truncate">{tool.name}</span>
                </div>
                {tool.description && (
                  <p className="text-zinc-600 truncate mt-0.5 ml-5">{tool.description}</p>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Execution Panel */}
        <div className="md:col-span-2 space-y-4">
          {!selectedTool ? (
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-8 text-center">
              <Terminal className="w-10 h-10 mx-auto mb-3 text-zinc-700" />
              <p className="text-zinc-500 text-sm">Select a tool from the left panel</p>
              <p className="text-xs text-zinc-700 mt-1">Choose a tool to inspect its schema and execute it</p>
            </div>
          ) : (
            <>
              {/* Tool Info + Tab Bar */}
              <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg">
                <div className="p-3 border-b border-zinc-800">
                  <div className="flex items-center justify-between">
                    <div>
                      <h2 className="text-sm font-medium text-white">{selectedTool.name}</h2>
                      {selectedTool.description && (
                        <p className="text-xs text-zinc-500 mt-0.5">{selectedTool.description}</p>
                      )}
                    </div>
                    {selectedTool.file && (
                      <span className="text-2xs text-zinc-700 font-mono">{selectedTool.file}</span>
                    )}
                  </div>
                  <div className="flex gap-2 mt-3 border-b border-zinc-800">
                    {(["execute", "schema", "history"] as const).map(tab => (
                      <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        className={`px-3 py-1.5 text-xs border-b-2 transition-colors capitalize ${
                          activeTab === tab ? "text-emerald-400 border-emerald-400" : "text-zinc-500 border-transparent hover:text-zinc-300"
                        }`}
                      >
                        {tab === "execute" && <Play className="w-3 h-3 inline mr-1" />}
                        {tab === "schema" && <FileJson className="w-3 h-3 inline mr-1" />}
                        {tab === "history" && <Clock className="w-3 h-3 inline mr-1" />}
                        {tab}
                      </button>
                    ))}
                  </div>
                </div>

                <div className="p-3">
                  {activeTab === "execute" && (
                    <div className="space-y-3">
                      <div>
                        <label className="text-xs text-zinc-500 block mb-1">Arguments (JSON)</label>
                        <textarea
                          value={args}
                          onChange={e => setArgs(e.target.value)}
                          className="w-full h-28 bg-zinc-950 border border-zinc-800 rounded p-2 text-xs font-mono text-zinc-300 focus:outline-none focus:border-zinc-600"
                          spellCheck={false}
                        />
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={execute}
                          disabled={executing}
                          className="px-4 py-2 bg-emerald-700 hover:bg-emerald-600 rounded text-xs font-medium text-white disabled:opacity-50 flex items-center gap-1.5"
                        >
                          {executing ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3" />}
                          {executing ? "Executing..." : "Execute"}
                        </button>
                        {selectedTool.handler && (
                          <span className="text-2xs text-zinc-700 font-mono">{selectedTool.handler}</span>
                        )}
                      </div>
                      {error && (
                        <div className="p-2 bg-red-950/30 border border-red-900/50 rounded text-xs text-red-400">
                          {error}
                        </div>
                      )}
                      {result && (
                        <div>
                          <label className="text-xs text-zinc-500 block mb-1">Result</label>
                          <pre className="bg-zinc-950 border border-zinc-800 rounded p-2 text-xs font-mono text-zinc-300 overflow-x-auto max-h-60">
                            {result}
                          </pre>
                        </div>
                      )}
                    </div>
                  )}

                  {activeTab === "schema" && (
                    <div className="space-y-2">
                      {Object.keys(schemaProps).length === 0 ? (
                        <p className="text-xs text-zinc-600 italic">No schema available for this tool</p>
                      ) : (
                        Object.entries(schemaProps).map(([key, val]) => (
                          <div key={key} className="flex gap-2 text-xs py-1 border-b border-zinc-800/50 last:border-0">
                            <span className="text-zinc-300 font-mono min-w-[24px]">{(val as any).type || "string"}</span>
                            <span className="text-zinc-400 font-mono min-w-[120px]">{key}</span>
                            <span className="text-zinc-600">{(val as any).description || ""}</span>
                          </div>
                        ))
                      )}
                    </div>
                  )}

                  {activeTab === "history" && (
                    <div className="space-y-1 max-h-60 overflow-y-auto">
                      {history.length === 0 ? (
                        <p className="text-xs text-zinc-600 italic">No executions yet</p>
                      ) : (
                        history.map(entry => (
                          <div key={entry.id} className="flex items-start gap-2 text-xs py-1.5 border-b border-zinc-800/30 last:border-0">
                            {entry.success ? (
                              <CheckCircle className="w-3 h-3 text-emerald-500 shrink-0 mt-0.5" />
                            ) : (
                              <XCircle className="w-3 h-3 text-red-500 shrink-0 mt-0.5" />
                            )}
                            <div className="flex-1 min-w-0">
                              <span className="text-zinc-300 font-medium">{entry.toolName}</span>
                              <span className="text-zinc-600 ml-1.5">{entry.duration}</span>
                              <p className="text-zinc-600 truncate">{entry.result?.slice(0, 80) || "no result"}</p>
                            </div>
                            <span className="text-zinc-700 shrink-0">{entry.timestamp}</span>
                          </div>
                        ))
                      )}
                    </div>
                  )}
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
