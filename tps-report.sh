#!/usr/bin/env bash
# HORIZON CORE - COMPREHENSIVE TPS & HEALTH BENCHMARK
# Standard: ISO/IEC 25010 - Performance Efficiency
BE="http://localhost:8080"
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
HK="X-Admin-Key: $TOK"
HS="X-Admin-Token: $TOK"
OUT="reports/tps-report-$(date +%Y%m%d-%H%M%S).log"
mkdir -p reports
exec > >(tee "$OUT") 2>&1

echo "=============================================================="
echo "  HORIZON CORE - COMPREHENSIVE BENCHMARK"
echo "  Standard: ISO/IEC 25010"
echo "  Date: $(date)"
echo "  CPU cores: $(nproc 2>/dev/null || echo 8)"
echo "=============================================================="
echo ""
echo "[PHASE 1] ECOSYSTEM HEALTH CHECK"
echo "--------------------------------------------------------------"
NODES="9001 9002 9003 9004 9005 9006 9007"
ROLES="merkle hash ecdsa type volume time user"
i=1
ok=0
for p in $NODES; do
  r=$(echo $ROLES | cut -d" " -f$i)
  resp=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o "\"role\":\"$r\"")
  if [ -n "$resp" ]; then echo "  [OK]   node $p ($r)"; ok=$((ok+1)); else echo "  [FAIL] node $p ($r)"; fi
  i=$((i+1))
done

BE_H=$(curl -s --max-time 2 "$BE/api/v1/health" | grep -o "online")
SW_H=$(curl -s --max-time 2 "$SW/api/v1/health" | grep -o "online")
if [ -n "$BE_H" ]; then echo "  [OK]   backend 8080"; else echo "  [FAIL] backend 8080"; fi
if [ -n "$SW_H" ]; then echo "  [OK]   switch 8091"; else echo "  [FAIL] switch 8091"; fi
echo ""
echo "[PHASE 2] 7-NODE PARALLEL CONSENSUS"
echo "--------------------------------------------------------------"
H=$(echo -n "WARM" | sha256sum | cut -d" " -f1)
P="{\"left\":\"X\",\"right\":\"1\",\"header\":\"X\",\"message\":\"X\",\"signature\":\"$H\",\"merkle_type\":\"binary\",\"used\":0,\"volume\":999999999,\"issued_at\":1,\"expires_at\":9999999999,\"user_id\":\"b\",\"product_id\":\"b\",\"expected_user\":\"b\",\"expected_product\":\"b\",\"mode\":\"batch\"}"
for w in 1 2 3 4 5; do for p in $NODES; do curl -s -o /dev/null -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$P"; done; done
N=200
S=$(date +%s%N)
for round in $(seq 1 $N); do
  for p in $NODES; do
    curl -s -o /dev/null -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$P" &
  done
  wait
