import { spawn, type ChildProcess } from 'child_process';

import { resolveCliEntryPath } from './orchestratorPaths.js';

export interface BackgroundCoreBootstrapOptions {
    healthUrl?: string;
    host?: string;
    startupTimeoutMs?: number;
    pollIntervalMs?: number;
    waitForReady?: boolean;
    cliEntryPath?: string | null;
    log?: (message?: unknown, ...optionalParams: unknown[]) => void;
}

export interface BackgroundCoreBootstrapResult {
    status: 'already-running' | 'spawned' | 'warming' | 'launch-unavailable';
    pid?: number;
    cliEntryPath?: string;
}

type FetchLike = typeof globalThis.fetch;
type SpawnLike = (
    command: string,
    args: ReadonlyArray<string>,
    options: {
        detached?: boolean;
        stdio?: 'ignore';
        windowsHide?: boolean;
    },
) => Pick<ChildProcess, 'pid' | 'unref'>;

interface BackgroundCoreBootstrapDeps {
    fetchImpl?: FetchLike;
    spawnImpl?: SpawnLike;
    waitImpl?: (ms: number) => Promise<void>;
}

const DEFAULT_HEALTH_URL = 'http://127.0.0.1:3001/health';
const DEFAULT_STARTUP_TIMEOUT_MS = 15_000;
const DEFAULT_POLL_INTERVAL_MS = 500;

function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function isCoreBridgeHealthy(
    healthUrl: string = DEFAULT_HEALTH_URL,
    fetchImpl: FetchLike = globalThis.fetch,
): Promise<boolean> {
    try {
        const response = await fetchImpl(healthUrl, { method: 'GET' });
        return response.ok;
    } catch {
        return false;
    }
}

export async function waitForCoreBridge(
    options: {
        healthUrl?: string;
        timeoutMs?: number;
        pollIntervalMs?: number;
    } = {},
    deps: Pick<BackgroundCoreBootstrapDeps, 'fetchImpl' | 'waitImpl'> = {},
): Promise<boolean> {
    const healthUrl = options.healthUrl ?? DEFAULT_HEALTH_URL;
    const timeoutMs = options.timeoutMs ?? DEFAULT_STARTUP_TIMEOUT_MS;
    const pollIntervalMs = options.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS;
    const fetchImpl = deps.fetchImpl ?? globalThis.fetch;
    const waitImpl = deps.waitImpl ?? delay;
    const deadline = Date.now() + timeoutMs;

    do {
        if (await isCoreBridgeHealthy(healthUrl, fetchImpl)) {
            return true;
        }

        if (Date.now() >= deadline) {
            break;
        }

        await waitImpl(pollIntervalMs);
    } while (true);

    return false;
}

export async function ensureBackgroundCoreRunning(
    options: BackgroundCoreBootstrapOptions = {},
    deps: BackgroundCoreBootstrapDeps = {},
): Promise<BackgroundCoreBootstrapResult> {
    const healthUrl = options.healthUrl ?? DEFAULT_HEALTH_URL;
    const log = options.log ?? (() => undefined);
    const fetchImpl = deps.fetchImpl ?? globalThis.fetch;
    const waitImpl = deps.waitImpl ?? delay;

    if (await isCoreBridgeHealthy(healthUrl, fetchImpl)) {
        return { status: 'already-running' };
    }

    const cliEntryPath = Object.prototype.hasOwnProperty.call(options, 'cliEntryPath')
        ? (options.cliEntryPath ?? null)
        : resolveCliEntryPath();
    if (!cliEntryPath) {
        log('[Hypercode Core] Background core bootstrap skipped: CLI entrypoint not found.');
        return { status: 'launch-unavailable' };
    }

    const spawnImpl = deps.spawnImpl ?? spawn;
    const args = [cliEntryPath, 'start'];

    if (options.host) {
        args.push('--host', options.host);
    }

    const child = spawnImpl(process.execPath, args, {
        detached: true,
        stdio: 'ignore',
        windowsHide: true,
    });
    child.unref?.();

    if (options.waitForReady === false) {
        return {
            status: 'spawned',
            pid: child.pid,
            cliEntryPath,
        };
    }

    const ready = await waitForCoreBridge(
        {
            healthUrl,
            timeoutMs: options.startupTimeoutMs ?? DEFAULT_STARTUP_TIMEOUT_MS,
            pollIntervalMs: options.pollIntervalMs ?? DEFAULT_POLL_INTERVAL_MS,
        },
        {
            fetchImpl,
            waitImpl,
        },
    );

    return {
        status: ready ? 'spawned' : 'warming',
        pid: child.pid,
        cliEntryPath,
    };
}