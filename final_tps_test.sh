#!/usr/bin/env bash

echo "=========================================================="
echo "  HORIZON CORE - FINAL TPS BENCHMARK"
echo "=========================================================="

export ADMIN_TOKEN=12345678901234567890123456789012

# ۱. گرفتن توکن نشست (Session Token) واقعی از سوییچ
echo "Authenticating..."
SESS_RESP=$(curl -s -X POST "http://localhost:8091/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{}')
SESS_TOKEN=$(echo "$SESS_RESP" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS_TOKEN" ]; then
  echo "❌ Failed to get session token. Response: $SESS_RESP"
  exit 1
fi
echo "✅ Authenticated successfully."

# ۲. ساخت فایل ۱۰۰,۰۰۰ تراکنشی با توکن واقعی
echo "Generating 100,000 transactions..."
N=100000
TMP="/tmp/tps_batch.json"
printf '{"session_token":"%s","txs":[' "$SESS_TOKEN" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

# ۳. اندازه‌گیری زمان دقیق
echo "Sending batch to Switch (:8091)..."
START=$(date +%s%N)
RESP=$(curl -s -X POST "http://localhost:8091/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" --data-binary @"$TMP" -w "\nHTTP_STATUS:%{http_code}")
END=$(date +%s%N)

# ۴. محاسبه TPS
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
