#!/usr/bin/env python3
"""
TormentNexus Subagent Swarm with P2P Messaging
================================================
Spawns multiple AI workers that concurrently work on tasks using the FreeLLM proxy.
Workers can send messages to each other via a shared message bus, enabling
coordination, knowledge sharing, conflict avoidance, and collaborative debugging.

P2P messaging model:
  - Each worker has a named mailbox (e.g. "W1", "W2")
  - Workers can broadcast to all, or send direct messages to specific workers
  - Messages are delivered on the next LLM turn as injected user context
  - A shared lock file prevents concurrent edits to the same files
  - Workers announce which files they're editing to avoid conflicts

Usage:
  python swarm.py --workers 5 --auto-discover
  python swarm.py --workers 3 --tasks swarm_tasks.json
  python swarm.py --dry-run --auto-discover
"""

import anthropic
import json
import os
import time
import subprocess
import queue
import argparse
import threading
import glob as glob_module
import re as regex_module
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime

# ============================================================
# Configuration
# ============================================================
FREELLM_BASE_URL = os.environ.get("FREELLM_BASE_URL", "http://localhost:4000")
FREELLM_API_KEY = os.environ.get("FREELLM_API_KEY", "sk-freellm-proxy")
MODEL = os.environ.get("SWARM_MODEL", "claude-sonnet-4-20250514")
MAX_TOKENS = int(os.environ.get("SWARM_MAX_TOKENS", "4096"))
MAX_RETRIES = int(os.environ.get("SWARM_MAX_RETRIES", "50"))
RETRY_BASE_DELAY = float(os.environ.get("SWARM_RETRY_DELAY", "2"))
MAX_TURNS = int(os.environ.get("SWARM_MAX_TURNS", "40"))


