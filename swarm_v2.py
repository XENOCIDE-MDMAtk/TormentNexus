#!/usr/bin/env python3
"""
TormentNexus Subagent Swarm v2 — Continuous Retry, Multi-Model, P2P
====================================================================

Spawns multiple AI workers that concurrently work on tasks using the FreeLLM proxy.
Workers continuously retry on errors with exponential backoff + jitter, and
automatically fall back to alternative models if the primary model fails.

Key improvements over v1:
  - Dual-mode API: Anthropic SDK + OpenAI SDK (some models only work with one)
  - Aggressive retry: 100+ retries with jittered exponential backoff
  - Model fallback: tries multiple models if primary fails
  - Worker-pull task queue: workers grab next task when done (no fixed assignment)
  - Graceful shutdown: SIGINT/SIGTERM stops workers cleanly
  - Auto git commit: after each task completion, auto-commits changes
  - Compact P2P context: injected as system-reminder, not conversation messages
  - Progress dashboard: periodic status updates

Usage:
  python swarm_v2.py --workers 5 --auto-discover
  python swarm_v2.py --workers 3 --tasks swarm_tasks.json
  python swarm_v2.py --dry-run --auto-discover
  python swarm_v2.py --model gpt-4o-mini --workers 8
"""

import anthropic
import openai
import json
import os
import time
import subprocess
import argparse
import threading
import signal
import random
import glob as glob_module
import re as regex_module
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime

# ============================================================
# Configuration
# ============================================================
FREELLM_BASE_URL = os.environ.get("FREELLM_BASE_URL", "http://localhost:4000")
FREELLM_API_KEY = os.environ.get("FREELLM_API_KEY", "sk-freellm-proxy")
OPENAI_API_KEY = os.environ.get("OPENAI_API_KEY", "sk-freellm")

# Model priority: try these in order. First available wins.
PREFERRED_MODELS = os.environ.get("SWARM_MODELS", "").split(",") if os.environ.get("SWARM_MODELS") else [
    "gemini-3-flash",
    "DeepSeek-V3.2",
    "gpt-4o-mini",
    "deepseek-v4-flash-free",
    "claude-sonnet-4-20250514",
    "free-llm",
]

MAX_TOKENS = int(os.environ.get("SWARM_MAX_TOKENS", "8192"))
MAX_RETRIES = int(os.environ.get("SWARM_MAX_RETRIES", "100"))
MAX_TURNS = int(os.environ.get("SWARM_MAX_TURNS", "50"))
RETRY_BASE_DELAY = float(os.environ.get("SWARM_RETRY_DELAY", "1.5"))
API_TIMEOUT = float(os.environ.get("SWARM_API_TIMEOUT", "180"))

# Global shutdown flag
shutdown_event = threading.Event()

def signal_handler(sig, frame):
    print("\n\nWARN:  Shutdown signal received! Stopping workers gracefully...")
    shutdown_event.set()

signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)


# ============================================================
# P2P Message Bus (same as v1, proven design)
# ============================================================
class MessageBus:
    """Thread-safe peer-to-peer message bus for inter-agent communication."""

    def __init__(self, worker_ids):
        self.lock = threading.Lock()
        self.mailboxes = {wid: [] for wid in worker_ids}
        self.file_locks = {}
        self.shared_knowledge = {}
        self.worker_status = {wid: {"status": "idle", "task": "", "files": set()} for wid in worker_ids}

    def send(self, sender, recipient, msg_type, content):
        with self.lock:
            if recipient in self.mailboxes:
                self.mailboxes[recipient].append({
                    "from": sender, "type": msg_type, "content": content,
                    "time": datetime.now().strftime("%H:%M:%S")
                })

    def broadcast(self, sender, msg_type, content):
        with self.lock:
            for wid in self.mailboxes:
                if wid != sender:
                    self.mailboxes[wid].append({
                        "from": sender, "type": msg_type, "content": content,
                        "time": datetime.now().strftime("%H:%M:%S")
                    })

    def drain(self, worker_id):
        with self.lock:
            msgs = self.mailboxes[worker_id][:]
            self.mailboxes[worker_id] = []
            return msgs

    def lock_file(self, worker_id, filepath):
        with self.lock:
            if filepath not in self.file_locks or self.file_locks[filepath] == worker_id:
                self.file_locks[filepath] = worker_id
                self.worker_status[worker_id]["files"].add(filepath)
                return True
            return False

    def unlock_file(self, worker_id, filepath):
        with self.lock:
            if self.file_locks.get(filepath) == worker_id:
                del self.file_locks[filepath]
                self.worker_status[worker_id]["files"].discard(filepath)

    def unlock_all(self, worker_id):
        with self.lock:
            to_remove = [f for f, w in self.file_locks.items() if w == worker_id]
            for f in to_remove:
                del self.file_locks[f]
            self.worker_status[worker_id]["files"] = set()

    def who_has_lock(self, filepath):
        with self.lock:
            return self.file_locks.get(filepath)

    def update_status(self, worker_id, status, task=""):
        with self.lock:
            self.worker_status[worker_id]["status"] = status
            if task:
                self.worker_status[worker_id]["task"] = task

    def get_all_status(self):
        with self.lock:
            result = {}
            for wid, info in self.worker_status.items():
                result[wid] = {
                    "status": info["status"], "task": info["task"],
                    "files": sorted(info["files"])
                }
            return result

    def share_knowledge(self, worker_id, topic, data):
        with self.lock:
            if topic not in self.shared_knowledge:
                self.shared_knowledge[topic] = []
            self.shared_knowledge[topic].append({
                "from": worker_id, "data": data,
                "time": datetime.now().strftime("%H:%M:%S")
            })

    def get_knowledge(self, topic=None):
        with self.lock:
            if topic:
                return self.shared_knowledge.get(topic, [])
            return dict(self.shared_knowledge)

    def format_status_for_injection(self, my_worker_id):
        status = self.get_all_status()
        lines = ["=== SWARM STATUS ==="]
        for wid, info in sorted(status.items()):
            marker = " (you)" if wid == my_worker_id else ""
            files_str = ", ".join(info["files"][:5]) if info["files"] else "none"
            lines.append(f"  {wid}{marker}: {info['status']} | editing: {files_str}")
        lines.append("====================")
        return "\n".join(lines)


