import { EventEmitter } from 'events';
import crypto from 'crypto';
import { Redis } from 'ioredis';
import { SwarmMessageSchema, type SwarmMessageMetadata as ProtocolSwarmMessageMetadata } from './protocol.js';

export enum SwarmMessageType {
    CAPABILITY_QUERY = 'CAPABILITY_QUERY',
    CAPABILITY_RESPONSE = 'CAPABILITY_RESPONSE',
    TASK_OFFER = 'TASK_OFFER',
    TASK_ACCEPT = 'TASK_ACCEPT',
    TASK_RESULT = 'TASK_RESULT',
    VERIFY_OFFER = 'VERIFY_OFFER',
    VERIFY_RESULT = 'VERIFY_RESULT',
    HEARTBEAT = 'HEARTBEAT',
    DIRECT_MESSAGE = 'DIRECT_MESSAGE',
    TASK_BID = 'TASK_BID',
    TASK_ASSIGN = 'TASK_ASSIGN',
    ARTIFACT_READ_REQUEST = 'ARTIFACT_READ_REQUEST',
    ARTIFACT_READ_RESPONSE = 'ARTIFACT_READ_RESPONSE',
    A2A_BRIDGE_SIGNAL = 'A2A_BRIDGE_SIGNAL'
}

export interface SwarmMessage {
    version?: '1.0';
    id: string;
    correlationId?: string;
    sender: string;
    target?: string;
    type: SwarmMessageType;
    payload: unknown;
    timestamp: number;
    metadata?: ProtocolSwarmMessageMetadata;
}

interface PendingMeshRequest {
    resolve: (message: SwarmMessage) => void;
    reject: (error: Error) => void;
    timer: NodeJS.Timeout;
    responseTypes?: Set<SwarmMessageType>;
    targetNodeId: string;
}

/**
 * MeshService
 * 
 * Provides a decentralized P2P event bus for agents to communicate across nodes.
 * In a real distributed system, this would be backed by Redis Pub/Sub, NATS, or WebRTC.
 * For Phase 77, it uses a global EventEmitter to simulate a mesh network within the same process,
 * laying the groundwork for true cross-process distribution.
 */

// A global singleton emitter to bridge multiple MeshService instances in the same Node process
const globalMeshBus = new EventEmitter();
globalMeshBus.setMaxListeners(100);

let redisSubscriber: Redis | null = null;
let redisPublisher: Redis | null = null;
const REDIS_CHANNEL = 'hypercode:swarm:mesh';

// Initialize Redis if REDIS_URL is provided in the environment
if (process.env.REDIS_URL) {
    try {
        redisSubscriber = new Redis(process.env.REDIS_URL, { lazyConnect: true });
        redisPublisher = new Redis(process.env.REDIS_URL, { lazyConnect: true });

        Promise.all([redisSubscriber.connect(), redisPublisher.connect()]).then(() => {
            console.log('[MeshService] 🟢 Connected to Redis Mesh Bus');

            redisSubscriber!.subscribe(REDIS_CHANNEL, (err: any) => {
                if (err) console.error('[MeshService] 🔴 Redis Subscribe Error:', err);
            });

            // Listen for mesh messages from the wider distributed swarm
            redisSubscriber!.on('message', (channel: string, message: string) => {
                if (channel === REDIS_CHANNEL) {
                    try {
                        const parsed: SwarmMessage = JSON.parse(message);
                        // Forward the remote message to all local MeshService instances
                        globalMeshBus.emit('mesh_message_inbound', parsed);
                    } catch (e: any) {
                        console.error('[MeshService] Failed to parse Redis message', e);
                    }
                }
            });

            // Listen to messages emitted by local MeshService instances and blast them to the wider swarm
            globalMeshBus.on('mesh_message_outbound', (msg: SwarmMessage) => {
                redisPublisher!.publish(REDIS_CHANNEL, JSON.stringify(msg)).catch((e: any) => {
                    console.error('[MeshService] Failed to publish message to Redis', e);
                });
            });

        }).catch((err: any) => {
            console.error('[MeshService] 🟡 Redis connection failed, falling back to local-only Mesh.', err.message);
            setupLocalFallback();
        });
    } catch (e: any) {
        console.warn(`[MeshService] 🟡 Redis initialization error: ${e.message}`);
        setupLocalFallback();
    }
} else {
    // If no Redis URL is given, run entirely locally across the single Node process
    setupLocalFallback();
}