# ============================================================
# P2P Message Bus
# ============================================================
class MessageBus:
    """
    Thread-safe peer-to-peer message bus for inter-agent communication.
    
    Each worker gets a named mailbox. Messages can be:
    - Direct: sender -> specific recipient
    - Broadcast: sender -> all other workers
    - Status: worker announces what it's doing (file locks, progress)
    
    Messages accumulate in mailboxes and are delivered as injected
    context on the next LLM turn.
    """

    def __init__(self, worker_ids):
        self.lock = threading.Lock()
        self.mailboxes = {wid: [] for wid in worker_ids}
        # File locks: {filepath: worker_id}
        self.file_locks = {}
        # Shared knowledge base: {topic: data}
        self.shared_knowledge = {}
        # Worker status: {worker_id: {status, task, files_being_edited}}
        self.worker_status = {wid: {"status": "starting", "task": "", "files": set()}
                              for wid in worker_ids}

    def send(self, sender, recipient, msg_type, content):
        """Send a direct message from sender to recipient."""
        with self.lock:
            if recipient in self.mailboxes:
                self.mailboxes[recipient].append({
                    "from": sender,
                    "type": msg_type,
                    "content": content,
                    "time": datetime.now().strftime("%H:%M:%S")
                })

    def broadcast(self, sender, msg_type, content):
        """Broadcast a message from sender to all other workers."""
        with self.lock:
            for wid in self.mailboxes:
                if wid != sender:
                    self.mailboxes[wid].append({
                        "from": sender,
                        "type": msg_type,
                        "content": content,
                        "time": datetime.now().strftime("%H:%M:%S")
                    })

    def drain(self, worker_id):
        """Get and clear all pending messages for a worker. Returns list of messages."""
        with self.lock:
            msgs = self.mailboxes[worker_id][:]
            self.mailboxes[worker_id] = []
            return msgs

    def lock_file(self, worker_id, filepath):
        """Try to claim exclusive edit access to a file. Returns True if acquired."""
        with self.lock:
            if filepath not in self.file_locks or self.file_locks[filepath] == worker_id:
                self.file_locks[filepath] = worker_id
                self.worker_status[worker_id]["files"].add(filepath)
                return True
            return False

    def unlock_file(self, worker_id, filepath):
        """Release edit access to a file."""
        with self.lock:
            if self.file_locks.get(filepath) == worker_id:
                del self.file_locks[filepath]
                self.worker_status[worker_id]["files"].discard(filepath)

    def unlock_all(self, worker_id):
        """Release all file locks held by a worker."""
        with self.lock:
            to_remove = [f for f, w in self.file_locks.items() if w == worker_id]
            for f in to_remove:
                del self.file_locks[f]
            self.worker_status[worker_id]["files"] = set()

    def who_has_lock(self, filepath):
        """Return which worker has a file locked, or None."""
        with self.lock:
            return self.file_locks.get(filepath)

    def update_status(self, worker_id, status, task=""):
        """Update a worker's current status."""
        with self.lock:
            self.worker_status[worker_id]["status"] = status
            if task:
                self.worker_status[worker_id]["task"] = task

    def get_all_status(self):
        """Get all workers' status (for context injection)."""
        with self.lock:
            result = {}
            for wid, info in self.worker_status.items():
                result[wid] = {
                    "status": info["status"],
                    "task": info["task"],
                    "files": sorted(info["files"])
                }
            return result

    def share_knowledge(self, worker_id, topic, data):
        """Add to the shared knowledge base that all workers can access."""
        with self.lock:
            if topic not in self.shared_knowledge:
                self.shared_knowledge[topic] = []
            self.shared_knowledge[topic].append({
                "from": worker_id,
                "data": data,
                "time": datetime.now().strftime("%H:%M:%S")
            })

    def get_knowledge(self, topic=None):
        """Get shared knowledge, optionally filtered by topic."""
        with self.lock:
            if topic:
                return self.shared_knowledge.get(topic, [])
            return dict(self.shared_knowledge)

    def format_messages_for_injection(self, messages):
        """Format pending messages as a text block for LLM context injection."""
        if not messages:
            return None
        lines = []
        for msg in messages:
            prefix = "BROADCAST" if msg["type"] == "broadcast" else msg["type"].upper()
            lines.append(f"[{msg['time']}] [{msg['from']}] [{prefix}] {msg['content']}")
        return "\n".join(lines)

    def format_status_for_injection(self, my_worker_id):
        """Format all workers' status as a text block for LLM context."""
        status = self.get_all_status()
        lines = ["=== SWARM STATUS ==="]
        for wid, info in sorted(status.items()):
            marker = " (you)" if wid == my_worker_id else ""
            files_str = ", ".join(info["files"][:5]) if info["files"] else "none"
            lines.append(f"  {wid}{marker}: {info['status']} | editing: {files_str}")
        lines.append("====================")
        return "\n".join(lines)