# ============================================================
# Tool definitions (Anthropic format)
# ============================================================
TOOL_DEFINITIONS = [
    {
        "name": "read_file",
        "description": "Read the contents of a file. Returns file content as text.",
        "input_schema": {
            "type": "object",
            "properties": {"path": {"type": "string", "description": "Path to the file"}},
            "required": ["path"]
        }
    },
    {
        "name": "write_file",
        "description": "Write content to a file. Creates parent directories if needed. "
                       "Automatically acquires a file lock to prevent conflicts.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Path to write"},
                "content": {"type": "string", "description": "Content to write"}
            },
            "required": ["path", "content"]
        }
    },
    {
        "name": "run_bash",
        "description": "Execute a bash command. Returns stdout and stderr.",
        "input_schema": {
            "type": "object",
            "properties": {
                "command": {"type": "string", "description": "Bash command"},
                "timeout": {"type": "integer", "description": "Timeout in seconds (default 120)"}
            },
            "required": ["command"]
        }
    },
    {
        "name": "grep_search",
        "description": "Search for a pattern in files.",
        "input_schema": {
            "type": "object",
            "properties": {
                "pattern": {"type": "string", "description": "Search pattern"},
                "path": {"type": "string", "description": "Directory to search"},
                "glob": {"type": "string", "description": "File glob filter"}
            },
            "required": ["pattern"]
        }
    },
    {
        "name": "send_message",
        "description": "Send a message to another worker or broadcast.",
        "input_schema": {
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Recipient (e.g. 'W2') or 'all'"},
                "content": {"type": "string", "description": "Message content"}
            },
            "required": ["to", "content"]
        }
    },
    {
        "name": "share_knowledge",
        "description": "Share a discovery with all workers in the shared knowledge base.",
        "input_schema": {
            "type": "object",
            "properties": {
                "topic": {"type": "string", "description": "Knowledge topic"},
                "data": {"type": "string", "description": "The knowledge to share"}
            },
            "required": ["topic", "data"]
        }
    },
    {
        "name": "query_knowledge",
        "description": "Query the shared knowledge base.",
        "input_schema": {
            "type": "object",
            "properties": {"topic": {"type": "string", "description": "Topic (empty = all)"}}
        }
    },
    {
        "name": "check_file_lock",
        "description": "Check if a file is locked by another worker.",
        "input_schema": {
            "type": "object",
            "properties": {"path": {"type": "string", "description": "File path"}},
            "required": ["path"]
        }
    },
    {
        "name": "task_complete",
        "description": "Mark the current task as complete with a summary.",
        "input_schema": {
            "type": "object",
            "properties": {
                "summary": {"type": "string", "description": "What was accomplished"},
                "files_changed": {
                    "type": "array", "items": {"type": "string"},
                    "description": "Files modified"
                },
                "discoveries": {"type": "string", "description": "Key discoveries to share"}
            },
            "required": ["summary"]
        }
    },
]

# OpenAI-format tool definitions (for openai SDK)
OPENAI_TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": t["name"],
            "description": t["description"],
            "parameters": t["input_schema"],
        }
    }
    for t in TOOL_DEFINITIONS
]


