import { VectorProvider, Document, SearchResult, Message } from '../interfaces/VectorProvider.js';
import path from 'path';
import fs from 'fs/promises';
import { existsSync } from 'fs';

import { ContextPruner, PruningOptions } from './ContextPruner.js';
import type { GraphMemory } from '@borg/memory';

type JsonRecord = Record<string, unknown>;

interface VectorStoreSearchRow extends JsonRecord {
    id?: string;
    text?: string;
    type?: string;
    namespace?: string;
    _distance?: number;
}

interface VectorStoreDocument extends JsonRecord {
    id: string;
    content: string;
    metadata?: JsonRecord;
    path?: string;
    hash?: string;
}

interface VectorStoreLike {
    initialize(): Promise<void>;
    addMemory(content: string, metadata: JsonRecord): Promise<unknown>;
    search(query: string, limit: number): Promise<VectorStoreSearchRow[]>;
    get(id: string): Promise<VectorStoreDocument | null>;
    delete(ids: string[]): Promise<void>;
    reset(): Promise<void>;
    listDocuments(where?: string, limit?: number): Promise<VectorStoreDocument[]>;
}

interface BorgMemoryModule {
    LanceDBStore: new (dbPath: string) => VectorStoreLike;
    MemoryVectorStore: new () => VectorStoreLike;
    GraphMemory: new () => GraphMemory;
}

interface CodeSplitterModule {
    CodeSplitter: {
        split(content: string, extension: string, maxChunkSize: number): string[];
    };
}

interface IndexerLike {
    indexSymbols?(rootDir: string): Promise<number>;
    indexDirectory(rootDir: string): Promise<number>;
}

interface IndexerModule {
    Indexer: new (store: unknown) => IndexerLike;
    VectorStore: new (dbPath: string) => unknown;
}

interface VectorProviderWithList extends VectorProvider {
    list(where?: string, limit?: number): Promise<Document[]>;
}

function sanitizeMetadataForVectorStore(metadata: JsonRecord): JsonRecord {
    const result: JsonRecord = {};

    for (const [key, value] of Object.entries(metadata)) {
        if (value === null || value === undefined) {
            result[key] = null;
        } else if (Array.isArray(value) || typeof value === 'object') {
            result[key] = JSON.stringify(value);
        } else {
            result[key] = value;
        }
    }

    return result;
}

function isArrowSchemaInferenceError(error: unknown): boolean {
    const text = error instanceof Error
        ? `${error.message}\n${error.stack ?? ''}`
        : String(error ?? '');

    const normalized = text.toLowerCase();
    return normalized.includes('failed to infer data type')
        || normalized.includes('consider providing an explicit schema')
        || normalized.includes('structuredobservation.filesread');
}

export class MemoryManager {
    private provider: VectorProvider | null = null;
    private pruner: ContextPruner;
    private initialized: boolean = false;
    private dbPath: string;
    private registryPath: string;
    public graph: GraphMemory | null = null;

    /**
     * Reason: Metadata across providers is loosely typed and may include non-string primitives.
     * What: Reads a metadata key and safely normalizes it to a lowercase string.
     * Why: Removes repeated ad-hoc casts while keeping symbol filtering behavior unchanged.
     */
    private getMetadataString(metadata: Record<string, unknown> | undefined, key: string): string {
        const value = metadata?.[key];
        return typeof value === 'string' ? value.toLowerCase() : '';
    }

    /**
     * Reason: `list` is an optional extension used by this adapter but not part of the base VectorProvider contract.
     * What: Runtime guard that narrows a provider to the extended list-capable variant.
     * Why: Allows optional capability checks without broad casts.
     */
    private hasList(provider: VectorProvider): provider is VectorProviderWithList {
        const maybeList = Reflect.get(provider as object, 'list');
        return typeof maybeList === 'function';
    }

    constructor(workspaceRoot: string = process.cwd()) {
        this.dbPath = path.join(workspaceRoot, '.borg', 'db');
        this.registryPath = path.join(workspaceRoot, '.borg', 'memory', 'contexts.json');
        this.pruner = new ContextPruner();
    }

