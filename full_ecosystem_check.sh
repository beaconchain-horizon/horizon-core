#!/usr/bin/env bash
set +e # ادامه کار حتی در صورت خطای یکی از سرویس‌ها

echo "============================================================"
echo "  HORIZON CORE - FULL ECOSYSTEM DIAGNOSTIC"
echo "============================================================"

# 1. Kill old processes safely
echo -e "\n[1/5] Stopping old instances (Safely)..."
taskkill //F //IM horizon-api.exe //T > /dev/null 2>&1
taskkill //F //IM horizon-switch.exe //T > /dev/null 2>&1
taskkill //F //IM angel-node.exe //T > /dev/null 2>&1
taskkill //F //IM go.exe //T > /dev/null 2>&1
sleep 2

# 2. Environment variables
export ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export HORIZON_DB_PATH="$HOME/Beaconchain/horizon-core/horizon-dev.db"
export SERVER_PORT=8080
export SWITCH_PORT=8091

# 3. Start Services
echo -e "\n[2/5] Starting Backend & Switch (using DB: $HORIZON_DB_PATH)..."
nohup go run ./cmd/api > api.log 2>&1 &
nohup go run ./cmd/switch > switch.log 2>&1 &

# 4. Start 7 Nodes
echo "[3/5] Starting 7 Horizon Angel Nodes..."
for port in 9001 9002 9003 9004 9005 9006 9007; do
    nohup go run ./cmd/angel-node -port $port > node_$port.log 2>&1 &
done

echo -e "\nWaiting 8 seconds for all services to boot..."
sleep 8

# 5. Diagnostics
echo -e "\n============================================================"
echo "  LIVE SYSTEM REPORT"
echo "============================================================"

echo -e "\n--- [A] Backend & Database Status (Port 8080) ---"
API_HEALTH=$(curl -s http://localhost:8080/api/v1/health)
CHAIN_INFO=$(curl -s http://localhost:8080/api/v1/chain/info)
echo "Health: $API_HEALTH"
echo "Chain Info: $CHAIN_INFO"
echo "DB Path Configured: $HORIZON_DB_PATH"

echo -e "\n--- [B] Switch Status (Port 8091) ---"
SW_HEALTH=$(curl -s http://localhost:8091/api/v1/health)
echo "Health: $SW_HEALTH"

echo -e "\n--- [C] 7-Node Consensus (Ports 9001-9007) ---"
LAWS=""
for p in 9001 9002 9003 9004 9005 9006 9007; do
    HASH=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$HASH" ]; then
        echo "  Node :$p -> Law Hash: ${HASH:0:16}..."
        LAWS="$LAWS$HASH\n"
    else
        echo "  Node :$p -> OFFLINE"
    fi
done
UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
if [ "$UNIQ" -eq 1 ]; then
    echo "  🏆 CONSENSUS: 7/7 nodes share IDENTICAL law_hash."
else
    echo "  ⚠️ WARNING: Nodes are NOT in sync (or some are offline)."
fi

echo -e "\n--- [D] API Check for Dashboards ---"
echo "Fetching Block 5 (The latest snapshot for your UI):"
curl -s http://localhost:8080/api/v1/blocks/5 | grep -o '"block_hash":"[^"]*"' || echo "Block 5 not found."
echo "Checking transactions:"
curl -s http://localhost:8080/api/v1/blocks/5 | grep -o '"tx_count":[0-9]*' || true

echo -e "\n============================================================"
echo "  DIAGNOSTIC COMPLETE."
echo "============================================================"
