export type BorgCapabilityStatus = 'shipped' | 'partial' | 'missing';

export type BorgCapability = {
    title: string;
    status: BorgCapabilityStatus;
    note: string;
    evidence: string;
};

export type BorgStartupSummary = {
    ready?: boolean;
    status?: string;
    summary?: string;
    checks?: {
        [key: string]: {
            ready?: boolean;
        } | undefined;
    };
};

export type BorgStatusSummary = {
    shippedCount: number;
    partialCount: number;
    missingCount: number;
    stage: 'full-parity' | 'parity-advancing' | 'compatibility-layer';
    stageLabel: string;
    coreReady: boolean;
    coreStatusLabel: string;
    coreStatusTone: 'ready' | 'pending' | 'warming' | 'degraded';
    coreStatusDetail: string | null;
    pendingStartupChecks: number;
};

export type BorgInstallSurfaceArtifact = {
    id: string;
    status: 'ready' | 'partial' | 'missing';
};

export type BorgStoreSnapshot = {
    exists?: boolean;
    totalEntries?: number;
    defaultSectionCount?: number;
    presentDefaultSectionCount?: number;
    populatedSectionCount?: number;
    missingSections?: string[];
    runtimePipeline?: {
        configuredMode?: string;
        providerNames?: string[];
        providerCount?: number;
        claudeMemEnabled?: boolean;
    };
};

export type BorgOperatorGuidance = {
    title: string;
    detail: string;
    tone: 'ready' | 'pending' | 'warning' | 'warming';
};

export const CLAUDE_MEM_CAPABILITIES: BorgCapability[] = [
    {
        title: 'Schema-inspired borg adapter',
        status: 'shipped',
        note: 'Hypercode ships a dedicated `BorgAdapter` that mirrors borg-style sections inside a Hypercode-managed local store.',
        evidence: 'packages/core/src/services/memory/BorgAdapter.ts',
    },
    {
        title: 'Redundant fan-out persistence',
        status: 'shipped',
        note: 'The default memory manager can fan out writes to both Hypercode JSON memory and the borg-inspired adapter.',
        evidence: 'packages/core/src/services/memory/RedundantMemoryManager.ts',
    },
    {
        title: 'Section-aware memory buckets',
        status: 'shipped',
        note: 'Current storage models project context, user facts, style preferences, commands, and general notes as borg-shaped sections.',
        evidence: 'packages/core/src/services/memory/BorgAdapter.ts',
    },
    {
        title: 'Dedicated operator parity surface',
        status: 'shipped',
        note: 'Hypercode now exposes a route that tells the truth about current borg assimilation instead of quietly forwarding to the generic vector explorer.',
        evidence: 'apps/web/src/app/dashboard/memory/borg/page.tsx',
    },
    {
        title: 'Canonical Hypercode observation schema',
        status: 'shipped',
        note: 'Hypercode defines shared observation input contracts in `@hypercode/types` and stores typed observation payloads with facts, concepts, files, hashes, and timestamps.',
        evidence: 'packages/types/src/schemas/memory.ts',
    },
    {
        title: 'Structured prompt and session summary capture',
        status: 'shipped',
        note: 'Hypercode natively records structured user prompts and supervised-session summaries alongside the adapter layer, instead of relying on the borg store alone.',
        evidence: 'packages/core/src/services/AgentMemoryService.ts',
    },
    {
        title: 'Generic Hypercode memory search foundation',
        status: 'partial',
        note: 'Hypercode can already search observations, prompts, summaries, and raw memory records from the main memory dashboard, but that is not yet a dedicated borg search/timeline workflow.',
        evidence: 'apps/web/src/app/dashboard/memory/page.tsx',
    },
    {
        title: 'Vector and graph memory primitives adjacent to the adapter',
        status: 'partial',
        note: 'Hypercode has broader memory infrastructure around the adapter, but it is not yet wired into a native borg runtime story.',
        evidence: 'apps/web/src/app/dashboard/memory/page.tsx',
    },
    {
        title: 'Claude Code lifecycle hooks',
        status: 'missing',
        note: 'Hypercode does not currently register SessionStart, UserPromptSubmit, PreToolUse, PostToolUse, Stop, or SessionEnd hooks into Claude Code.',
        evidence: 'Gap vs upstream borg hook system',
    },
    {
        title: 'Structured observation compression pipeline',
        status: 'partial',
        note: 'Hypercode already records heuristic typed observations with facts, concepts, files, and deduplicated hashes, but it does not yet have borg-style model-driven observation workers or response processors.',
        evidence: 'packages/core/src/services/AgentMemoryService.ts',
    },
    {
        title: 'Progressive-disclosure memory injection',
        status: 'missing',
        note: 'Hypercode does not yet assemble borg-style session context with index/detail/source layers and token-budgeted injection.',
        evidence: 'Gap vs upstream ContextBuilder / ObservationCompiler pipeline',
    },
    {
        title: 'Observation-centric search and timeline workflow',
        status: 'missing',
        note: 'Upstream tools like `search`, `timeline`, and `get_observations` do not have Hypercode-native borg equivalents yet.',
        evidence: 'Gap vs upstream memory MCP toolset',
    },
    {
        title: 'Transcript compression / Endless Mode',
        status: 'missing',
        note: 'Hypercode does not currently rewrite long-running transcripts in place to replace bulky tool output with compressed memories.',
        evidence: 'Gap vs upstream transcript transformer and watcher',
    },
    {
        title: 'Relational session-observation storage model',
        status: 'missing',
        note: 'There is no Hypercode-native borg schema yet for sessions, observations, summaries, prompts, correlations, and a persistent pending queue.',
        evidence: 'Gap vs upstream SQLite schema and queueing model',
    },
];

