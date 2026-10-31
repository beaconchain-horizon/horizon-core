#!/usr/bin/env bash
set -e

echo "1) Killing old node processes..."
taskkill //F //IM angel-node.exe //T 2>/dev/null || true
taskkill //F //IM go.exe //T 2>/dev/null || true
sleep 2

echo "2) Starting 7 Nodes with ALL required variables..."

ports=(9001 9002 9003 9004 9005 9006 9007)
roles=("merkle" "hash" "ecdsa" "type" "volume" "time" "user")
PUBKEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"

for i in ${!ports[@]}; do
    port=${ports[$i]}
    role=${roles[$i]}
    
    # تنظیم تمام متغیرهای محیطی مورد نیاز
    export NODE_PORT=$port
    export NODE_ROLE=$role
    export NODE_PUBKEY=$PUBKEY
    
    echo "   Starting Node on port $port (Role: $role)..."
    nohup ./angel-node.exe > node_$port.log 2>&1 &
    sleep 1
done

echo "3) Waiting 6 seconds for nodes to sync..."
sleep 6

echo "4) Strict Consensus Check..."
LAWS=""
ALIVE_COUNT=0

for p in 9001 9002 9003 9004 9005 9006 9007; do
    HASH=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$HASH" ]; then
        echo "  ✅ Node :$p -> Law Hash: ${HASH:0:16}..."
        LAWS="$LAWS$HASH\n"
        ALIVE_COUNT=$((ALIVE_COUNT+1))
    else
        echo "  ❌ Node :$p -> OFFLINE"
    fi
done

echo ""
# اصلاح منطق: اول چک می‌کنیم آیا همه ۷ نود آنلاین هستند
if [ "$ALIVE_COUNT" -eq 7 ]; then
    UNIQ=$(echo -e "$LAWS" | sort -u | grep -c .)
    if [ "$UNIQ" -eq 1 ]; then
        echo "🏆 PROOF: All 7 nodes are ONLINE and share IDENTICAL law_hash."
    else
        echo "⚠️ WARNING: All nodes are online, but their hashes DO NOT MATCH!"
    fi
else
    echo "🚨 FAIL-CLOSED: Only $ALIVE_COUNT/7 nodes are online. System will NOT build blocks."
    # چاپ لاگ برای نودی که آفلاین است
    echo "--- Checking logs for offline nodes ---"
    for p in 9001 9002 9003 9004 9005 9006 9007; do
        HASH=$(curl -s --max-time 1 "http://localhost:$p/health")
        if [ -z "$HASH" ]; then
            echo "Log for node :$p -> $(tail -n 1 node_$p.log)"
        fi
    done
fi
