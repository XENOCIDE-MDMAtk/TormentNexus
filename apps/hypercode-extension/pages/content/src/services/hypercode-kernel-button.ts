/**
 * Hypercode Kernel Button Injection Service
 *
 * Detects supported AI chat websites (Claude.ai, ChatGPT) and injects
 * a "Hypercode Kernel" button that attaches the local Hypercode Kernel to the
 * active conversation session.
 *
 * The button provides:
 * - One-click attachment to the local Hypercode Kernel (port 4300)
 * - Visual status indicator (connected/disconnected/warming)
 * - Context bridge: sends conversation context to Hypercode for memory & healing
 */

import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('HypercodeKernelButton');

// ─── Types ───────────────────────────────────────────────────────────────────

export type HypercodeKernelStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export interface HypercodeKernelConfig {
  kernelUrl: string;
  bridgeUrl: string;
  autoAttach: boolean;
  pollIntervalMs: number;
}

export interface HypercodeKernelButtonState {
  status: HypercodeKernelStatus;
  kernelUrl: string;
  sessionId: string | null;
  connectedAt: number | null;
  error: string | null;
}

// ─── Constants ───────────────────────────────────────────────────────────────

const HYPERCODE_KERNEL_BUTTON_ID = 'hypercode-hypercode-kernel-btn';
const HYPERCODE_KERNEL_STATUS_ID = 'hypercode-hypercode-kernel-status';
const HYPERCODE_KERNEL_PANEL_ID = 'hypercode-hypercode-kernel-panel';

const DEFAULT_CONFIG: HypercodeKernelConfig = {
  kernelUrl: 'http://127.0.0.1:4300',
  bridgeUrl: 'http://127.0.0.1:4100',
  autoAttach: false,
  pollIntervalMs: 5000,
};

// Supported websites with their injection selectors
const SUPPORTED_SITES = {
  'claude.ai': {
    name: 'Claude',
    buttonAnchor: '[data-testid="model-selector"], .claude-header, header',
    conversationSelector: '[data-testid="conversation-turn"], .conversation-turn',
    inputSelector: '[contenteditable="true"], textarea, [data-testid="chat-input"]',
  },
  'chatgpt.com': {
    name: 'ChatGPT',
    buttonAnchor: '#__next header, nav, [class*="header"], [class*="nav"]',
    conversationSelector: '[data-testid="conversation-turn"], [class*="conversation"], [class*="message"]',
    inputSelector: '#prompt-textarea, textarea[placeholder], [contenteditable="true"]',
  },
  'chat.openai.com': {
    name: 'ChatGPT',
    buttonAnchor: '#__next header, nav',
    conversationSelector: '[class*="conversation"], [class*="message"]',
    inputSelector: '#prompt-textarea, textarea, [contenteditable="true"]',
  },
} as const;

type SupportedSite = keyof typeof SUPPORTED_SITES;

// ─── Button Styles ───────────────────────────────────────────────────────────

