
function redirectProtocolUnsafeConsoleMethods(): void {
    const stderr = console.error.bind(console);

    console.log = stderr;
    console.info = stderr;
    console.debug = stderr;
    console.trace = stderr;
    console.dir = ((...args: unknown[]) => stderr(...args)) as typeof console.dir;
}

async function main() {
    // MCP stdio requires stdout to remain pristine JSON-RPC output only.
    redirectProtocolUnsafeConsoleMethods();

    process.on('unhandledRejection', (reason) => {
        console.error('[Hypercode Core] Unhandled promise rejection:', reason);
    });

    process.on('uncaughtException', (error) => {
        console.error('[Hypercode Core] Uncaught exception:', error);
        process.exit(1);
    });

    try {
        const { ensureBackgroundCoreRunning } = await import('./backgroundCoreBootstrap.js');
        void ensureBackgroundCoreRunning({
            waitForReady: false,
            log: (message, ...optionalParams) => console.error(message, ...optionalParams),
        }).then((result) => {
            if (result.status === 'spawned') {
                console.error(`[Hypercode Core] Background control-plane bootstrap requested (PID: ${result.pid ?? 'unknown'}).`);
            }
        }).catch((error) => {
            console.error('[Hypercode Core] Background control-plane bootstrap failed:', error);
        });

        const { MCPServer } = await import('./MCPServer.js');
        const mcp = new MCPServer({ skipWebsocket: true });
        await mcp.start();
        console.error("[Hypercode Core] MCP Server Stdio Entry Point Started.");
    } catch (err) {
        console.error("Failed to start MCP server:", err);
        process.exit(1);
    }
}

main();
