#!/usr/bin/env bash
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
N_PER_BATCH=100000
BATCHES=5

echo "=========================================="
echo "  🚀 PARALLEL TPS TEST (500,000 TX)"
echo "=========================================="

echo "[1/3] Preparing licenses & accounts..."
REAL_LICENSE='{"license_id":"HORIZON-GOLD-004","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"1c70ad2e7293e37ab06728938d29db9c76c658cfc659c360805b5262316442276c80049c3d951347162909cdf64da5f37083835c1792fc5552c972b060305393","signed_data":"{\"license_id\":\"HORIZON-GOLD-004\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"292d2e3f85f281465998355924778226f4aec9f182b1a2c3ade823cc07970c18\",\"issued_at\":1792164282,\"expires_at\":1823700282,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
curl -s -X POST "$SW/api/v1/license/save" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d "$REAL_LICENSE" > /dev/null
curl -s -X POST "$SW/api/v1/account/seed" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"bank_id":"bench","amount":999999999999}' > /dev/null

echo "[2/3] Generating $BATCHES JSON files (100K each)..."
TMP_DIR="/tmp/horizon_bench"
mkdir -p "$TMP_DIR"
rm -f "$TMP_DIR"/*.json

for b in $(seq 1 $BATCHES); do
    SESS=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
    awk -v sess="$SESS" -v n="$N_PER_BATCH" 'BEGIN {
        printf "{\"session_token\":\"%s\",\"txs\":[", sess;
        for (i = 1; i <= n; i++) {
            if (i > 1) printf ",";
            printf "{\"from\":\"bench\",\"to\":\"r%d\",\"amount\":1}", i;
        }
        printf "]}";
    }' > "$TMP_DIR/batch_$b.json"
done

echo "[3/3] Sending $BATCHES requests in parallel..."
START=$(date +%s%N)

for b in $(seq 1 $BATCHES); do
    curl -s -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" --data-binary @"$TMP_DIR/batch_$b.json" > "$TMP_DIR/resp_$b.txt" &
done
wait # منتظر ماندن برای اتمام همه درخواست‌ها

END=$(date +%s%N)
DUR=$(( (END - START) / 1000000 ))
[ $DUR -eq 0 ] && DUR=1
TOTAL=$(( BATCHES * N_PER_BATCH ))
TPS=$(( TOTAL * 1000 / DUR ))

echo "------------------------------------------"
echo "  ⏱️  Time:     ${DUR} ms"
echo "  📦 Processed: $TOTAL transactions"
echo "  🚀 REAL TPS:  $TPS"
echo "------------------------------------------"
echo "Sample Response: $(cat $TMP_DIR/resp_1.txt)"
rm -rf "$TMP_DIR"
echo "=========================================="