const BUTTON_STYLES = `
  #${HYPERCODE_KERNEL_BUTTON_ID} {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 8px;
    border: 1px solid rgba(56, 189, 248, 0.3);
    background: rgba(56, 189, 248, 0.08);
    color: #7dd3fc;
    font-size: 12px;
    font-weight: 600;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
    user-select: none;
    letter-spacing: 0.02em;
    z-index: 9999;
  }
  #${HYPERCODE_KERNEL_BUTTON_ID}:hover {
    background: rgba(56, 189, 248, 0.15);
    border-color: rgba(56, 189, 248, 0.5);
    color: #38bdf8;
  }
  #${HYPERCODE_KERNEL_BUTTON_ID}.hypercode-connected {
    border-color: rgba(52, 211, 153, 0.4);
    background: rgba(52, 211, 153, 0.08);
    color: #6ee7b7;
  }
  #${HYPERCODE_KERNEL_BUTTON_ID}.hypercode-connected:hover {
    background: rgba(52, 211, 153, 0.15);
    border-color: rgba(52, 211, 153, 0.5);
  }
  #${HYPERCODE_KERNEL_BUTTON_ID}.hypercode-error {
    border-color: rgba(248, 113, 113, 0.4);
    background: rgba(248, 113, 113, 0.08);
    color: #fca5a5;
  }
  #${HYPERCODE_KERNEL_STATUS_ID} {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #64748b;
    flex-shrink: 0;
    transition: background 0.2s;
  }
  .hypercode-connected #${HYPERCODE_KERNEL_STATUS_ID} {
    background: #34d399;
    box-shadow: 0 0 6px rgba(52, 211, 153, 0.5);
    animation: hypercode-pulse 2s ease-in-out infinite;
  }
  .hypercode-error #${HYPERCODE_KERNEL_STATUS_ID} {
    background: #f87171;
  }
  .hypercode-connecting #${HYPERCODE_KERNEL_STATUS_ID} {
    background: #fbbf24;
    animation: hypercode-blink 1s ease-in-out infinite;
  }
  @keyframes hypercode-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.6; }
  }
  @keyframes hypercode-blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
  #${HYPERCODE_KERNEL_PANEL_ID} {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 8px;
    width: 280px;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid rgba(56, 189, 248, 0.2);
    background: rgba(15, 23, 42, 0.95);
    backdrop-filter: blur(12px);
    color: #e2e8f0;
    font-size: 12px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    z-index: 10000;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  }
`;

// ─── Hypercode Kernel Button Service ─────────────────────────────────────────────

export class HypercodeKernelButtonService {
  private config: HypercodeKernelConfig;
  private state: HypercodeKernelButtonState;
  private button: HTMLElement | null = null;
  private panel: HTMLElement | null = null;
  private pollTimer: ReturnType<typeof setInterval> | null = null;
  private siteConfig: (typeof SUPPORTED_SITES)[SupportedSite] | null = null;

  constructor(config: Partial<HypercodeKernelConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.state = {
      status: 'disconnected',
      kernelUrl: this.config.kernelUrl,
      sessionId: null,
      connectedAt: null,
      error: null,
    };
  }

  /** Detect if the current page is a supported site */
  detectSite(): SupportedSite | null {
    const hostname = window.location.hostname.replace(/^www\./, '');
    if (hostname in SUPPORTED_SITES) {
      return hostname as SupportedSite;
    }
    return null;
  }

  /** Initialize the button injection */
  async initialize(): Promise<boolean> {
    const site = this.detectSite();
    if (!site) {
      logger.debug('HypercodeKernel: unsupported site, skipping injection');
      return false;
    }

    this.siteConfig = SUPPORTED_SITES[site];
    logger.info(`HypercodeKernel: detected ${SUPPORTED_SITES[site].name}, injecting button`);

    // Inject styles
    this.injectStyles();

    // Wait for the anchor element to appear
    const anchor = await this.waitForElement(this.siteConfig.buttonAnchor, 10000);
    if (!anchor) {
      logger.warn('HypercodeKernel: could not find button anchor element');
      return false;
    }

    // Create and inject the button
    this.createButton(anchor);

    // Start health polling
    this.startPolling();

    // If auto-attach, connect immediately
    if (this.config.autoAttach) {
      this.attach();
    }

    return true;
  }

  /** Inject the button styles into the page */
  private injectStyles(): void {
    if (document.getElementById('hypercode-hypercode-kernel-styles')) return;

    const style = document.createElement('style');
    style.id = 'hypercode-hypercode-kernel-styles';
    style.textContent = BUTTON_STYLES;
    document.head.appendChild(style);
  }

  /** Create and inject the Hypercode Kernel button */
  private createButton(anchor: Element): void {
    if (document.getElementById(HYPERCODE_KERNEL_BUTTON_ID)) return;

    const btn = document.createElement('button');
    btn.id = HYPERCODE_KERNEL_BUTTON_ID;
    btn.type = 'button';
    btn.title = 'Attach to Hypercode Kernel';

    btn.innerHTML = `
      <span id="${HYPERCODE_KERNEL_STATUS_ID}"></span>
      <span>Hypercode Kernel</span>
    `;

    btn.addEventListener('click', (e) => {
      e.preventDefault();
      e.stopPropagation();
      this.handleClick();
    });

    // Insert the button near the anchor
    // Try to append to the anchor's parent or after the anchor
    const container = anchor.parentElement || anchor;
    container.style.position = 'relative';
    container.appendChild(btn);
    this.button = btn;

    logger.info('HypercodeKernel: button injected');
  }

