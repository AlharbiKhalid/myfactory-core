"""`myfactory discover` - print the discovery prompt for AI agents.

The CLI never runs discovery itself: it produces the prompt that the user
pastes into Claude, Codex, or another agent.
"""

from __future__ import annotations

import argparse

from myfactory.core import paths

DISCOVERY_PROMPT = """\
# MyFactory Product Discovery Prompt

You are an AI product discovery agent working inside a MyFactory-enabled
repository at: {target}

MyFactory rules: Git is the source of truth. Chat is not. Your job is to
understand the product by talking with the user, then write everything into
the source-of-truth docs.

## Step 1 - Inspect

- Read `.ApplicationFactory/product.yaml` and `.ApplicationFactory/config.yaml`.
- Read the existing repository: README, code, docs. If this is an existing
  application, infer as much as possible before asking questions.
- Read `docs/00-product/` to see what is already filled vs `CHANGE_ME`.

## Step 2 - Talk with the user

Ask focused questions (a few at a time) about the problem, target users, core
journeys, MVP scope, what is out of scope, constraints, and success criteria.
Propose drafts for confirmation instead of interrogating.

## Step 3 - Fill the product docs

Write your findings into:

- docs/00-product/idea-brief.md
- docs/00-product/prd.md
- docs/00-product/user-journeys.md
- docs/00-product/acceptance-criteria.md

Requirements:

- Record every assumption in an "Assumptions" section.
- Record unresolved items in an "Open Questions" section. Never invent answers.
- Preserve real existing content; only replace placeholders.

## Step 4 - Boundaries

- Do NOT create implementation tasks unless the user explicitly asks.
- Do NOT write application code.
- Do NOT edit business rules, architecture, or delivery files in this session
  (those have their own commands/agents).

## Step 5 - Finish

Summarize what you wrote and what remains open. If the repo uses git, offer to
commit the doc changes with message:
`docs(product): populate product discovery docs`

Afterwards the user should continue with business rules discovery and then
architecture (in a Claude session: /myfactory-business-rules and
/myfactory-architecture).
{agent_note}"""

AGENT_NOTES = {
    "claude": (
        "\nNote for Claude: this repo ships ready-made commands in "
        ".claude/commands/ - /myfactory-discover is the interactive version "
        "of this prompt.\n"
    ),
    "codex": (
        "\nNote for Codex: read AGENTS.md in the repo root first; it defines "
        "the factory rules you operate under.\n"
    ),
}


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "discover",
        help="Print the AI discovery prompt (the CLI never runs discovery itself).",
        description=(
            "Prints a prompt to paste into Claude/Codex. The agent then talks "
            "with you and fills the product source-of-truth docs."
        ),
    )
    p.add_argument("--target", default=None, help="Target directory (default: current directory).")
    p.add_argument("--agent", choices=["claude", "codex"], default=None,
                   help="Tailor the prompt for a specific agent.")
    p.add_argument("--print-prompt", action="store_true",
                   help="Print the discovery prompt (default action).")
    p.set_defaults(func=run)


def run(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)
    note = AGENT_NOTES.get(args.agent or "", "")
    print(DISCOVERY_PROMPT.format(target=target, agent_note=note))
    return 0
