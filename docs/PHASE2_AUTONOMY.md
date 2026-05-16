# Phase 2: Autonomy Loop

## 1. Autonomous Healer Loop
The "Immune System" has been upgraded from a single-pass diagnosis to a multi-turn autonomous cycle.

### HealerService
- **`healAndVerify(error, context, maxAttempts)`**: Orchestrates the loop.
    1. **Diagnose**: Analyze error logs via LLM.
    2. **Fix**: Generate and apply code changes.
    3. **Verify**: Run tests (`vitest`) or type checks (`tsc`) to confirm the fix.
    4. **Retry**: If verification fails, feed the new error back into the loop.

### HealerReactor
- **StopHook**: Agents can emit `agent:stop_healing` to suspend auto-fixes during sensitive operations.
- **IdleHealer**: Triggers routine diagnostics when the system detects inactivity (`IDLE_THRESHOLD_MS`).
- **Exponential Backoff**: Prevents log spam and budget exhaustion if LLM providers are failing.

## 2. Skills Progressive Disclosure
Nexus now employs the "AI Hypervisor" pattern for Skills management.

### SkillRegistry (TS) & SkillDecisionSystem (Go)
- **Ranked Retrieval**: Skills are ranked using BM25 relevance scoring via the Go sidecar.
- **Predictive Pre-loading**: High-confidence skills (>7.0 score) are automatically loaded into the active context.
- **Context Optimization**: Instead of listing all available skills, only the most relevant are disclosed to the agent based on the `chatHistory` and `activeGoal`.