/**
 * Routes outbound emissions straight back into inbound listeners.
 * Emulates the mesh network internally if there is no Redis backend present.
 */
function setupLocalFallback() {
    globalMeshBus.on('mesh_message_outbound', (msg: SwarmMessage) => {
        globalMeshBus.emit('mesh_message_inbound', msg);
    });
}

// Forward all mesh traffic to the central EventBus for SSE Visualization
globalMeshBus.on('mesh_message_inbound', (msg: SwarmMessage) => {
    if (global.mcpServerInstance?.eventBus) {
        global.mcpServerInstance.eventBus.emitEvent('mesh:traffic', 'MeshService', msg);
    }
});

export class MeshService extends EventEmitter {
    public readonly nodeId: string;
    private readonly knownNodes: Set<string> = new Set();
    private readonly nodeCapabilities: Map<string, string[]> = new Map();
    private heartbeatInterval?: NodeJS.Timeout;
    private readonly inboundHandler: (msg: SwarmMessage) => void;
    private readonly pendingRequests = new Map<string, PendingMeshRequest>();

    constructor() {
        super();
        this.nodeId = crypto.randomUUID();
        this.inboundHandler = this.handleGlobalMessage.bind(this);

        // Listen to global inbound network traffic
        globalMeshBus.on('mesh_message_inbound', this.inboundHandler);

        this.startHeartbeat();
    }

