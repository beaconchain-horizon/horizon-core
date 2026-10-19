#!/usr/bin/env bash
BE="http://localhost:8080"
HK="X-Admin-Key: $TOK"
HS="X-Admin-Token: $TOK"
exec > >(tee "$OUT") 2>&1
echo "==================================================="
echo "  Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo "==================================================="
echo ""

# Setup Switch
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$HS" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"b"}' >/dev/null 2>&1

echo "[TEST 1] Switch /tx/batch - Raw Ledger"
for N in 1000 10000 50000 100000; do
  TOK_S=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "$HS" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
  TMP="/tmp/tb.json"
  printf '{"session_token":"%s","txs":[' "$TOK_S" > "$TMP"
  i=1
  while [ $i -le $N ]; do
    [ $i -gt 1 ] && printf ',' >> "$TMP"
    printf '{"from":"b","to":"r%s","amount":1}' "$i" >> "$TMP"
    i=$((i+1))
  done
  S=$(date +%s%N)
  curl -s -o /dev/null -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "$HS" --data-binary @"$TMP"
  E=$(date +%s%N)
  MS=$(( (E - S) / 1000000 ))
  [ $MS -eq 0 ] && MS=1
  TPS=$(( N * 1000 / MS ))
  printf "  n=%-8s %sms    %s TPS\n" "$N" "$MS" "$TPS"
  rm -f "$TMP"
done
echo ""

echo "[TEST 2] Backend /license/register"
SIG="5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e"
ST=$(date +%s%N)
LIC='{"license_id":"TPSF-'$ST'","user_id":"bench","product_id":"horizon-core","volume":50000000,"signature":"'$SIG'","signed_data":"{\"license_id\":\"TPSF-'$ST'\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"'$MR'\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
E=$(date +%s%N)
MS=$(( (E - ST) / 1000000 ))
[ $MS -eq 0 ] && MS=1
echo "  HTTP: $CODE   Latency: ${MS}ms   TPS: $((1000 / MS))"
echo ""

echo "[TEST 3] E2E: register -> block"
BEFORE=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)
ST=$(date +%s%N)
STAMP=$(date +%s%N)
LIC='{"license_id":"E2E-'$STAMP'","user_id":"bench","product_id":"horizon-core","volume":50000000,"signature":"'$SIG'","signed_data":"{\"license_id\":\"E2E-'$STAMP'\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"'$MR'\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
REG_MS=$(( (REG_E - ST) / 1000000 ))
AFTER=$BEFORE
for r in 1 2 3 4 5; do
  sleep 1
  AFTER=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)
done
BLK_E=$(date +%s%N)
BLK_MS=$(( (BLK_E - ST) / 1000000 ))
echo "  Register: ${REG_MS}ms   Block finality: ${BLK_MS}ms   Height: $BEFORE -> $AFTER"
echo ""

echo "==================================================="
echo "  FINAL TABLE"
echo "==================================================="
echo "  Component                  | Latency   | TPS"
echo "  ---------------------------|-----------|-----------"
echo "  Backend register           | ${REG_MS}ms     | $((1000 / REG_MS))"
echo "  E2E register->block        | ${BLK_MS}ms     | -"
echo "  Report: $OUT"
echo "==================================================="
