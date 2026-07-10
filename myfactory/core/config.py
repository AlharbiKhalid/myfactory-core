"""Configuration loading.

Uses PyYAML when available. Falls back to a minimal YAML-subset parser that
understands the files MyFactory itself generates: nested mappings, lists of
scalars, lists of mappings, quoted strings, booleans, null, and numbers.
The fallback keeps the CLI dependency-free on machines without PyYAML.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

try:
    import yaml as _pyyaml  # type: ignore
except ImportError:  # pragma: no cover
    _pyyaml = None


def load_yaml(path: Path) -> Any:
    """Load a YAML file into Python data. Returns {} for missing/empty files."""
    if not path.exists():
        return {}
    text = path.read_text(encoding="utf-8")
    if not text.strip():
        return {}
    if _pyyaml is not None:
        return _pyyaml.safe_load(text)
    return _parse_yaml_subset(text)


def get_path(data: Any, dotted: str, default: Any = None) -> Any:
    """Fetch a nested key like 'plane.enabled' from parsed YAML data."""
    current = data
    for part in dotted.split("."):
        if not isinstance(current, dict) or part not in current:
            return default
        current = current[part]
    return current


# ---------------------------------------------------------------------------
# Minimal YAML-subset fallback parser
# ---------------------------------------------------------------------------


def _parse_scalar(token: str) -> Any:
    token = token.strip()
    if token == "" or token == "~" or token == "null":
        return None
    if token in ("true", "True"):
        return True
    if token in ("false", "False"):
        return False
    if token in ("[]",):
        return []
    if token in ("{}",):
        return {}
    if len(token) >= 2 and token[0] == token[-1] and token[0] in ("'", '"'):
        return token[1:-1]
    try:
        return int(token)
    except ValueError:
        pass
    try:
        return float(token)
    except ValueError:
        pass
    return token


def _strip_comment(line: str) -> str:
    in_single = False
    in_double = False
    for i, ch in enumerate(line):
        if ch == "'" and not in_double:
            in_single = not in_single
        elif ch == '"' and not in_single:
            in_double = not in_double
        elif ch == "#" and not in_single and not in_double:
            if i == 0 or line[i - 1] in (" ", "\t"):
                return line[:i]
    return line


def _parse_yaml_subset(text: str) -> Any:
    lines: list[tuple[int, str]] = []
    for raw in text.splitlines():
        stripped = _strip_comment(raw).rstrip()
        if not stripped.strip():
            continue
        indent = len(stripped) - len(stripped.lstrip(" "))
        lines.append((indent, stripped.strip()))
    value, consumed = _parse_block(lines, 0, 0)
    if consumed != len(lines):
        # Best effort: return what parsed cleanly.
        pass
    return value


def _parse_block(lines: list[tuple[int, str]], index: int, indent: int) -> tuple[Any, int]:
    if index >= len(lines):
        return {}, index
    if lines[index][1].startswith("- ") or lines[index][1] == "-":
        return _parse_list(lines, index, indent)
    return _parse_mapping(lines, index, indent)


def _parse_mapping(lines: list[tuple[int, str]], index: int, indent: int) -> tuple[dict, int]:
    result: dict[str, Any] = {}
    while index < len(lines):
        line_indent, content = lines[index]
        if line_indent < indent or content.startswith("- ") or content == "-":
            break
        if line_indent > indent:
            # Unexpected deeper indent without a parent key; stop.
            break
        if ":" not in content:
            break
        key, _, rest = content.partition(":")
        key = key.strip()
        rest = rest.strip()
        index += 1
        if rest in ("", "|", ">", "|-", ">-"):
            if rest in ("|", ">", "|-", ">-"):
                # Multiline scalar: consume deeper-indented lines as text.
                parts = []
                while index < len(lines) and lines[index][0] > line_indent:
                    parts.append(lines[index][1])
                    index += 1
                result[key] = " ".join(parts)
            elif index < len(lines) and lines[index][0] > line_indent:
                value, index = _parse_block(lines, index, lines[index][0])
                result[key] = value
            else:
                result[key] = None
        else:
            result[key] = _parse_scalar(rest)
    return result, index


def _parse_list(lines: list[tuple[int, str]], index: int, indent: int) -> tuple[list, int]:
    result: list[Any] = []
    while index < len(lines):
        line_indent, content = lines[index]
        if line_indent != indent or not (content.startswith("- ") or content == "-"):
            break
        item_text = content[1:].strip()
        index += 1
        if not item_text:
            if index < len(lines) and lines[index][0] > indent:
                value, index = _parse_block(lines, index, lines[index][0])
                result.append(value)
            else:
                result.append(None)
        elif ":" in item_text and not item_text.startswith(("'", '"')):
            # Inline mapping start: '- key: value' — re-parse as a mapping
            # whose first line is item_text at a virtual deeper indent.
            item: dict[str, Any] = {}
            key, _, rest = item_text.partition(":")
            rest = rest.strip()
            if rest:
                item[key.strip()] = _parse_scalar(rest)
            else:
                if index < len(lines) and lines[index][0] > indent + 2:
                    value, index = _parse_block(lines, index, lines[index][0])
                    item[key.strip()] = value
                else:
                    item[key.strip()] = None
            while index < len(lines) and lines[index][0] == indent + 2 and not lines[index][1].startswith("- "):
                sub, index = _parse_mapping(lines, index, indent + 2)
                item.update(sub)
            result.append(item)
        else:
            result.append(_parse_scalar(item_text))
    return result, index
