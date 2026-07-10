#!/usr/bin/env bash
#
# Build MyFactory from the current local checkout and install the binary
# (for development/testing). End users should use scripts/install.sh, which
# downloads prebuilt release binaries and needs no toolchain.
#
# Requires: Go 1.23+.
#
# Usage:
#   bash scripts/install-local.sh
#
# Configuration:
#   MYFACTORY_INSTALL_DIR  Install directory (default: $HOME/.local/bin)

set -euo pipefail

INSTALL_DIR="${MYFACTORY_INSTALL_DIR:-$HOME/.local/bin}"

log() { printf '[myfactory-install-local] %s\n' "$*"; }
fail() { printf '[myfactory-install-local] ERROR: %s\n' "$*" >&2; exit 1; }

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" >/dev/null 2>&1 && pwd)"
core_dir="$(cd "$script_dir/.." && pwd)"
[ -f "$core_dir/cmd/myfactory/main.go" ] || fail "run this from a myfactory-core checkout."
command -v go >/dev/null 2>&1 || fail "Go 1.23+ is required to build from source. End users: use scripts/install.sh instead."

BIN_NAME="myfactory"
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) BIN_NAME="myfactory.exe" ;;
esac

VERSION="dev-local"
COMMIT="$(git -C "$core_dir" rev-parse --short HEAD 2>/dev/null || echo unknown)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-X github.com/AlharbiKhalid/myfactory-core/internal/version.Version=$VERSION \
 -X github.com/AlharbiKhalid/myfactory-core/internal/version.Commit=$COMMIT \
 -X github.com/AlharbiKhalid/myfactory-core/internal/version.Date=$DATE"

log "Building $BIN_NAME from $core_dir"
mkdir -p "$INSTALL_DIR"
CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o "$INSTALL_DIR/$BIN_NAME" "$core_dir/cmd/myfactory"
log "Installed: $INSTALL_DIR/$BIN_NAME"

"$INSTALL_DIR/$BIN_NAME" version

case ":$PATH:" in
    *":$INSTALL_DIR:"*)
        log "Done. Try: myfactory --help"
        ;;
    *)
        cat <<EOF

$INSTALL_DIR is not on your PATH. Add it, e.g.:

    echo 'export PATH="$INSTALL_DIR:\$PATH"' >> ~/.bashrc && source ~/.bashrc
EOF
        ;;
esac
