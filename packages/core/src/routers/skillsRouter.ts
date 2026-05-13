
import { z } from 'zod';
import {
    t,
    publicProcedure,
    getSkillRegistry,
    getSkillAssimilationService,
} from '../lib/trpc-core.js';

const skillSchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z.string(),
    content: z.string(),
    path: z.string(),
});

export const skillsRouter = t.router({
    list: publicProcedure.output(z.array(skillSchema)).query(async () => {
        return getSkillRegistry()?.getSkills() ?? [];
    }),

    read: publicProcedure.input(z.object({
        name: z.string()
    })).query(async ({ input }) => {
        const registry = getSkillRegistry();
        if (!registry) {
            return { content: [{ type: "text", text: "Skill registry not available" }] };
        }
        return registry.readSkill(input.name);
    }),

    create: publicProcedure.input(z.object({
        id: z.string(),
        name: z.string(),
        description: z.string()
    })).mutation(async ({ input }) => {
        const registry = getSkillRegistry();
        if (!registry) {
            return { content: [{ type: "text", text: "Skill registry not available" }] };
        }
        return registry.createSkill(input.id, input.name, input.description);
    }),

    save: publicProcedure.input(z.object({
        id: z.string(),
        content: z.string()
    })).mutation(async ({ input }) => {
        const registry = getSkillRegistry();
        if (!registry) {
            return { content: [{ type: "text", text: "Skill registry not available" }] };
        }
        return registry.saveSkill(input.id, input.content);
    }),

    assimilate: publicProcedure.input(z.object({
        topic: z.string(),
        docsUrl: z.string().optional()
    })).mutation(async ({ input }) => {
        const service = getSkillAssimilationService();
        if (!service) {
            return { success: false, logs: ["Service not ready"] };
        }

        return await service.assimilate({
            topic: input.topic,
            docsUrl: input.docsUrl,
            autoInstall: true
        });
    }),
});