# ============================================================
# Tool execution engine
# ============================================================
def execute_tool(name, input_data, work_dir, worker_id, bus):
    """Execute a tool call and return the result string."""
    try:
        if name == "read_file":
            path = input_data["path"]
            if not os.path.isabs(path):
                path = os.path.join(work_dir, path)
            if not os.path.exists(path):
                return f"ERROR: File not found: {path}"
            with open(path, "r", encoding="utf-8", errors="replace") as f:
                return f.read(50000)

        elif name == "write_file":
            path = input_data["path"]
            if not os.path.isabs(path):
                path = os.path.join(work_dir, path)
            locked_by = bus.who_has_lock(path)
            if locked_by and locked_by != worker_id:
                return (f"CONFLICT: File {path} is locked by {locked_by}. "
                        f"Use send_message to coordinate.")
            bus.lock_file(worker_id, path)
            os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
            with open(path, "w", encoding="utf-8", newline="\n") as f:
                f.write(input_data["content"])
            bus.broadcast(worker_id, "info", f"Edited file: {path}")
            return f"OK: Wrote {len(input_data['content'])} chars to {path}"

        elif name == "run_bash":
            cmd = input_data["command"]
            timeout = input_data.get("timeout", 120)
            try:
                result = subprocess.run(
                    cmd, shell=True, capture_output=True, text=True,
                    cwd=work_dir, timeout=timeout
                )
                output = ""
                if result.stdout:
                    output += result.stdout[:10000]
                if result.stderr:
                    output += f"\nSTDERR:\n{result.stderr[:5000]}"
                if result.returncode != 0:
                    output += f"\nEXIT CODE: {result.returncode}"
                return output or "(no output)"
            except subprocess.TimeoutExpired:
                return f"ERROR: Command timed out after {timeout}s"

        elif name == "grep_search":
            pattern = input_data["pattern"]
            search_path = input_data.get("path", work_dir)
            if not os.path.isabs(search_path):
                search_path = os.path.join(work_dir, search_path)
            glob_filter = input_data.get("glob", "*")
            matches = []
            for filepath in glob_module.glob(os.path.join(search_path, "**", glob_filter), recursive=True):
                if not os.path.isfile(filepath) or len(filepath) > 300:
                    continue
                try:
                    with open(filepath, "r", encoding="utf-8", errors="replace") as f:
                        for i, line in enumerate(f, 1):
                            if regex_module.search(pattern, line):
                                rel = os.path.relpath(filepath, work_dir)
                                matches.append(f"{rel}:{i}: {line.rstrip()[:200]}")
                                if len(matches) >= 50:
                                    return "\n".join(matches)
                except Exception:
                    pass
            return "\n".join(matches) or "No matches found"

        elif name == "send_message":
            to = input_data.get("to", "all")
            content = input_data["content"]
            if to == "all":
                bus.broadcast(worker_id, "info", content)
                return "OK: Broadcast sent"
            else:
                bus.send(worker_id, to, "info", content)
                return f"OK: Message sent to {to}"

        elif name == "share_knowledge":
            bus.share_knowledge(worker_id, input_data["topic"], input_data["data"])
            bus.broadcast(worker_id, "discovery",
                         f"Shared knowledge on '{input_data['topic']}': {input_data['data'][:100]}")
            return f"OK: Knowledge shared on '{input_data['topic']}'"

        elif name == "query_knowledge":
            topic = input_data.get("topic", "")
            knowledge = bus.get_knowledge(topic if topic else None)
            if not knowledge:
                return "No shared knowledge found."
            lines = []
            for t, entries in (knowledge.items() if not topic else [(topic, knowledge)]):
                for entry in entries:
                    lines.append(f"[{entry['from']}] [{entry['time']}] {entry['data'][:200]}")
            return "\n".join(lines) or "No entries found."

        elif name == "check_file_lock":
            path = input_data["path"]
            if not os.path.isabs(path):
                path = os.path.join(work_dir, path)
            locked_by = bus.who_has_lock(path)
            if locked_by:
                return f"LOCKED by {locked_by}. Send them a message to coordinate."
            return f"File {path} is available."

        elif name == "task_complete":
            discoveries = input_data.get("discoveries", "")
            if discoveries:
                bus.share_knowledge(worker_id, f"task-result-{worker_id}", discoveries)
                bus.broadcast(worker_id, "discovery", f"Completed task. Findings: {discoveries[:200]}")
            bus.unlock_all(worker_id)
            bus.update_status(worker_id, "completed")
            return "TASK_COMPLETE"

        else:
            return f"ERROR: Unknown tool: {name}"

    except Exception as e:
        return f"ERROR: {type(e).__name__}: {e}"


# ============================================================
# Auto git commit after task completion
# ============================================================
def auto_git_commit(work_dir, worker_id, task_id, summary):
    """Auto-commit any changes made by a worker."""
    try:
        # Stage all changes
        subprocess.run(["git", "add", "-A"], cwd=work_dir, capture_output=True, timeout=10)
        # Check if there's anything to commit
        result = subprocess.run(
            ["git", "diff", "--cached", "--quiet"],
            cwd=work_dir, capture_output=True, timeout=10
        )
        if result.returncode != 0:
            # There are staged changes - commit them
            msg = f"swarm({worker_id}): {task_id} - {summary[:72]}"
            subprocess.run(
                ["git", "commit", "-m", msg, "--allow-empty"],
                cwd=work_dir, capture_output=True, timeout=30
            )
            return True
        return False
    except Exception as e:
        print(f"  [{worker_id}] git commit error: {e}")
        return False


