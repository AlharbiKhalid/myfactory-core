"""`myfactory plan` — planning readiness report and planning prompt.

The CLI does not generate plans with AI. It reports whether the source-of-truth
docs are ready for planning and prints the prompt an AI agent uses to populate
work-breakdown, missions, and sprints.
"""

from __future__ import annotations

import argparse
from pathlib import Path

from myfactory.core import paths

PLACEHOLDER = "CHANGE_ME"

READINESS_FILES = {
    "product": [
        "docs/00-product/idea-brief.md",
        "docs/00-product/prd.md",
        "docs/00-product/user-journeys.md",
        "docs/00-product/acceptance-criteria.md",
    ],
    "business": [
        "docs/01-business/business-rules.yaml",
        "docs/01-business/decision-tables.md",
    ],
    "architecture": [
        "docs/02-architecture/system-overview.md",
        "docs/02-architecture/api-contracts.md",
    ],
    "qa": [
        "docs/04-qa/test-strategy.md",
    ],
}

PLANNING_PROMPT = """\
# MyFactory Planning Prompt

You are an AI planning agent inside a MyFactory-enabled repository at: {target}

Git is the source of truth. Plane is only the execution tracker.
Mapping: MyFactory Mission = larger goal. MyFactory Sprint = Plane Cycle.
MyFactory Task = Plane Issue.

## Read first

1. `.ApplicationFactory/product.yaml` and `.ApplicationFactory/config.yaml`
2. `docs/00-product/prd.md` and `docs/00-product/acceptance-criteria.md`
3. `docs/01-business/business-rules.yaml`
4. `docs/02-architecture/system-overview.md` and `api-contracts.md`
5. `docs/04-qa/test-strategy.md`
6. Existing `docs/03-delivery/` files — extend them; never renumber IDs.

If product/business/architecture docs are still placeholders, stop and tell the
user to run discovery first.

## Produce

1. `docs/03-delivery/work-breakdown.yaml` — work items following the file's
   `work_item_schema` (id, type, title, module, priority, state, source_docs,
   acceptance_criteria, definition_of_done, dependencies). Use the task ID
   convention with the project key. Reference BR-* business rule IDs where
   business logic is touched. Dependencies must form a DAG.
2. `docs/03-delivery/missions.yaml` — MISSION-### entries with goal, status,
   source_docs, success_criteria, and the sprints that deliver them.
3. `docs/03-delivery/sprints.yaml` — SPRINT-### entries with mission_id, scope
   (work item IDs), entry/exit criteria, and validation_required (functional QA
   always; business QA when BR-* rules are in scope).

## Rules

- Every work item must trace to at least one source doc.
- Size items so one agent finishes one item in one session.
- Respect dependencies across sprints.
- Do not implement anything. Do not modify product/business/architecture docs.
- Record assumptions and open questions in the delivery files as comments.

## Finish

Summarize the plan and offer to commit with:
`docs(delivery): populate work breakdown, missions, and sprints`
Then the user runs: myfactory plane sync --dry-run
"""


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "plan",
        help="Report planning readiness; print the planning prompt for AI agents.",
        description=(
            "Checks whether source-of-truth docs are ready for planning "
            "(--dry-run) and prints the planning prompt (--print-prompt)."
        ),
    )
    p.add_argument("--target", default=None, help="Target directory (default: current directory).")
    p.add_argument("--dry-run", action="store_true", help="Report readiness without changing anything (default).")
    p.add_argument("--print-prompt", action="store_true", help="Print the planning prompt for Claude/Codex.")
    p.set_defaults(func=run)


def _real_items(path: Path, list_key: str) -> list:
    """Return non-placeholder entries of a delivery file's main list."""
    from myfactory.core.config import load_yaml

    try:
        data = load_yaml(path) or {}
    except Exception:
        return []
    items = data.get(list_key) or []
    return [i for i in items
            if isinstance(i, dict) and i.get("title") not in (None, "", PLACEHOLDER)]


def file_state(path: Path) -> str:
    if not path.is_file():
        return "missing"
    try:
        text = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return "unreadable"
    if PLACEHOLDER in text:
        return "placeholder"
    return "filled"


def run(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)

    if args.print_prompt:
        print(PLANNING_PROMPT.format(target=target))
        return 0

    print(f"MyFactory planning readiness for: {target}\n")
    all_ready = True
    for section, files in READINESS_FILES.items():
        section_ready = True
        print(f"{section}:")
        for rel in files:
            state = file_state(target / rel)
            marker = {"filled": "ready      ", "placeholder": "placeholder",
                      "missing": "missing    ", "unreadable": "unreadable "}[state]
            print(f"  [{marker}] {rel}")
            if state != "filled":
                section_ready = False
        if not section_ready:
            all_ready = False
        print()

    delivery = [("docs/03-delivery/work-breakdown.yaml", "work_items"),
                ("docs/03-delivery/missions.yaml", "missions"),
                ("docs/03-delivery/sprints.yaml", "sprints")]
    print("delivery (planning output):")
    for rel, list_key in delivery:
        path = target / rel
        if not path.is_file():
            label = "missing    "
        else:
            items = _real_items(path, list_key)
            label = "populated  " if items else "template   "
        print(f"  [{label}] {rel}")
    print()

    if all_ready:
        print("Source docs are ready for planning.")
        print("Next: myfactory plan --print-prompt  (paste the prompt into Claude/Codex)")
    else:
        print("Source docs are not fully ready. Missing/placeholder files above need")
        print("discovery first: myfactory discover --print-prompt")
    return 0
