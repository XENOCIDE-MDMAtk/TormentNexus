import { spawn, ChildProcess } from 'child_process';
import { EventEmitter } from 'events';

export interface ProcessConfig {
    sessionId: string;
    command: string;
    args: string[];
    cwd: string;
    env?: Record<string, string>;
}

export interface ProcessOutput {
    sessionId: string;
    data: string;
    type: 'stdout' | 'stderr';
}

export class ProcessManager extends EventEmitter {
    private activeProcesses: Map<string, ChildProcess> = new Map();

    /**
     * Spawns a new process and tracks its output.
     */
    async spawn(config: ProcessConfig): Promise<{ pid: number; success: boolean }> {
        console.log(`[Core:Process] Spawning: ${config.command} ${config.args.join(' ')}`);
        
        try {
            const child = spawn(config.command, config.args, {
                cwd: config.cwd,
                env: { ...process.env, ...config.env },
                shell: true
            });

            if (!child.pid) {
                return { pid: -1, success: false };
            }

            this.activeProcesses.set(config.sessionId, child);

            child.stdout?.on('data', (data) => {
                this.emit('output', {
                    sessionId: config.sessionId,
                    data: data.toString(),
                    type: 'stdout'
                });
            });

            child.stderr?.on('data', (data) => {
                this.emit('output', {
                    sessionId: config.sessionId,
                    data: data.toString(),
                    type: 'stderr'
                });
            });

            child.on('close', (code) => {
                console.log(`[Core:Process] Process ${config.sessionId} exited with code ${code}`);
                this.activeProcesses.delete(config.sessionId);
                this.emit('exit', { sessionId: config.sessionId, code });
            });

            return { pid: child.pid, success: true };
        } catch (error) {
            console.error(`[Core:Process] Failed to spawn ${config.command}:`, error);
            return { pid: -1, success: false };
        }
    }

    /**
     * Writes data to a process's stdin.
     */
    write(sessionId: string, data: string): boolean {
        const child = this.activeProcesses.get(sessionId);
        if (child && child.stdin && child.stdin.writable) {
            child.stdin.write(data);
            return true;
        }
        return false;
    }

    /**
     * Kills an active process.
     */
    kill(sessionId: string): boolean {
        const child = this.activeProcesses.get(sessionId);
        if (child) {
            child.kill();
            this.activeProcesses.delete(sessionId);
            return true;
        }
        return false;
    }

    /**
     * Lists all active process session IDs.
     */
    listActiveSessions(): string[] {
        return Array.from(this.activeProcesses.keys());
    }
}
