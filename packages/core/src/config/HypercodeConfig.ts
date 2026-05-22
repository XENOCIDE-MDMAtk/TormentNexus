import { z } from 'zod';
import fs from 'fs';
import path from 'path';

export const MCPServerConfigSchema = z.object({
    command: z.string(),
    args: z.array(z.string()).optional().default([]),
    env: z.record(z.string()).optional().default({}),
    enabled: z.boolean().optional().default(true)
});

export type MCPServerConfig = z.infer<typeof MCPServerConfigSchema>;

export const HypercodeConfigSchema = z.object({
    mcpServers: z.record(MCPServerConfigSchema).default({})
});

export type HypercodeConfig = z.infer<typeof HypercodeConfigSchema>;

export class HypercodeConfigLoader {
    private static getConfigPath(): string {
        // Look in current working directory (project root)
        return path.join(process.cwd(), 'hypercode.config.json');
    }

    public static loadConfig(): HypercodeConfig {
        const configPath = this.getConfigPath();
        if (!fs.existsSync(configPath)) {
            console.warn(`[HypercodeConfig] No config found at ${configPath}. Using defaults.`);
            return { mcpServers: {} };
        }

        try {
            const raw = fs.readFileSync(configPath, 'utf-8');
            const json = JSON.parse(raw);
            return HypercodeConfigSchema.parse(json);
        } catch (error) {
            console.error(`[HypercodeConfig] Failed to load config:`, error);
            return { mcpServers: {} };
        }
    }
}
