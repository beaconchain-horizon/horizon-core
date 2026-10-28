#!/usr/bin/env bash

echo "=========================================================="
echo "  HORIZON CORE - REAL TPS BENCHMARK (WITH SEED)"
echo "=========================================================="

export ADMIN_TOKEN=12345678901234567890123456789012
BE="http://localhost:8091"

# ۱. شارژ حساب bench با موجودی اولیه
echo "Step 1: Seeding sender account (bench)..."
SEED_RESP=$(curl -s -X POST "$BE/api/v1/account/seed" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{"bank_id":"bench","amount":999999999}')
echo "  → $SEED_RESP"

# ۲. گرفتن توکن نشست
echo "Step 2: Authenticating..."
SESS_RESP=$(curl -s -X POST "$BE/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" -d '{}')
SESS_TOKEN=$(echo "$SESS_RESP" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS_TOKEN" ]; then
  echo "❌ Auth failed: $SESS_RESP"
  exit 1
fi
echo "  → Authenticated OK"

# ۳. ساخت فایل ۱۰۰,۰۰۰ تراکنشی
echo "Step 3: Generating 100,000 transactions..."
N=100000
TMP="/tmp/real_tps.json"
printf '{"session_token":"%s","txs":[' "$SESS_TOKEN" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

# ۴. اندازه‌گیری
echo "Step 4: Sending batch..."
START=$(date +%s%N)
RESP=$(curl -s -X POST "$BE/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $ADMIN_TOKEN" --data-binary @"$TMP" -w "\nHTTP_STATUS:%{http_code}")
END=$(date +%s%N)

DURATION_MS=$(( (END - START) / 1000000 ))
[ $DURATION_MS -eq 0 ] && DURATION_MS=1
TPS=$(( N * 1000 / DURATION_MS ))
STATUS=$(echo "$RESP" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
BODY=$(echo "$RESP" | grep -v "HTTP_STATUS")

echo ""
echo "=========================================================="
echo "  Transactions: $N"
echo "  Time:         ${DURATION_MS} ms"
echo "  HTTP Status:  $STATUS"
echo "  Response:     $BODY"
echo "=========================================================="

if echo "$BODY" | grep -q '"processed":100000'; then
  echo "  🚀 REAL TPS: $TPS (ALL $N PROCESSED)"
  echo ""
  echo "  COMPARISON:"
  echo "  - Bitcoin:    ~7 TPS"
  echo "  - Ethereum:   ~30 TPS"
  echo "  - Solana:     ~65,000 TPS"
  echo "  - VisaNet:    ~65,000 TPS (Peak)"
  echo "  - Horizon:    $TPS TPS   ← TOP OF THE WORLD"
else
  echo "  ⚠️  Partial success. Check response above."
fi
echo "=========================================================="

rm -f "$TMP"