  /** Handle button click — toggle attach/detach or show panel */
  private handleClick(): void {
    if (this.state.status === 'connected') {
      this.showPanel();
    } else {
      this.attach();
    }
  }

  /** Attach to the local Hypercode Kernel */
  async attach(): Promise<void> {
    this.setState({ status: 'connecting', error: null });

    try {
      const response = await fetch(`${this.config.kernelUrl}/api/native/status`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      });

      if (!response.ok) {
        throw new Error(`Kernel returned ${response.status}`);
      }

      const data = await response.json();

      this.setState({
        status: 'connected',
        sessionId: data.sessionId || this.generateSessionId(),
        connectedAt: Date.now(),
        error: null,
      });

      logger.info('HypercodeKernel: attached to kernel', data);

      // Register this chat session with the kernel
      await this.registerSession();
    } catch (err: any) {
      this.setState({
        status: 'error',
        error: err.message || 'Connection failed',
      });
      logger.warn('HypercodeKernel: attach failed', err.message);
    }
  }

  /** Detach from the kernel */
  detach(): void {
    this.setState({
      status: 'disconnected',
      sessionId: null,
      connectedAt: null,
      error: null,
    });
    logger.info('HypercodeKernel: detached');
  }

  /** Register this session with the Hypercode Kernel */
  private async registerSession(): Promise<void> {
    if (!this.state.sessionId) return;

    try {
      const site = this.detectSite();
      const payload = {
        sessionId: this.state.sessionId,
        source: site ? SUPPORTED_SITES[site].name : 'unknown',
        url: window.location.href,
        attachedAt: Date.now(),
      };

      await fetch(`${this.config.kernelUrl}/api/native/session/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: AbortSignal.timeout(3000),
      });
    } catch (err: any) {
      logger.debug('HypercodeKernel: session registration failed (non-critical)', err.message);
    }
  }

  /** Show the info panel */
  private showPanel(): void {
    if (this.panel) {
      this.panel.remove();
      this.panel = null;
      return;
    }

    const panel = document.createElement('div');
    panel.id = HYPERCODE_KERNEL_PANEL_ID;

    const uptime = this.state.connectedAt
      ? Math.round((Date.now() - this.state.connectedAt) / 1000)
      : 0;
    const uptimeStr = uptime >= 60 ? `${Math.floor(uptime / 60)}m ${uptime % 60}s` : `${uptime}s`;

    panel.innerHTML = `
      <div style="margin-bottom: 8px; font-weight: 700; color: #38bdf8;">
        Hypercode Kernel Connected
      </div>
      <div style="color: #94a3b8; line-height: 1.6;">
        <div>Session: <span style="color: #e2e8f0;">${this.state.sessionId?.slice(0, 12)}...</span></div>
        <div>Uptime: <span style="color: #e2e8f0;">${uptimeStr}</span></div>
        <div>Kernel: <span style="color: #e2e8f0;">${this.state.kernelUrl}</span></div>
      </div>
      <div style="margin-top: 10px; display: flex; gap: 6px;">
        <button id="hypercode-detach-btn" style="
          flex: 1; padding: 6px; border-radius: 6px;
          border: 1px solid rgba(248,113,113,0.3);
          background: rgba(248,113,113,0.08);
          color: #fca5a5; font-size: 11px; cursor: pointer;
        ">Detach</button>
        <button id="hypercode-reconnect-btn" style="
          flex: 1; padding: 6px; border-radius: 6px;
          border: 1px solid rgba(56,189,248,0.3);
          background: rgba(56,189,248,0.08);
          color: #7dd3fc; font-size: 11px; cursor: pointer;
        ">Reconnect</button>
      </div>
    `;

    this.button?.appendChild(panel);
    this.panel = panel;

    // Wire up panel buttons
    document.getElementById('hypercode-detach-btn')?.addEventListener('click', () => {
      this.detach();
      this.showPanel(); // Toggle panel off
    });

    document.getElementById('hypercode-reconnect-btn')?.addEventListener('click', () => {
      this.detach();
      this.attach();
      this.showPanel(); // Toggle panel off
    });

    // Close panel when clicking outside
    const closePanel = (e: MouseEvent) => {
      if (!panel.contains(e.target as Node) && !this.button?.contains(e.target as Node)) {
        this.showPanel(); // Toggle off
        document.removeEventListener('click', closePanel);
      }
    };
    setTimeout(() => document.addEventListener('click', closePanel), 0);
  }

  /** Start polling for kernel health */
  private startPolling(): void {
    this.stopPolling();
    this.pollTimer = setInterval(() => this.healthCheck(), this.config.pollIntervalMs);
  }

  /** Stop polling */
  private stopPolling(): void {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
    }
  }

  /** Check kernel health */
  private async healthCheck(): Promise<void> {
    if (this.state.status === 'disconnected') return;

    try {
      const response = await fetch(`${this.config.kernelUrl}/api/native/status`, {
        method: 'GET',
        signal: AbortSignal.timeout(3000),
      });

      if (!response.ok) {
        throw new Error(`Kernel returned ${response.status}`);
      }

      if (this.state.status === 'error') {
        // Kernel recovered
        this.setState({ status: 'connected', error: null });
      }
    } catch (err: any) {
      if (this.state.status === 'connected') {
        this.setState({ status: 'error', error: 'Kernel unreachable' });
      }
    }
  }

  /** Update state and re-render button */
  private setState(partial: Partial<HypercodeKernelButtonState>): void {
    this.state = { ...this.state, ...partial };

    if (this.button) {
      this.button.className = '';
      switch (this.state.status) {
        case 'connected':
          this.button.classList.add('hypercode-connected');
          this.button.title = 'Hypercode Kernel: Connected';
          break;
        case 'connecting':
          this.button.classList.add('hypercode-connecting');
          this.button.title = 'Hypercode Kernel: Connecting...';
          break;
        case 'error':
          this.button.classList.add('hypercode-error');
          this.button.title = `Hypercode Kernel: Error - ${this.state.error}`;
          break;
        default:
          this.button.title = 'Attach to Hypercode Kernel';
      }
    }
  }

  /** Generate a session ID */
  private generateSessionId(): string {
    const site = this.detectSite();
    const prefix = site ? SUPPORTED_SITES[site].name.toLowerCase() : 'unknown';
    return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
  }

  /** Wait for an element matching a selector to appear in the DOM */
  private waitForElement(selector: string, timeoutMs: number): Promise<Element | null> {
    // Check if already present
    const existing = document.querySelector(selector);
    if (existing) return Promise.resolve(existing);

    return new Promise((resolve) => {
      const observer = new MutationObserver(() => {
        const el = document.querySelector(selector);
        if (el) {
          observer.disconnect();
          resolve(el);
        }
      });

      observer.observe(document.body, { childList: true, subtree: true });

      setTimeout(() => {
        observer.disconnect();
        resolve(document.querySelector(selector));
      }, timeoutMs);
    });
  }

  /** Clean up the button and stop polling */
  destroy(): void {
    this.stopPolling();
    this.button?.remove();
    this.panel?.remove();
    this.button = null;
    this.panel = null;

    const styles = document.getElementById('hypercode-hypercode-kernel-styles');
    styles?.remove();
  }
}

// ─── Singleton ───────────────────────────────────────────────────────────────

let instance: HypercodeKernelButtonService | null = null;

/** Get or create the Hypercode Kernel button service */
export function getHypercodeKernelService(config?: Partial<HypercodeKernelConfig>): HypercodeKernelButtonService {
  if (!instance) {
    instance = new HypercodeKernelButtonService(config);
  }
  return instance;
}

/** Auto-initialize the Hypercode Kernel button if on a supported site */
export async function initHypercodeKernelButton(config?: Partial<HypercodeKernelConfig>): Promise<boolean> {
  const service = getHypercodeKernelService(config);
  return service.initialize();
}
