"""MyFactory CLI entrypoint.

Usage:
  python -m myfactory --help
  myfactory init
  myfactory doctor
  myfactory discover --print-prompt
  myfactory plan --dry-run
  myfactory plane sync --dry-run
  myfactory orchestrator prompt
"""

from __future__ import annotations

import argparse
import sys

from myfactory import __version__
from myfactory.commands import discover, doctor, init, orchestrator, plan, plane


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="myfactory",
        description=(
            "MyFactory: reusable AI software factory CLI. "
            "init sets up structure; discovery/planning are done by AI agents "
            "using generated prompts."
        ),
    )
    parser.add_argument("--version", action="version", version=f"myfactory {__version__}")

    subparsers = parser.add_subparsers(dest="command", metavar="COMMAND")

    init.register(subparsers)
    doctor.register(subparsers)
    discover.register(subparsers)
    plan.register(subparsers)
    plane.register(subparsers)
    orchestrator.register(subparsers)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    if not getattr(args, "func", None):
        parser.print_help()
        return 0
    try:
        return int(args.func(args) or 0)
    except KeyboardInterrupt:
        print("Interrupted.", file=sys.stderr)
        return 130
    except Exception as exc:  # keep CLI failures readable
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
