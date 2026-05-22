import { existsSync, readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';

export interface HypercodeStartupProvenance {
    requestedRuntime?: string;
    activeRuntime?: string;
    requestedPort?: number;
    activePort?: number;
    portDecision?: string;
    portReason?: string;
    launchMode?: string;
    dashboardMode?: string;
    installDecision?: string;
    installReason?: string;
    buildDecision?: string;
    buildReason?: string;
    updatedAt?: string;
}

interface HypercodeStartLockRecord {
    instanceId: string;
    pid: number;
    port: number;
    host: string;
    createdAt: string;
    startup?: HypercodeStartupProvenance;
}

function normalizeStartupProvenance(record: HypercodeStartLockRecord | null): HypercodeStartupProvenance | null {
    if (!record || typeof record.port !== 'number' || record.port <= 0) {
        return null;
    }

    const startup = record.startup ?? {};
    const activePort = typeof startup.activePort === 'number' && startup.activePort > 0
        ? startup.activePort
        : record.port;
    const requestedPort = typeof startup.requestedPort === 'number' && startup.requestedPort > 0
        ? startup.requestedPort
        : activePort;

    return {
        ...startup,
        requestedPort,
        activePort,
        portDecision: startup.portDecision?.trim() || 'derived from lock record',
        portReason: startup.portReason?.trim() || 'Detailed startup port provenance was unavailable; using the current control-plane lock port.',
        updatedAt: startup.updatedAt?.trim() || record.createdAt,
    };
}

function resolveDataDir(dataDir: string, homeDirectory: string = homedir()): string {
    if (dataDir === '~') {
        return homeDirectory;
    }

    if (dataDir.startsWith('~/') || dataDir.startsWith('~\\') || dataDir.startsWith(`~${path.sep}`)) {
        return path.resolve(homeDirectory, dataDir.slice(2));
    }

    return path.isAbsolute(dataDir) ? dataDir : path.resolve(dataDir);
}

function readStartLockRecord(dataDir: string): HypercodeStartLockRecord | null {
    const lockPath = path.join(resolveDataDir(dataDir), 'lock');
    if (!existsSync(lockPath)) {
        return null;
    }

    try {
        const parsed = JSON.parse(readFileSync(lockPath, 'utf8')) as Partial<HypercodeStartLockRecord>;
        if (
            typeof parsed.instanceId !== 'string'
            || typeof parsed.pid !== 'number'
            || typeof parsed.port !== 'number'
            || typeof parsed.host !== 'string'
            || typeof parsed.createdAt !== 'string'
        ) {
            return null;
        }

        return parsed as HypercodeStartLockRecord;
    } catch {
        return null;
    }
}

export function readLocalStartupProvenance(dataDir: string = process.env.HYPERCODE_DATA_DIR ?? '~/.hypercode'): HypercodeStartupProvenance | null {
    return normalizeStartupProvenance(readStartLockRecord(dataDir));
}
