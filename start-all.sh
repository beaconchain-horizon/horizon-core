#!/usr/bin/env bash
cd "$(dirname "$0")"

export ADMIN_TOKEN="12345678901234567890123456789012"
export SWITCH_URL="http://localhost:8091"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export NODE_PUBKEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export HORIZON_DB_PATH="./horizon-dev.db"

echo "=== Horizon Angel - Full Stack Startup ==="
echo ""

echo "[1/4] Cleaning old processes..."
for p in 8080 8091 9001 9002 9003 9004 9005 9006 9007; do
  pid=$(netstat -ano 2>/dev/null | grep ":$p " | grep LISTENING | awk '{print $5}' | head -1)
  if [ -n "$pid" ]; then
    taskkill //PID $pid //F 2>/dev/null
  fi
done
sleep 1

echo "[2/4] Building binaries..."
go build -o switch.exe ./cmd/switch/ 2>&1 | head -3
go build -o angel-node.exe ./cmd/angel-node/ 2>&1 | head -3
go build -o angel-receive.exe ./cmd/angel-receive/ 2>&1 | head -3
echo "     done."

echo "[3/4] Starting 7 engine nodes..."
NODE_ROLE=merkle NODE_PORT=9001 ./angel-node.exe > /tmp/n1.log 2>&1 &
NODE_ROLE=hash   NODE_PORT=9002 ./angel-node.exe > /tmp/n2.log 2>&1 &
NODE_ROLE=ecdsa  NODE_PORT=9003 ./angel-node.exe > /tmp/n3.log 2>&1 &
NODE_ROLE=type   NODE_PORT=9004 ./angel-node.exe > /tmp/n4.log 2>&1 &
NODE_ROLE=volume NODE_PORT=9005 ./angel-node.exe > /tmp/n5.log 2>&1 &
NODE_ROLE=time   NODE_PORT=9006 ./angel-node.exe > /tmp/n6.log 2>&1 &
NODE_ROLE=user   NODE_PORT=9007 ./angel-node.exe > /tmp/n7.log 2>&1 &
sleep 2

echo "[4/4] Starting Switch on :8091..."
SWITCH_PORT=8091 ./switch.exe > /tmp/switch.log 2>&1 &
sleep 3

echo ""
echo "=== Health Check ==="
echo "Engine nodes:"
for p in 9001 9002 9003 9004 9005 9006 9007; do
  r=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"role":"[^"]*"' | cut -d'"' -f4)
  if [ -z "$r" ]; then echo "  :$p DOWN"; else echo "  :$p $r"; fi
done

echo ""
echo "Switch:"
curl -s --max-time 2 "http://localhost:8091/api/v1/health"
echo ""

echo ""
echo "=== Done ==="
echo "Logs: /tmp/n1.log .. /tmp/n7.log and /tmp/switch.log"
