
import { DEFAULT_OPENROUTER_FREE_MODEL, LLMService } from '@borg/ai';
import type { MCPServer } from '../MCPServer.js';
import path from 'path';
import fs from 'fs';
import { EventEmitter } from 'events';

export interface Diagnosis {
    errorType: string;
    description: string;
    file?: string;
    line?: number;
    suggestedFix: string;
    confidence: number;
}

export interface FixPlan {
    id: string;
    diagnosis: Diagnosis;
    filesToModify: { path: string; content: string }[];
    explanation: string;
}

export interface HealRecord {
    timestamp: number;
    error: string;
    fix: FixPlan;
    success: boolean;
}

interface LlmTextResponse {
    text?: string;
    content?: string;
}

function extractJsonObject(text: string): string {
    const fencedMatch = text.match(/```(?:json)?\s*([\s\S]*?)\s*```/i);
    if (fencedMatch?.[1]) {
        return fencedMatch[1].trim();
    }

    const firstBrace = text.indexOf('{');
    const lastBrace = text.lastIndexOf('}');
    if (firstBrace >= 0 && lastBrace > firstBrace) {
        return text.slice(firstBrace, lastBrace + 1).trim();
    }

    return text.trim();
}

/**
 * Reason: LLM providers in this codebase can return either `{ text }`, `{ content }`, or a raw string.
 * What: Normalizes heterogeneous provider responses into a single string extraction path.
 * Why: Avoids repeated unsafe casts and keeps JSON parsing behavior consistent across healer flows.
 */
function extractLlmText(response: unknown): string {
    if (typeof response === 'string') {
        return response;
    }

    if (response && typeof response === 'object') {
        const maybeResponse = response as LlmTextResponse;
        if (typeof maybeResponse.text === 'string') {
            return maybeResponse.text;
        }
        if (typeof maybeResponse.content === 'string') {
            return maybeResponse.content;
        }
    }

    return String(response);
}

export class HealerService extends EventEmitter {
    private llm: LLMService;
    private server: MCPServer;
    private history: HealRecord[] = [];

    public getHistory() {
        return this.history;
    }

    constructor(llm: LLMService, server: MCPServer) {
        super();
        this.llm = llm;
        this.server = server;
    }

    public async autoHeal(errorLog: string): Promise<{ success: boolean; file?: string; fix?: string }> {
        console.log("[HealerService] 🚑 Auto-healing initialized...");

        // 1. Diagnose
        const diagnosis = await this.analyzeError(errorLog);
        console.log("[HealerService] 🔍 Diagnosis:", diagnosis);

        if (!diagnosis.file || !diagnosis.suggestedFix) {
            console.log("[HealerService] 🤷 Could not identify file or fix.");
            return { success: false };
        }

        // 2. Locate File
        const filePath = path.isAbsolute(diagnosis.file)
            ? diagnosis.file
            : path.join(process.cwd(), diagnosis.file);

        if (!fs.existsSync(filePath)) {
            console.log(`[HealerService] ❌ File not found: ${filePath}`);
            return { success: false };
        }

        // 3. Apply Fix
        // Ask LLM to generate the FULL file content with the fix applied
        const currentContent = await fs.readFileSync(filePath, 'utf-8');
        const prompt = `
        You are an Expert Linter and Fixer.
        Original File Content:
        \`\`\`typescript
        ${currentContent}
        \`\`\`

        Diagnosis: ${diagnosis.description}
        Suggested Fix: ${diagnosis.suggestedFix}

        Output the COMPLETE, CORRECTED file content. Do not include markdown fences.
        `;

        const response = await this.llm.generateText('openrouter', DEFAULT_OPENROUTER_FREE_MODEL, 'You are a code fixer.', prompt);
        const newContent = response.content.replace(/```typescript|```/g, '').trim();

        // 4. Write
        await fs.writeFileSync(filePath, newContent);
        console.log(`[HealerService] 💉 Fix applied to ${filePath}`);

        return { success: true, file: filePath, fix: diagnosis.suggestedFix };
    }

