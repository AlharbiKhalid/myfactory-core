"""Safe file operations with created/skipped/overwritten tracking."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path


def yaml_string(value: str) -> str:
    """Return a YAML-safe string using JSON string escaping."""
    return json.dumps(value, ensure_ascii=False)


@dataclass
class FileActions:
    """Records what happened to each file during an init/copy run."""

    created: list[Path] = field(default_factory=list)
    skipped: list[Path] = field(default_factory=list)
    overwritten: list[Path] = field(default_factory=list)

    def summary_lines(self, base: Path | None = None) -> list[str]:
        def rel(p: Path) -> str:
            if base is not None:
                try:
                    return str(p.relative_to(base))
                except ValueError:
                    pass
            return str(p)

        lines = [
            f"Files created:     {len(self.created)}",
            f"Files skipped:     {len(self.skipped)}",
            f"Files overwritten: {len(self.overwritten)}",
        ]
        for label, items in (
            ("created", self.created),
            ("skipped", self.skipped),
            ("overwritten", self.overwritten),
        ):
            for p in items:
                lines.append(f"  [{label}] {rel(p)}")
        return lines


def write_file(
    path: Path,
    content: str,
    actions: FileActions,
    force: bool = False,
    dry_run: bool = False,
) -> None:
    """Write content to path, respecting no-overwrite-by-default."""
    if path.exists():
        if not force:
            actions.skipped.append(path)
            return
        if not dry_run:
            path.write_text(content, encoding="utf-8")
        actions.overwritten.append(path)
        return
    if not dry_run:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
    actions.created.append(path)


def copy_file(
    src: Path,
    dest: Path,
    actions: FileActions,
    force: bool = False,
    dry_run: bool = False,
) -> None:
    """Copy a single file, respecting no-overwrite-by-default."""
    content = src.read_bytes()
    if dest.exists():
        if not force:
            actions.skipped.append(dest)
            return
        if not dry_run:
            dest.write_bytes(content)
        actions.overwritten.append(dest)
        return
    if not dry_run:
        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_bytes(content)
    actions.created.append(dest)


def ensure_dir(path: Path, dry_run: bool = False) -> None:
    if not dry_run:
        path.mkdir(parents=True, exist_ok=True)


def replace_placeholders_if_safe(path: Path, replacements: list[tuple[str, str]], dry_run: bool = False) -> bool:
    """Replace placeholder strings in a file only where the placeholder exists.

    Returns True if any replacement was applied. Never raises when a
    placeholder is absent (the file may have been customized already).
    """
    if not path.exists():
        return False
    text = path.read_text(encoding="utf-8")
    changed = False
    for old, new in replacements:
        if old in text:
            text = text.replace(old, new, 1)
            changed = True
    if changed and not dry_run:
        path.write_text(text, encoding="utf-8")
    return changed
