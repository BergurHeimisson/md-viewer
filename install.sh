#!/bin/sh
# install.sh — build and install md-viewer
#
# Installs a single static binary. By default that is /usr/local/bin, which
# needs sudo; --user installs into your Go bin directory instead and needs no
# password, which is what to reach for when sudo has no terminal to prompt on.
#
#   ./install.sh           system-wide, /usr/local/bin/md-viewer
#   ./install.sh --user    per-user, $(go env GOBIN) or $(go env GOPATH)/bin
#
# Requires: Go 1.26+

set -e

BIN_NAME="md-viewer"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
USER_INSTALL=0

usage() {
    sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
}

for arg in "$@"; do
    case "$arg" in
        --user) USER_INSTALL=1 ;;
        -h|--help) usage; exit 0 ;;
        *)
            echo "install.sh: unknown option: $arg" >&2
            echo "Try: ./install.sh --help" >&2
            exit 1
            ;;
    esac
done

cd "$SCRIPT_DIR"

if ! command -v go > /dev/null 2>&1; then
    echo "install.sh: Go is not installed or not on PATH." >&2
    echo "  macOS:  brew install go" >&2
    echo "  Linux:  https://go.dev/dl/" >&2
    exit 1
fi

# Resolve the destination before building, so a bad target fails fast.
if [ "$USER_INSTALL" -eq 1 ]; then
    BIN_DIR="$(go env GOBIN)"
    if [ -z "$BIN_DIR" ]; then
        BIN_DIR="$(go env GOPATH)/bin"
    fi
else
    BIN_DIR="/usr/local/bin"
fi

# Tag the binary with the current commit so --version is meaningful.
VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

echo "==> Building $BIN_NAME $VERSION..."
go build -trimpath -ldflags "-s -w -X main.Version=$VERSION" \
    -o "$SCRIPT_DIR/$BIN_NAME" ./cmd/md-viewer

echo "==> Installing to $BIN_DIR/$BIN_NAME..."
if [ "$USER_INSTALL" -eq 1 ]; then
    mkdir -p "$BIN_DIR"
    install -m 755 "$SCRIPT_DIR/$BIN_NAME" "$BIN_DIR/$BIN_NAME"
else
    # sudo cannot prompt without a terminal — say so rather than failing with
    # its own opaque "a password is required".
    if ! sudo -v -p "Root Password: "; then
        echo "" >&2
        echo "install.sh: could not get root. Either run this from a terminal," >&2
        echo "or install without sudo:  ./install.sh --user" >&2
        exit 1
    fi
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
fi

# An install nobody can run is not an install.
case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *)
        echo ""
        echo "Note: $BIN_DIR is not on your PATH. Add it to your shell profile:"
        echo "  export PATH=\"$BIN_DIR:\$PATH\""
        ;;
esac

echo ""
echo "Done. Run:  $BIN_NAME path/to/file.md"
if [ "$USER_INSTALL" -eq 1 ]; then
    echo "Uninstall: rm -f $BIN_DIR/$BIN_NAME"
else
    echo "Uninstall: sudo rm -f $BIN_DIR/$BIN_NAME"
fi