# ============================================================
# LLM Client — with model fallback and aggressive retry
# ============================================================
class SwarmLLMClient:
    """LLM client that tries Anthropic format first, then OpenAI format."""

    def __init__(self, models: list[str]):
        self.models = models
        self.anthropic_client = anthropic.Anthropic(
            base_url=FREELLM_BASE_URL,
            api_key=FREELLM_API_KEY,
            max_retries=0,
            timeout=API_TIMEOUT,
        )
        self.openai_client = openai.OpenAI(
            base_url=f"{FREELLM_BASE_URL}/v1",
            api_key=OPENAI_API_KEY,
            max_retries=0,
            timeout=API_TIMEOUT,
        )
        # Track which models work with which SDK
        self.model_sdk_preference = {}  # model -> "anthropic" or "openai"

    def _is_anthropic_model(self, model: str) -> bool:
        """Guess if a model name is Anthropic-compatible."""
        return any(kw in model.lower() for kw in ["claude", "anthropic"])

    def call_anthropic(self, model, system, messages, tools, max_tokens):
        """Call via Anthropic SDK."""
        return self.anthropic_client.messages.create(
            model=model,
            max_tokens=max_tokens,
            system=system,
            tools=tools,
            messages=messages,
        )

    def call_openai(self, model, system, messages, tools, max_tokens):
        """Call via OpenAI SDK (translates Anthropic-format messages)."""
        # Convert messages to OpenAI format
        oai_messages = [{"role": "system", "content": system}]
        for msg in messages:
            role = msg["role"]
            content = msg["content"]
            if isinstance(content, str):
                oai_messages.append({"role": role, "content": content})
            elif isinstance(content, list):
                # Anthropic tool_use blocks -> OpenAI format
                oai_content = []
                for block in content:
                    if hasattr(block, "type"):
                        if block.type == "text":
                            oai_content.append({"type": "text", "text": block.text})
                        elif block.type == "tool_use":
                            oai_content.append({
                                "type": "function",
                                "id": block.id,
                                "function": {
                                    "name": block.name,
                                    "arguments": json.dumps(block.input)
                                }
                            })
                    elif isinstance(block, dict):
                        if block.get("type") == "tool_result":
                            oai_content.append({
                                "role": "tool",
                                "tool_call_id": block.get("tool_use_id", ""),
                                "content": block.get("content", "")
                            })
                        elif block.get("type") == "text":
                            oai_content.append({"type": "text", "text": block.get("text", "")})
                        else:
                            oai_content.append(block)
                if oai_content:
                    last_msg = oai_messages[-1] if oai_messages else None
                    if last_msg and last_msg.get("role") == "assistant":
                        last_msg.setdefault("tool_calls", [])
                        for item in oai_content:
                            if item.get("type") == "function":
                                last_msg["tool_calls"].append(item)
                            elif item.get("type") == "text" and item.get("text"):
                                last_msg.setdefault("content", "")
                                last_msg["content"] += item["text"]
                    else:
                        text_parts = [c for c in oai_content if isinstance(c, str) or c.get("type") == "text"]
                        if text_parts:
                            combined = " ".join(
                                c if isinstance(c, str) else c.get("text", "") for c in text_parts
                            )
                            oai_messages.append({"role": role, "content": combined})

        resp = self.openai_client.chat.completions.create(
            model=model,
            messages=oai_messages,
            tools=tools if tools else openai.NOT_GIVEN,
            max_tokens=max_tokens,
        )
        return resp

    def create(self, model, system, messages, tools, max_tokens, retry_count=0):
        """
        Make an LLM call with automatic SDK selection and retry.
        Returns (response, sdk_used) or raises after MAX_RETRIES.
        """
        # Check known SDK preference
        preferred_sdk = self.model_sdk_preference.get(model)

        if preferred_sdk == "anthropic":
            return self.call_anthropic(model, system, messages, tools, max_tokens), "anthropic"
        elif preferred_sdk == "openai":
            return self.call_openai(model, system, messages, tools, max_tokens), "openai"

        # Try Anthropic first for claude models, OpenAI first for others
        if self._is_anthropic_model(model):
            try:
                resp = self.call_anthropic(model, system, messages, tools, max_tokens)
                self.model_sdk_preference[model] = "anthropic"
                return resp, "anthropic"
            except Exception as e:
                # If Anthropic format fails, try OpenAI
                try:
                    resp = self.call_openai(model, system, messages, tools, max_tokens)
                    self.model_sdk_preference[model] = "openai"
                    return resp, "openai"
                except Exception:
                    raise e
        else:
            try:
                resp = self.call_openai(model, system, messages, tools, max_tokens)
                self.model_sdk_preference[model] = "openai"
                return resp, "openai"
            except Exception as e:
                try:
                    resp = self.call_anthropic(model, system, messages, tools, max_tokens)
                    self.model_sdk_preference[model] = "anthropic"
                    return resp, "anthropic"
                except Exception:
                    raise e


# ============================================================
# Shared Task Queue — workers pull tasks when available
# ============================================================
class TaskQueue:
    """Thread-safe task queue. Workers pull tasks and push results."""

    def __init__(self, tasks):
        self.lock = threading.Lock()
        self.pending = list(tasks)
        self.in_progress = {}  # task_id -> (worker_id, start_time)
        self.completed = []
        self.failed = []

    def pull(self, worker_id):
        """Pull the next available task. Returns None if queue empty."""
        with self.lock:
            if not self.pending:
                return None
            task = self.pending.pop(0)
            self.in_progress[task["id"]] = (worker_id, time.time())
            return task

    def complete(self, task_id, result):
        """Mark a task as completed."""
        with self.lock:
            if task_id in self.in_progress:
                del self.in_progress[task_id]
            result["id"] = task_id
            self.completed.append(result)

    def fail(self, task_id, result):
        """Mark a task as failed. Re-queue it for retry if retries remain."""
        with self.lock:
            if task_id in self.in_progress:
                del self.in_progress[task_id]
            result["id"] = task_id
            retries = result.get("retries", 0) + 1
            result["retries"] = retries
            if retries < 3:  # Re-queue up to 3 times
                task = {"id": task_id, "description": result.get("summary", "retry"),
                        "files": result.get("files_changed", []),
                        "work_dir": result.get("work_dir", ""), "_retry_count": retries}
                self.pending.append(task)
            else:
                self.failed.append(result)

    @property
    def stats(self):
        with self.lock:
            return {
                "pending": len(self.pending),
                "in_progress": len(self.in_progress),
                "completed": len(self.completed),
                "failed": len(self.failed),
            }


