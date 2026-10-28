#!/usr/bin/env bash
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
BATCHES=3
N_PER_BATCH=100000

echo "🚀 FINAL TPS TEST (300,000 TX)"

# آماده‌سازی (لایسنس، حساب و کلید)
REAL_LICENSE='{"license_id":"HORIZON-GOLD-004",...}' # همان لایسنس قبلی
curl -s -X POST "$SW/api/v1/license/save" -H "X-Admin-Token: $TOK" -d "$REAL_LICENSE" > /dev/null
curl -s -X POST "$SW/api/v1/account/seed" -H "X-Admin-Token: $TOK" -d '{"bank_id":"bench","amount":999999999999}' > /dev/null

TMP_DIR="/tmp/horizon_final_bench"
mkdir -p "$TMP_DIR"
for b in $(seq 1 $BATCHES); do
    SESS=$(curl -s -X POST "$SW/api/v1/signing/session" -H "X-Admin-Token: $TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
    awk -v sess="$SESS" -v n="$N_PER_BATCH" 'BEGIN {
        printf "{\"session_token\":\"%s\",\"txs\":[", sess;
        for (i = 1; i <= n; i++) { if (i > 1) printf ","; printf "{\"from\":\"bench\",\"to\":\"r%d\",\"amount\":1}", i; }
        printf "]}";
    }' > "$TMP_DIR/batch_$b.json"
done

START=$(date +%s%N)
for b in $(seq 1 $BATCHES); do
    curl -s -X POST "$SW/api/v1/tx/batch" -H "X-Admin-Token: $TOK" --data-binary @"$TMP_DIR/batch_$b.json" > "$TMP_DIR/resp_$b.txt" &
done
wait
END=$(date +%s%N)

DUR=$(( (END - START) / 1000000 ))
TPS=$(( (BATCHES * N_PER_BATCH) * 1000 / DUR ))
echo "⏱️ Time: ${DUR} ms | 🚀 TPS: ${TPS}"
rm -rf "$TMP_DIR"
