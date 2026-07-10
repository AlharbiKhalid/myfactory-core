"""`myfactory plane sync` - dry-run synchronization plan for Plane.

Mapping:
- MyFactory Mission  -> Plane Module (or label; future mapping configurable)
- MyFactory Sprint   -> Plane Cycle
- MyFactory Task     -> Plane Issue / Work Item

Default is always dry-run. Live sync requires --apply AND plane.enabled: true
AND the configured API key environment variable. Live API calls are not
implemented in this milestone; --apply explains what is missing instead.
"""

from __future__ import annotations

import argparse
import os

from myfactory.core import paths
from myfactory.core.config import get_path, load_yaml


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "plane",
        help="Plane execution-tracker integration (dry-run by default).",
        description="Plane integration. Plane is the execution tracker; Git remains the source of truth.",
    )
    sub = p.add_subparsers(dest="plane_command", metavar="SUBCOMMAND")

    sync = sub.add_parser(
        "sync",
        help="Show what would be created/updated in Plane (dry-run by default).",
        description="Reads config, work breakdown, missions, and sprints, and prints the sync plan.",
    )
    sync.add_argument("--target", default=None, help="Target directory (default: current directory).")
    sync.add_argument("--dry-run", action="store_true", help="Print the sync plan without calling Plane (default).")
    sync.add_argument("--apply", action="store_true", help="Attempt live sync (requires Plane config + API key).")
    sync.set_defaults(func=run_sync)

    p.set_defaults(func=lambda args: (p.print_help(), 0)[1])


def _items(data, key: str) -> list[dict]:
    value = (data or {}).get(key) or []
    return [v for v in value if isinstance(v, dict)]


def run_sync(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)
    config = load_yaml(paths.project_config_path(target)) or {}

    missions = _items(load_yaml(target / "docs" / "03-delivery" / "missions.yaml"), "missions")
    sprints = _items(load_yaml(target / "docs" / "03-delivery" / "sprints.yaml"), "sprints")
    work = _items(load_yaml(target / "docs" / "03-delivery" / "work-breakdown.yaml"), "work_items")

    enabled = bool(get_path(config, "plane.enabled", False))
    key_env = str(get_path(config, "plane.api_key_env", "PLANE_API_KEY") or "PLANE_API_KEY")
    has_key = bool(os.environ.get(key_env))
    base_url = get_path(config, "plane.base_url", "CHANGE_ME")

    mode = "APPLY" if args.apply else "DRY RUN"
    print(f"Plane sync plan ({mode}) for: {target}")
    print(f"Plane enabled in config: {enabled}")
    print(f"Plane base_url: {base_url}")
    print(f"API key (${key_env}) present: {has_key}")
    print()
    print("Mapping: Mission -> Plane Module/Label | Sprint -> Plane Cycle | Task -> Plane Issue")
    print()

    def describe(kind: str, items: list[dict], id_field: str = "id", title_field: str = "title") -> None:
        real = [i for i in items if i.get(title_field) not in (None, "", "CHANGE_ME")]
        placeholders = len(items) - len(real)
        print(f"{kind}: {len(items)} defined"
              + (f" ({placeholders} still placeholder)" if placeholders else ""))
        for item in items:
            item_id = item.get(id_field, "?")
            title = item.get(title_field, "?")
            status = item.get("status", item.get("state", "-"))
            print(f"  would create/update: [{item_id}] {title} (status: {status})")
        if not items:
            print("  nothing to sync.")
        print()

    describe("Missions (-> Plane Modules/Labels)", missions)
    describe("Sprints (-> Plane Cycles)", sprints)
    describe("Tasks (-> Plane Issues)", work)

    print("Sync rules: create_missing=true, update_existing=true, delete_from_plane=false (never deletes).")

    if args.apply:
        missing = []
        if not enabled:
            missing.append("plane.enabled is false in .ApplicationFactory/config.yaml")
        if not base_url or base_url == "CHANGE_ME":
            missing.append("plane.base_url is not configured")
        if not has_key:
            missing.append(f"${key_env} environment variable is not set")
        if missing:
            print("\nCannot apply - missing requirements:")
            for m in missing:
                print(f"  - {m}")
            return 1
        print("\nLive Plane sync is not implemented in this milestone.")
        print("The dry-run plan above is what a live sync would perform.")
        return 1

    print("\nThis was a dry run. No Plane API calls were made.")
    return 0
