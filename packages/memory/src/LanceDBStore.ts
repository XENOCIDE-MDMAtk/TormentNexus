import { connect } from '@lancedb/lancedb';
import { pipeline } from '@xenova/transformers';
import path from 'path';
import fs from 'fs';
import { IVectorStore } from './IVectorStore.js';

function sanitizeMetadataForArrow(metadata: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(metadata)) {
        if (value === null || value === undefined) {
            result[key] = null;
        } else if (Array.isArray(value) || (typeof value === 'object')) {
            result[key] = JSON.stringify(value);
        } else {
            result[key] = value;
        }
    }
    return result;
}

// Global lock for table creation across all instances sharing the same DB path
const tableInitializationLocks = new Map<string, Promise<any>>();

export class LanceDBStore implements IVectorStore {
    private dbPath: string;
    private db: any;
    private embedder: any;
    private readonly HEAT_DECAY_HALFLIFE_MS = 1000 * 60 * 60 * 24;

    constructor(rootPath: string) {
        this.dbPath = path.resolve(rootPath, 'data', 'lancedb');
        if (!fs.existsSync(this.dbPath)) fs.mkdirSync(this.dbPath, { recursive: true });
    }

    async initialize() {
        this.db = await connect(this.dbPath);
        this.embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
    }

    private async ensureTable(initialData?: any[]) {
        const lockKey = this.dbPath;
        let promise = tableInitializationLocks.get(lockKey);

        if (!promise) {
            promise = (async () => {
                try {
                    return await this.db.openTable('memories');
                } catch (e) {
                    if (initialData && initialData.length > 0) {
                        try {
                            return await this.db.createTable('memories', initialData);
                        } catch (createErr: any) {
                            if (createErr.message?.includes('already exists')) {
                                return await this.db.openTable('memories');
                            }
                            throw createErr;
                        }
                    }
                    throw e;
                }
            })();
            tableInitializationLocks.set(lockKey, promise);
        }

        return promise;
    }

    async createEmbeddings(text: string): Promise<number[]> {
        const output = await this.embedder(text, { pooling: 'mean', normalize: true });
        return Array.from(output.data);
    }

    async addMemory(content: string, metadata: any) {
        const vector = await this.createEmbeddings(content);
        const sanitizedMetadata = sanitizeMetadataForArrow(metadata);
        const data = [{ 
            vector, 
            text: content, 
            ...sanitizedMetadata, 
            heat_score: metadata.heat_score ?? 50, 
            last_accessed_at: Date.now(), 
            timestamp: Date.now() 
        }];
        const table = await this.ensureTable(data);
        try {
            await table.add(data);
        } catch (err: any) {
            if (err.message?.includes('not in schema')) {
                // If schema mismatch, we might need to be more careful.
                // For now, let's just log and skip or we could drop and recreate (expensive).
                console.error('[LanceDBStore] Schema mismatch, skipping entry:', err.message);
            } else {
                throw err;
            }
        }
    }

    async addDocuments(docs: any[]) {
        const processed = await Promise.all(docs.map(async d => {
            const sanitizedMetadata = sanitizeMetadataForArrow(d.metadata || {});
            return {
                ...d, 
                ...sanitizedMetadata,
                vector: d.vector || await this.createEmbeddings(d.text || d.content),
                heat_score: d.heat_score ?? 50, 
                last_accessed_at: d.last_accessed_at ?? Date.now(), 
                timestamp: d.timestamp || Date.now()
            };
        }));
        const table = await this.ensureTable(processed);
        try {
            await table.add(processed);
        } catch (err: any) {
            if (err.message?.includes('not in schema')) {
                console.error('[LanceDBStore] Schema mismatch in batch, skipping:', err.message);
            } else {
                throw err;
            }
        }
    }

    async get(id: string) {
        try {
            const table = await this.ensureTable();
            const res = await table.search(await this.createEmbeddings('')).where(`id = '${id}'`).limit(1).toArray();
            return res.length > 0 ? res[0] : null;
        } catch (e) { return null; }
    }

    async delete(ids: string[]) {
        const table = await this.ensureTable();
        await table.delete(ids.map(id => `id = '${id}'`).join(' OR '));
    }

    async reset() { 
        tableInitializationLocks.delete(this.dbPath);
        await this.db.dropTable('memories'); 
    }

    async listDocuments(where?: string, limit: number = 100) {
        const table = await this.ensureTable();
        let q = table.search(await this.createEmbeddings('query')).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async search(query: string, limit: number = 5, where?: string) {
        const table = await this.ensureTable();
        let q = table.search(await this.createEmbeddings(query)).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async maintenance() {
        const table = await this.ensureTable();
        const all = await table.search(await this.createEmbeddings('')).limit(10000).toArray();
        const now = Date.now();
        const updates = all.map((item: any) => {
            const elapsed = now - (item.last_accessed_at || item.timestamp);
            const decay = Math.pow(0.5, elapsed / this.HEAT_DECAY_HALFLIFE_MS);
            return { ...item, heat_score: (item.heat_score || 50) * decay };
        });

        const maintenancePromise = (async () => {
            await this.db.dropTable('memories');
            return await this.db.createTable('memories', updates);
        })();

        tableInitializationLocks.set(this.dbPath, maintenancePromise);
        await maintenancePromise;
    }
}
