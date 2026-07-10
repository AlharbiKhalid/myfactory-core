"""`myfactory doctor` — readiness report for a MyFactory-enabled repo."""

from __future__ import annotations

import argparse
import os
from pathlib import Path

from myfactory.core import git as gitutil
from myfactory.core import paths
from myfactory.core.config import get_path, load_yaml

OK = "OK     "
MISSING = "MISSING"
WARN = "WARN   "
INFO = "INFO   "


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "doctor",
        help="Report MyFactory readiness of a repo.",
        description="Inspects the target repo and prints a readiness report. Never fails harshly.",
    )
    p.add_argument("--target", default=None, help="Target directory (default: current directory).")
    p.set_defaults(func=run)


def check(present: bool, label: str, detail: str = "") -> tuple[bool, str]:
    status = OK if present else MISSING
    suffix = f" — {detail}" if detail else ""
    return present, f"[{status}] {label}{suffix}"


def run(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)
    print(f"MyFactory doctor report for: {target}\n")

    lines: list[str] = []
    ok_count = 0
    total = 0

    def add(present: bool, label: str, detail: str = "") -> bool:
        nonlocal ok_count, total
        present, line = check(present, label, detail)
        lines.append(line)
        total += 1
        if present:
            ok_count += 1
        return present

    meta = paths.metadata_dir(target)
    add(meta.is_dir(), ".ApplicationFactory/ metadata directory")
    add(paths.product_manifest_path(target).is_file(), ".ApplicationFactory/product.yaml")
    has_config = add(paths.project_config_path(target).is_file(), ".ApplicationFactory/config.yaml")
    add(paths.task_packages_dir(target).is_dir(), ".ApplicationFactory/task-packages/")
    add(paths.orchestrator_dir(target).is_dir(), ".ApplicationFactory/orchestrator/")
    add((paths.orchestrator_dir(target) / "HERMES-CONTROLLER-PROMPT.md").is_file(),
        "Hermes controller prompt")

    docs = target / "docs"
    for sub in ("00-product", "01-business", "02-architecture", "03-delivery", "04-qa", "05-operations"):
        add((docs / sub).is_dir(), f"docs/{sub}/")

    required_files = [
        "docs/00-product/prd.md",
        "docs/00-product/acceptance-criteria.md",
        "docs/01-business/business-rules.yaml",
        "docs/02-architecture/system-overview.md",
        "docs/03-delivery/work-breakdown.yaml",
        "docs/03-delivery/missions.yaml",
        "docs/03-delivery/sprints.yaml",
        "docs/04-qa/test-strategy.md",
    ]
    for rel in required_files:
        add((target / rel).is_file(), rel)

    add((target / "AGENTS.md").is_file(), "AGENTS.md (Codex overlay)")
    claude_cmds = target / ".claude" / "commands"
    add(claude_cmds.is_dir() and any(claude_cmds.glob("myfactory-*.md")),
        ".claude/commands/myfactory-*.md (Claude overlay)")

    is_repo = gitutil.is_git_repo(target)
    add(is_repo, "Git repository")
    if is_repo:
        provider = gitutil.detect_provider(target)
        lines.append(f"[{INFO}] Git provider detected: {provider}")

    # Plane configuration.
    if has_config:
        config = load_yaml(paths.project_config_path(target)) or {}
        plane_enabled = bool(get_path(config, "plane.enabled", False))
        if plane_enabled:
            key_env = get_path(config, "plane.api_key_env", "PLANE_API_KEY") or "PLANE_API_KEY"
            base_url = get_path(config, "plane.base_url", "CHANGE_ME")
            if not base_url or base_url == "CHANGE_ME":
                lines.append(f"[{WARN}] Plane enabled but base_url is not configured")
            if os.environ.get(str(key_env)):
                lines.append(f"[{OK}] Plane API key present in ${key_env}")
            else:
                lines.append(f"[{WARN}] Plane enabled but ${key_env} is not set")
        else:
            lines.append(f"[{INFO}] Plane disabled in config (dry-run only)")
        hermes = bool(get_path(config, "orchestration.hermes.enabled", False))
        lines.append(f"[{INFO}] Hermes orchestrator: {'enabled' if hermes else 'disabled'}")

    for line in lines:
        print(line)

    print()
    print(f"Readiness: {ok_count}/{total} checks passed.")
    if ok_count < total:
        print("Run `myfactory init` to add missing structure (existing files are preserved).")
    else:
        print("Structure looks ready. Next: myfactory discover --print-prompt")
    return 0
