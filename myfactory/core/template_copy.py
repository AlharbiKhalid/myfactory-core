"""Recursive template copying that never overwrites without force."""

from __future__ import annotations

from pathlib import Path

from myfactory.core.files import FileActions, copy_file

# Files that should never be copied out of templates.
IGNORED_NAMES = {".DS_Store", "__pycache__", ".git"}


def copy_tree(
    src_dir: Path,
    dest_dir: Path,
    actions: FileActions,
    force: bool = False,
    dry_run: bool = False,
    exclude: set[str] | None = None,
) -> None:
    """Copy every file under src_dir into dest_dir.

    - Existing destination files are skipped unless force is set.
    - Nothing is ever deleted.
    - `exclude` holds POSIX-style relative paths to skip entirely.
    """
    exclude = exclude or set()
    for src in sorted(src_dir.rglob("*")):
        if any(part in IGNORED_NAMES for part in src.parts):
            continue
        rel = src.relative_to(src_dir)
        if rel.as_posix() in exclude:
            continue
        dest = dest_dir / rel
        if src.is_dir():
            if not dry_run:
                dest.mkdir(parents=True, exist_ok=True)
            continue
        copy_file(src, dest, actions, force=force, dry_run=dry_run)