# ============================================================
# Tool definitions (Anthropic tool_use format)
# ============================================================
TOOL_DEFINITIONS = [
    {
        "name": "read_file",
        "description": "Read the contents of a file. Returns file content as text.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Path to the file to read"}
            },
            "required": ["path"]
        }
    },
    {
        "name": "write_file",
        "description": "Write content to a file. Creates parent directories if needed. "
                       "Automatically acquires a file lock to prevent other workers from editing simultaneously.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Path to the file to write"},
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
                "command": {"type": "string", "description": "Bash command to execute"},
                "timeout": {"type": "integer", "description": "Timeout in seconds (default 60)"}
            },
            "required": ["command"]
        }
    },
    {
        "name": "grep_search",
        "description": "Search for a pattern in files. Returns matching lines with file paths.",
        "input_schema": {
            "type": "object",
            "properties": {
                "pattern": {"type": "string", "description": "Search pattern"},
                "path": {"type": "string", "description": "Directory to search in"},
                "glob": {"type": "string", "description": "File glob filter (e.g. '*.go')"}
            },
            "required": ["pattern"]
        }
    },
    {
        "name": "send_message",
        "description": "Send a message to another worker or broadcast to all workers. "
                       "Use this to coordinate, share findings, warn about conflicts, or ask for help.",
        "input_schema": {
            "type": "object",
            "properties": {
                "to": {"type": "string", "description": "Recipient worker ID (e.g. 'W2') or 'all' to broadcast"},
                "type": {"type": "string", "enum": ["info", "warning", "help", "broadcast", "discovery"],
                         "description": "Message type"},
                "content": {"type": "string", "description": "Message content"}
            },
            "required": ["to", "content"]
        }
    },
    {
        "name": "share_knowledge",
        "description": "Share a discovery or finding with all workers. This persists in a shared "
                       "knowledge base that any worker can query. Use for architectural insights, "
                       "API patterns, gotchas, or useful commands you discover.",
        "input_schema": {
            "type": "object",
            "properties": {
                "topic": {"type": "string", "description": "Knowledge topic (e.g. 'build-errors', 'api-patterns')"},
                "data": {"type": "string", "description": "The knowledge to share"}
            },
            "required": ["topic", "data"]
        }
    },
    {
        "name": "query_knowledge",
        "description": "Query the shared knowledge base for information shared by other workers.",
        "input_schema": {
            "type": "object",
            "properties": {
                "topic": {"type": "string", "description": "Topic to query (empty = all topics)"}
            }
        }
    },
    {
        "name": "check_file_lock",
        "description": "Check if a file is currently being edited by another worker. "
                       "Always check before editing to avoid conflicts.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "File path to check"}
            },
            "required": ["path"]
        }
    },
    {
        "name": "task_complete",
        "description": "Mark the current task as complete with a summary.",
        "input_schema": {
            "type": "object",
            "properties": {
                "summary": {"type": "string", "description": "Summary of what was accomplished"},
                "files_changed": {"type": "array", "items": {"type": "string"}, "description": "Files modified"},
                "discoveries": {"type": "string", "description": "Key discoveries to share with other workers"}
            },
            "required": ["summary"]
        }
    }
]


# ============================================================
# Tool execution (with P2P integration)
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
            # Try to acquire file lock
            locked_by = bus.who_has_lock(path)
            if locked_by and locked_by != worker_id:
                return (f"CONFLICT: File {path} is currently being edited by {locked_by}. "
                        f"Wait for them to finish, or send them a message to coordinate. "
                        f"Use send_message to ask {locked_by} to release the lock.")
            bus.lock_file(worker_id, path)
            os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
            with open(path, "w", encoding="utf-8") as f:
                f.write(input_data["content"])
            # Broadcast that we edited this file
            bus.broadcast(worker_id, "info",
                          f"Edited file: {path}")
            return f"OK: Wrote {len(input_data['content'])} chars to {path}"

        elif name == "run_bash":
            cmd = input_data["command"]
            timeout = input_data.get("timeout", 60)
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
                if not os.path.isfile(filepath):
                    continue
                if len(filepath) > 300:
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
            msg_type = input_data.get("type", "info")
            if to == "all":
                bus.broadcast(worker_id, msg_type, content)
                return "OK: Broadcast sent to all workers"
            else:
                bus.send(worker_id, to, msg_type, content)
                return f"OK: Message sent to {to}"

        elif name == "share_knowledge":
            topic = input_data["topic"]
            data = input_data["data"]
            bus.share_knowledge(worker_id, topic, data)
            bus.broadcast(worker_id, "discovery", f"Shared knowledge on '{topic}': {data[:100]}")
            return f"OK: Knowledge shared on topic '{topic}' and broadcast to all workers"

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
                return f"LOCKED by {locked_by}. Send them a message if you need access."
            return f"File {path} is available (no lock)."

        elif name == "task_complete":
            # Share any discoveries before completing
            discoveries = input_data.get("discoveries", "")
            if discoveries:
                bus.share_knowledge(worker_id, f"task-result-{worker_id}", discoveries)
                bus.broadcast(worker_id, "discovery", f"Completed task. Key findings: {discoveries[:200]}")
            # Release all file locks
            bus.unlock_all(worker_id)
            bus.update_status(worker_id, "completed")
            return "TASK_COMPLETE"

        else:
            return f"ERROR: Unknown tool: {name}"

    except Exception as e:
        return f"ERROR: {type(e).__name__}: {e}"


