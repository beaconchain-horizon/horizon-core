#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════
# ═══════════════════════════════════════════════════════════════════════
BE="http://localhost:8080"
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
TH="X-Admin-Key: $TOK"
TS="X-Admin-Token: $TOK"
OUT="reports/tps-full-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee "$OUT") 2>&1

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║  Date: $(date '+%Y-%m-%d %H:%M:%S')                                       ║"
echo "║  CPU:  $(nproc 2>/dev/null || echo 8) cores                                              ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"

# ─── تست ۱: سرعت ۷ نود موازی (consensus) ───
echo "════════════════════════════════════════════════════════════════════"
echo "════════════════════════════════════════════════════════════════════"

HASH=$(echo -n "BATCH-ROOT-TEST" | sha256sum | cut -d' ' -f1)
# Warmup
for p in 9001 9002 9003 9004 9005 9006 9007; do
  curl -s -o /dev/null -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$PAYLOAD"

N=200
START=$(date +%s.%N)
for ((i=0; i<N; i++)); do
  for p in 9001 9002 9003 9004 9005 9006 9007; do
  done
  wait
done
END=$(date +%s.%N)
DUR=$(echo "$END - $START" | bc)
TOTAL=$((N * 7))
NODE_TPS=$(echo "scale=0; $TOTAL / $DUR" | bc)
NODE_MS=$(echo "scale=2; $DUR * 1000 / $N" | bc)

printf "  Iterations:     %d (each iteration = 7 nodes parallel)\n" "$N"
printf "  Duration:       %ss\n" "$DUR"
printf "  Avg per cycle:  %sms\n" "$NODE_MS"
printf "  Consensus TPS:  %s\n" "$NODE_TPS"
echo ""

# ─── تست ۲: Backend license/register latency ───
echo "════════════════════════════════════════════════════════════════════"
echo " [TEST 2] Backend /license/register — End-to-End Latency"
echo "════════════════════════════════════════════════════════════════════"
STAMP=$(date +%s)
S=$(date +%s.%N)
RESP=$(curl -s -w "%{http_code}" -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$TH" -d "$LIC")
DUR=$(echo "$E - $S" | bc)
MS=$(echo "scale=1; $DUR * 1000" | bc)
REG_TPS=$(echo "scale=0; 1 / $DUR" | bc)

printf "  HTTP code:       %s\n" "${RESP: -3}"
printf "  Latency:         %sms\n" "$MS"
printf "  Theoretical:     %s TPS (single-thread)\n" "$REG_TPS"
echo ""

# ─── تست ۳: Switch tx/batch (raw ledger, multi-size) ───
echo "════════════════════════════════════════════════════════════════════"
echo "════════════════════════════════════════════════════════════════════"
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$TS" \
  -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "$TS" -d '{"password":"bench"}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/admin/license/issue" -H "Content-Type: application/json" -H "$TS" \
declare -a SIZES=(1000 5000 10000 25000 50000 100000 200000)
declare -a TPS_ARR=()
declare -a MS_ARR=()

for N in "${SIZES[@]}"; do
  TOKEN=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "$TS" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

  TMP="/tmp/tpsb.json"
  {
    printf '{"session_token":"%s","txs":[' "$TOKEN"
      [ $i -gt 1 ] && printf ','
      printf '{"from":"bench","to":"r%s","amount":1}' "$i"
    printf ']}'
  } > "$TMP"

  S=$(date +%s.%N)
  curl -s -o /dev/null -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "$TS" --data-binary @"$TMP"
  E=$(date +%s.%N)
  DUR=$(echo "$E - $S" | bc)
  TPS=$(echo "scale=0; $N / $DUR" | bc)
  MS=$(echo "scale=0; $DUR * 1000" | bc)
  TPS_ARR+=("$TPS")
  MS_ARR+=("$MS")
  rm -f "$TMP"
done
echo ""

# ─── تست ۴: E2E — از register تا بلاک ساخته‌شده ───
echo "════════════════════════════════════════════════════════════════════"
STAMP=$(date +%s)
LIC='{"license_id":"E2E-'$STAMP'","user_id":"bench","product_id":"horizon-core","volume":50000000,"signature":"5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e","signed_data":"{\"license_id\":\"E2E-'$STAMP'\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"d6372d10b57a40c70a4808eac836cedebd8f4028c701d3e9e49d0443b1ab30c7\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
BEFORE_H=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)

S=$(date +%s.%N)
curl -s -o /dev/null -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$TH" -d "$LIC"

# انتظار برای بلاک
for i in 1 2 3 4 5 6; do
  sleep 1
  AFTER_H=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)
  [ "$AFTER_H" -gt "$BEFORE_H" ] && break
done
E_BLK=$(date +%s.%N)

REG_MS=$(echo "scale=0; ($E_REG - $S) * 1000" | bc)
BLK_MS=$(echo "scale=0; ($E_BLK - $S) * 1000" | bc)

printf "  Register latency:     %sms\n" "$REG_MS"
printf "  Block finality:       %sms\n" "$BLK_MS"
echo ""

# ─── جدول نهایی ───
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║  FINAL RESULTS                                                    ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo "  ┌────────────────────────────────────┬────────────┬───────────────┐"
echo "  ├────────────────────────────────────┼────────────┼───────────────┤"
printf "  │ 7 nodes consensus (parallel)       │ %-10s │ %-13s │\n" "${NODE_MS}ms" "$NODE_TPS"
printf "  │ Backend register (per-tx)          │ %-10s │ %-13s │\n" "${MS}ms" "$REG_TPS"
printf "  │ E2E: register → block finality     │ %-10s │ %-13s │\n" "${BLK_MS}ms" "—"
echo "  ├────────────────────────────────────┼────────────┼───────────────┤"
for i in "${!SIZES[@]}"; do
  printf "  │   batch n=%-24d │ %-10s │ %-13s │\n" "${SIZES[$i]}" "${MS_ARR[$i]}ms" "${TPS_ARR[$i]}"
done
echo "  └────────────────────────────────────┴────────────┴───────────────┘"
echo ""
echo "  ┌──────────────────────┬──────────────┬──────────────────────────┐"
echo "  ├──────────────────────┼──────────────┼──────────────────────────┤"
echo "  │ Bitcoin              │ 7            │ 60 min                   │"
echo "  │ Solana               │ 65,000       │ ~13 sec                  │"
echo "  │ Visa (VisaNet peak)  │ 24,000-65,000│ sub-second               │"
echo "  │ Horizon (7-node)     │ ${NODE_TPS}     │ ${BLK_MS}ms               │"
echo ""
echo "  Report: $OUT"
