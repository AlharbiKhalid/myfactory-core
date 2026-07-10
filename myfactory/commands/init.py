"""`myfactory init` — non-interactive project setup.

Adds MyFactory structure to an existing repository:
- copies templates/product-repo (never overwriting existing files),
- applies Codex (AGENTS.md) and Claude (.claude/commands) overlays,
- creates .ApplicationFactory metadata,
- fills placeholders where safe.

It asks no questions, calls no AI tools, and contacts no external services.
Product/business/architecture discovery is done afterwards by AI agents.
"""

from __future__ import annotations

import argparse
import re

from myfactory.core import git as gitutil
from myfactory.core import paths
from myfactory.core.files import FileActions, ensure_dir, replace_placeholders_if_safe, yaml_string
from myfactory.core.template_copy import copy_tree

KEY_PATTERN = re.compile(r"[A-Z][A-Z0-9_]{1,15}")


def register(subparsers: argparse._SubParsersAction) -> None:
    p = subparsers.add_parser(
        "init",
        help="Set up MyFactory structure in a repo (non-interactive).",
        description=(
            "Non-interactive setup. Copies missing factory files into the target "
            "repo. Existing files are always preserved unless --force is given. "
            "Nothing is ever deleted. Discovery is done later by AI agents."
        ),
    )
    p.add_argument("--target", default=None, help="Target directory (default: current directory).")
    p.add_argument("--key", default=None, help="Stable project key (default: derived from folder name).")
    p.add_argument("--name", default=None, help="Product name (default: derived from folder name).")
    p.add_argument("--description", default="", help="Short product description.")
    p.add_argument(
        "--git-provider",
        choices=["github", "gitlab", "none"],
        default=None,
        help="Git provider (default: auto-detect, falling back to github).",
    )
    codex = p.add_mutually_exclusive_group()
    codex.add_argument("--with-codex", dest="codex", action="store_true", default=True,
                       help="Include Codex overlay (AGENTS.md). Default.")
    codex.add_argument("--no-codex", dest="codex", action="store_false",
                       help="Skip Codex overlay.")
    claude = p.add_mutually_exclusive_group()
    claude.add_argument("--with-claude", dest="claude", action="store_true", default=True,
                        help="Include Claude overlay (.claude/commands). Default.")
    claude.add_argument("--no-claude", dest="claude", action="store_false",
                        help="Skip Claude overlay.")
    p.add_argument("--with-github", action="store_true", help="Force-include GitHub helper files.")
    p.add_argument("--with-gitlab", action="store_true", help="Force-include GitLab helper files.")
    p.add_argument("--force", action="store_true", help="Overwrite existing files (default: never).")
    p.add_argument("--dry-run", action="store_true", help="Print planned actions without changing anything.")
    p.set_defaults(func=run)


def derive_key(folder_name: str) -> str:
    """Derive a project key like CLINIC_BOOKING from a folder name."""
    cleaned = re.sub(r"[^A-Za-z0-9]+", "_", folder_name).strip("_").upper()
    if not cleaned:
        return "APP"
    if not cleaned[0].isalpha():
        cleaned = "P_" + cleaned
    cleaned = cleaned[:16]
    if not KEY_PATTERN.fullmatch(cleaned):
        return "APP"
    return cleaned


def derive_name(folder_name: str) -> str:
    words = re.sub(r"[-_]+", " ", folder_name).strip()
    return words.title() if words else "New Product"


