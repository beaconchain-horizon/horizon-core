#!/usr/bin/env bash

echo "=========================================================="
echo "  HORIZON CORE - RAW LEDGER TPS BENCHMARK (FIXED LICENSE)"
echo "=========================================================="

# ۱. ری‌استارت سرویس‌ها
echo "[1/5] Resetting and starting services..."
pid_sw=$(netstat -ano | grep ":8091 " | grep LISTENING | awk '{print $5}' | head -1)
[ -n "$pid_sw" ] && taskkill //PID $pid_sw //F >/dev/null 2>&1
pid_be=$(netstat -ano | grep ":8080 " | grep LISTENING | awk '{print $5}' | head -1)
[ -n "$pid_be" ] && taskkill //PID $pid_be //F >/dev/null 2>&1
sleep 1
rm -f data/horizon-switch.db* test-tps.db

export ADMIN_TOKEN=12345678901234567890123456789012
export SWITCH_PORT=8091
export HORIZON_DB_PATH=./test-tps.db
export HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

./switch.exe > /tmp/switch-tps.log 2>&1 &
sleep 2
./backend.exe > /tmp/backend-tps.log 2>&1 &
sleep 3

# ۲. باز کردن کلید و ثبت لایسنس معتبر (gold4.json)
echo "[2/5] Unlocking key and loading REAL license..."
curl -s -X POST "http://localhost:8091/api/v1/key/setup" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null
curl -s -X POST "http://localhost:8091/api/v1/key/unlock" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"password":"bench"}' > /dev/null

# بارگذاری لایسنس معتبر از فایل gold4.json (کلید اصلی)
if [ -f "/tmp/gold4.json" ]; then
    curl -s -X POST "http://localhost:8091/api/v1/license/save" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" --data-binary @/tmp/gold4.json > /dev/null
    echo "License loaded from /tmp/gold4.json"
else
    echo "❌ /tmp/gold4.json not found. Please create a valid license first."
    kill %1 2>/dev/null; kill %2 2>/dev/null
    exit 1
fi

# ۳. شارژ حساب فرستنده
echo "[3/5] Seeding sender account..."
curl -s -X POST "http://localhost:8091/api/v1/account/seed" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"bank_id":"bench","amount":99999999999}' > /dev/null

# ۴. گرفتن توکن نشست
echo "[4/5] Authenticating..."
SESS_RESP=$(curl -s -X POST "http://localhost:8091/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{}')
SESS_TOKEN=$(echo "$SESS_RESP" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS_TOKEN" ]; then
    echo "❌ Authentication failed: $SESS_RESP"
    kill %1 2>/dev/null; kill %2 2>/dev/null
    exit 1
fi
echo "Authenticated."

# ۵. ساخت بسته ۱۰۰,۰۰۰ تراکنشی و اندازه‌گیری زمان
echo "[5/5] Generating 100,000 transactions and sending..."
N=100000
TMP="/tmp/tps_batch.json"
printf '{"session_token":"%s","txs":[' "$SESS_TOKEN" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

START=$(date +%s%N)
RESP=$(curl -s -X POST "http://localhost:8091/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" --data-binary @"$TMP")
END=$(date +%s%N)

DURATION_MS=$(( (END - START) / 1000000 ))
[ $DURATION_MS -eq 0 ] && DURATION_MS=1
TPS=$(( N * 1000 / DURATION_MS ))

echo ""
echo "=========================================================="
echo "  BENCHMARK RESULTS"
echo "=========================================================="
echo "  Transactions: $N"
echo "  Time Elapsed: ${DURATION_MS} ms"
echo "  Server Reply: $RESP"
echo "----------------------------------------------------------"
if echo "$RESP" | grep -q '"processed":100000'; then
    echo "  🚀 REAL TPS: $TPS"
    echo ""
    echo "  COMPARISON (Peak TPS):"
    echo "  - Bitcoin:    ~7 TPS"
    echo "  - Ethereum:   ~30 TPS"
    echo "  - VisaNet:    ~65,000 TPS"
    echo "  - Solana:     ~65,000 TPS"
    echo "  - Horizon:    $TPS TPS  🔥"
else
    echo "  ⚠️  Batch processed partially. Check response above."
fi
echo "=========================================================="

# پاکسازی (با try-catch برای جلوگیری از خطای فایل قفل شده)
kill %1 2>/dev/null
kill %2 2>/dev/null
sleep 1
rm -f "$TMP" test-tps.db 2>/dev/null
echo "Done."
