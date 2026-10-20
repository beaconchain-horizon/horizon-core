#!/usr/bin/env bash

echo "=========================================================="
echo "  HORIZON CORE - RAW LEDGER TPS BENCHMARK (BATCH MODE)"
echo "=========================================================="

# 1. ریست کردن سوییچ
pid=$(netstat -ano | grep ":8091 " | grep LISTENING | awk '{print $5}' | head -1)
[ -n "$pid" ] && taskkill //PID $pid //F >/dev/null 2>&1
sleep 1
rm -f data/horizon-switch.db data/horizon-switch.db-shm data/horizon-switch.db-wal

# 2. اجرای سوییچ
export ADMIN_TOKEN=12345678901234567890123456789012
export SWITCH_PORT=8091
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
./switch.exe > /tmp/switch-bench.log 2>&1 &
sleep 2

# 3. باز کردن کلید (Key Unlock)
curl -s -X POST "http://localhost:8091/api/v1/key/setup" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null
curl -s -X POST "http://localhost:8091/api/v1/key/unlock" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"password":"bench"}' > /dev/null

# 4. صدور یک لایسنس معتبر داخلی برای پاس شدن چک امنیتی
echo "Issuing local benchmark license..."
curl -s -X POST "http://localhost:8091/api/v1/admin/license/issue" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"license_id":"BENCH-001","user_id":"bench","volume":999999999,"duration":365}' > /dev/null

# 5. ساخت فایل ۱۰۰,۰۰۰ تراکنشی
echo "Generating 100,000 transactions..."
N=100000
TMP="/tmp/tps_batch.json"
printf '{"session_token":"bench","txs":[' > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

# 6. اندازه‌گیری زمان دقیق
echo "Sending batch to Switch (:8091)..."
START=$(date +%s%N)
RESP=$(curl -s -X POST "http://localhost:8091/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" --data-binary @"$TMP" -w "\nHTTP_STATUS:%{http_code}")
END=$(date +%s%N)

# 7. محاسبه TPS
DURATION_MS=$(( (END - START) / 1000000 ))
[ $DURATION_MS -eq 0 ] && DURATION_MS=1
TPS=$(( N * 1000 / DURATION_MS ))
STATUS=$(echo "$RESP" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESP" | grep -v "HTTP_STATUS")

echo "----------------------------------------------------------"
echo "  Transactions: $N"
echo "  Time Elapsed: ${DURATION_MS} ms"
echo "  HTTP Status:  $STATUS"
echo "  Response:     $BODY"
echo "----------------------------------------------------------"
if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
  echo "  🚀 REAL TPS: $TPS"
  echo ""
  echo "  COMPARISON:"
  echo "  - Bitcoin:  ~7 TPS"
  echo "  - Solana:   ~65,000 TPS"
  echo "  - VisaNet:  ~65,000 TPS (Peak)"
  echo "  - Horizon:  $TPS TPS"
else
  echo "  ❌ Failed to process batch."
fi
echo "=========================================================="

rm -f "$TMP"
