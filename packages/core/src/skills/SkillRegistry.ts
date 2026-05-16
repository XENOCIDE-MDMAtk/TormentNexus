import glob from 'fast-glob';
import fs from 'fs/promises';
import path from 'path';
import matter from 'gray-matter';

export interface Skill {
    id: string;
    name: string;
    description: string;
    content: string;
    path: string;
    score?: number;
}

function getErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
}

export class SkillRegistry {
    private skills: Map<string, Skill> = new Map();
    private searchPaths: string[];
    private masterIndexPath?: string;

    constructor(searchPaths: string[]) {
        this.searchPaths = searchPaths;
    }

    setMasterIndexPath(indexPath: string) {
        this.masterIndexPath = indexPath;
    }

    async getLibraryIndex() {
        if (!this.masterIndexPath) {
            return { categories: { mcp_servers: [], universal_harness: [], skills: [] } };
        }

        try {
            const content = await fs.readFile(this.masterIndexPath, 'utf-8');
            const cleanJSON = content.replace(/\\"|"(?:\\"|[^"])*"|(\/\/.*|\/\*[\s\S]*?\*\/)/g, (m, g) => g ? "" : m);
            return JSON.parse(cleanJSON);
        } catch (e) {
            console.error("Error reading master index:", e);
            return { categories: { mcp_servers: [], universal_harness: [], skills: [] } };
        }
    }

    hasSkill(id: string): boolean {
        return this.skills.has(id);
    }

    getSkills(): Skill[] {
        return Array.from(this.skills.values());
    }

    async loadSkills() {
        this.skills.clear();

        for (const loc of this.searchPaths) {
            try {
                const entries = await glob(['**/SKILL.md', '**/skill.md'], {
                    cwd: loc,
                    absolute: true,
                    deep: 3
                });

                for (const file of entries) {
                    await this.parseSkill(file);
                }
            } catch (e) {
            }
        }
    }

    private async parseSkill(filePath: string) {
        try {
            const raw = await fs.readFile(filePath, 'utf-8');
            const { data, content } = matter(raw);
            const id = data.name || path.basename(path.dirname(filePath));

            const skill: Skill = {
                id,
                name: data.name || id,
                description: data.description || "No description provided",
                content,
                path: filePath
            };

            this.skills.set(id, skill);
        } catch (e) {
            console.error(`Failed to parse skill at ${filePath}`, e);
        }
    }

    async getPredictedSkills(chatHistory: string, activeGoal: string): Promise<Skill[]> {
        const SIDECAR_URL = process.env.BORG_SIDECAR_URL || 'http://localhost:4300';
        try {
            const response = await fetch(`${SIDECAR_URL}/api/skills/predict`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ chatHistory, activeGoal }),
            });

            if (!response.ok) return [];
            const result = (await response.json()) as any;
            return result?.data?.predictedSkills || [];
        } catch (e) {
            console.warn('[SkillRegistry] Sidecar skill prediction failed, using local fallback.');
            return this.localSearch(activeGoal || chatHistory.slice(-500), 5);
        }
    }

    private localSearch(query: string, limit: number): Skill[] {
        const queryLower = query.toLowerCase();
        return Array.from(this.skills.values())
            .map(s => {
                let score = 0;
                if (s.name.toLowerCase().includes(queryLower)) score += 10;
                if (s.description.toLowerCase().includes(queryLower)) score += 5;
                if (s.content.toLowerCase().includes(queryLower)) score += 1;
                return { ...s, score };
            })
            .filter(s => (s.score || 0) > 0)
            .sort((a, b) => (b.score || 0) - (a.score || 0))
            .slice(0, limit);
    }

    getSkillTools() {
        return [
            {
                name: "list_skills",
                description: "List all available skills (runbooks). Uses progressive disclosure.",
                inputSchema: {
                    type: "object",
                    properties: {
                        query: { type: "string", description: "Optional filter/context for ranking" }
                    }
                }
            },
            {
                name: "read_skill",
                description: "Read the content/instructions of a specific skill",
                inputSchema: {
                    type: "object",
                    properties: {
                        skillName: { type: "string" }
                    },
                    required: ["skillName"]
                }
            },
            {
                name: "create_skill",
                description: "Create a new skill (runbook)",
                inputSchema: {
                    type: "object",
                    properties: {
                        id: { type: "string", description: "Unique ID (folder name)" },
                        name: { type: "string" },
                        description: { type: "string" }
                    },
                    required: ["id", "name", "description"]
                }
            },
            {
                name: "search_skills",
                description: "Search for skills (runbooks) by name or description",
                inputSchema: {
                    type: "object",
                    properties: {
                        query: { type: "string" }
                    },
                    required: ["query"]
                }
            },
            {
                name: "update_skill",
                description: "Update content of an existing skill",
                inputSchema: {
                    type: "object",
                    properties: {
                        id: { type: "string" },
                        content: { type: "string" }
                    },
                    required: ["id", "content"]
                }
            }
        ];
    }

    async searchSkills(query: string) {
        const matches = this.localSearch(query, 10).map(s => ({
            id: s.id,
            name: s.name,
            description: s.description
        }));

        return {
            content: [{
                type: "text",
                text: JSON.stringify({ matches }, null, 2)
            }]
        };
    }

    async listSkills(query?: string) {
        const skills = query ? this.localSearch(query, 5) : Array.from(this.skills.values()).slice(0, 10);
        const skillList = skills.map(s => s.id);

        return {
            content: [{
                type: "text",
                text: JSON.stringify({ skills: skillList, note: query ? "Ranked by relevance" : "Showing first 10" }, null, 2)
            }]
        };
    }

    async readSkill(skillName: string) {
        const skill = this.skills.get(skillName);
        if (!skill) {
            return {
                content: [{ type: "text", text: `Skill '${skillName}' not found.` }]
            };
        }

        return {
            content: [{
                type: "text",
                text: skill.content
            }]
        };
    }

    async createSkill(id: string, name: string, description: string) {
        const targetDir = this.searchPaths[0];
        const skillDir = path.join(targetDir, id);
        const skillFile = path.join(skillDir, 'SKILL.md');

        try {
            await fs.mkdir(skillDir, { recursive: true });

            const content = `---
name: ${name}
description: ${description}
---

# ${name}

${description}

## Instructions
1. ...
`;
            await fs.writeFile(skillFile, content, 'utf-8');
            await this.parseSkill(skillFile);

            return {
                content: [{ type: "text", text: `Created skill '${name}' at ${skillFile}` }]
            };
        } catch (e: unknown) {
            return { content: [{ type: "text", text: `Error creating skill: ${getErrorMessage(e)}` }] };
        }
    }

    async saveSkill(id: string, content: string) {
        const skill = this.skills.get(id);
        if (!skill) {
            return { content: [{ type: "text", text: `Skill '${id}' not found.` }] };
        }

        try {
            await fs.writeFile(skill.path, content, 'utf-8');
            skill.content = content;
            await this.parseSkill(skill.path);

            return { content: [{ type: "text", text: `Saved skill '${id}'.` }] };
        } catch (e: unknown) {
            return { content: [{ type: "text", text: `Error saving skill: ${getErrorMessage(e)}` }] };
        }
    }
}
