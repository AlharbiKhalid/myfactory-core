"""`myfactory orchestrator prompt` - print the Hermes controller prompt.

Prefers the project's own copy under .ApplicationFactory/orchestrator/;
falls back to the core template so the command works before init.
"""

from __future__ import annotations

import argparse

from myfactory.core import paths

PROMPT_FILENAME = "HERMES-CONTROLLER-PROMPT.md"


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "orchestrator",
        help="Hermes orchestrator/controller helpers.",
        description="Hermes controls sprint execution: it delegates to agents and enforces QA gates.",
    )
    sub = p.add_subparsers(dest="orchestrator_command", metavar="SUBCOMMAND")

    prompt = sub.add_parser(
        "prompt",
        help="Print the Hermes controller prompt.",
        description=(
            "Prints the prompt to give the Hermes controller agent. Uses the "
            "project's .ApplicationFactory/orchestrator/ copy when present, "
            "otherwise the core template."
        ),
    )
    prompt.add_argument("--target", default=None, help="Target directory (default: current directory).")
    prompt.set_defaults(func=run_prompt)

    p.set_defaults(func=lambda args: (p.print_help(), 0)[1])


def run_prompt(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)

    project_prompt = paths.orchestrator_dir(target) / PROMPT_FILENAME
    template_prompt = (
        paths.product_template_dir() / ".ApplicationFactory" / "orchestrator" / PROMPT_FILENAME
    )

    if project_prompt.is_file():
        source = project_prompt
        origin = "project"
    elif template_prompt.is_file():
        source = template_prompt
        origin = "core template (project not initialized - run `myfactory init`)"
    else:
        print("ERROR: Hermes controller prompt not found in project or core templates.")
        return 1

    print(f"# Source: {origin}: {source}")
    print()
    print(source.read_text(encoding="utf-8"))
    return 0
