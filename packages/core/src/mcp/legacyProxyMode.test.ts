import { describe, expect, it } from 'vitest';

import { isLegacyProxyDisabled, shouldUseLegacyProxy } from './legacyProxyMode.js';

describe('legacyProxyMode', () => {
  it('keeps the legacy proxy disabled by default', () => {
    expect(shouldUseLegacyProxy({} as NodeJS.ProcessEnv)).toBe(false);
  });

  it('enables the legacy proxy only when explicitly opted in', () => {
    expect(
      shouldUseLegacyProxy({ MCP_ENABLE_LEGACY_METAMCP_PROXY: 'true' } as NodeJS.ProcessEnv),
    ).toBe(true);
    expect(
      shouldUseLegacyProxy({ MCP_ENABLE_METAMCP_PROXY: '1' } as NodeJS.ProcessEnv),
    ).toBe(true);
  });

  it('honors the explicit disable flag over any opt-in flags', () => {
    expect(
      shouldUseLegacyProxy({
        MCP_ENABLE_LEGACY_METAMCP_PROXY: 'true',
        MCP_DISABLE_METAMCP: 'true',
      } as NodeJS.ProcessEnv),
    ).toBe(false);
    expect(isLegacyProxyDisabled({ MCP_DISABLE_METAMCP: 'yes' } as NodeJS.ProcessEnv)).toBe(true);
  });
});