export const CLAUDE_MEM_IMPLEMENTATION_FILES = [
    {
        label: 'Current adapter implementation',
        path: 'packages/core/src/services/memory/BorgAdapter.ts',
        note: 'Flat-file JSON provider inspired by borg sections, not the full upstream runtime.',
    },
    {
        label: 'Redundant write manager',
        path: 'packages/core/src/services/memory/RedundantMemoryManager.ts',
        note: 'Fans out reads/writes across Hypercode JSON memory and the borg-inspired adapter.',
    },
    {
        label: 'Primary Hypercode memory dashboard',
        path: 'apps/web/src/app/dashboard/memory/page.tsx',
        note: 'Hypercode-native view for observations, prompts, session summaries, search, and provider interchange.',
    },
    {
        label: 'This parity page',
        path: 'apps/web/src/app/dashboard/memory/borg/page.tsx',
        note: 'Operator-facing truth table for what Hypercode has and has not assimilated from borg yet.',
    },
];

function getPendingStartupChecks(startupStatus?: BorgStartupSummary | null): number {
    if (!startupStatus?.checks) {
        return 0;
    }

    return Object.values(startupStatus.checks).filter((check) => check?.ready === false).length;
}

const BROWSER_EXTENSION_SURFACE_IDS = [
    'browser-extension-chromium',
    'browser-extension-firefox',
] as const;

function hasStartupInstallArtifactCheck(startupStatus?: BorgStartupSummary | null): boolean {
    const keys = Object.keys(startupStatus?.checks ?? {});
    return keys.some((key) => /artifact|installsurface/i.test(key));
}

function getPendingInstallArtifactCheckCount(installSurfaceArtifacts?: BorgInstallSurfaceArtifact[] | null): number {
    const relevantArtifacts = (installSurfaceArtifacts ?? []).filter((artifact) => BROWSER_EXTENSION_SURFACE_IDS.includes(artifact.id as (typeof BROWSER_EXTENSION_SURFACE_IDS)[number]));
    if (relevantArtifacts.length === 0) {
        return 1;
    }

    const allReady = relevantArtifacts.length === BROWSER_EXTENSION_SURFACE_IDS.length && relevantArtifacts.every((artifact) => artifact.status === 'ready');
    return allReady ? 0 : 1;
}