    private async initialize() {
        if (this.initialized) return;

        console.log("[MemoryManager] Initializing Vector Backend...");

        // Lazy load the providers from @borg/memory
        // This allows switching to Chroma/SQLite/Memory based on config
        const { LanceDBStore, MemoryVectorStore, GraphMemory } = await import('@borg/memory') as unknown as BorgMemoryModule;

        // Select Backend
        let store: VectorStoreLike;
        const useMemoryStore = process.env.MEMORY_BACKEND === 'memory';

        if (useMemoryStore) {
            console.log("[MemoryManager] Using MemoryVectorStore fallback backend.");
            store = new MemoryVectorStore();
        } else {
            console.log("[MemoryManager] Using LanceDBStore backend.");
            store = new LanceDBStore(this.dbPath);
        }

        this.graph = new GraphMemory();
        await this.graph.initialize();

        const provider: VectorProviderWithList = {
            initialize: async () => store.initialize(),
            add: async (docs: Document[]) => {
                for (const d of docs) {
                    const metadata = (d.metadata ?? {}) as JsonRecord;
                    const pathValue = typeof metadata.path === 'string' ? metadata.path : d.id;
                    const vectorMetadata = {
                        ...metadata,
                        id: d.id,
                        path: pathValue
                    };

                    try {
                        await store.addMemory(d.content, vectorMetadata);
                    } catch (error) {
                        if (!useMemoryStore && isArrowSchemaInferenceError(error)) {
                            console.warn('[MemoryManager] Retrying vector-store write with scalar-safe metadata after Arrow schema inference failure.');
                            await store.addMemory(d.content, sanitizeMetadataForVectorStore(vectorMetadata));
                            continue;
                        }

                        throw error;
                    }
                }
            },
            search: async (query: string, limit: number = 5) => {
                const results = await store.search(query, limit);
                console.log(`[MemoryManager] Raw Search Results:`, JSON.stringify(results, null, 2));
                // LanceDB results: [{ vector, text, ...metadata, _distance }]
                return results.map((r) => ({
                    id: r.id || 'unknown',
                    content: r.text ?? '',
                    metadata: { type: r.type, namespace: r.namespace, ...r },
                    score: 1 - (r._distance || 0) // Convert distance to similarity
                }));
            },
            get: async (id: string) => {
                const doc = await store.get(id);
                if (!doc) return null;
                return {
                    id: doc.id,
                    content: doc.content,
                    metadata: { ...doc.metadata, path: doc.path, hash: doc.hash },
                    score: 1
                };
            },
            delete: async (ids: string[]) => store.delete(ids),
            reset: async () => store.reset(),
            list: async (where?: string, limit?: number) => {
                const docs = await store.listDocuments(where, limit);
                return docs.map((d) => ({
                    id: d.id,
                    content: d.content,
                    metadata: { ...d.metadata, path: d.path, hash: d.hash }
                }));
            }
        };

        this.provider = provider;

        if (this.provider) {
            await this.provider.initialize();
        }

        this.initialized = true;
        console.log("[MemoryManager] Ready.");
    }

    public async saveContext(content: string, metadata: Record<string, unknown> = {}) {
        if (!this.initialized) await this.initialize();
        if (!this.provider) throw new Error("Provider failed to init");

        const docId = `ctx/${Date.now()}/${Math.random().toString(36).substring(7)}`;
        const docs = await this.prepareContextDocuments(docId, content, metadata);

        await this.provider.add(docs);

        await this.addToRegistry({
            id: docId,
            title: metadata.title || 'Untitled',
            source: metadata.source || 'unknown',
            createdAt: Date.now(),
            chunks: docs.length,
            metadata
        });

        return docId;
    }

    public async search(query: string, limit: number = 5): Promise<SearchResult[]> {
        if (!this.initialized) await this.initialize();
        if (!this.provider) return [];

        return await this.provider!.search(query, limit);
    }

