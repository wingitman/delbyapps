#!/usr/bin/env sh
set -eu

INSTALL_DIR="${INSTALL_DIR:-${1:-$HOME/.local/bin}}"
REPO_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
WORK_DIR="$(cd "$REPO_DIR/.." && pwd)"

cd "$REPO_DIR"
go build -ldflags "-X 'main.defaultWorkDir=$WORK_DIR'" -o delbyapps .
mkdir -p "$INSTALL_DIR"
cp delbyapps "$INSTALL_DIR/delbyapps"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) printf '%s\n' "Add $INSTALL_DIR to your PATH if delbyapps is not found." ;;
esac
