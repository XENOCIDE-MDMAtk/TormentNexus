import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { SSEClientTransport } from "@modelcontextprotocol/sdk/client/sse.js";
import type { MCPServerConfig } from "./types.js";

export class SSEClient {
    private client: Client | null = null;
    private transport: SSEClientTransport | null = null;
    public readonly name: string;
    private config: MCPServerConfig;

    constructor(name: string, config: MCPServerConfig) {
        this.name = name;
        this.config = config;
    }

    public async connect(): Promise<void> {
        if (!this.config.url) {
            throw new Error(`[SSEClient:${this.name}] Cannot connect: missing URL.`);
        }

        console.log(`[SSEClient:${this.name}] Connecting to ${this.config.url}...`);

        const headers: Record<string, string> = { ...(this.config.headers || {}) };
        if (this.config.bearerToken) {
            headers.Authorization = `Bearer ${this.config.bearerToken}`;
        }

        const hasHeaders = Object.keys(headers).length > 0;

        if (hasHeaders) {
            // Since the standard EventSource polyfill provided by the SDK doesn't natively support custom headers,
            // we configure the requestInit to pass headers for the HTTP POST requests (tool calls)
            // and optionally rely on native platform EventSource behavior if it allows header injection or cookie auth.
            // A common workaround for node environments is passing `headers` in eventSourceInit if the underlying polyfill (like eventsource package) supports it.
            // If the @modelcontextprotocol/sdk explicitly doesn't type 'headers' in eventSourceInit, we cast it to any.
            this.transport = new SSEClientTransport(new URL(this.config.url), {
                eventSourceInit: {
                    headers: headers
                } as any,
                requestInit: { headers }
            });
        } else {
            this.transport = new SSEClientTransport(new URL(this.config.url));
        }

        this.client = new Client(
            { name: "hypernexus-client", version: "1.0.0" },
            { capabilities: {} }
        );

        try {
            await this.client.connect(this.transport);
            console.log(`[SSEClient:${this.name}] Connected!`);
        } catch (error) {
            console.error(`[SSEClient:${this.name}] Connection failed:`, error);
            throw error;
        }
    }

    public async listTools() {
        if (!this.client) throw new Error(`[SSEClient:${this.name}] Not connected.`);
        const response = await this.client.listTools();
        return response.tools || [];
    }

    public async callTool(toolName: string, args: unknown) {
        if (!this.client) throw new Error(`[SSEClient:${this.name}] Not connected.`);
        return await this.client.callTool({ name: toolName, arguments: args as any });
    }

    public async close(): Promise<void> {
        if (this.client) {
            await this.client.close();
            this.client = null;
            this.transport = null;
            console.log(`[SSEClient:${this.name}] Closed.`);
        }
    }

    public async listPrompts() {
        if (!this.client) throw new Error(`[SSEClient:${this.name}] Not connected.`);
        const response = await this.client.listPrompts();
        return response.prompts || [];
    }
    public async listResources() {
        if (!this.client) throw new Error(`[SSEClient:${this.name}] Not connected.`);
        const response = await this.client.listResources();
        return response.resources || [];
    }
    public async ping() {
        if (!this.client) throw new Error(`[SSEClient:${this.name}] Not connected.`);
        await this.client.ping();
        return true;
    }
}