def run(args: argparse.Namespace) -> int:
    target = paths.resolve_target(args.target)
    if not target.exists():
        print(f"ERROR: target directory does not exist: {target}")
        return 1

    key = (args.key or derive_key(target.name)).strip().upper()
    if not KEY_PATTERN.fullmatch(key):
        print(
            "ERROR: project key must be 2-16 chars, uppercase, A-Z/0-9/_ and start "
            f"with a letter. Got: {key}"
        )
        return 1
    name = (args.name or derive_name(target.name)).strip()
    description = args.description.strip()

    if args.git_provider:
        provider = args.git_provider
    else:
        detected = gitutil.detect_provider(target)
        provider = detected if detected != "unknown" else "github"
    include_github = args.with_github or provider == "github"
    include_gitlab = args.with_gitlab or provider == "gitlab"

    dry_run = args.dry_run
    force = args.force
    actions = FileActions()

    if dry_run:
        print("DRY RUN: no files will be changed.\n")

    # 1. Product template (docs, .ApplicationFactory, .github when selected).
    exclude: set[str] = set()
    if not include_github:
        template_root = paths.product_template_dir()
        gh_dir = template_root / ".github"
        if gh_dir.exists():
            for f in gh_dir.rglob("*"):
                if f.is_file():
                    exclude.add(f.relative_to(template_root).as_posix())
    copy_tree(paths.product_template_dir(), target, actions, force=force, dry_run=dry_run, exclude=exclude)

    # 2. Ensure metadata directories exist even if templates change.
    ensure_dir(paths.task_packages_dir(target), dry_run=dry_run)
    ensure_dir(paths.orchestrator_dir(target), dry_run=dry_run)

    # 3. GitLab helper files (template ships GitHub ones; GitLab is generated).
    if include_gitlab:
        from myfactory.core.files import write_file

        mr_template = target / ".gitlab" / "merge_request_templates" / "Default.md"
        gh_pr = paths.product_template_dir() / ".github" / "pull_request_template.md"
        content = gh_pr.read_text(encoding="utf-8") if gh_pr.exists() else "# Merge Request\n"
        write_file(mr_template, content, actions, force=force, dry_run=dry_run)

    # 4. Codex overlay: AGENTS.md only if absent (or --force).
    if args.codex and paths.codex_overlay_dir().exists():
        copy_tree(paths.codex_overlay_dir(), target, actions, force=force, dry_run=dry_run)

    # 5. Claude overlay: .claude/commands/*.
    if args.claude and paths.claude_overlay_dir().exists():
        copy_tree(paths.claude_overlay_dir(), target, actions, force=force, dry_run=dry_run)

    # 6. Fill placeholders where the placeholder text still exists.
    replacements = [
        ("key: CHANGE_ME", f"key: {yaml_string(key)}"),
        ("name: CHANGE_ME", f"name: {yaml_string(name)}"),
        ("description: CHANGE_ME", f"description: {yaml_string(description or 'CHANGE_ME')}"),
    ]
    placeholder_files = [
        paths.product_manifest_path(target),
        paths.project_config_path(target),
        target / "docs" / "03-delivery" / "work-breakdown.yaml",
    ]
    updated = []
    if not dry_run:
        for f in placeholder_files:
            if replace_placeholders_if_safe(f, list(replacements)):
                updated.append(f)
        config_file = paths.project_config_path(target)
        if provider != "github":
            replace_placeholders_if_safe(
                config_file,
                [("provider: github        # github | gitlab | none",
                  f"provider: {provider}        # github | gitlab | none")],
            )

    # 7. Summary.
    print(f"MyFactory init {'planned' if dry_run else 'complete'} for: {target}")
    print(f"Project key:  {key}")
    print(f"Product name: {name}")
    print(f"Git provider: {provider}"
          + (" (detected)" if not args.git_provider and gitutil.detect_provider(target) != "unknown" else ""))
    print(f"Git repo:     {'yes' if gitutil.is_git_repo(target) else 'no'}")
    print(f"Overlays:     codex={'on' if args.codex else 'off'}, claude={'on' if args.claude else 'off'}")
    print()
    for line in actions.summary_lines(base=target):
        print(line)
    if updated:
        print()
        print("Placeholders filled in:")
        for f in updated:
            print(f"  {f.relative_to(target)}")
    print()
    print("Next steps:")
    print("  myfactory doctor")
    print("  myfactory discover --print-prompt")
    print("  myfactory plan --dry-run")
    return 0
