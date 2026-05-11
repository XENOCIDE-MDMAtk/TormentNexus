import { getLLMService } from '../../../lib/trpc-core.js';
import type { Supervisor, SupervisorConfig, Message } from './types.js';

export class BorgSupervisor implements Supervisor {
  name: string;
  provider: string;
  config: SupervisorConfig;

  constructor(config: SupervisorConfig) {
    this.name = config.name;
    this.provider = config.provider;
    this.config = config;
  }

  async isAvailable(): Promise<boolean> {
    const llm = getLLMService();
    // In Borg, availability is checked via ProviderTruth
    const quotaService = (llm.modelSelector as any)?.getQuotaService?.();
    const quota = quotaService?.getQuota?.(this.config.provider);
    return !!quota && quota.authTruth === 'VALID';
  }

  async chat(messages: Message[]): Promise<string> {
    const llm = getLLMService();
    
    // Convert council messages to Borg format if needed
    // (Assuming Borg generate accepts prompt string or similar)
    const prompt = messages.map(m => `[${m.role.toUpperCase()}]: ${m.content}`).join('\n\n');
    
    const response = await (llm as any).generateText(
      this.config.provider,
      this.config.model || 'default',
      '', // System prompt handled in council logic or here
      prompt,
      {
        temperature: this.config.temperature,
        taskComplexity: 'high',
      }
    );

    return response.content;
  }
}