    public async searchSymbols(query: string, limit: number = 5): Promise<SearchResult[]> {
        const candidates = await this.search(query, Math.max(limit * 3, limit));
        const symbols = candidates.filter((r) => {
            const metadata = (r.metadata ?? {}) as Record<string, unknown>;
            const type = this.getMetadataString(metadata, 'type');
            const hash = this.getMetadataString(metadata, 'hash');
            return type === 'symbol' || hash === 'symbol';
        });

        if (symbols.length > 0) {
            return symbols.slice(0, limit);
        }

        return candidates.slice(0, limit);
    }

    private async prepareContextDocuments(baseId: string, content: string, metadata: Record<string, unknown>): Promise<Document[]> {
        const pathHint = String(metadata?.path ?? metadata?.source ?? metadata?.title ?? '');
        const extension = path.extname(pathHint) || '.md';
        const looksCode = ['.ts', '.tsx', '.js', '.jsx', '.py', '.rs', '.go', '.java', '.cs', '.cpp', '.c', '.h'].includes(extension);
        const shouldSplit = looksCode || content.length > 2000;

        if (!shouldSplit) {
            return [{ id: baseId, content, metadata }];
        }

        try {
            const { CodeSplitter } = await import('@borg/memory') as unknown as CodeSplitterModule;
            const chunks = CodeSplitter
                .split(content, extension, 1000)
                .map((chunk: string) => chunk.trim())
                .filter((chunk: string) => chunk.length > 0);

            if (chunks.length <= 1) {
                return [{ id: baseId, content, metadata }];
            }

            return chunks.map((chunk: string, index: number) => ({
                id: `${baseId}#${index + 1}`,
                content: chunk,
                metadata: {
                    ...metadata,
                    parentId: baseId,
                    chunkIndex: index,
                    totalChunks: chunks.length,
                },
            }));
        } catch (e) {
            console.warn('[MemoryManager] CodeSplitter unavailable, falling back to single document context save.', e);
            return [{ id: baseId, content, metadata }];
        }
    }

    public async getAllSymbols(): Promise<Document[]> {
        if (!this.initialized) await this.initialize();
        if (!this.provider) return [];

        if (this.hasList(this.provider)) {
            return await this.provider.list("hash = 'symbol'", 5000);
        }

        console.warn("[MemoryManager] Provider does not support list capability.");
        return [];
    }

    /**
     * Index structured symbols (Classes, Functions) using AST parsing.
     */
    public async indexSymbols(rootDir: string): Promise<number> {
        if (!this.initialized) await this.initialize();
        if (!this.provider) throw new Error("Provider failed to init");

        console.log(`[MemoryManager] Indexing symbols at ${rootDir}...`);

        // Lazy load Indexer from @borg/memory
        const { Indexer } = await import('@borg/memory') as unknown as IndexerModule;

        // Adapter to make VectorProvider look like IndexerStorage
        const storageAdapter = {
            initialize: async () => { /* provider assumed initialized */ },
            addDocuments: async (docs: any[]) => {
                await this.provider!.add(docs.map(doc => ({
                    id: doc.id,
                    content: doc.content,
                    metadata: {
                        path: doc.file_path,
                        hash: doc.hash,
                        ...doc.metadata
                    }
                })));
            }
        };

        const indexer = new Indexer(storageAdapter);

        if (!indexer.indexSymbols) {
            throw new Error("Indexer does not utilize indexSymbols (check build?)");
        }

        return await indexer.indexSymbols(rootDir);
    }
    public async indexCodebase(rootDir: string): Promise<number> {
        if (!this.initialized) await this.initialize();
        if (!this.provider) throw new Error("Provider failed to init");

        console.log(`[MemoryManager] Indexing codebase at ${rootDir}...`);

        // Lazy load Indexer from @borg/memory
        const { Indexer } = await import('@borg/memory') as unknown as IndexerModule;

        // Adapter to make VectorProvider look like IndexerStorage
        const storageAdapter = {
            initialize: async () => { /* provider assumed initialized or auto-init */ },
            addDocuments: async (docs: any[]) => {
                await this.provider!.add(docs.map(doc => ({
                    id: doc.id,
                    content: doc.content,
                    metadata: {
                        path: doc.file_path,
                        hash: doc.hash,
                        ...doc.metadata
                    }
                })));
            }
        };

        const indexer = new Indexer(storageAdapter);

        return await indexer.indexDirectory(rootDir);
    }
    public async getContext(id: string) {
        if (!this.initialized) await this.initialize();
        if (!this.provider || !this.provider.get) return null;
        return await this.provider.get(id);
    }