# ============================================================
# Worker loop — continuously pulls tasks and works until done
# ============================================================
def run_worker(worker_id: int, task_queue: TaskQueue, bus: MessageBus,
               llm_client: SwarmLLMClient, work_dir: str, models: list[str]):
    """
    Worker loop: pulls tasks from queue, executes them with continuous retry,
    and auto-commits changes on completion.
    """
    wid = f"W{worker_id}"

    while not shutdown_event.is_set():
        # Pull next task
        task = task_queue.pull(wid)
        if task is None:
            bus.update_status(wid, "idle")
            time.sleep(2)
            continue

        task_id = task.get("id", f"task-{worker_id}")
        task_desc = task.get("description", "")
        task_files = task.get("files", [])
        task_work_dir = task.get("work_dir", work_dir)
        log_prefix = f"[{wid}|{task_id}]"

        print(f"{log_prefix} Starting: {task_desc[:80]}...")
        bus.update_status(wid, "working", task_id)
        bus.broadcast(wid, "info", f"Starting task: {task_id}")

        # ---- Execute the task with retry loop ----
        result = execute_task_with_retry(
            task, wid, bus, llm_client, task_work_dir, models, log_prefix
        )

        if result["status"] == "completed":
            task_queue.complete(task_id, result)
            # Auto git commit
            committed = auto_git_commit(task_work_dir, wid, task_id, result.get("summary", ""))
            if committed:
                print(f"{log_prefix} OK Auto-committed changes")
            print(f"{log_prefix} OK COMPLETED: {result.get('summary', '')[:60]} "
                  f"({result.get('turns', 0)} turns, {result.get('api_calls', 0)} calls, "
                  f"{result.get('elapsed', 0):.0f}s)")
        elif result["status"] == "shutdown":
            # Put task back for later
            task["_retry_count"] = task.get("_retry_count", 0)
            with task_queue.lock:
                task_queue.pending.insert(0, task)
                if task_id in task_queue.in_progress:
                    del task_queue.in_progress[task_id]
            print(f"{log_prefix} PAUSE Task returned to queue (shutdown)")
            break
        else:
            result["work_dir"] = task_work_dir
            result["files_changed"] = task_files
            task_queue.fail(task_id, result)
            print(f"{log_prefix} FAIL FAILED: {result.get('error', result.get('summary', ''))[:60]}")

        bus.unlock_all(wid)
        bus.update_status(wid, "idle")

    print(f"[{wid}] Worker exiting.")


