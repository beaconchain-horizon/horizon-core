#!/usr/bin/env bash
# موتور: ۷ نود، ۷ پورت، ۷ معادله
set -e
GO=go

export NODE_PUBKEY="${HORIZON_LICENSE_PUBLIC_KEY:-}"

start_node() {
  local role=$1 port=$2
  NODE_ROLE=$role NODE_PORT=$port $GO run ./cmd/angel-node &
  echo "  → $role on :$port (pid=$!)"
}

echo "═══ Starting Horizon Angel Engine — 7 nodes ═══"
start_node merkle 9001
start_node hash   9002
start_node ecdsa  9003
start_node type   9004
start_node volume 9005
start_node time   9006
start_node user   9007
echo "═══ All 7 nodes launched ═══"
wait
