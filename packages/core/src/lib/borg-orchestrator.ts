import { existsSync, readFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';

type BorgLockRecord = {
    port?: number;
    host?: string;
};

type OrchestratorEnv = Record<string, string | undefined>;

function normalizeBaseURL(value?: string): string | null {
    const trimmed = value?.trim();
    if (!trimmed) {
        return null;
    }

    const withoutTrailingSlash = trimmed.replace(/\/$/, '');
    return withoutTrailingSlash.endsWith('/trpc')
        ? withoutTrailingSlash.slice(0, -5)
        : withoutTrailingSlash;
}

function resolveBrowserHost(host: string): string {
    return host === '0.0.0.0' || host === '::' || host === '[::]'
        ? '127.0.0.1'
        : host;
}

export function resolveBorgConfigDir(env: OrchestratorEnv = process.env): string {
    const configuredDir = env.BORG_CONFIG_DIR?.trim();
    if (configuredDir) {
        return configuredDir;
    }

    return path.join(os.homedir(), '.borg');
}

export function resolveBorgLockPath(env: OrchestratorEnv = process.env): string {
    return path.join(resolveBorgConfigDir(env), 'lock');
}

export function resolveLockedBorgBase(env: OrchestratorEnv = process.env): string | null {
    const lockPath = resolveBorgLockPath(env);
    if (!existsSync(lockPath)) {
        return null;
    }

    try {
        const parsed = JSON.parse(readFileSync(lockPath, 'utf8')) as BorgLockRecord;
        if (!parsed || typeof parsed.port !== 'number' || parsed.port <= 0) {
            return null;
        }

        const host = typeof parsed.host === 'string' && parsed.host.trim().length > 0
            ? resolveBrowserHost(parsed.host.trim())
            : '127.0.0.1';

        return `http://${host}:${parsed.port}`;
    } catch {
        return null;
    }
}

export function resolveOrchestratorBase(env: OrchestratorEnv = process.env): string | null {
    return normalizeBaseURL(env.BORG_ORCHESTRATOR_URL)
        ?? normalizeBaseURL(env.BORG_TRPC_UPSTREAM)
        ?? resolveLockedBorgBase(env)
        ?? normalizeBaseURL(env.NEXT_PUBLIC_BORG_ORCHESTRATOR_URL)
        ?? normalizeBaseURL(env.NEXT_PUBLIC_AUTOPILOT_URL);
}
