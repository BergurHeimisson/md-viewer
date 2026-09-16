#!/bin/sh
# uninstall.sh — remove md-viewer, and the Java mdviewer it replaced
#
# Finds every install this project could have created, shows you the list,
# and removes it. Paths you own are removed directly; only the root-owned
# ones escalate, so sudo is asked for once and only when it is needed.
#
#   ./uninstall.sh          remove md-viewer and any legacy mdviewer
#   ./uninstall.sh --legacy remove only the old Java mdviewer
#   ./uninstall.sh --yes    skip the confirmation prompt
#
# Nothing outside these paths is ever touched:
#   /usr/local/bin/md-viewer          $(go env GOPATH)/bin/md-viewer
#   /usr/local/bin/mdviewer           /usr/local/lib/mdviewer/

set -e

LEGACY_ONLY=0
ASSUME_YES=0

usage() {
    # Print the header comment block, minus the shebang, and stop at the
    # first line that is not a comment.
    awk 'NR > 1 { if (!/^#/) exit; sub(/^# ?/, ""); print }' "$0"
}

for arg in "$@"; do
    case "$arg" in
        --legacy) LEGACY_ONLY=1 ;;
        --yes|-y) ASSUME_YES=1 ;;
        -h|--help) usage; exit 0 ;;
        *)
            echo "uninstall.sh: unknown option: $arg" >&2
            echo "Try: ./uninstall.sh --help" >&2
            exit 1
            ;;
    esac
done

LEGACY_LIB="/usr/local/lib/mdviewer"
LEGACY_BIN="/usr/local/bin/mdviewer"

# Targets are collected before anything is removed, so the confirmation shows
# the whole picture and a half-finished run cannot surprise anyone.
plain=""   # removable without root
rooted=""  # needs sudo

add() {
    if [ -w "$(dirname "$1")" ]; then
        plain="$plain $1"
    else
        rooted="$rooted $1"
    fi
}

if [ "$LEGACY_ONLY" -eq 0 ]; then
    if command -v go > /dev/null 2>&1; then
        user_bin="$(go env GOBIN)"
        if [ -z "$user_bin" ]; then
            user_bin="$(go env GOPATH)/bin"
        fi
        [ -e "$user_bin/md-viewer" ] && add "$user_bin/md-viewer"
    fi
    [ -e "/usr/local/bin/md-viewer" ] && add "/usr/local/bin/md-viewer"
fi

# The JAR is what identifies a genuine old install. Without it, an unrelated
# mdviewer on PATH is somebody else's program and none of our business.
if [ -f "$LEGACY_LIB/mdviewer.jar" ]; then
    add "$LEGACY_LIB"
    # Only unlink a symlink that actually points into the directory we are
    # deleting; a real file there belongs to someone else.
    if [ -L "$LEGACY_BIN" ]; then
        case "$(readlink "$LEGACY_BIN")" in
            "$LEGACY_LIB"/*) add "$LEGACY_BIN" ;;
        esac
    fi
elif [ -e "$LEGACY_BIN" ] || [ -d "$LEGACY_LIB" ]; then
    echo "Note: found $LEGACY_BIN but no $LEGACY_LIB/mdviewer.jar."
    echo "      That is not this project's install — leaving it alone."
fi

if [ -z "$plain$rooted" ]; then
    echo "Nothing to remove."
    exit 0
fi

echo "Will remove:"
for p in $plain $rooted; do
    echo "  $p"
done

if [ -n "$rooted" ]; then
    echo ""
    echo "These are root-owned and need your password:"
    for p in $rooted; do
        echo "  $p"
    done
fi

if [ "$ASSUME_YES" -eq 0 ]; then
    echo ""
    printf "Proceed? [y/N] "
    read -r reply
    case "$reply" in
        y|Y|yes|YES) ;;
        *) echo "Aborted."; exit 1 ;;
    esac
fi

for p in $plain; do
    echo "==> Removing $p"
    rm -rf "$p"
done

if [ -n "$rooted" ]; then
    if ! sudo -v -p "Root Password: "; then
        echo "" >&2
        echo "uninstall.sh: could not get root, so these are still present:" >&2
        for p in $rooted; do
            echo "  $p" >&2
        done
        echo "Run this script from a terminal where sudo can prompt." >&2
        exit 1
    fi
    for p in $rooted; do
        echo "==> Removing $p"
        sudo rm -rf "$p"
    done
fi

echo ""
echo "Done."
