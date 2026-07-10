#!/usr/bin/env bash
#
# MyFactory installer: downloads a prebuilt release binary from GitHub
# Releases, verifies its SHA-256 checksum, and installs it into a
# user-writable directory. No Python, Go, or other runtime is required.
#
# Usage:
#   curl -fsSL "https://raw.githubusercontent.com/AlharbiKhalid/myfactory-core/main/scripts/install.sh" | bash
#
# Configuration via environment variables:
#   MYFACTORY_REPOSITORY   GitHub "owner/repo" (default: AlharbiKhalid/myfactory-core)
#   MYFACTORY_VERSION      Release tag to install, e.g. v0.3.0 (default: latest)
#   MYFACTORY_INSTALL_DIR  Install directory (default: $HOME/.local/bin)
#
# Security posture:
#   - curl -f everywhere: HTTP errors fail closed and are never executed.
#   - TLS verification is never disabled.
#   - SHA-256 checksums are verified before installing.
#   - No root required; no credentials stored.
#   - Only versioned GitHub Release assets are downloaded, never branch source.

set -euo pipefail

REPO="${MYFACTORY_REPOSITORY:-AlharbiKhalid/myfactory-core}"
INSTALL_DIR="${MYFACTORY_INSTALL_DIR:-$HOME/.local/bin}"
REQUESTED_VERSION="${MYFACTORY_VERSION:-}"

log() { printf '[myfactory-install] %s\n' "$*"; }
fail() { printf '[myfactory-install] ERROR: %s\n' "$*" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || fail "curl is required."

# --- Detect platform ----------------------------------------------------------

detect_os() {
    case "$(uname -s)" in
        Linux)                     echo linux ;;
        Darwin)                    echo darwin ;;
        MINGW*|MSYS*|CYGWIN*)      echo windows ;;
        *) fail "Unsupported operating system: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo amd64 ;;
        aarch64|arm64)  echo arm64 ;;
        *) fail "Unsupported CPU architecture: $(uname -m)" ;;
    esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# --- Resolve version ----------------------------------------------------------

if [ -z "$REQUESTED_VERSION" ]; then
    log "Resolving latest release of $REPO"
    # GitHub redirects releases/latest to the tagged release URL.
    latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")" \
        || fail "Could not resolve the latest release. Set MYFACTORY_VERSION explicitly."
    VERSION="${latest_url##*/}"
    case "$VERSION" in
        v*) : ;;
        *) fail "Could not parse latest release tag from: $latest_url" ;;
    esac
else
    VERSION="$REQUESTED_VERSION"
    case "$VERSION" in
        v*) : ;;
        *) VERSION="v$VERSION" ;;
    esac
fi
log "Installing MyFactory $VERSION for $OS/$ARCH"

# --- Download and verify ------------------------------------------------------

if [ "$OS" = "windows" ]; then
    ASSET="myfactory_${VERSION}_${OS}_${ARCH}.zip"
    BIN_NAME="myfactory.exe"
else
    ASSET="myfactory_${VERSION}_${OS}_${ARCH}.tar.gz"
    BIN_NAME="myfactory"
fi
BASE_URL="https://github.com/$REPO/releases/download/$VERSION"

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

log "Downloading $ASSET"
curl -fsSL -o "$TMP_DIR/$ASSET" "$BASE_URL/$ASSET" \
    || fail "Download failed: $BASE_URL/$ASSET (does this release ship $OS/$ARCH?)"
curl -fsSL -o "$TMP_DIR/checksums.txt" "$BASE_URL/checksums.txt" \
    || fail "Download failed: $BASE_URL/checksums.txt"

log "Verifying SHA-256 checksum"
expected="$(awk -v asset="$ASSET" '$2 == asset { print $1 }' "$TMP_DIR/checksums.txt")"
[ -n "$expected" ] || fail "checksums.txt has no entry for $ASSET"
if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$TMP_DIR/$ASSET" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$TMP_DIR/$ASSET" | awk '{print $1}')"
else
    fail "Neither sha256sum nor shasum is available to verify the download."
fi
[ "$expected" = "$actual" ] || fail "Checksum mismatch for $ASSET (expected $expected, got $actual). Aborting."
log "Checksum OK"

# --- Extract and install ------------------------------------------------------

case "$ASSET" in
    *.tar.gz) tar -xzf "$TMP_DIR/$ASSET" -C "$TMP_DIR" "$BIN_NAME" ;;
    *.zip)
        command -v unzip >/dev/null 2>&1 || fail "unzip is required to extract $ASSET."
        unzip -q -o "$TMP_DIR/$ASSET" "$BIN_NAME" -d "$TMP_DIR"
        ;;
esac
[ -f "$TMP_DIR/$BIN_NAME" ] || fail "Archive did not contain $BIN_NAME."

mkdir -p "$INSTALL_DIR"
install_path="$INSTALL_DIR/$BIN_NAME"
mv -f "$TMP_DIR/$BIN_NAME" "$install_path"
chmod +x "$install_path"
log "Installed: $install_path"

# --- Verify and report --------------------------------------------------------

"$install_path" version || fail "Installed binary failed to run."

case ":$PATH:" in
    *":$INSTALL_DIR:"*)
        log "Done. Try: myfactory --help"
        ;;
    *)
        cat <<EOF

$INSTALL_DIR is not on your PATH. Add it, e.g.:

    echo 'export PATH="$INSTALL_DIR:\$PATH"' >> ~/.bashrc && source ~/.bashrc

Then run: myfactory --help
EOF
        ;;
esac
