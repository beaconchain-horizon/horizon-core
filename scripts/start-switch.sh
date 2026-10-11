#!/usr/bin/env bash
# Starts Switch with the public key auto-detected from any USB drive.
# Falls back to HORIZON_LICENSE_PUBLIC_KEY if already set.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${HORIZON_LICENSE_PUBLIC_KEY:-}" ]; then
  echo "[start-switch] HORIZON_LICENSE_PUBLIC_KEY not set — auto-detecting from USB..."
  DETECTED=$(go run ./cmd/horizon-usbkey 2>/dev/null || true)
  if [ -n "$DETECTED" ]; then
    export HORIZON_LICENSE_PUBLIC_KEY="$DETECTED"
    echo "[start-switch] using USB key: ${DETECTED:0:16}..."
  else
    echo "[start-switch] ERROR: no USB key found" >&2
    exit 1
  fi
fi

export SWITCH_PORT="${SWITCH_PORT:-8081}"
export SWITCH_DB="${SWITCH_DB:-horizon-switch-dev.db}"
export ADMIN_TOKEN="${ADMIN_TOKEN:-testadmin1234567890abcdef1234567890abcdef}"

echo "[start-switch] port=$SWITCH_PORT db=$SWITCH_DB"
exec go run ./cmd/switch
