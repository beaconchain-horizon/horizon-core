#!/usr/bin/env bash
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"

echo "=========================================="
echo "  🚀 PURE TPS TEST (NO RESTART)"
echo "=========================================="

# ۱. چک کردن اینکه سیستم زنده است یا نه
echo "[1/4] Checking if system is running..."
IS_SW_ONLINE=$(curl -s --max-time 2 $SW/api/v1/health | grep -o '"status":"online"')

if [ -z "$IS_SW_ONLINE" ]; then
    echo "  ❌ SYSTEM IS OFFLINE."
    echo "  Please make sure your Guardian / Services are running first."
    echo "  (I am not starting them for you, as requested)."
    exit 1
fi
echo "  ✅ System is ONLINE. Proceeding with test..."

# ۲. باز کردن کلید (اگر قبلا باز نشده باشد)
echo "[2/4] Unlocking Switch Key..."
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"password":"bench"}' > /dev/null

# ۳. اطمینان از وجود لایسنس و شارژ حساب
echo "[3/4] Ensuring license & account balance..."
REAL_LICENSE='{"license_id":"HORIZON-GOLD-004","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"1c70ad2e7293e37ab06728938d29db9c76c658cfc659c360805b5262316442276c80049c3d951347162909cdf64da5f37083835c1792fc5552c972b060305393","signed_data":"{\"license_id\":\"HORIZON-GOLD-004\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"292d2e3f85f281465998355924778226f4aec9f182b1a2c3ade823cc07970c18\",\"issued_at\":1792164282,\"expires_at\":1823700282,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
curl -s -X POST "$SW/api/v1/license/save" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d "$REAL_LICENSE" > /dev/null
curl -s -X POST "$SW/api/v1/account/seed" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{"bank_id":"bench","amount":999999999999}' > /dev/null
sleep 1

# ۴. گرفتن توکن و اجرای تست
echo "[4/4] Generating 100,000 transactions and sending..."
SESS=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
N=100000
TMP="/tmp/pure_tps.json"
printf '{"session_token":"%s","txs":[' "$SESS" > "$TMP"
for i in $(seq 1 $N); do
  [ $i -gt 1 ] && printf ',' >> "$TMP"
  printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
done
printf ']}' >> "$TMP"

START=$(date +%s%N)
RESP=$(curl -s -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "X-Admin-Token: $TOK" --data-binary @"$TMP")
END=$(date +%s%N)

DUR=$(( (END - START) / 1000000 ))
[ $DUR -eq 0 ] && DUR=1
TPS=$(( N * 1000 / DUR ))
rm -f "$TMP"

echo "------------------------------------------"
echo "  ⏱️  Time:     ${DUR} ms"
echo "  📦 Response: $RESP"
echo "------------------------------------------"

if echo "$RESP" | grep -q '"processed":100000'; then
  echo -e "  🚀 REAL TPS: \033[1;32m$TPS\033[0m"
else
  echo "  ❌ Failed to process batch."
fi
echo "=========================================="
