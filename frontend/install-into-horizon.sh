#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET="$ROOT/frontend-next"

if [ -d "$TARGET" ]; then
  echo "[ERROR] $TARGET already exists. Remove/rename it first if you want a clean install."
  exit 1
fi

mkdir -p "$TARGET"
cp -R "$SRC_DIR"/. "$TARGET"/
rm -rf "$TARGET/node_modules" "$TARGET/.next"

echo "=============================================="
echo " Horizon Store Next.js installed"
echo "=============================================="
echo "Target: $TARGET"
echo
cd "$TARGET"

if [ ! -f .env.local ]; then
  cp .env.example .env.local
fi

echo "Next: 16.3.6"
echo "Run:"
echo "  cd frontend-next"
echo "  npm install"
echo "  npm run dev"
echo
printf '%s\n' "No Git commit/push was performed."
