#!/usr/bin/env bash
#
# Install MyFactory from the current local checkout (for development/testing).
#
# Creates a `myfactory` shim in ~/.local/bin (or $MYFACTORY_BIN_DIR) that runs
# `python -m myfactory` against this checkout. No copying, no root, no secrets.
#
# Usage:
#   bash scripts/install-local.sh

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" >/dev/null 2>&1 && pwd)"
core_dir="$(cd "$script_dir/.." && pwd)"

[ -f "$core_dir/myfactory/cli.py" ] || {
    echo "ERROR: run this from a myfactory-core checkout (myfactory/cli.py not found)." >&2
    exit 1
}

# Delegate to the main installer in local mode (no repo URL => local checkout).
MYFACTORY_REPO_URL="" bash "$script_dir/install.sh"