# ============================================================
# Worker - runs a single task with P2P messaging
# ============================================================
def run_worker(task, worker_id, results_queue, bus):
    """Run a single task using the Anthropic SDK through FreeLLM proxy with P2P messaging."""
    wid = f"W{worker_id}"
    task_id = task.get("id", f"task-{worker_id}")
    task_desc = task.get("description", "")
    task_files = task.get("files", [])
    work_dir = task.get("work_dir", os.getcwd())

    log_prefix = f"[{wid}|{task_id}]"
    print(f"{log_prefix} Starting: {task_desc[:80]}...")

    # Register with the message bus
    bus.update_status(wid, "working", task_id)

    # Announce ourselves
    bus.broadcast(wid, "info", f"Starting task: {task_id} - {task_desc[:60]}")

    client = anthropic.Anthropic(
        base_url=FREELLM_BASE_URL,
        api_key=FREELLM_API_KEY,
        max_retries=0
    )

    system_prompt = f"""You are an autonomous AI worker ({wid}) in a swarm working on the TormentNexus project.
You are part of a team of workers running in parallel. You CAN communicate with other workers.

Your task: {task_desc}

Working directory: {work_dir}

AVAILABLE TOOLS:
- read_file: Read file contents
- write_file: Write to files (auto-acquires file locks to prevent conflicts)
- run_bash: Execute bash commands
- grep_search: Search code patterns
- send_message: Send direct or broadcast messages to other workers
- share_knowledge: Share discoveries in a shared knowledge base
- query_knowledge: Query what other workers have discovered
- check_file_lock: Check if another worker is editing a file
- task_complete: Mark your task as complete

P2P MESSAGING:
- Use send_message(to="W2", content="...") for direct messages
- Use send_message(to="all", content="...") to broadcast
- Use share_knowledge(topic="...", data="...") for persistent shared findings
- Use query_knowledge() to see what others have discovered
- Use check_file_lock(path) before editing files to avoid conflicts

RULES:
1. Always read existing code before modifying it
2. Check file locks before writing to avoid conflicts with other workers
3. Share useful discoveries (build commands, gotchas, patterns) via share_knowledge
4. If stuck, broadcast for help - another worker may have the answer
5. If another worker has a file locked, send them a message to coordinate
6. Run tests/builds after making changes
7. If something fails, diagnose and fix it - keep trying until it works
8. Use task_complete when done with a summary AND discoveries"""

    if task_files:
        system_prompt += "\n\nKey files to focus on:\n"
        for f in task_files:
            system_prompt += f"- {f}\n"

    messages = []
    retry_count = 0
    total_api_calls = 0
    start_time = time.time()

    for turn in range(MAX_TURNS):
        retry_delay = RETRY_BASE_DELAY

        # ---- Drain P2P messages and inject as context ----
        pending_msgs = bus.drain(wid)
        p2p_context = ""
        if pending_msgs:
            p2p_context = bus.format_messages_for_injection(pending_msgs)

        # Always inject current swarm status
        swarm_status = bus.format_status_for_injection(wid)

        # Build context injection message
        context_parts = []
        if p2p_context:
            context_parts.append(f"=== MESSAGES FROM OTHER WORKERS ===\n{p2p_context}\n==================================")
        context_parts.append(swarm_status)

        # Check shared knowledge for relevant info
        knowledge = bus.get_knowledge()
        if knowledge:
            kb_lines = []
            for topic, entries in knowledge.items():
                for entry in entries[-5:]:  # Last 5 entries per topic
                    kb_lines.append(f"  [{entry['from']}] {entry['data'][:100]}")
            if kb_lines:
                context_parts.append("=== SHARED KNOWLEDGE ===\n" + "\n".join(kb_lines[-10:]) + "\n========================")

        context_injection = "\n\n".join(context_parts)

        # Inner retry loop for API errors
        while True:
            try:
                total_api_calls += 1

                # Build messages list with optional context injection
                api_messages = list(messages)

                # If we have P2P context, inject it as a user message at the end
                if context_injection and turn > 0:
                    # We'll prepend context to the last user message or add a new one
                    api_messages.append({
                        "role": "user",
                        "content": f"[SWARM CONTEXT - this is automated system context, not a user request]\n{context_injection}\n[END SWARM CONTEXT - continue your work]"
                    })

                response = client.messages.create(
                    model=MODEL,
                    max_tokens=MAX_TOKENS,
                    system=system_prompt,
                    tools=TOOL_DEFINITIONS,
                    messages=api_messages
                )
                retry_count = 0
                break

            except anthropic.APIConnectionError as e:
                retry_count += 1
                if retry_count > MAX_RETRIES:
                    bus.update_status(wid, "failed")
                    bus.unlock_all(wid)
                    bus.broadcast(wid, "warning", f"Failed - connection errors: {e}")
                    results_queue.put({"id": task_id, "status": "failed", "error": str(e)})
                    return
                delay = min(retry_delay * (2 ** min(retry_count, 6)), 120)
                print(f"{log_prefix} Connection error #{retry_count}, retry {delay:.0f}s")
                time.sleep(delay)

            except anthropic.RateLimitError:
                retry_count += 1
                delay = min(retry_delay * (2 ** min(retry_count, 4)), 60)
                print(f"{log_prefix} Rate limited #{retry_count}, retry {delay:.0f}s")
                time.sleep(delay)

            except anthropic.APIStatusError as e:
                if e.status_code in (429, 503, 502, 500):
                    retry_count += 1
                    delay = min(retry_delay * (2 ** min(retry_count, 5)), 90)
                    print(f"{log_prefix} API {e.status_code} #{retry_count}, retry {delay:.0f}s")
                    time.sleep(delay)
                else:
                    bus.update_status(wid, "failed")
                    bus.unlock_all(wid)
                    results_queue.put({"id": task_id, "status": "failed", "error": str(e)})
                    return

            except Exception as e:
                retry_count += 1
                delay = min(retry_delay * (2 ** min(retry_count, 4)), 60)
                print(f"{log_prefix} Error #{retry_count}: {type(e).__name__}: {e}")
                time.sleep(delay)

        # Process response
        assistant_content = response.content
        messages.append({"role": "assistant", "content": assistant_content})

        if response.stop_reason == "end_turn":
            text_parts = [b.text for b in assistant_content if hasattr(b, "text")]
            summary = " ".join(text_parts) if text_parts else "Completed (end_turn)"
            print(f"{log_prefix} Done: {summary[:80]}")
            bus.update_status(wid, "completed")
            bus.unlock_all(wid)
            results_queue.put({
                "id": task_id, "status": "completed", "summary": summary,
                "api_calls": total_api_calls, "turns": turn + 1,
                "elapsed": time.time() - start_time
            })
            return

        # Process tool calls
        tool_results = []
        task_done = False
        for block in assistant_content:
            if block.type == "tool_use":
                tool_name = block.name
                tool_input = block.input
                tool_input_str = json.dumps(tool_input)
                print(f"{log_prefix} Tool: {tool_name}({tool_input_str[:80]})")

                result = execute_tool(tool_name, tool_input, work_dir, wid, bus)

                if result == "TASK_COMPLETE":
                    task_done = True
                    summary = tool_input.get("summary", "Task completed")
                    files_changed = tool_input.get("files_changed", [])
                    results_queue.put({
                        "id": task_id, "status": "completed", "summary": summary,
                        "files_changed": files_changed, "api_calls": total_api_calls,
                        "turns": turn + 1, "elapsed": time.time() - start_time
                    })
                    tool_results.append({
                        "type": "tool_result", "tool_use_id": block.id,
                        "content": "Task marked as complete. Good work."
                    })
                else:
                    tool_results.append({
                        "type": "tool_result", "tool_use_id": block.id,
                        "content": result[:15000]
                    })

        messages.append({"role": "user", "content": tool_results})

        if task_done:
            print(f"{log_prefix} COMPLETED: {turn+1} turns, {total_api_calls} calls, {time.time()-start_time:.0f}s")
            return

    # Max turns
    bus.update_status(wid, "max_turns")
    bus.unlock_all(wid)
    bus.broadcast(wid, "warning", f"Reached max turns ({MAX_TURNS}) without completing")
    results_queue.put({
        "id": task_id, "status": "max_turns",
        "summary": f"Reached {MAX_TURNS} turns without completing",
        "api_calls": total_api_calls, "turns": MAX_TURNS,
        "elapsed": time.time() - start_time
    })


