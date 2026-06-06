#!/usr/bin/env python3
"""
TormentNexus Subagent Swarm
============================
Spawns multiple AI workers that concurrently work on tasks using the FreeLLM proxy.
Each worker has file read/write and bash tools, retries continuously on errors,
and picks tasks from a shared queue.

Usage:
  python swarm.py --workers 5 --tasks tasks.json
  python swarm.py --workers 3 --task-dir ./swarm_tasks
"""

import anthropic
import json
import os
import sys
import time
import subprocess
import threading
import queue
import argparse
import traceback
import signal
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed

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
WORK_DIR = os.environ.get("SWARM_WORK_DIR", os.getcwd())

# ============================================================
# Tools available to each worker
# ============================================================
TOOL_DEFINITIONS = [
    {
        "name": "read_file",
        "description": "Read the contents of a file. Returns file content as text.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Path to the file to read (relative to work dir or absolute)"}
            },
            "required": ["path"]
        }
    },
    {
        "name": "write_file",
        "description": "Write content to a file. Creates parent directories if needed.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Path to the file to write"},
                "content": {"type": "string", "description": "Content to write to the file"}
            },
            "required": ["path", "content"]
        }
    },
    {
        "name": "run_bash",
        "description": "Execute a bash command. Returns stdout and stderr. Working directory is the project root.",
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
        "name": "list_files",
        "description": "List files in a directory recursively. Returns file paths.",
        "input_schema": {
            "type": "object",
            "properties": {
                "path": {"type": "string", "description": "Directory path to list (default: work dir)"},
                "pattern": {"type": "string", "description": "Glob pattern to filter (e.g. '*.go', '*.ts')"}
            }
        }
    },
    {
        "name": "grep_search",
        "description": "Search for a pattern in files. Returns matching lines with file paths.",
        "input_schema": {
            "type": "object",
            "properties": {
                "pattern": {"type": "string", "description": "Search pattern (regex or literal)"},
                "path": {"type": "string", "description": "Directory or file to search in"},
                "glob": {"type": "string", "description": "File glob filter (e.g. '*.go')"}
            },
            "required": ["pattern"]
        }
    },
    {
        "name": "task_complete",
        "description": "Mark the current task as complete with a summary.",
        "input_schema": {
            "type": "object",
            "properties": {
                "summary": {"type": "string", "description": "Summary of what was accomplished"},
                "files_changed": {"type": "array", "items": {"type": "string"}, "description": "List of files that were modified"}
            },
            "required": ["summary"]
        }
    }
]

# ============================================================
# Tool execution
# ============================================================
def execute_tool(name, input_data, work_dir):
    """Execute a tool call and return the result string."""
    try:
        if name == "read_file":
            path = input_data["path"]
            if not os.path.isabs(path):
                path = os.path.join(work_dir, path)
            if not os.path.exists(path):
                return f"ERROR: File not found: {path}"
            with open(path, "r", encoding="utf-8", errors="replace") as f:
                content = f.read(50000)  # Limit to 50KB
            return content

        elif name == "write_file":
            path = input_data["path"]
            if not os.path.isabs(path):
                path = os.path.join(work_dir, path)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "w", encoding="utf-8") as f:
                f.write(input_data["content"])
            return f"OK: Wrote {len(input_data['content'])} chars to {path}"

        elif name == "run_bash":
            cmd = input_data["command"]
            timeout = input_data.get("timeout", 60)
            try:
                result = subprocess.run(
                    cmd, shell=True, capture_output=True, text=True,
                    cwd=work_dir, timeout=timeout,
                    env={**os.environ, "PATH": os.environ.get("PATH", "")}
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

        elif name == "list_files":
            search_path = input_data.get("path", work_dir)
            if not os.path.isabs(search_path):
                search_path = os.path.join(work_dir, search_path)
            pattern = input_data.get("pattern", "*")
            import glob
            matches = glob.glob(os.path.join(search_path, "**", pattern), recursive=True)
            # Limit output
            return "\n".join(matches[:200]) or "No files found"

        elif name == "grep_search":
            import re as regex_module
            pattern = input_data["pattern"]
            search_path = input_data.get("path", work_dir)
            if not os.path.isabs(search_path):
                search_path = os.path.join(work_dir, search_path)
            glob_filter = input_data.get("glob", "*")
            import glob as glob_module
            matches = []
            for filepath in glob_module.glob(os.path.join(search_path, "**", glob_filter), recursive=True):
                if not os.path.isfile(filepath):
                    continue
                try:
                    with open(filepath, "r", encoding="utf-8", errors="replace") 
