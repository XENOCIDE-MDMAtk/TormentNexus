import os

memory_path = r"c:\Users\hyper\workspace\borg\MEMORY.md"

new_observation = """

## Multi-Agent Systemic Observation (2026-06-01) - v1.0.0-alpha.88

1. **NPM Lock Conflicts & ECOMPROMISED Errors**:
   - Widespread write access concurrency on Windows can yield `ECOMPROMISED Lock compromised` NPM CLI errors when executing concurrent child process runs.
   - **Resolution**: Gracefully catch compromised locks inside the automated child process exit handlers, logging the conflict outcome and continuing down the queue to guarantee uninterrupted batch processing.
"""

# Read existing content in utf-16-le
with open(memory_path, "r", encoding="utf-16le", errors="replace") as f:
    existing = f.read()

# Append new observation
updated = existing + new_observation

# Write back in utf-16-le
with open(memory_path, "w", encoding="utf-16le") as f:
    f.write(updated)

print("Successfully appended new systemic observations to MEMORY.md!")
