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

export class LanceDBStore implements IVectorStore {
    private dbPath: string;
    private db: any;
    private embedder: any;
    private readonly HEAT_DECAY_HALFLIFE_MS = 1000 * 60 * 60 * 24;

    constructor(rootPath: string) {
        this.dbPath = path.join(rootPath, 'data', 'lancedb');
        if (!fs.existsSync(this.dbPath)) fs.mkdirSync(this.dbPath, { recursive: true });
    }

    async initialize() {
        this.db = await connect(this.dbPath);
        this.embedder = await pipeline('feature-extraction', 'Xenova/all-MiniLM-L6-v2');
    }

    async createEmbeddings(text: string): Promise<number[]> {
        const output = await this.embedder(text, { pooling: 'mean', normalize: true });
        return Array.from(output.data);
    }

    async addMemory(content: string, metadata: any) {
        const vector = await this.createEmbeddings(content);
        const data = [{ vector, text: content, ...metadata, heat_score: metadata.heat_score ?? 50, last_accessed_at: Date.now(), timestamp: Date.now() }];
        try {
            const table = await this.db.openTable('memories');
            await table.add(data);
        } catch (e) {
            await this.db.createTable('memories', data);
        }
    }

    async addDocuments(docs: any[]) {
        const processed = await Promise.all(docs.map(async d => ({
            ...d, vector: d.vector || await this.createEmbeddings(d.text || d.content),
            heat_score: d.heat_score ?? 50, last_accessed_at: d.last_accessed_at ?? Date.now(), timestamp: d.timestamp || Date.now()
        })));
        try {
            const table = await this.db.openTable('memories');
            await table.add(processed);
        } catch (e) {
            await this.db.createTable('memories', processed);
        }
    }

    async get(id: string) {
        try {
            const table = await this.db.openTable('memories');
            const res = await table.search(await this.createEmbeddings('')).where(`id = '${id}'`).limit(1).toArray();
            return res.length > 0 ? res[0] : null;
        } catch (e) { return null; }
    }

    async delete(ids: string[]) {
        const table = await this.db.openTable('memories');
        await table.delete(ids.map(id => `id = '${id}'`).join(' OR '));
    }

    async reset() { await this.db.dropTable('memories'); }

    async listDocuments(where?: string, limit: number = 100) {
        const table = await this.db.openTable('memories');
        let q = table.search(await this.createEmbeddings('query')).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async search(query: string, limit: number = 5, where?: string) {
        const table = await this.db.openTable('memories');
        let q = table.search(await this.createEmbeddings(query)).limit(limit);
        if (where) q = q.where(where);
        return await q.toArray();
    }

    async maintenance() {
        const table = await this.db.openTable('memories');
        const all = await table.search(await this.createEmbeddings('')).limit(10000).toArray();
        const now = Date.now();
        const updates = all.map((item: any) => {
            const elapsed = now - (item.last_accessed_at || item.timestamp);
            const decay = Math.pow(0.5, elapsed / this.HEAT_DECAY_HALFLIFE_MS);
            return { ...item, heat_score: (item.heat_score || 50) * decay };
        });
        await this.db.dropTable('memories');
        await this.db.createTable('memories', updates);
    }
}