    private startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            // Include local capabilities in heartbeat (Phase 80)
            const localTools = this.getLocalCapabilities();
            this.broadcast(SwarmMessageType.HEARTBEAT, {
                status: 'alive',
                capabilities: localTools
            });
        }, 15000);
    }

    private getLocalCapabilities(): string[] {
        // Query the local MCPServer for registered tools if available
        if (global.mcpServerInstance) {
            // This is a simplified list of tool names
            // In a better implementation, we'd pull the full schema from the registry
            return ['fs', 'terminal', 'git', 'research', 'coder'];
        }
        return [];
    }

    private handleGlobalMessage(msg: SwarmMessage) {
        const parsed = SwarmMessageSchema.safeParse(msg);
        if (!parsed.success) {
            return;
        }

        const normalizedMsg = parsed.data as SwarmMessage;

        // Ignore our own broadcast messages
        if (normalizedMsg.sender === this.nodeId) return;

        // Keep track of known peers
        this.knownNodes.add(normalizedMsg.sender);

        // Update capabilities registry on heartbeat
        if (normalizedMsg.type === SwarmMessageType.HEARTBEAT) {
            const payload = normalizedMsg.payload as any;
            if (payload?.capabilities) {
                this.nodeCapabilities.set(normalizedMsg.sender, payload.capabilities);
            }
        }

        const correlationKey = normalizedMsg.correlationId ?? normalizedMsg.id;
        const pending = this.pendingRequests.get(correlationKey);
        if (pending && (!normalizedMsg.target || normalizedMsg.target === this.nodeId)) {
            if (!pending.responseTypes || pending.responseTypes.has(normalizedMsg.type)) {
                clearTimeout(pending.timer);
                this.pendingRequests.delete(correlationKey);
                pending.resolve(normalizedMsg);
            }
        }

        // If it's a direct message tailored to us, or a broadcast (no target)
        if (!normalizedMsg.target || normalizedMsg.target === this.nodeId) {
            this.emit('message', normalizedMsg);
        }
    }

    private buildMessage(input: {
        type: SwarmMessageType;
        payload: unknown;
        target?: string;
        id?: string;
        correlationId?: string;
        metadata?: ProtocolSwarmMessageMetadata;
    }): SwarmMessage {
        const message: SwarmMessage = {
            version: '1.0',
            id: input.id ?? crypto.randomUUID(),
            correlationId: input.correlationId,
            sender: this.nodeId,
            target: input.target,
            type: input.type,
            payload: input.payload,
            timestamp: Date.now(),
            metadata: input.metadata,
        };
        return SwarmMessageSchema.parse(message) as SwarmMessage;
    }

    public broadcast(type: SwarmMessageType, payload: unknown, metadata?: ProtocolSwarmMessageMetadata) {
        const msg = this.buildMessage({
            type,
            payload,
            metadata,
        });
        globalMeshBus.emit('mesh_message_outbound', msg);
    }

    public sendDirect(
        targetNodeId: string,
        type: SwarmMessageType,
        payload: unknown,
        metadata?: ProtocolSwarmMessageMetadata,
    ) {
        const msg = this.buildMessage({
            target: targetNodeId,
            type,
            payload,
            metadata,
        });
        globalMeshBus.emit('mesh_message_outbound', msg);
    }

    public sendResponse(originalMsg: SwarmMessage, type: SwarmMessageType, payload: unknown) {
        // Respond directly to the sender of the original message
        const response = this.buildMessage({
            id: originalMsg.id,
            correlationId: originalMsg.correlationId ?? originalMsg.id,
            target: originalMsg.sender,
            type,
            payload,
            metadata: originalMsg.metadata,
        });
        globalMeshBus.emit('mesh_message_outbound', response);
    }

    public async request(
        targetNodeId: string,
        type: SwarmMessageType,
        payload: unknown,
        options: {
            timeoutMs?: number;
            responseTypes?: SwarmMessageType[];
            metadata?: ProtocolSwarmMessageMetadata;
        } = {},
    ): Promise<SwarmMessage> {
        const requestId = crypto.randomUUID();
        const correlationId = requestId;
        const timeoutMs = options.timeoutMs ?? 30_000;
        const responseTypes = options.responseTypes ? new Set(options.responseTypes) : undefined;

        return await new Promise<SwarmMessage>((resolve, reject) => {
            const timer = setTimeout(() => {
                this.pendingRequests.delete(correlationId);
                reject(new Error(`Mesh request timed out after ${timeoutMs}ms (${type} -> ${targetNodeId})`));
            }, timeoutMs);

            this.pendingRequests.set(correlationId, {
                resolve,
                reject,
                timer,
                responseTypes,
                targetNodeId,
            });

            const request = this.buildMessage({
                id: requestId,
                correlationId,
                target: targetNodeId,
                type,
                payload,
                metadata: options.metadata,
            });
            globalMeshBus.emit('mesh_message_outbound', request);
        });
    }

    public getPeers(): string[] {
        return Array.from(this.knownNodes);
    }

    /**
     * Returns a map of all known nodes and their advertised capabilities.
     * Phase 80: Dynamic capability discovery.
     */
    public getMeshCapabilities(): Record<string, string[]> {
        const result: Record<string, string[]> = {};
        this.nodeCapabilities.forEach((caps, id) => {
            result[id] = caps;
        });
        // Also include our own
        result[this.nodeId] = this.getLocalCapabilities();
        return result;
    }

    public destroy() {
        if (this.heartbeatInterval) clearInterval(this.heartbeatInterval);
        for (const pending of this.pendingRequests.values()) {
            clearTimeout(pending.timer);
            pending.reject(new Error('MeshService destroyed before request completed.'));
        }
        this.pendingRequests.clear();
        globalMeshBus.off('mesh_message_inbound', this.inboundHandler);
        this.removeAllListeners();
    }
}
