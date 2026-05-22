import { z } from 'zod';
import { t, publicProcedure, getHealerService } from '../lib/trpc-core.js';
import { observable } from '@trpc/server/observable';

import type { AnyTRPCRouter } from '@trpc/server';
import type * as _TRPCCore from '@trpc/server/unstable-core-do-not-import';

export const healerRouter = t.router({
    diagnose: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        return getHealerService().analyzeError(input.error, input.context || "");
    }),
    heal: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        const success = await getHealerService().heal(input.error, input.context || "");
        return { success };
    }),
    getHistory: t.procedure.query(async () => {
        try { return getHealerService()?.getHistory() ?? []; } catch { return []; }
    }),
    vaultRecords: publicProcedure.input(z.object({ limit: z.number().optional() }).optional()).query(async ({ input }) => {
        const SIDECAR_URL = process.env.HYPERCODE_SIDECAR_URL || 'http://127.0.0.1:4300';
        const limit = input?.limit || 50;
        try {
            const res = await fetch(`${SIDECAR_URL}/api/native/healer/vault?limit=${limit}`);
            if (res.ok) {
                const json = await res.json();
                if (json.records && json.records.length > 0) {
                    return json.records;
                }
            }
        } catch (e) {
            console.warn('[healerRouter] Failed to fetch vault records from sidecar, trying fallback:', e);
        }

        try {
            // Fallback for getting records if go backend isn't mapped
            const service = getHealerService() as any;
            if (service && typeof service.getVaultRecords === 'function') {
                return service.getVaultRecords();
            }
        } catch { }
        return [];
    }),
    subscribe: publicProcedure.subscription(() => {
        return observable<unknown>((emit) => {
            const onHeal = (data: unknown) => {
                emit.next(data);
            };
            const service = getHealerService();
            service.on('heal', onHeal);
            return () => {
                service.off('heal', onHeal);
            };
        });
    }),
    subscribeQuotaEvents: publicProcedure.subscription(() => {
        return observable<unknown>((emit) => {
            const onEvent = (eventData: unknown) => {
                emit.next(eventData);
            };
            // Safely fetch internal EventBus from active Orchestrator Context
            const server = getHealerService() as any; 
            // In Hypercode, either the server or generic system holds event bus. We use MCPServer fallback.
            const mcpServer = (global as any).mcpServerInstance;
            if (mcpServer && mcpServer.eventBus) {
                mcpServer.eventBus.on('system:llm_quota_exhausted', onEvent);
                mcpServer.eventBus.on('system:llm_fallback', onEvent);
                mcpServer.eventBus.on('system:healer_halted', onEvent);
                return () => {
                    mcpServer.eventBus.off('system:llm_quota_exhausted', onEvent);
                    mcpServer.eventBus.off('system:llm_fallback', onEvent);
                    mcpServer.eventBus.off('system:healer_halted', onEvent);
                };
            }
            return () => {}; // No-op teardown if missing context
        });
    })
});
