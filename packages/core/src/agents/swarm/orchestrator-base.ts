import { resolveOrchestratorBase } from '../../lib/borg-orchestrator.js';

export function resolveSwarmOrchestratorBase(explicitBase?: string): string | null {
    const explicit = explicitBase?.trim();
    if (explicit) {
        return explicit.replace(/\/$/, '');
    }

    return resolveOrchestratorBase();
}
