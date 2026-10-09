#!/usr/bin/env bash
# Horizon Core installer (Linux/macOS; Docker Compose required)
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${HORIZON_HOME:-/opt/horizon}"
DATA_DIR="${INSTALL_DIR}/data"
LOG_DIR="${INSTALL_DIR}/logs"

echo "============================================"
echo " Horizon Core installer"
echo " Source: $SCRIPT_DIR"
echo " Target: $INSTALL_DIR"
echo " Go toolchain: 1.24 (container build)"
echo "============================================"

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: Docker is required." >&2
  exit 1
fi
if ! docker compose version >/dev/null 2>&1 && ! command -v docker-compose >/dev/null 2>&1; then
  echo "ERROR: Docker Compose is required." >&2
  exit 1
fi
if ! command -v openssl >/dev/null 2>&1; then
  echo "ERROR: openssl is required to generate secrets." >&2
  exit 1
fi
if ! git -C "$SCRIPT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "ERROR: Run this installer from a Git checkout." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR" "$DATA_DIR" "$LOG_DIR"

# Copy only tracked files. Local databases, private keys, .env files and scratch
# directories are not copied into the installation.
git -C "$SCRIPT_DIR" archive --format=tar HEAD | tar -xf - -C "$INSTALL_DIR"

if [ ! -f "$INSTALL_DIR/.env" ]; then
  cat > "$INSTALL_DIR/.env" <<EOF
POSTGRES_PASSWORD=$(openssl rand -hex 32)
ADMIN_TOKEN=$(openssl rand -hex 32)
API_KEY=$(openssl rand -hex 32)
EOF
  chmod 600 "$INSTALL_DIR/.env"
  echo "Created $INSTALL_DIR/.env with random secrets."
else
  echo "Existing $INSTALL_DIR/.env preserved."
fi

cd "$INSTALL_DIR"
if command -v docker-compose >/dev/null 2>&1; then
  docker-compose build
  docker-compose up -d
else
  docker compose build
  docker compose up -d
fi

echo "Services started. Verify /health and /api/v1/health before use."
