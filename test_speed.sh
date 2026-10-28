#!/usr/bin/env bash
# ==========================================
# Horizon Core - FINAL TPS BENCHMARK (100K)
# ==========================================
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"

echo "=========================================="
echo "  🚀 HORIZON CORE SPEED TEST"
echo "=========================================="

# ۱. ریست کردن دیتابیس سوییچ
echo "1. Resetting Switch DB for a clean test..."
pid=$(netstat -ano | grep ":8091 " | grep LISTENING | awk '{print $5}' | head -1)
[ -n "$pid" ] && taskkill //PID $pid //F > /dev/null 2>&1
sleep 1
rm -f data/horizon-switch.db data/horizon-switch.db-shm data/horizon-switch.db-wal

# ۲. اجرای سوییچ
echo "2. Starting Switch on :8091..."
export ADMIN_TOKEN=$TOK
export SWITCH_PORT=8091
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
./switch.exe > /tmp/switch-tps.log 2>&1 &
sleep 2

# ۳. باز کردن کلید سوییچ
echo "3. Unlocking Switch Key..."
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"password":"bench"}' > /dev/null

# ۴. ثبت لایسنس واقعی USB (GOLD-004)
echo "4. Saving REAL license (signed by USB)..."
if [ -f "/tmp/gold4.json" ]; then
    curl -s -X POST "$SW/api/v1/license/save" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" --data-binary @/tmp/gold4.json > /dev/null
else
    echo "❌ gold4.json not found in /tmp/"
    exit 1
fi
sleep 1 # زمان برای اینکه سیستم لایسنس را تأیید کند
# ۴.۵. شارژ حساب تستی
echo "4.5 Seeding test account..."
curl -s -X POST "$SW/api/v1/account/seed" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"bank_id":"bench","amount":999999999999}' > /dev/null


# ۵. گرفتن توکن معتبر
echo "5. Getting valid session token..."
SESSION_TOKEN=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$SESSION_TOKEN" ]; then
    echo "❌ Failed to get session token."
    exit 1
fi

# ۶. ساخت فایل ۱۰۰,۰۰۰ تراکنشی
echo "6. Generating 100,000 transactions (please wait)..."
N=100000
TMP="/tmp/tps_100k.json"
printf '{"session_token":"%s","txs":[' "$SESSION_TOKEN" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

# ۷. ارسال و اندازه‌گیری زمان
echo "7. Sending batch to Switch..."
START=$(date +%s%N)
RESP=$(curl -s -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" --data-binary @"$TMP" -w "\nHTTP_STATUS:%{http_code}")
END=$(date +%s%N)

# ۸. محاسبه TPS
DURATION_MS=$(( (END - START) / 1000000 ))
[ $DURATION_MS -eq 0 ] && DURATION_MS=1
TPS=$(( N * 1000 / DURATION_MS ))
STATUS=$(echo "$RESP" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESP" | grep -v "HTTP_STATUS")

echo ""
echo "------------------------------------------"
echo "  ⏱️  Time Elapsed: ${DURATION_MS} ms"
echo "  📡 HTTP Status:  $STATUS"
echo "  📦 Response:     $BODY"
echo "------------------------------------------"

if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
  echo -e "  🚀 REAL TPS: \033[1;32m$TPS\033[0m"
  echo "  (VisaNet Peak: ~65,000 TPS)"
else
  echo -e "  ❌ Failed to process batch. Check logs at /tmp/switch-tps.log"
fi

rm -f "$TMP"
echo "=========================================="
