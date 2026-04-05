#!/bin/sh
# install.sh — build and install mdViewer system-wide
#
# Installs to:
#   /usr/local/lib/mdviewer/mdviewer.jar  (the fat JAR)
#   /usr/local/bin/mdviewer               (symlink to the launcher script)
#
# Requires: Java 25+, Maven 3.6+
# Uninstall: sudo rm -rf /usr/local/lib/mdviewer /usr/local/bin/mdviewer

set -e

LIB_DIR="/usr/local/lib/mdviewer"
BIN_DIR="/usr/local/bin"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "==> Building mdViewer..."
cd "$SCRIPT_DIR"
mvn -q package

# Prompt once with a clear label; subsequent sudo calls reuse the cached token
sudo -v -p "Root Password: "

echo "==> Installing to $LIB_DIR..."
sudo mkdir -p "$LIB_DIR"
sudo cp target/mdviewer.jar "$LIB_DIR/mdviewer.jar"
sudo cp mdviewer "$LIB_DIR/mdviewer"
sudo chmod +x "$LIB_DIR/mdviewer"

echo "==> Linking to $BIN_DIR/mdviewer..."
sudo mkdir -p "$BIN_DIR"
sudo ln -sf "$LIB_DIR/mdviewer" "$BIN_DIR/mdviewer"

echo ""
echo "Done. Run:  mdviewer path/to/file.md"
echo "Uninstall: sudo rm -rf $LIB_DIR $BIN_DIR/mdviewer"
