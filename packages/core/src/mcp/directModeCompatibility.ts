import type { CallToolResult, Tool } from '@modelcontextprotocol/sdk/types.js';
import type { SavedScriptConfig, SavedToolSetConfig } from '../interfaces/IConfigProvider.js';
import { getCompatibilityToolDefinitions } from './compatibilityToolDefinitions.js';
import {
    createCompatibleAgentRunner,
    executeCompatibleImportConfig,
    executeCompatibleRunCode,
    executeCompatibleRunAgent,
    executeCompatibleRunPython,
    executeCompatibleSearchMemory,
    executeCompatibleSaveMemory,
    executeCompatibleSaveScript,
} from './compatibilityToolRuntime.js';
import { getAllowedToolsMetadataGuardResult } from './toolAccessGuards.js';
import { executeSavedScriptTool, listSavedScriptTools } from './savedScriptExecution.js';
import {
    executeCompatibleListToolSets,
    executeCompatibleLoadToolSet,
    executeCompatibleSaveToolSet,
} from './toolSetCompatibility.js';

interface CodeModeServiceLike {
    enable(): void;
    disable(): void;
    isEnabled(): boolean;
    executeCode(code: string, context?: Record<string, unknown>): Promise<unknown>;
}

interface SandboxServiceLike {
    execute(language: 'node' | 'python', code: string, timeoutMs?: number, context?: Record<string, unknown>): Promise<{ output: string; result?: unknown; error?: string }>;
}

interface AgentMemoryServiceLike {
    add(
        content: string,
        type: 'session' | 'working' | 'long_term',
        namespace: 'user' | 'agent' | 'project',
        metadata?: Record<string, unknown>,
    ): Promise<{ id: string }>;
    search(query: string, options?: { limit?: number }): Promise<unknown>;
}

interface SavedScriptStoreLike {
    loadScripts(): Promise<SavedScriptConfig[]>;
    saveScript(script: SavedScriptConfig): Promise<unknown>;
}

interface ToolSetStoreLike {
    loadToolSets(): Promise<SavedToolSetConfig[]>;
    saveToolSet(toolSet: SavedToolSetConfig): Promise<unknown>;
}

interface NativeSessionCompatibilityLike {
    getLoadedToolNames(): string[];
    hasTool(name: string): boolean;
    loadToolIntoSession(name: string): { loaded: boolean; evicted: string[] };
}

interface ConfigImportServiceLike {
    importClaudeConfig(configJson: string): Promise<{ imported: number; skipped: string[] }>;
}

interface DirectModeAgentRunnerLike {
    runAgent(
        task: string,
        toolCallback: (toolName: string, toolArgs: unknown, meta?: Record<string, unknown>) => Promise<unknown>,
        policyId?: string,
    ): Promise<unknown>;
}

type DirectModeToolDelegate = (toolName: string, toolArgs: Record<string, unknown>, meta?: Record<string, unknown>) => Promise<unknown>;

export function getDirectModeMetadataGuardResult(
    toolName: string,
    meta?: Record<string, unknown>,
): CallToolResult | null {
    return getAllowedToolsMetadataGuardResult(toolName, meta);
}

export function createDirectModeAgentRunner(llm: import('./compatibilityToolRuntime.js').CompatibleLlmServiceLike): DirectModeAgentRunnerLike {
    return createCompatibleAgentRunner(llm, { includePolicyIdInToolMeta: true });
}

export function getDirectModeCompatibilityTools(): Tool[] {
    return getCompatibilityToolDefinitions({
        descriptions: {
            run_code: 'MetaMCP-compatible alias for one-shot Borg code execution without manually enabling Code Mode.',
            run_python: 'MetaMCP-compatible alias for Borg sandboxed Python execution.',
            run_agent: 'MetaMCP-compatible autonomous tool-using agent loop backed by Borg native LLM and tool execution surfaces.',
            save_memory: 'MetaMCP-compatible alias for persisting agent memory through Borg native memory services.',
            search_memory: 'MetaMCP-compatible alias for searching Borg native memory services.',
            save_script: 'MetaMCP-compatible alias for persisting reusable scripts in Borg managed config.',
            save_tool_set: 'MetaMCP-compatible alias for saving the currently loaded Borg session tools as a named tool set.',
            load_tool_set: 'MetaMCP-compatible alias for loading a saved Borg tool set into the current session working set.',
            toolset_list: 'List Borg-managed saved tool sets available to the current direct-mode session.',
            import_mcp_config: 'MetaMCP-compatible alias for importing MCP servers from Claude-style JSON config content.',
        },
    });
}

export async function getDirectModeSavedScriptTools(store: SavedScriptStoreLike): Promise<Tool[]> {
    return await listSavedScriptTools(store, (script) => script.description ?? `Run saved script '${script.name}'.`);
}

export async function tryHandleDirectModeCompatibilityTool(
    name: string,
    args: Record<string, unknown>,
    codeModeService: CodeModeServiceLike,
    sandboxService: SandboxServiceLike,
    agentMemoryService: AgentMemoryServiceLike,
    savedScriptStore: SavedScriptStoreLike,
    toolSetStore: ToolSetStoreLike,
    nativeSessionMetaTools: NativeSessionCompatibilityLike,
    agentRunner?: DirectModeAgentRunnerLike,
    delegatedToolCaller?: DirectModeToolDelegate,
    configImportService?: ConfigImportServiceLike,
): Promise<CallToolResult | null> {
    if (name === 'run_code') {
        return await executeCompatibleRunCode(args, {
            execute: async ({ code, context }) => await codeModeService.executeCode(code, context),
        }, {
            beforeExecute: () => {
                const wasEnabled = codeModeService.isEnabled();
                if (!wasEnabled) {
                    codeModeService.enable();
                    return () => {
                        codeModeService.disable();
                    };
                }

                return undefined;
            },
            isExecutionError: isErrorLike,
        });
    }

    if (name === 'run_python') {
        return await executeCompatibleRunPython(args, {
            execute: async (code) => await sandboxService.execute('python', code),
        });
    }

    if (name === 'run_agent') {
        return await executeCompatibleRunAgent(
            args,
            agentRunner,
            delegatedToolCaller,
            'Agent runner not available in Borg direct mode.',
        );
    }

    if (name === 'save_memory') {
        return await executeCompatibleSaveMemory(args, agentMemoryService);
    }

    if (name === 'search_memory') {
        return await executeCompatibleSearchMemory(args, agentMemoryService);
    }

    if (name === 'save_script') {
        return await executeCompatibleSaveScript(args, savedScriptStore);
    }

    if (name === 'save_tool_set') {
        return await executeCompatibleSaveToolSet(args, nativeSessionMetaTools, toolSetStore);
    }

    if (name === 'load_tool_set') {
        return await executeCompatibleLoadToolSet(args, nativeSessionMetaTools, toolSetStore);
    }

    if (name === 'toolset_list') {
        return await executeCompatibleListToolSets(toolSetStore);
    }

    if (name === 'import_mcp_config') {
        return await executeCompatibleImportConfig(
            args,
            configImportService,
            'Config import service not available in Borg direct mode.',
        );
    }

    return await executeSavedScriptTool(name, savedScriptStore);
}
function isErrorLike(value: unknown): boolean {
    return typeof value === 'object' && value !== null && 'success' in value && (value as { success?: unknown }).success === false;
}