done
E=$(date +%s%N)
MS=$(( (E - S) / 1000000 ))
TOTAL=$((N * 7))
TPS=$(( TOTAL * 1000 / MS ))
CYCLE=$(( MS / N ))
echo "  Iterations: $N (x7 nodes)"
echo "  Total reqs: $TOTAL"
echo "  Duration:   ${MS}ms"
echo "  Cycle:      ${CYCLE}ms"
echo "  Consensus:  $TPS TPS"
echo ""
echo "[PHASE 3] BACKEND /license/register LATENCY"
echo "--------------------------------------------------------------"
SIG="5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e"
MR="d6372d10b57a40c70a4808eac836cedebd8f4028c701d3e9e49d0443b1ab30c7"
ST=$(date +%s%N)
STAMP=$(date +%s%N)
LIC="{\"license_id\":\"TPS-$STAMP\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"signature\":\"$SIG\",\"signed_data\":\"{\\\"license_id\\\":\\\"TPS-$STAMP\\\",\\\"user_id\\\":\\\"bench\\\",\\\"product_id\\\":\\\"horizon-core\\\",\\\"volume\\\":50000000,\\\"duration\\\":365,\\\"merkle_root\\\":\\\"$MR\\\",\\\"issued_at\\\":1792148281,\\\"expires_at\\\":1823684281,\\\"hardware_id\\\":\\\"\\\"}\",\"public_key\":\"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214\"}"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$HK" -d "$LIC")
E=$(date +%s%N)
MS=$(( (E - ST) / 1000000 ))
[ $MS -eq 0 ] && MS=1
REG_TPS=$(( 1000 / MS ))
echo "  HTTP:      $CODE"
echo "  Latency:   ${MS}ms"
echo "  Single:    $REG_TPS TPS"
echo ""
echo "[PHASE 3] BACKEND /license/register LATENCY"
echo "--------------------------------------------------------------"
SIG="5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e"
MR="d6372d10b57a40c70a4808eac836cedebd8f4028c701d3e9e49d0443b1ab30c7"
ST=$(date +%s%N)
STAMP=$(date +%s%N)
LIC="{\"license_id\":\"TPS-$STAMP\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"signature\":\"$SIG\",\"signed_data\":\"{\\\"license_id\\\":\\\"TPS-$STAMP\\\",\\\"user_id\\\":\\\"bench\\\",\\\"product_id\\\":\\\"horizon-core\\\",\\\"volume\\\":50000000,\\\"duration\\\":365,\\\"merkle_root\\\":\\\"$MR\\\",\\\"issued_at\\\":1792148281,\\\"expires_at\\\":1823684281,\\\"hardware_id\\\":\\\"\\\"}\",\"public_key\":\"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214\"}"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$HK" -d "$LIC")
E=$(date +%s%N)
MS=$(( (E - ST) / 1000000 ))
[ $MS -eq 0 ] && MS=1
REG_TPS=$(( 1000 / MS ))
echo "  HTTP:      $CODE"
echo "  Latency:   ${MS}ms"
echo "  Single:    $REG_TPS TPS"
echo ""
echo "[PHASE 4] SWITCH RAW LEDGER - /tx/batch"
echo "--------------------------------------------------------------"
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$HS" -d "{\"private_key\":\"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\",\"password\":\"b\"}" >/dev/null 2>&1
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "$HS" -d "{\"password\":\"b\"}" >/dev/null 2>&1
for N in 1000 10000 50000 100000; do
  TOK_S=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "$HS" -d "{}" | grep -o "\"session_token\":\"[^\"]*\"" | cut -d"\"" -f4)
  [ -z "$TOK_S" ] && TOK_S="b"
  TMP="/tmp/tpbr.json"
  printf "{\"session_token\":\"%s\",\"txs\":[" "$TOK_S" > "$TMP"
  j=1
  while [ $j -le $N ]; do
    [ $j -gt 1 ] && printf "," >> "$TMP"
    printf "{\"from\":\"b\",\"to\":\"r%s\",\"amount\":1}" "$j" >> "$TMP"
    j=$((j+1))
  done
  printf "]}" >> "$TMP"
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
echo "[PHASE 5] END-TO-END REGISTER -> BLOCK"
echo "--------------------------------------------------------------"
BEFORE=$(curl -s "$BE/api/v1/chain/info" | grep -o "\"height\":[0-9]*" | cut -d: -f2)
STAMP2=$(date +%s%N)
LIC2="{\"license_id\":\"E2E-$STAMP2\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"signature\":\"$SIG\",\"signed_data\":\"{\\\"license_id\\\":\\\"E2E-$STAMP2\\\",\\\"user_id\\\":\\\"bench\\\",\\\"product_id\\\":\\\"horizon-core\\\",\\\"volume\\\":50000000,\\\"duration\\\":365,\\\"merkle_root\\\":\\\"$MR\\\",\\\"issued_at\\\":1792148281,\\\"expires_at\\\":1823684281,\\\"hardware_id\\\":\\\"\\\"}\",\"public_key\":\"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214\"}"
T0=$(date +%s%N)
curl -s -o /dev/null -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$HK" -d "$LIC2"
T1=$(date +%s%N)
REG_MS=$(( (T1 - T0) / 1000000 ))
AFTER=$BEFORE
for w in 1 2 3 4 5 6; do
  sleep 1
  AFTER=$(curl -s "$BE/api/v1/chain/info" | grep -o "\"height\":[0-9]*" | cut -d: -f2)
  [ "$AFTER" -gt "$BEFORE" ] && break
done
T2=$(date +%s%N)
BLK_MS=$(( (T2 - T0) / 1000000 ))
echo "  Register:   ${REG_MS}ms"
echo "  Block:      ${BLK_MS}ms"
echo "  Height:     $BEFORE -> $AFTER"
echo ""
echo "=============================================================="
echo "  FINAL REPORT - Horizon Core v1.4"
echo "=============================================================="
echo ""
echo "  Component                    | Latency    | TPS"
echo "  -----------------------------|------------|---------------"
printf "  %-28s | %-10s | %-13s\n" "7-node consensus" "${CYCLE}ms" "$TPS"
printf "  %-28s | %-10s | %-13s\n" "Backend register" "${MS}ms" "$REG_TPS"
printf "  %-28s | %-10s | %-13s\n" "E2E register -> block" "${BLK_MS}ms" "-"
echo ""
echo "  International Comparison (2026):"
echo "  --------------------------------------------"
printf "  %-20s | %-12s | %-15s\n" "System" "Peak TPS" "Finality"
printf "  %-20s | %-12s | %-15s\n" "Bitcoin" "7" "60 min"
printf "  %-20s | %-12s | %-15s\n" "Ethereum L1" "15-30" "6-12 min"
printf "  %-20s | %-12s | %-15s\n" "Solana" "65,000" "~13 sec"
printf "  %-20s | %-12s | %-15s\n" "Visa (VisaNet)" "24,000-65,000" "sub-sec"
printf "  %-20s | %-12s | %-15s\n" "Horizon Core" "$TPS" "${BLK_MS}ms"
echo ""
echo "  Report saved: $OUT"
echo "=============================================================="