# ============================================================
# Auto-discover tasks from the repo
# ============================================================
def discover_tasks(repo_dir):
    """Auto-discover tasks by scanning the repo for incomplete work."""
    tasks = []

    # 1. Find Go files with TODOs
    for root, dirs, files in os.walk(os.path.join(repo_dir, "go")):
        dirs[:] = [d for d in dirs if d not in ("vendor", "node_modules", ".git")]
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
                            "description": f"Resolve TODOs in {relpath}: {'; '.join(todos[:3])}",
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
    for root, dirs, files in os.walk(os.path.join(repo_dir, "go")):
        dirs[:] = [d for d in dirs if d not in ("vendor", "node_modules", ".git")]
        if any(f.endswith("_test.go") for f in files):
            pkg_path = "./" + os.path.relpath(root, repo_dir).replace("\\", "/")
            tasks.append({
                "id": f"go-test-{os.path.basename(root)}",
                "description": f"Run 'go test {pkg_path}/...' and fix any test failures.",
                "files": [pkg_path],
                "work_dir": repo_dir
            })

    # 4. Security audit
    tasks.append({
        "id": "security-audit",
        "description": "Scan for security issues: hardcoded secrets, SQL injection, XSS, insecure deps.",
        "files": [],
        "work_dir": repo_dir
    })

    # 5. Dead code cleanup
    tasks.append({
        "id": "dead-code-cleanup",
        "description": "Find and remove dead code, unused imports, unreachable functions.",
        "files": [],
        "work_dir": repo_dir
    })

    return tasks