def execute_task_with_retry(task, wid, bus, llm_client, work_dir, models, log_prefix):
    """Execute a single task with aggressive retry and model fallback."""
    task_desc = task.get("description", "")
    task_files = task.get("files", [])
    task_id = task.get("id", "unknown")

    system_prompt = f"""You are an autonomous AI worker ({wid}) in a swarm working on the TormentNexus project.
You are part of a team of workers running in parallel. You CAN communicate with other workers.

Your task: {task_desc}
Working directory: {work_dir}

AVAILABLE TOOLS:
- read_file: Read file contents
- write_file: Write to files (auto-acquires file locks)
- run_bash: Execute bash commands
- grep_search: Search code patterns
- send_message: Send direct or broadcast messages
- share_knowledge: Share discoveries in shared knowledge base
- query_knowledge: Query what others discovered
- check_file_lock: Check if another worker is editing a file
- task_complete: Mark task as complete

RULES:
1. Always read existing code before modifying it
2. Check file locks before writing to avoid conflicts
3. Share discoveries (build commands, gotchas, patterns) via share_knowledge
4. If stuck, broadcast for help
5. Run tests/builds after making changes
6. If something fails, diagnose and fix it - keep retrying until it works
7. Use task_complete when done with summary AND discoveries"""

    if task_files:
        system_prompt += "\n\nKey files:\n" + "\n".join(f"- {f}" for f in task_files)

    messages = []
    retry_count = 0
    total_api_calls = 0
    start_time = time.time()
    current_model_idx = 0

    for turn in range(MAX_TURNS):
        if shutdown_event.is_set():
            return {"status": "shutdown"}

        # ---- Drain P2P messages ----
        pending_msgs = bus.drain(wid)
        context_parts = []

        if pending_msgs:
            msg_lines = []
            for msg in pending_msgs:
                prefix = msg["type"].upper()
                msg_lines.append(f"[{msg['time']}][{msg['from']}][{prefix}] {msg['content'][:200]}")
            context_parts.append("=== MESSAGES ===\n" + "\n".join(msg_lines) + "\n===============")

        context_parts.append(bus.format_status_for_injection(wid))

        knowledge = bus.get_knowledge()
        if knowledge:
            kb_lines = []
            for topic, entries in knowledge.items():
                for entry in entries[-3:]:
                    kb_lines.append(f"  [{entry['from']}] {entry['data'][:120]}")
            if kb_lines:
                context_parts.append("=== KNOWLEDGE ===\n" + "\n".join(kb_lines[-8:]) + "\n=================")

        p2p_context = "\n\n".join(context_parts)

        # ---- Call LLM with retry + model fallback ----
        response = None
        sdk_used = "anthropic"
        model = models[current_model_idx % len(models)]

        while True:
            if shutdown_event.is_set():
                return {"status": "shutdown"}

            try:
                total_api_calls += 1
                # Build messages with optional P2P context
                api_messages = list(messages)
                if p2p_context and turn > 0:
                    api_messages.append({
                        "role": "user",
                        "content": f"[SWARM CONTEXT]\n{p2p_context}\n[END CONTEXT - continue work]"
                    })

                response, sdk_used = llm_client.create(
                    model=model,
                    system=system_prompt,
                    messages=api_messages,
                    tools=TOOL_DEFINITIONS if sdk_used != "openai" else None,
                    max_tokens=MAX_TOKENS,
                )
                retry_count = 0  # Reset on success
                break

            except (anthropic.APIConnectionError, openai.APIConnectionError) as e:
                retry_count += 1
                if retry_count > MAX_RETRIES:
                    return {"status": "failed", "error": f"Connection errors: {e}",
                            "api_calls": total_api_calls, "turns": turn, "elapsed": time.time() - start_time}
                delay = min(RETRY_BASE_DELAY * (2 ** min(retry_count, 7)), 120) * (0.5 + random.random())
                print(f"{log_prefix} Connection error #{retry_count}, retry {delay:.1f}s")
                time.sleep(delay)

            except (anthropic.RateLimitError, openai.RateLimitError):
                retry_count += 1
                delay = min(RETRY_BASE_DELAY * (2 ** min(retry_count, 5)), 60) * (0.5 + random.random())
                print(f"{log_prefix} Rate limited #{retry_count}, retry {delay:.1f}s")
                time.sleep(delay)

            except (anthropic.APIStatusError, openai.APIStatusError) as e:
                status = getattr(e, 'status_code', 0)
                if status in (429, 502, 503, 500, 529):
                    retry_count += 1
                    delay = min(RETRY_BASE_DELAY * (2 ** min(retry_count, 6)), 90) * (0.5 + random.random())
                    print(f"{log_prefix} API {status} #{retry_count}, retry {delay:.1f}s")
                    time.sleep(delay)
                elif status == 401:
                    # Auth error - try next model
                    print(f"{log_prefix} Auth error on {model}, trying next model")
                    current_model_idx += 1
                    if current_model_idx >= len(models) * 2:
                        return {"status": "failed", "error": "All models failed auth",
                                "api_calls": total_api_calls, "turns": turn, "elapsed": time.time() - start_time}
                    model = models[current_model_idx % len(models)]
                    retry_count = 0
                    time.sleep(2)
                else:
                    return {"status": "failed", "error": f"API {status}: {e}",
                            "api_calls": total_api_calls, "turns": turn, "elapsed": time.time() - start_time}

            except Exception as e:
                retry_count += 1
                err_name = type(e).__name__
                if retry_count > MAX_RETRIES:
                    return {"status": "failed", "error": f"{err_name}: {e}",
                            "api_calls": total_api_calls, "turns": turn, "elapsed": time.time() - start_time}
                delay = min(RETRY_BASE_DELAY * (2 ** min(retry_count, 5)), 60) * (0.5 + random.random())
                print(f"{log_prefix} {err_name} #{retry_count}, retry {delay:.1f}s: {str(e)[:60]}")
                time.sleep(delay)

        # ---- Process Anthropic-format response ----
        if sdk_used == "anthropic":
            assistant_content = response.content
            messages.append({"role": "assistant", "content": assistant_content})

            if response.stop_reason == "end_turn":
                text_parts = [b.text for b in assistant_content if hasattr(b, "text")]
                summary = " ".join(text_parts) if text_parts else "Completed (end_turn)"
                return {
                    "status": "completed", "summary": summary,
                    "api_calls": total_api_calls, "turns": turn + 1,
                    "elapsed": time.time() - start_time, "model": model
                }

            # Process tool calls
            tool_results = []
            task_done = False
            for block in assistant_content:
                if block.type == "tool_use":
                    print(f"{log_prefix} Tool: {block.name}({json.dumps(block.input)[:80]})")
                    result = execute_tool(block.name, block.input, work_dir, wid, bus)
                    if result == "TASK_COMPLETE":
                        task_done = True
                        summary = block.input.get("summary", "Task completed")
                        files_changed = block.input.get("files_changed", [])
                        tool_results.append({
                            "type": "tool_result", "tool_use_id": block.id,
                            "content": "Task marked as complete."
                        })
                        return {
                            "status": "completed", "summary": summary,
                            "files_changed": files_changed,
                            "api_calls": total_api_calls, "turns": turn + 1,
                            "elapsed": time.time() - start_time, "model": model
                        }
                    tool_results.append({
                        "type": "tool_result", "tool_use_id": block.id,
                        "content": result[:15000]
                    })
            messages.append({"role": "user", "content": tool_results})

        else:
            # OpenAI format response
            choice = response.choices[0]
            msg = choice.message

            if msg.content:
                messages.append({"role": "assistant", "content": msg.content})

            if choice.finish_reason == "stop":
                return {
                    "status": "completed", "summary": msg.content or "Completed",
                    "api_calls": total_api_calls, "turns": turn + 1,
                    "elapsed": time.time() - start_time, "model": model
                }

            # Process tool calls
            if msg.tool_calls:
                tool_results = []
                assistant_msg = {"role": "assistant", "content": msg.content or "",
                                "tool_calls": []}
                for tc in msg.tool_calls:
                    func_name = tc.function.name
                    func_args = json.loads(tc.function.arguments)
                    print(f"{log_prefix} Tool: {func_name}({json.dumps(func_args)[:80]})")
                    result = execute_tool(func_name, func_args, work_dir, wid, bus)
                    assistant_msg["tool_calls"].append({
                        "id": tc.id, "type": "function",
                        "function": {"name": func_name, "arguments": tc.function.arguments}
                    })
                    tool_results.append({
                        "role": "tool", "tool_call_id": tc.id,
                        "content": result[:15000]
                    })
                    if result == "TASK_COMPLETE":
                        return {
                            "status": "completed",
                            "summary": func_args.get("summary", "Task completed"),
                            "files_changed": func_args.get("files_changed", []),
                            "api_calls": total_api_calls, "turns": turn + 1,
                            "elapsed": time.time() - start_time, "model": model
                        }
                messages.append(assistant_msg)
                messages.extend(tool_results)

    # Max turns reached
    return {
        "status": "max_turns",
        "summary": f"Reached {MAX_TURNS} turns without completing",
        "api_calls": total_api_calls, "turns": MAX_TURNS,
        "elapsed": time.time() - start_time
    }


