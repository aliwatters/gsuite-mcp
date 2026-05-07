#!/usr/bin/env bash
#
# install.sh — Standalone installer for gsuite-mcp
#
# Builds and installs the gsuite-mcp binary with a recorded version stamp so
# future runs can skip a no-op rebuild.
#
# Usage:
#   ./install.sh              # build + install if HEAD changed since last install
#   ./install.sh --force      # force rebuild and reinstall
#   FORCE_REBUILD=1 ./install.sh   # same as --force
#   INSTALL_PREFIX=/opt ./install.sh
#
# Environment:
#   INSTALL_PREFIX   Install root (default: $HOME/.local). gsuite-mcp goes to
#                    $INSTALL_PREFIX/bin/gsuite-mcp.
#   XDG_DATA_HOME    Where the version stamp is recorded. Defaults to
#                    $HOME/.local/share. The stamp file is at
#                    $XDG_DATA_HOME/gsuite-mcp/version.
#   FORCE_REBUILD    When set to 1 (or --force passed), forces rebuild.
#
# Exit codes:
#   0  success (or no-op when up to date)
#   1  build/install failure or invalid environment
#
# Notes:
#   - Idempotent: running twice in a row is a no-op the second time.
#   - Standalone: works without dotfiles. Requires Go, git, and bash (3.2+).
#   - Rebuild is triggered only by committed changes (new git HEAD). Uncommitted
#     edits do not change HEAD, so run with --force if you want to rebuild after
#     local changes without committing first.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# --- Argument parsing -------------------------------------------------------

FORCE=0
case "${1:-}" in
    --force|-f)
        FORCE=1
        shift
        ;;
    --help|-h)
        sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
    "")
        ;;
    *)
        echo "Error: unknown argument: $1" >&2
        echo "Usage: $0 [--force|-f] [--help|-h]" >&2
        exit 1
        ;;
esac

# Reject extra positional args so typos don't pass silently.
if [[ $# -gt 0 ]]; then
    echo "Error: unexpected extra arguments: $*" >&2
    echo "Usage: $0 [--force|-f] [--help|-h]" >&2
    exit 1
fi

if [[ "${FORCE_REBUILD:-0}" == "1" ]]; then
    FORCE=1
fi

# --- Configuration ----------------------------------------------------------

INSTALL_PREFIX="${INSTALL_PREFIX:-$HOME/.local}"
BIN_DIR="$INSTALL_PREFIX/bin"
BINARY="$BIN_DIR/gsuite-mcp"

XDG_DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
STAMP_DIR="$XDG_DATA_HOME/gsuite-mcp"
STAMP_FILE="$STAMP_DIR/version"

# --- Pre-flight -------------------------------------------------------------

if ! command -v go >/dev/null 2>&1; then
    echo "Error: go not found on PATH; install Go before running this script" >&2
    exit 1
fi

if ! command -v git >/dev/null 2>&1; then
    echo "Error: git not found on PATH; install git before running this script" >&2
    exit 1
fi

# .git is a directory in a normal clone, but a *file* in a worktree.
if [[ ! -e "$SCRIPT_DIR/.git" ]]; then
    echo "Error: $SCRIPT_DIR is not a git repo (no .git)" >&2
    echo "       install.sh must run from inside a clone of aliwatters/gsuite-mcp" >&2
    exit 1
fi

if [[ ! -f "$SCRIPT_DIR/go.mod" ]]; then
    echo "Error: go.mod not found at $SCRIPT_DIR" >&2
    exit 1
fi

if [[ ! -d "$SCRIPT_DIR/cmd/gsuite-mcp" ]]; then
    echo "Error: expected ./cmd/gsuite-mcp directory in $SCRIPT_DIR" >&2
    exit 1
fi

# --- Version detection ------------------------------------------------------

GIT_HEAD="$(git -C "$SCRIPT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)"
if [[ "$GIT_HEAD" == "unknown" ]]; then
    echo "Error: could not resolve git HEAD" >&2
    exit 1
fi

INSTALLED_HEAD=""
if [[ -f "$STAMP_FILE" ]]; then
    INSTALLED_HEAD="$(cat "$STAMP_FILE" 2>/dev/null || true)"
fi

# --- Idempotency check ------------------------------------------------------

if [[ "$FORCE" -ne 1 ]] \
   && [[ -x "$BINARY" ]] \
   && [[ -n "$INSTALLED_HEAD" ]] \
   && [[ "$INSTALLED_HEAD" == "$GIT_HEAD" ]]; then
    echo "gsuite-mcp already installed at $GIT_HEAD; nothing to do"
    echo "  binary: $BINARY"
    echo "(re-run with --force or FORCE_REBUILD=1 to rebuild)"
    exit 0
fi

# --- Build ------------------------------------------------------------------

mkdir -p "$BIN_DIR" "$STAMP_DIR"

echo "Building gsuite-mcp (HEAD=$GIT_HEAD)"
if ! go build \
        -ldflags "-X main.GitCommit=$GIT_HEAD" \
        -o "$BINARY" \
        ./cmd/gsuite-mcp/; then
    echo "Error: gsuite-mcp build failed" >&2
    exit 1
fi
chmod +x "$BINARY"

# --- Record version stamp ---------------------------------------------------

printf '%s\n' "$GIT_HEAD" > "$STAMP_FILE"

# --- Summary ----------------------------------------------------------------

echo
echo "Installed gsuite-mcp at $GIT_HEAD"
echo "  binary: $BINARY"
echo "  stamp:  $STAMP_FILE"

# Hint about PATH
case ":$PATH:" in
    *":$BIN_DIR:"*)
        ;;
    *)
        echo
        echo "Note: $BIN_DIR is not on your PATH."
        echo "      Add it to your shell rc:"
        echo "        export PATH=\"$BIN_DIR:\$PATH\""
        ;;
esac
