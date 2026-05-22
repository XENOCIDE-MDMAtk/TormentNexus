import { describe, expect, it } from 'vitest';

import { CLAUDE_MEM_CAPABILITIES, getBorgOperatorGuidance, getBorgStatusSummary } from './borg-status';

describe('borg status helpers', () => {
    it('summarizes the current Hypercode borg parity state honestly', () => {
        expect(getBorgStatusSummary({ ready: true }, [
            { id: 'browser-extension-chromium', status: 'ready' },
            { id: 'browser-extension-firefox', status: 'ready' },
        ])).toEqual({
            shippedCount: CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'shipped').length,
            partialCount: CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'partial').length,
            missingCount: CLAUDE_MEM_CAPABILITIES.filter((item) => item.status === 'missing').length,
            stage: 'compatibility-layer',
            stageLabel: 'Compatibility layer',
            coreReady: true,
            coreStatusLabel: 'Core ready',
            coreStatusTone: 'ready',
            coreStatusDetail: null,
            pendingStartupChecks: 0,
        });
    });

    it('reports startup state when core readiness is still pending', () => {
        expect(getBorgStatusSummary({ ready: false }).coreStatusLabel).toBe('Core warming up');
        expect(getBorgStatusSummary({ ready: false }).coreStatusTone).toBe('warming');
        expect(getBorgStatusSummary(null).coreReady).toBe(false);
    });

    it('treats degraded startup compat fallback as a first-class operator state', () => {
        expect(getBorgStatusSummary({
            ready: false,
            status: 'degraded',
            summary: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        })).toMatchObject({
            coreReady: false,
            coreStatusLabel: 'Core running in compat fallback',
            coreStatusTone: 'degraded',
            coreStatusDetail: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        });
    });

    it('surfaces pending startup checks even after core reaches ready state', () => {
        expect(getBorgStatusSummary({
            ready: true,
            checks: {
                configSync: { ready: true },
                extensionBridge: { ready: false },
            },
        }, [
            { id: 'browser-extension-chromium', status: 'ready' },
            { id: 'browser-extension-firefox', status: 'ready' },
        ])).toMatchObject({
            coreReady: true,
            pendingStartupChecks: 1,
            coreStatusTone: 'pending',
            coreStatusLabel: 'Core ready · 1 startup check pending',
            coreStatusDetail: null,
        });
    });

    it('counts extension install artifacts as a pending startup check until both bundles are ready', () => {
        expect(getBorgStatusSummary({
            ready: true,
            checks: {
                configSync: { ready: true },
            },
        }, null)).toMatchObject({
            coreReady: true,
            pendingStartupChecks: 1,
            coreStatusTone: 'pending',
            coreStatusLabel: 'Core ready · 1 startup check pending',
        });

        expect(getBorgStatusSummary({
            ready: true,
            checks: {
                configSync: { ready: true },
            },
        }, [
            { id: 'browser-extension-chromium', status: 'ready' },
            { id: 'browser-extension-firefox', status: 'ready' },
        ])).toMatchObject({
            coreReady: true,
            pendingStartupChecks: 0,
            coreStatusTone: 'ready',
            coreStatusLabel: 'Core ready',
        });
    });

    it('does not double-count install artifacts when startup telemetry already reports that check', () => {
        expect(getBorgStatusSummary({
            ready: true,
            checks: {
                extensionInstallArtifacts: { ready: false },
            },
        }, [
            { id: 'browser-extension-chromium', status: 'missing' },
            { id: 'browser-extension-firefox', status: 'missing' },
        ])).toMatchObject({
            pendingStartupChecks: 1,
            coreStatusLabel: 'Core ready · 1 startup check pending',
        });
    });

    it('guides operators when the adapter store has not been created yet', () => {
        expect(getBorgOperatorGuidance({
            exists: false,
            defaultSectionCount: 5,
            presentDefaultSectionCount: 0,
            populatedSectionCount: 0,
            missingSections: ['project_context', 'user_facts', 'style_preferences', 'commands', 'general'],
            runtimePipeline: {
                configuredMode: 'redundant',
                providerNames: ['json', 'borg'],
                providerCount: 2,
                claudeMemEnabled: true,
            },
        })).toEqual({
            title: 'Adapter store not created yet',
            detail: 'No Hypercode-managed claude_mem store exists yet. When the adapter initializes, it seeds 5 default buckets for project context, user facts, style preferences, commands, and general notes.',
            tone: 'warning',
        });
    });

    it('guides operators when data exists but default bucket coverage is incomplete', () => {
        expect(getBorgOperatorGuidance({
            exists: true,
            totalEntries: 3,
            defaultSectionCount: 5,
            presentDefaultSectionCount: 2,
            populatedSectionCount: 2,
            missingSections: ['user_facts', 'style_preferences', 'general'],
            runtimePipeline: {
                configuredMode: 'redundant',
                providerNames: ['json', 'borg'],
                providerCount: 2,
                claudeMemEnabled: true,
            },
        })).toEqual({
            title: 'Adapter store active, bucket coverage incomplete',
            detail: '2 buckets currently hold data, but 3 default buckets are still missing: user_facts, style_preferences, general.',
            tone: 'pending',
        });
    });

    it('warns when borg is not part of the active memory pipeline', () => {
        expect(getBorgOperatorGuidance({
            exists: true,
            totalEntries: 3,
            defaultSectionCount: 5,
            presentDefaultSectionCount: 5,
            populatedSectionCount: 2,
            missingSections: [],
            runtimePipeline: {
                configuredMode: 'json',
                providerNames: ['json'],
                providerCount: 1,
                claudeMemEnabled: false,
            },
        })).toEqual({
            title: 'Claude-mem adapter not active in the runtime pipeline',
            detail: 'Core reports the active memory pipeline as json with json. The adapter file can still exist on disk, but Hypercode is not currently writing new memories through borg.',
            tone: 'warning',
        });
    });
});