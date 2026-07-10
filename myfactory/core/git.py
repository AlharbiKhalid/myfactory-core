"""Git detection helpers. Read-only; never mutates the repository."""

from __future__ import annotations

import subprocess
from pathlib import Path


def is_git_repo(target: Path) -> bool:
    if (target / ".git").exists():
        return True
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--is-inside-work-tree"],
            cwd=target,
            capture_output=True,
            text=True,
            timeout=10,
        )
        return result.returncode == 0 and result.stdout.strip() == "true"
    except (OSError, subprocess.TimeoutExpired):
        return False


def remote_urls(target: Path) -> list[str]:
    try:
        result = subprocess.run(
            ["git", "remote", "-v"],
            cwd=target,
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            return []
        urls = []
        for line in result.stdout.splitlines():
            parts = line.split()
            if len(parts) >= 2:
                urls.append(parts[1])
        return urls
    except (OSError, subprocess.TimeoutExpired):
        return []


def detect_provider(target: Path) -> str:
    """Detect git provider from remotes. Returns 'github', 'gitlab', or 'unknown'."""
    for url in remote_urls(target):
        lowered = url.lower()
        if "github.com" in lowered:
            return "github"
        if "gitlab" in lowered:
            return "gitlab"
    return "unknown"
