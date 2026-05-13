// @borg/memory -- stub with vector store types

export interface GraphEdge { id?: string; source: string; target: any; type?: string; relation?: string; [key: string]: any; }
export interface GraphNode { id: string; type: string; data: Record<string, any>; [key: string]: any; }
export interface GraphMemory { nodes: GraphNode[]; edges: GraphEdge[]; initialize(): Promise<void>; [key: string]: any; }
export interface LanceDBStore { initialize(): Promise<void>; add(docs: any[]): Promise<void>; search(query: any, limit?: number): Promise<any[]>; delete(ids: string[]): Promise<void>; list(): Promise<any[]>; }
export interface MemoryVectorStore { initialize(): Promise<void>; add(docs: any[]): Promise<void>; search(query: any, limit?: number): Promise<any[]>; delete(ids: string[]): Promise<void>; list(): Promise<any[]>; }

export const GraphEdge: any = undefined;
export const GraphMemory: any = undefined;
export const GraphNode: any = undefined;
export const LanceDBStore: any = undefined;
export const MemoryVectorStore: any = undefined;
