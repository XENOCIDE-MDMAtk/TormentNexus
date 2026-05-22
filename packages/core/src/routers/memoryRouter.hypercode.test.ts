import path from 'path';
import { describe, expect, it } from 'vitest';

import { summarizeBorgRuntimePipeline, summarizeBorgStore } from './memoryRouter.borg.js';

describe('summarizeBorgStore', () => {
    it('summarizes section counts and latest update timestamps', () => {
        const result = summarizeBorgStore(
            path.join('C:', 'hypercode', '.hypercode', 'claude_mem.json'),
            {
                sections: [
                    {
                        section: 'project_context',
                        entries: [
                            { createdAt: '2026-03-10T10:00:00.000Z' },
                            { createdAt: '2026-03-10T11:00:00.000Z' },
                        ],
                    },
                    {
                        section: 'commands',
                        entries: [{ createdAt: '2026-03-09T09:00:00.000Z' }],
                    },
                ],
            },
        );

        expect(result).toEqual({
            exists: true,
            storePath: path.join('C:', 'hypercode', '.hypercode', 'claude_mem.json'),
            totalEntries: 3,
            sectionCount: 2,
            defaultSectionCount: 5,
            presentDefaultSectionCount: 2,
            populatedSectionCount: 2,
            missingSections: ['user_facts', 'style_preferences', 'general'],
            runtimePipeline: {
                configuredMode: 'unknown',
                providerNames: [],
                providerCount: 0,
                sectionedStoreEnabled: false,
            },
            sections: [
                { section: 'project_context', entryCount: 2 },
                { section: 'commands', entryCount: 1 },
            ],
            lastUpdatedAt: '2026-03-10T11:00:00.000Z',
        });
    });

    it('returns an empty status shape when the store is missing', () => {
        expect(summarizeBorgStore('C:/hypercode/.hypercode/claude_mem.json', null)).toEqual({
            exists: false,
            storePath: 'C:/hypercode/.hypercode/claude_mem.json',
            totalEntries: 0,
            sectionCount: 0,
            defaultSectionCount: 5,
            presentDefaultSectionCount: 0,
            populatedSectionCount: 0,
            missingSections: ['project_context', 'user_facts', 'style_preferences', 'commands', 'general'],
            runtimePipeline: {
                configuredMode: 'unknown',
                providerNames: [],
                providerCount: 0,
                sectionedStoreEnabled: false,
            },
            sections: [],
            lastUpdatedAt: null,
        });
    });

    it('summarizes when borg is part of the active memory pipeline', () => {
        expect(summarizeBorgRuntimePipeline({
            configuredMode: 'redundant',
            providerNames: ['json', 'borg'],
            providerCount: 2,
            sectionedStoreEnabled: true,
        })).toEqual({
            configuredMode: 'redundant',
            providerNames: ['json', 'borg'],
            providerCount: 2,
            sectionedStoreEnabled: true,
        });
    });
});