# ============================================================
# Main orchestrator
# ============================================================
def main():
    parser = argparse.ArgumentParser(description="TormentNexus Subagent Swarm with P2P Messaging")
    parser.add_argument("--workers", type=int, default=3, help="Number of parallel workers")
    parser.add_argument("--tasks", type=str, help="JSON file with task definitions")
    parser.add_argument("--auto-discover", action="store_true", help="Auto-discover tasks")
    parser.add_argument("--repo-dir", type=str, default=os.getcwd(), help="Repository directory")
    parser.add_argument("--dry-run", action="store_true", help="Show tasks without running")
    parser.add_argument("--model", type=str, help="Override model name")
    args = parser.parse_args()

    global MODEL
    if args.model:
        MODEL = args.model

    # Load or discover tasks
    tasks = []
    if args.tasks:
        with open(args.tasks) as f:
            tasks = json.load(f)
    else:
        print(f"Auto-discovering tasks in {args.repo_dir}...")
        tasks = discover_tasks(args.repo_dir)

    if not tasks:
        print("No tasks found. Exiting.")
        return

    # Create worker IDs
    worker_ids = [f"W{i+1}" for i in range(min(args.workers, len(tasks)))]

    # Create the P2P message bus
    bus = MessageBus(worker_ids)

    print(f"\n{'='*60}")
    print("  TORMENTNEXUS SUBAGENT SWARM (P2P)")
    print(f"{'='*60}")
    print(f"  Tasks:   {len(tasks)}")
    print(f"  Workers: {len(worker_ids)}")
    print(f"  Model:   {MODEL}")
    print(f"  Proxy:   {FREELLM_BASE_URL}")
    print(f"  Repo:    {args.repo_dir}")
    print("  P2P:     enabled (message bus + file locks + knowledge base)")
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
        print(f"Proxy: {resp.read().decode()}\n")
    except Exception as e:
        print(f"ERROR: FreeLLM proxy not reachable: {e}")
        return

    # Run workers in parallel
    results_queue = queue.Queue()
    start_time = time.time()

    with ThreadPoolExecutor(max_workers=len(worker_ids)) as executor:
        futures = {}
        for i, task in enumerate(tasks[:len(worker_ids)]):
            wid = worker_ids[i]
            future = executor.submit(run_worker, task, i + 1, results_queue, bus)
            futures[future] = (task, wid)

        # For remaining tasks, submit as workers finish
        remaining = tasks[len(worker_ids):]
        task_idx = 0

        completed = 0
        failed = 0
        total_tasks = len(tasks)

        while completed + failed < total_tasks:
            try:
                result = results_queue.get(timeout=600)
                if result["status"] == "completed":
                    completed += 1
                    elapsed = result.get("elapsed", 0)
                    print(f"\n  DONE {result['id']}: {result.get('summary','')[:80]} ({elapsed:.0f}s)")
                else:
                    failed += 1
                    print(f"\n  FAIL {result['id']}: {result.get('error', result.get('summary',''))[:80]}")

                # Submit next task if any remain
                if task_idx < len(remaining):
                    next_task = remaining[task_idx]
                    next_wid = worker_ids[completed + failed - 1] if completed + failed - 1 < len(worker_ids) else worker_ids[-1]
                    # Find a free worker slot
                    free_worker = (completed + failed) % len(worker_ids) + 1
                    future = executor.submit(run_worker, next_task, free_worker, results_queue, bus)
                    futures[future] = (next_task, f"W{free_worker}")
                    task_idx += 1

            except queue.Empty:
                if all(f.done() for f in futures):
                    break

    elapsed = time.time() - start_time

    # Print shared knowledge summary
    knowledge = bus.get_knowledge()
    if knowledge:
        print(f"\n=== SHARED KNOWLEDGE ({sum(len(v) for v in knowledge.values())} entries) ===")
        for topic, entries in knowledge.items():
            print(f"  [{topic}]: {len(entries)} entries")
            for entry in entries[:3]:
                print(f"    {entry['from']}: {entry['data'][:80]}")

    print(f"\n{'='*60}")
    print("  SWARM COMPLETE")
    print(f"{'='*60}")
    print(f"  Completed: {completed}/{total_tasks}")
    print(f"  Failed:    {failed}/{total_tasks}")
    print(f"  Elapsed:   {elapsed:.0f}s")
    print(f"  Knowledge: {sum(len(v) for v in bus.get_knowledge().values())} shared entries")
    print(f"{'='*60}")

    # Save results and knowledge
    all_results = []
    while not results_queue.empty():
        all_results.append(results_queue.get())

    results_data = {
        "summary": {"completed": completed, "failed": failed, "elapsed": elapsed},
        "results": all_results,
        "shared_knowledge": bus.get_knowledge()
    }
    with open("swarm_results.json", "w") as f:
        json.dump(results_data, f, indent=2)
    print("Results + knowledge saved to swarm_results.json")


if __name__ == "__main__":
    main()
