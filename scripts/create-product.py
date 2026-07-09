#!/usr/bin/env python3

"""
Create a new product repository from the reusable MyFactory product template.

Usage:

  python scripts/create-product.py \
    --target ../my-new-product \
    --key APP \
    --name "My New Product" \
    --description "Short product description"

This script:
- Copies templates/product-repo into the target directory.
- Preserves hidden folders like .ApplicationFactory and .github.
- Updates basic project metadata.
- Does not create Plane issues.
- Does not call agents.
- Does not initialize Git unless --init-git is provided.
"""

from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
import sys
from pathlib import Path


def yaml_string(value: str) -> str:
    """
    Return a YAML-safe string using JSON string escaping.
    YAML accepts JSON-style quoted strings.
    """
    return json.dumps(value, ensure_ascii=False)


def replace_once(text: str, old: str, new: str, file_path: Path) -> str:
    if old not in text:
        raise RuntimeError(f"Expected text not found in {file_path}: {old}")
    return text.replace(old, new, 1)


def update_file(path: Path, replacements: list[tuple[str, str]]) -> None:
    text = path.read_text(encoding="utf-8")
    for old, new in replacements:
        text = replace_once(text, old, new, path)
    path.write_text(text, encoding="utf-8")


def validate_project_key(project_key: str) -> None:
    if not re.fullmatch(r"[A-Z][A-Z0-9_]{1,15}", project_key):
        raise ValueError(
            "Project key must be 2-16 characters, uppercase, and contain only A-Z, 0-9, or _. "
            "Examples: APP, CRM, CLINIC, BILLING_1"
        )


def copy_template(template_dir: Path, target_dir: Path) -> None:
    if not template_dir.exists():
        raise FileNotFoundError(f"Template directory does not exist: {template_dir}")

    if target_dir.exists():
        existing = list(target_dir.iterdir())
        if existing:
            raise FileExistsError(
                f"Target directory already exists and is not empty: {target_dir}"
            )
        target_dir.rmdir()

    shutil.copytree(template_dir, target_dir)


def init_git_repo(target_dir: Path) -> None:
    subprocess.run(["git", "init"], cwd=target_dir, check=True)
    subprocess.run(["git", "add", "."], cwd=target_dir, check=True)
    subprocess.run(
        ["git", "commit", "-m", "Initialize product repository from MyFactory template"],
        cwd=target_dir,
        check=True,
    )


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Create a new product repository from the MyFactory template."
    )

    parser.add_argument(
        "--target",
        required=True,
        help="Target directory for the new product repository.",
    )

    parser.add_argument(
        "--key",
        required=True,
        help="Stable project key. Example: APP, CRM, CLINIC.",
    )

    parser.add_argument(
        "--name",
        required=True,
        help='Human-readable product name. Example: "Clinic Booking System".',
    )

    parser.add_argument(
        "--description",
        default="",
        help="Short product description.",
    )

    parser.add_argument(
        "--init-git",
        action="store_true",
        help="Initialize a Git repository and create the first commit.",
    )

    args = parser.parse_args()

    project_key = args.key.strip().upper()
    product_name = args.name.strip()
    description = args.description.strip() or "CHANGE_ME"

    validate_project_key(project_key)

    core_dir = Path(__file__).resolve().parents[1]
    template_dir = core_dir / "templates" / "product-repo"
    target_dir = Path(args.target).expanduser().resolve()

    copy_template(template_dir, target_dir)

    product_manifest = target_dir / ".ApplicationFactory" / "product.yaml"
    work_breakdown = target_dir / "docs" / "03-delivery" / "work-breakdown.yaml"
    readme = target_dir / "README.md"

    update_file(
        product_manifest,
        [
            ("key: CHANGE_ME", f"key: {yaml_string(project_key)}"),
            ("name: CHANGE_ME", f"name: {yaml_string(product_name)}"),
            ("description: CHANGE_ME", f"description: {yaml_string(description)}"),
        ],
    )

    update_file(
        work_breakdown,
        [
            ("key: CHANGE_ME", f"key: {yaml_string(project_key)}"),
            ("name: CHANGE_ME", f"name: {yaml_string(product_name)}"),
            ("description: CHANGE_ME", f"description: {yaml_string(description)}"),
        ],
    )

    if readme.exists():
        text = readme.read_text(encoding="utf-8")
        text = text.replace("# Product Repository", f"# {product_name}", 1)
        readme.write_text(text, encoding="utf-8")

    if args.init_git:
        init_git_repo(target_dir)

    print("MyFactory product repository created.")
    print(f"Target: {target_dir}")
    print(f"Project key: {project_key}")
    print(f"Product name: {product_name}")
    print("")
    print("Next files to edit:")
    print("- docs/00-product/idea-brief.md")
    print("- docs/00-product/prd.md")
    print("- docs/01-business/business-rules.yaml")
    print("- docs/03-delivery/work-breakdown.yaml")

    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
