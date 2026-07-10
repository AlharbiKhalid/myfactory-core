"""Path resolution for MyFactory core, templates, and project metadata."""

from __future__ import annotations

from pathlib import Path

METADATA_DIR_NAME = ".ApplicationFactory"


def core_root() -> Path:
    """Return the root of the myfactory-core checkout (parent of the package)."""
    return Path(__file__).resolve().parents[2]


def templates_dir() -> Path:
    return core_root() / "templates"


def product_template_dir() -> Path:
    return templates_dir() / "product-repo"


def overlays_dir() -> Path:
    return templates_dir() / "project-overlays"


def codex_overlay_dir() -> Path:
    return overlays_dir() / "codex"


def claude_overlay_dir() -> Path:
    return overlays_dir() / "claude"


def metadata_dir(target: Path) -> Path:
    return target / METADATA_DIR_NAME


def product_manifest_path(target: Path) -> Path:
    return metadata_dir(target) / "product.yaml"


def project_config_path(target: Path) -> Path:
    return metadata_dir(target) / "config.yaml"


def task_packages_dir(target: Path) -> Path:
    return metadata_dir(target) / "task-packages"


def orchestrator_dir(target: Path) -> Path:
    return metadata_dir(target) / "orchestrator"


def resolve_target(target: str | None) -> Path:
    """Resolve a --target argument (default: current working directory)."""
    if target:
        return Path(target).expanduser().resolve()
    return Path.cwd().resolve()
