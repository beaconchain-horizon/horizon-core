#!/usr/bin/env bash
# ==========================================
# Horizon Core - Raw TPS Benchmark (100K)
# ==========================================
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"

echo "=========================================="
echo "  HORIZON CORE TPS BENCHMARK"
echo "=========================================="

# ۱. ریست کردن دیتابیس
echo "1. Resetting Switch DB..."
pid=$(netstat -ano | grep ":8091 " | grep LISTENING | awk '{print $5}' | head -1)
[ -n "$pid" ] && taskkill //PID $pid //F > /dev/null 2>&1
sleep 1
rm -f data/horizon-switch.db data/horizon-switch.db-shm data/horizon-switch.db-wal

# ۲. اجرای سوییچ
echo "2. Starting Switch..."
export ADMIN_TOKEN=$TOK
export SWITCH_PORT=8091
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
./switch.exe > /tmp/switch-tps.log 2>&1 &
sleep 2

# ۳. باز کردن کلید
echo "3. Unlocking Key..."
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"password":"bench"}' > /dev/null

# ۴. ساخت لایسنس تستی با کلید خود سوییچ (دور زدن باگ امضا)
echo "4. Issuing test license directly from Switch..."
curl -s -X POST "$SW/api/v1/admin/license/issue" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"license_id":"TPS-TEST","user_id":"bench","volume":999999999,"duration":365}' > /dev/null
sleep 1 # صبر برای اینکه چکر لایسنس آن را تایید کند

# ۵. گرفتن توکن واقعی
echo "5. Getting session token..."
SESSION_TOKEN=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$SESSION_TOKEN" ]; then
    echo "❌ Failed to get session token."
    exit 1
fi

# ۶. ساخت فایل ۱۰۰,۰۰۰ تراکنشی با توکن واقعی
echo "6. Generating 100,000 transactions..."
N=100000
TMP="/tmp/tps_100k.json"
printf '{"session_token":"%s","txs":[' "$SESSION_TOKEN" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

# ۷. اندازه‌گیری و ارسال
echo "7. Sending to Switch..."
START=$(date +%s%N)
RESP=$(curl -s -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" --data-binary @"$TMP" -w "\nHTTP_STATUS:%{http_code}")
END=$(date +%s%N)

# ۸. محاسبه
DURATION_MS=$(( (END - START) / 1000000 ))
[ $DURATION_MS -eq 0 ] && DURATION_MS=1
TPS=$(( N * 1000 / DURATION_MS ))
STATUS=$(echo "$RESP" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESP" | grep -v "HTTP_STATUS")

echo "------------------------------------------"
echo "  Time Elapsed: ${DURATION_MS} ms"
echo "  HTTP Status:  $STATUS"
echo "  Response:     $BODY"
echo "------------------------------------------"

if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
  echo -e "  🚀 REAL TPS: \033[1;32m$TPS\033[0m"
  echo "  (Compared to Visa: ~65,000 TPS)"
else
  echo -e "  ❌ Failed to process batch. Check logs at /tmp/switch-tps.log"
fi

rm -f "$TMP"
echo "=========================================="
