
import { EventBus, SystemEvent } from '../services/EventBus.js';
import { HealerService } from '../services/HealerService.js';

export function shouldIgnoreExpectedStartupError(errorLog: string): boolean {
    const normalized = errorLog.toLowerCase();

    const ignoredFragments = [
        'openai api key not configured',
        'bad control character in string literal in json',
        'failed to load mcp.jsonc',
        'error fetching saved scripts',
        'error fetching mcp servers from config',
        'sqliteerror: no such table: config',
        // Billing/quota errors — all downstream LLM providers are unavailable;
        // the healer cannot generate a fix without a working LLM. Retrying would just
        // exhaust the fallback chain repeatedly and spam the logs.
        'insufficient balance',
        'credit balance is too low',
        'quota exceeded',
        'you exceeded your current quota',
        'too many requests',
        'rate limit',
        'retry in ',
        'fetch failed',
        'cannot find module',
        'is not recognized as an internal or external command',
        'does not provide any executables',
        'econnrefused',
        'no connection could be made because the target machine actively refused it',
        'could not connect to a chroma server',
        'not available on path',
        'failed to capture tool observation',
        'failed to infer data type',
        'sqlite runtime is unavailable',
        'could not locate the bindings file',
        'better_sqlite3.node',
    ];

    return ignoredFragments.some((fragment) => normalized.includes(fragment));
}

function getErrorLog(payload: unknown): string {
    if (!payload || typeof payload !== 'object') {
        return '';
    }

    const record = payload as Record<string, unknown>;
    const message = typeof record.message === 'string' ? record.message : '';
    const error = typeof record.error === 'string' ? record.error : '';
    const nestedErrorMessage = record.error && typeof record.error === 'object' && typeof (record.error as Record<string, unknown>).message === 'string'
        ? String((record.error as Record<string, unknown>).message)
        : '';
    const stack = typeof record.stack === 'string' ? record.stack : '';

    return [message, error, nestedErrorMessage, stack]
        .filter((value) => typeof value === 'string' && value.trim().length > 0)
        .join('\n');
}

export class HealerReactor {
    private eventBus: EventBus;
    private healerService: HealerService;
    private isHealing: boolean = false;
    private lastErrorTime: number = 0;
    private consecutiveFailures: number = 0;
    private readonly BASE_COOLDOWN_MS = 10000; // 10s base cooldown to prevent loops
    // After repeated healer failures (e.g., all LLM providers dead), back off exponentially
    // up to 5 minutes to avoid continuous log spam.
    private readonly MAX_COOLDOWN_MS = 300000;

    constructor(eventBus: EventBus, healerService: HealerService) {
        this.eventBus = eventBus;
        this.healerService = healerService;
    }

    public start() {
        console.log("[HealerReactor] 🛡️ Immune System Active. Listening for pathogens...");
        this.eventBus.subscribe('terminal:error', this.handleError.bind(this));
    }

    private async handleError(event: SystemEvent) {
        const errorLog = getErrorLog(event.payload);

        if (shouldIgnoreExpectedStartupError(errorLog)) {
            console.log('[HealerReactor] ℹ️ Ignoring expected startup/configuration error.');
            return;
        }

        // Prevent reaction loops (e.g., if the healer itself causes an error).
        // Cooldown grows with consecutive failures to back off when all LLM providers are dead.
        const cooldown = Math.min(
            this.BASE_COOLDOWN_MS * Math.pow(2, this.consecutiveFailures),
            this.MAX_COOLDOWN_MS
        );
        const now = Date.now();
        if (this.isHealing || (now - this.lastErrorTime < cooldown)) {
            console.log("[HealerReactor] ⏳ Healing in progress or cooldown. Skipping error.");
            return;
        }

        console.log(`[HealerReactor] 🩺 Detected Pathogen! Initiating immune response...`);
        this.isHealing = true;
        this.lastErrorTime = now;

        try {
            this.eventBus.emitEvent('task:update', 'HealerReactor', { message: 'Diagnosing error...' });

            // Trigger the Healer Service
            const report = await this.healerService.autoHeal(errorLog);

            if (report.success) {
                console.log(`[HealerReactor] ✅ Pathogen neutralized. Fix applied to ${report.file}`);
                this.eventBus.emitEvent('system:healed', 'HealerReactor', { file: report.file, fix: report.fix });
                this.consecutiveFailures = 0; // reset backoff on success
            } else {
                console.log(`[HealerReactor] ⚠️ Integration failed. Could not auto-heal.`);
                this.consecutiveFailures++;
            }

        } catch (error: any) {
            if (shouldIgnoreExpectedStartupError(String(error?.message || error))) {
                console.log(`[HealerReactor] ⏳ Healer API Limit Reached (Waiting for budget/quota...)`);
                this.eventBus.emitEvent('system:healer_halted', 'HealerReactor', { reason: 'CRITICAL_BUDGET_HALT', detail: error?.message });
            } else {
                console.error("[HealerReactor] ❌ Immune System Failure:", error);
            }
            this.consecutiveFailures++;
        } finally {
            this.isHealing = false;
        }
    }
}
