#!/usr/bin/env bash
cd ~/Beaconchain/horizon-core
export NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
export ADMIN_TOKEN=12345678901234567890123456789012
export SWITCH_URL=http://localhost:8091
export HORIZON_DB_PATH=./horizon-dev.db
export HORIZON_LICENSE_PUBLIC_KEY=$NODE_PUBKEY
export GIN_MODE=release
export SWITCH_PORT=8091
NODE_ROLE=merkle NODE_PORT=9001 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=hash NODE_PORT=9002 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=ecdsa NODE_PORT=9003 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=type NODE_PORT=9004 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=volume NODE_PORT=9005 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=time NODE_PORT=9006 ./angel-node.exe >/dev/null 2>&1 &
NODE_ROLE=user NODE_PORT=9007 ./angel-node.exe >/dev/null 2>&1 &
sleep 3
./switch.exe >/dev/null 2>&1 &
sleep 2
./backend.exe >/dev/null 2>&1 &
sleep 4
echo "=== HEALTH ==="
for p in 8080 8091 9001 9002 9003 9004 9005 9006 9007; do
  r=$(curl -s --max-time 2 "http://localhost:$p/health" 2>/dev/null | head -c 25)
  echo "  :$p -> $r"
done