# ============================================================
# Auto-discover tasks from the repo
# ============================================================
def discover_tasks(repo_dir):
    """Auto-discover tasks by scanning the repo for incomplete work."""
    tasks = []

    # 1. Find Go files with TODOs
    go_dir = os.path.join(repo_dir, "go")
    if os.path.isdir(go_dir):
        for root, dirs, files in os.walk(go_dir):
            dirs[:] = [d for d in dirs if d not in ("vendor", "node_modules", ".git", "build_output")]
            for fname in files:
                if not fname.endswith(".go") or fname.endswith("_test.go"):
                    continue
                fpath = os.path.join(root, fname)
                relpath = os.path.relpath(fpath, repo_dir).replace("\\", "/")
                try:
                    with open(fpath, "r", encoding="utf-8", errors="replace") as f:
                        todos = []
                        for i, line in enumerate(f, 1):
                            if any(kw in line.upper() for kw in ["TODO", "FIXME", "HACK", "NOT IMPLEMENTED"]):
                                todos.append(f"L{i}: {line.strip()[:80]}")
                        if todos:
                            tasks.append({
                                "id": f"fix-todos-{fname.replace('.go','')}",
                                "description": f"Resolve TODOs in {relpath}: {'; '.join(todos[:5])}",
                                "files": [relpath],
                                "work_dir": repo_dir
                            })
                except Exception:
                    pass

    # 2. Find TS packages needing typecheck fixes
    packages_dir = os.path.join(repo_dir, "packages")
    if os.path.isdir(packages_dir):
        for pkg in sorted(os.listdir(packages_dir)):
            pkg_dir = os.path.join(packages_dir, pkg)
            if os.path.isfile(os.path.join(pkg_dir, "tsconfig.json")):
                tasks.append({
                    "id": f"typecheck-{pkg}",
                    "description": f"Run typecheck on packages/{pkg} and fix all TS errors.",
                    "files": [f"packages/{pkg}"],
                    "work_dir": repo_dir
                })

    # 3. Find Go packages needing test fixes
    if os.path.isdir(go_dir):
        for root, dirs, files in os.walk(go_dir):
            dirs[:] = [d for d in dirs if d not in ("vendor", "node_modules", ".git", "build_output")]
            if any(f.endswith("_test.go") for f in files):
                pkg_path = "./" + os.path.relpath(root, repo_dir).replace("\\", "/")
                tasks.append({
                    "id": f"go-test-{os.path.basename(root)}",
                    "description": f"Run 'go test {pkg_path}' and fix any test failures.",
                    "files": [pkg_path],
                    "work_dir": repo_dir
                })

    # 4. TODO.md-based tasks
    todo_path = os.path.join(repo_dir, "TODO.md")
    if os.path.isfile(todo_path):
        with open(todo_path, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line.startswith("- [ ]"):
                    desc = line[5:].strip().strip("*")
                    if desc:
                        # Extract task name from markdown
                        clean = regex_module.sub(r'\*\*([^*]+)\*\*', r'\1', desc)
                        task_id = regex_module.sub(r'[^a-z0-9]+', '-', clean.lower())[:40]
                        tasks.append({
                            "id": f"todo-{task_id}",
                            "description": f"From TODO.md: {clean}",
                            "files": [],
                            "work_dir": repo_dir
                        })

    # 5. Security audit
    tasks.append({
        "id": "security-audit",
        "description": "Scan for security issues: hardcoded secrets, SQL injection, XSS, insecure deps. "
                       "Check .env files, API endpoints, and authentication code.",
        "files": [],
        "work_dir": repo_dir
    })

    # 6. Dead code cleanup
    tasks.append({
        "id": "dead-code-cleanup",
        "description": "Find and remove dead code, unused imports, unreachable functions in Go and TS.",
        "files": [],
        "work_dir": repo_dir
    })

    # 7. Build verification
    tasks.append({
        "id": "go-build-verify",
        "description": "Run 'go build ./...' in the go/ directory and fix any compilation errors.",
        "files": ["go/"],
        "work_dir": repo_dir
    })

    return tasks


# ============================================================
# Progress dashboard (background thread)
# ============================================================
def dashboard_printer(task_queue, bus, start_time):
    """Periodically print swarm status."""
    while not shutdown_event.is_set():
        time.sleep(15)
        if shutdown_event.is_set():
            break
        stats = task_queue.stats
        elapsed = time.time() - start_time
        status = bus.get_all_status()
        workers_str = " | ".join(
            f"{wid}: {info['status'][:12]}" for wid, info in sorted(status.items())
        )
        print(f"\n  [STATS] [{elapsed:.0f}s] pending:{stats['pending']} "
              f"active:{stats['in_progress']} done:{stats['completed']} "
              f"fail:{stats['failed']} | {workers_str}\n")


# ============================================================
# Main orchestrator
# ============================================================
def main():
    parser = argparse.ArgumentParser(description="TormentNexus Subagent Swarm v2")
    parser.add_argument("--workers", type=int, default=3, help="Parallel workers")
    parser.add_argument("--tasks", type=str, help="JSON file with tasks")
    parser.add_argument("--auto-discover", action="store_true", help="Auto-discover tasks")
    parser.add_argument("--repo-dir", type=str, default=os.getcwd(), help="Repository directory")
    parser.add_argument("--dry-run", action="store_true", help="Show tasks without running")
    parser.add_argument("--model", type=str, help="Override model (comma-separated for fallback chain)")
    args = parser.parse_args()

    # Build model list
    models = PREFERRED_MODELS
    if args.model:
        models = [m.strip() for m in args.model.split(",")]

    # Load or discover tasks
    tasks = []
    if args.tasks:
        with open(args.tasks) as f:
            tasks = json.load(f)
    else:
        if not args.auto_discover:
            args.auto_discover = True
        print(f"Auto-discovering tasks in {args.repo_dir}...")
        tasks = discover_tasks(args.repo_dir)

    if not tasks:
        print("No tasks found. Exiting.")
        return

    # Create worker IDs and infrastructure
    num_workers = min(args.workers, len(tasks))
    worker_ids = [f"W{i+1}" for i in range(num_workers)]
    bus = MessageBus(worker_ids)
    task_queue = TaskQueue(tasks)
    llm_client = SwarmLLMClient(models)

    # Header
    print(f"\n{'='*60}")
    print("  TORMENTNEXUS SWARM v2 (P2P + Continuous Retry)")
    print(f"{'='*60}")
    print(f"  Tasks:    {len(tasks)}")
    print(f"  Workers:  {num_workers}")
    print(f"  Models:   {' | '.join(models)}")
    print(f"  Proxy:    {FREELLM_BASE_URL}")
    print(f"  Repo:     {args.repo_dir}")
    print(f"  MaxRetry: {MAX_RETRIES}")
    print(f"  MaxTurns: {MAX_TURNS}")
    print("  P2P:      enabled (message bus + file locks + knowledge)")
    print(f"{'='*60}\n")

    for i, t in enumerate(tasks):
        print(f"  [{i+1}/{len(tasks)}] {t['id']}: {t['description'][:70]}...")
    print()

    if args.dry_run:
        with open("swarm_tasks_discovered.json", "w") as f:
            json.dump(tasks, f, indent=2)
        print(f"Saved {len(tasks)} tasks to swarm_tasks_discovered.json")
        return

    # Check proxy
    try:
        import urllib.request
        resp = urllib.request.urlopen(f"{FREELLM_BASE_URL}/health", timeout=5)
        print(f"Proxy health: {resp.read().decode()}\n")
    except Exception as e:
        print(f"ERROR: FreeLLM proxy not reachable at {FREELLM_BASE_URL}: {e}")
        print("Start it with: cd litellm_control_panel && ./freellm.exe")
        return

    # Start dashboard thread
    start_time = time.time()
    dash_thread = threading.Thread(
        target=dashboard_printer, args=(task_queue, bus, start_time), daemon=True
    )
    dash_thread.start()

    # Launch workers
    with ThreadPoolExecutor(max_workers=num_workers) as executor:
        futures = []
        for i in range(num_workers):
            future = executor.submit(
                run_worker, i + 1, task_queue, bus, llm_client, args.repo_dir, models
            )
            futures.append(future)

        # Wait for all workers to finish (they exit when no more tasks or shutdown)
        for future in futures:
            try:
                future.result(timeout=7200)  # 2h max per worker
            except Exception as e:
                print(f"Worker exception: {e}")

    elapsed = time.time() - start_time
    stats = task_queue.stats
    knowledge = bus.get_knowledge()

    # Final summary
    print(f"\n{'='*60}")
    print("  SWARM COMPLETE")
    print(f"{'='*60}")
    print(f"  Completed: {stats['completed']}/{len(tasks)}")
    print(f"  Failed:    {stats['failed']}/{len(tasks)}")
    print(f"  Pending:   {stats['pending']}")
    print(f"  Elapsed:   {elapsed:.0f}s")
    print(f"  Knowledge: {sum(len(v) for v in knowledge.values())} shared entries")
    print(f"{'='*60}")

    # Save results + knowledge
    results_data = {
        "summary": {
            "completed": stats["completed"], "failed": stats["failed"],
            "pending": stats["pending"], "elapsed": elapsed,
            "models_used": models, "timestamp": datetime.now().isoformat()
        },
        "completed_tasks": task_queue.completed,
        "failed_tasks": task_queue.failed,
        "shared_knowledge": knowledge,
    }
    with open("swarm_results.json", "w") as f:
        json.dump(results_data, f, indent=2, default=str)
    print("Results saved to swarm_results.json")

    # Push any remaining git changes
    try:
        subprocess.run(["git", "push", "origin", "main"], cwd=args.repo_dir,
                      capture_output=True, timeout=30)
        print("Git changes pushed to origin.")
    except Exception:
        pass


if __name__ == "__main__":
    main()