export function getBorgOperatorGuidance(storeStatus?: BorgStoreSnapshot | null): BorgOperatorGuidance {
    if (!storeStatus) {
        return {
            title: 'Reading adapter state',
            detail: 'Waiting for core to report whether the Hypercode-managed borg store exists and how many default buckets are already seeded.',
            tone: 'warming',
        };
    }

    const runtimePipeline = storeStatus.runtimePipeline;
    const defaultSectionCount = storeStatus.defaultSectionCount ?? 0;
    const presentDefaultSectionCount = storeStatus.presentDefaultSectionCount ?? 0;
    const populatedSectionCount = storeStatus.populatedSectionCount ?? 0;
    const missingSections = storeStatus.missingSections ?? [];

    if (runtimePipeline && runtimePipeline.claudeMemEnabled === false) {
        const providerLabel = runtimePipeline.providerNames?.length ? runtimePipeline.providerNames.join(', ') : 'no active providers reported';
        return {
            title: 'Claude-mem adapter not active in the runtime pipeline',
            detail: `Core reports the active memory pipeline as ${runtimePipeline.configuredMode ?? 'unknown'} with ${providerLabel}. The adapter file can still exist on disk, but Hypercode is not currently writing new memories through borg.`,
            tone: 'warning',
        };
    }

    if (!storeStatus.exists) {
        return {
            title: 'Adapter store not created yet',
            detail: `No Hypercode-managed claude_mem store exists yet. When the adapter initializes, it seeds ${defaultSectionCount} default buckets for project context, user facts, style preferences, commands, and general notes.`,
            tone: 'warning',
        };
    }

    if ((storeStatus.totalEntries ?? 0) === 0) {
        return {
            title: 'Adapter store seeded, waiting for entries',
            detail: `${presentDefaultSectionCount}/${defaultSectionCount} default buckets exist, but none contain entries yet. The adapter shell is ready; the workflow data is not.`,
            tone: 'pending',
        };
    }

    if (missingSections.length > 0) {
        return {
            title: 'Adapter store active, bucket coverage incomplete',
            detail: `${populatedSectionCount} bucket${populatedSectionCount === 1 ? '' : 's'} currently hold data, but ${missingSections.length} default bucket${missingSections.length === 1 ? '' : 's'} are still missing: ${missingSections.join(', ')}.`,
            tone: 'pending',
        };
    }

    return {
        title: 'Adapter store active',
        detail: `${populatedSectionCount} populated bucket${populatedSectionCount === 1 ? '' : 's'} across all ${presentDefaultSectionCount}/${defaultSectionCount} default borg buckets.`,
        tone: 'ready',
    };
}

export function getBorgStatusSummary(
    startupStatus?: BorgStartupSummary | null,
    installSurfaceArtifacts?: BorgInstallSurfaceArtifact[] | null,
): BorgStatusSummary {
    const shippedCount = CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'shipped').length;
    const partialCount = CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'partial').length;
    const missingCount = CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'missing').length;
    const coreReady = Boolean(startupStatus?.ready);
    const startupPendingChecks = getPendingStartupChecks(startupStatus);
    const installArtifactPendingChecks = startupStatus && !hasStartupInstallArtifactCheck(startupStatus)
        ? getPendingInstallArtifactCheckCount(installSurfaceArtifacts)
        : 0;
    const pendingStartupChecks = startupPendingChecks + installArtifactPendingChecks;
    const startupSummary = startupStatus?.summary?.trim() || null;

    const stage = missingCount === 0 && partialCount === 0
        ? 'full-parity'
        : missingCount <= partialCount
            ? 'parity-advancing'
            : 'compatibility-layer';

    const coreStatusLabel = !startupStatus
        ? 'Core warming up'
        : startupStatus.status === 'degraded'
            ? 'Core running in compat fallback'
            : coreReady && pendingStartupChecks > 0
                ? `Core ready · ${pendingStartupChecks} startup check${pendingStartupChecks === 1 ? '' : 's'} pending`
                : coreReady
                    ? 'Core ready'
                    : 'Core warming up';

    const coreStatusTone = !startupStatus
        ? 'warming'
        : startupStatus.status === 'degraded'
            ? 'degraded'
            : coreReady && pendingStartupChecks > 0
                ? 'pending'
                : coreReady
                    ? 'ready'
                    : 'warming';

    const coreStatusDetail = !startupStatus
        ? null
        : startupStatus.status === 'degraded'
            ? (startupSummary || 'Live startup telemetry is unavailable, so Hypercode is serving a cached compatibility snapshot.')
            : !coreReady && startupSummary
                ? startupSummary
                : null;

    return {
        shippedCount,
        partialCount,
        missingCount,
        stage,
        stageLabel: stage === 'full-parity'
            ? 'Full parity'
            : stage === 'parity-advancing'
                ? 'Parity advancing'
                : 'Compatibility layer',
        coreReady,
        coreStatusLabel,
        coreStatusTone,
        coreStatusDetail,
        pendingStartupChecks,
    };
}