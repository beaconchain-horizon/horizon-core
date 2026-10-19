#!/usr/bin/env bash
set -Eeuo pipefail

TUNNEL_NAME="horizon-backend"
TUNNEL_ID="9a1cf0a8-02de-40c2-aa41-4c5c8d2c6c0b"
LOCAL_URL="http://127.0.0.1:8090"
HEALTH_PATH="/api/v1/health"

echo "===== HORIZON CLOUDFLARE PATCH ====="
echo "Tunnel: $TUNNEL_NAME"
echo "Tunnel ID: $TUNNEL_ID"
echo "Local target: $LOCAL_URL"
echo

if ! command -v cloudflared >/dev/null 2>&1; then
  echo "ERROR: cloudflared is not in PATH."
  echo "Install the Windows MSI, then reopen Git Bash."
  exit 1
fi

echo "===== CLOUDFLARED VERSION ====="
cloudflared --version

echo
echo "===== LOCAL HEALTH ====="
if ! curl -fsS --max-time 5   "$LOCAL_URL$HEALTH_PATH"; then
  echo
  echo "ERROR: Local health check failed."
  echo "No tunnel was started."
  exit 1
fi
echo

echo
echo "===== TUNNEL CREDENTIALS ====="
echo "Checking whether cloudflared can access the configured tunnel..."
if ! cloudflared tunnel info "$TUNNEL_ID"; then
  echo "ERROR: Tunnel info failed."
  echo "Check your existing Cloudflare login and tunnel credentials."
  echo "No credentials or tokens were modified."
  exit 1
fi

echo
echo "===== STARTING TUNNEL ====="
echo "Press Ctrl+C to stop only this foreground cloudflared process."
echo "This command does not change DNS or create a public hostname."
exec cloudflared tunnel run "$TUNNEL_NAME"