    public async analyzeError(error: Error | string, context?: string): Promise<Diagnosis> {
        const errorStr = typeof error === 'string' ? error : error.message + '\n' + error.stack;

        const prompt = `
        You are The Healer, an expert debugging agent.
        Analyze the following error and context.
        Provide a diagnosis and a suggested fix.
        
        Error:
        ${errorStr}
        
        Context:
        ${context || 'No additional context.'}
        
        Return JSON format:
        {
            "errorType": "SyntaxError|RuntimeError|LogicError|...",
            "description": "Short explanation",
            "file": "path/to/culprit.ts (if known)",
            "line": 123 (if known),
            "suggestedFix": "Code snippet or description of fix",
            "confidence": 0.0 to 1.0
        }
        `;

        const response = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "You are a JSON-only debugging tool.", prompt, {});

        try {
            return JSON.parse(extractJsonObject(extractLlmText(response)));
        } catch (e) {
            console.error("Failed to parse Healer diagnosis", response);
            return {
                errorType: "Unknown",
                description: "Failed to parse LLM diagnosis",
                suggestedFix: "Manual review required",
                confidence: 0
            };
        }
    }

    public async generateFix(diagnosis: Diagnosis): Promise<FixPlan> {
        if (!diagnosis.file) {
            throw new Error("Cannot generate fix without file path.");
        }

        // 1. Read File Content
        const fs = await import('fs/promises');
        let content = '';
        try {
            content = await fs.readFile(diagnosis.file, 'utf-8');
        } catch (e) {
            throw new Error(`Failed to read file: ${diagnosis.file}`);
        }

        // 2. Generate Fix via LLM
        const prompt = `
        You are The Healer.
        Generate a fix for the following file based on the diagnosis.
        
        Diagnosis: ${diagnosis.description}
        Suggested Fix: ${diagnosis.suggestedFix}
        
        File Content:
        ${content}
        
        Return JSON format:
        {
            "explanation": "Why this fix works",
            "newContent": "The entire new file content"
        }
        `;

        const response = await this.llm.generateText("openrouter", DEFAULT_OPENROUTER_FREE_MODEL, "You are a code repair agent. Return only JSON with 'explanation' and 'newContent'.", prompt, {});

        try {
            const result = JSON.parse(extractLlmText(response));
            return {
                id: Math.random().toString(36).substring(7),
                diagnosis,
                filesToModify: [{ path: diagnosis.file, content: result.newContent }],
                explanation: result.explanation
            };
        } catch (e) {
            console.error("Failed to parse fix plan", response);
            throw new Error("Failed to generate valid fix plan.");
        }
    }

    public async applyFix(plan: FixPlan): Promise<boolean> {
        const fs = await import('fs/promises');
        try {
            for (const file of plan.filesToModify) {
                await fs.writeFile(file.path, file.content, 'utf-8');
            }
            return true;
        } catch (e) {
            console.error("Failed to apply fix", e);
            return false;
        }
    }

    public async heal(error: Error | string, context?: string): Promise<boolean> {
        console.log("🚑 Healer Activated...");
        const diagnosis = await this.analyzeError(error, context);
        console.log("📋 Diagnosis:", diagnosis);

        if (diagnosis.confidence < 0.8) {
            console.warn("⚠️ Confidence too low for auto-heal.");
            return false;
        }

        if (!diagnosis.file) {
            console.warn("⚠️ No file identified to fix.");
            return false;
        }

        try {
            const plan = await this.generateFix(diagnosis);
            console.log("🛠️ Fix Generated:", plan.explanation);

            const success = await this.applyFix(plan);

            const historyItem = {
                timestamp: Date.now(),
                error: typeof error === 'string' ? error : error.message,
                fix: plan,
                success
            };
            this.history.push(historyItem);
            this.emit('heal', historyItem);

            return success;
        } catch (e) {
            console.error("❌ Healer Failed:", e);
            const historyItem = {
                timestamp: Date.now(),
                error: typeof error === 'string' ? error : error.message,
                fix: { id: 'failed', diagnosis, filesToModify: [], explanation: 'Fix generation failed' },
                success: false
            };
            this.history.push(historyItem);
            this.emit('heal', historyItem);
            return false;
        }
    }


}
