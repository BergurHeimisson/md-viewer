#!/bin/sh
# install.sh — build and install md-viewer system-wide
#
# Installs a single static binary to /usr/local/bin/md-viewer.
# No PATH or shell configuration is needed afterwards.
#
# Requires: Go 1.26+
# Uninstall: sudo rm -f /usr/local/bin/md-viewer

set -e

BIN_DIR="/usr/local/bin"
BIN_NAME="md-viewer"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

cd "$SCRIPT_DIR"

if ! command -v go > /dev/null 2>&1; then
    echo "install.sh: Go is not installed or not on PATH." >&2
    echo "  macOS:  brew install go" >&2
    echo "  Linux:  https://go.dev/dl/" >&2
    exit 1
fi

# Tag the binary with the current commit so --version is meaningful.
VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

echo "==> Building $BIN_NAME $VERSION..."
go build -trimpath -ldflags "-s -w -X main.Version=$VERSION" \
    -o "$SCRIPT_DIR/$BIN_NAME" ./cmd/md-viewer

# Prompt once with a clear label; subsequent sudo calls reuse the cached token
sudo -v -p "Root Password: "

echo "==> Installing to $BIN_DIR/$BIN_NAME..."
sudo mkdir -p "$BIN_DIR"
sudo install -m 755 "$SCRIPT_DIR/$BIN_NAME" "$BIN_DIR/$BIN_NAME"

# Clean up the Java install if it is still around from a previous version.
# Gated on the JAR so an unrelated mdviewer on PATH is never deleted.
if [ -f "/usr/local/lib/mdviewer/mdviewer.jar" ]; then
    echo "==> Removing the old Java mdviewer install..."
    sudo rm -rf /usr/local/lib/mdviewer
    if [ -L "/usr/local/bin/mdviewer" ]; then
        sudo rm -f /usr/local/bin/mdviewer
    fi
fi

echo ""
echo "Done. Run:  $BIN_NAME path/to/file.md"
echo "Uninstall: sudo rm -f $BIN_DIR/$BIN_NAME"
