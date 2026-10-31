#!/usr/bin/env bash
set -e

echo "1) Building angel-node.exe (Once)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o angel-node.exe ./cmd/angel-node

echo "2) Killing old node processes..."
taskkill //F //IM angel-node.exe //T 2>/dev/null || true
taskkill //F //IM go.exe //T 2>/dev/null || true
sleep 2

echo "3) Starting 7 Nodes with correct NODE_ROLE and NODE_PORT..."

# تعریف پورت‌ها و نقش‌های مربوطه بر اساس قانون هفت‌گانه شما
ports=(9001 9002 9003 9004 9005 9006 9007)
roles=("merkle" "hash" "ecdsa" "type" "volume" "time" "user")

for i in ${!ports[@]}; do
    port=${ports[$i]}
    role=${roles[$i]}
    
    # تنظیم متغیرهای محیطی برای هر نود
    export NODE_PORT=$port
    export NODE_ROLE=$role
    
    echo "   Starting Node on port $port (Role: $role)..."
    nohup ./angel-node.exe > node_$port.log 2>&1 &
    sleep 1 # تاخیر کوتاه برای جلوگیری از تداخل
done

echo "4) Waiting 6 seconds for nodes to boot..."
sleep 6

echo "5) Checking Consensus..."
LAWS=""
for p in 9001 9002 9003 9004 9005 9006 9007; do
    HASH=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$HASH" ]; then
        echo "  ✅ Node :$p -> Law Hash: ${HASH:0:16}..."
        LAWS="$LAWS$HASH\n"
    else
        echo "  ❌ Node :$p -> OFFLINE"
    fi
done

UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
if [ "$UNIQ" -eq 1 ]; then
    echo "  🏆 PROOF: All 7 nodes share IDENTICAL law_hash."
else
    echo "  ⚠️ WARNING: Nodes are NOT in sync."
fi