    public async deleteContext(id: string) {
        if (!this.initialized) await this.initialize();
        if (!this.provider) return;

        await this.provider.delete([id]);
        await this.removeFromRegistry(id);
    }

    public async listContexts() {
        try {
            if (!existsSync(this.registryPath)) return [];
            const data = await fs.readFile(this.registryPath, 'utf-8');
            return JSON.parse(data);
        } catch (e) {
            return [];
        }
    }

    private async addToRegistry(entry: any) {
        try {
            const dir = path.dirname(this.registryPath);
            if (!existsSync(dir)) await fs.mkdir(dir, { recursive: true });

            let current = await this.listContexts();
            current.unshift(entry); // Newest first
            await fs.writeFile(this.registryPath, JSON.stringify(current, null, 2));
        } catch (e) {
            console.error("Failed to update memory registry:", e);
        }
    }

    private async removeFromRegistry(id: string) {
        try {
            let current = await this.listContexts();
            current = current.filter((c: unknown) => {
                if (!c || typeof c !== 'object') {
                    return true;
                }
                const candidateId = Reflect.get(c, 'id');
                return candidateId !== id;
            });
            await fs.writeFile(this.registryPath, JSON.stringify(current, null, 2));
        } catch (e) {
            console.error("Failed to update memory registry:", e);
        }
    }

    public async exportMemory(destPath: string): Promise<number> {
        if (!this.initialized) await this.initialize();
        if (!this.provider || !this.hasList(this.provider)) throw new Error("Export requires list capability from provider.");

        console.log(`[MemoryManager] Exporting full memory snapshot to ${destPath}...`);
        const allDocs = await this.provider.list(undefined, 50000);

        const dir = path.dirname(destPath);
        if (!existsSync(dir)) await fs.mkdir(dir, { recursive: true });

        await fs.writeFile(destPath, JSON.stringify(allDocs, null, 2));
        console.log(`[MemoryManager] Exported ${allDocs.length} vectors to ${destPath}`);
        return allDocs.length;
    }

    public async importMemory(srcPath: string): Promise<number> {
        if (!this.initialized) await this.initialize();
        if (!this.provider) throw new Error("Provider not initialized");

        if (!existsSync(srcPath)) {
            console.warn(`[MemoryManager] Import file not found: ${srcPath}`);
            return 0;
        }

        console.log(`[MemoryManager] Importing memory snapshot from ${srcPath}...`);
        const data = await fs.readFile(srcPath, 'utf-8');
        const docs = JSON.parse(data) as Document[];

        if (Array.isArray(docs) && docs.length > 0) {
            await this.provider.add(docs);
            console.log(`[MemoryManager] Imported ${docs.length} vectors.`);
            return docs.length;
        }
        return 0;
    }


    /**
     * Infinite Context V3: Prune a conversation history to fit within limits.
     */
    public pruneContext(messages: Message[], options?: Partial<PruningOptions>): Message[] {
        if (options) {
            // Re-instantiate pruner if temporary options provided
            // Or just use a temporary instance
            const tempPruner = new ContextPruner(options);
            return tempPruner.prune(messages);
        }
        return this.pruner.prune(messages);
    }

    /**
     * Calculate token usage for observability
     */
    public getContextSize(messages: Message[]): number {
        return this.pruner.estimateTokens(messages);
    }
}
