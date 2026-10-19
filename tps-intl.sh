#!/usr/bin/env bash
# ============================================================================
# HORIZON CORE - INTERNATIONAL TPS BENCHMARK SUITE
# Version: 1.0
# Standard: Hyperledger Caliper concepts / ISO-IEC 25010
# ============================================================================

BE="http://localhost:8080"
SW="http://localhost:8091"
TOKEN="12345678901234567890123456789012"
H_KEY="X-Admin-Key: $TOKEN"
H_TOK="X-Admin-Token: $TOKEN"
OUT="$REPORT_DIR/tps-intl-$(date +%Y%m%d-%H%M%S).log"
mkdir -p "$REPORT_DIR"
exec > >(tee "$OUT") 2>&1
echo "==================================================================="
echo "  HORIZON CORE - INTERNATIONAL TPS BENCHMARK SUITE v1.0"
echo "  CPU:  $(nproc 2>/dev/null || echo 8) cores"
echo "==================================================================="
echo ""

echo "[PHASE 0] Ecosystem Health Check"
echo "-------------------------------------------------------------------"
NODES="9001 9002 9003 9004 9005 9006 9007"
ROLES="merkle hash ecdsa type volume time user"
i=1
for p in $NODES; do
  resp=$(curl -s --max-time 2 "http://localhost:$p/health" | grep -o "\"role\":\"$r\"")
  if [ -n "$resp" ]; then echo "  [OK]   node $p ($r)"; nodes_ok=$((nodes_ok+1))
  else echo "  [FAIL] node $p ($r)"; fi
  i=$((i+1))
done

BE_H=$(curl -s --max-time 2 "$BE/api/v1/health" | grep -o '"status":"online"')
SW_H=$(curl -s --max-time 2 "$SW/api/v1/health" | grep -o '"status":"online"')
if [ -n "$BE_H" ]; then echo "  [OK]   backend 8080"; else echo "  [FAIL] backend 8080"; fi
if [ -n "$SW_H" ]; then echo "  [OK]   switch 8091"; else echo "  [FAIL] switch 8091"; fi

if [ "$nodes_ok" -lt 7 ] || [ -z "$BE_H" ] || [ -z "$SW_H" ]; then
  echo ""; echo "ERROR: ecosystem incomplete - abort."; exit 1
fi
echo ""

echo "[WARMUP] Priming nodes..."
HASH=$(echo -n "WARMUP" | sha256sum | cut -d' ' -f1)
WM='{"left":"W","right":"1","header":"W","message":"W","signature":"'$HASH'","merkle_type":"binary","used":0,"volume":999999999,"issued_at":1,"expires_at":9999999999,"user_id":"w","product_id":"w","expected_user":"w","expected_product":"w","mode":"batch"}'
for round in 1 2 3 4 5; do
  for p in $NODES; do
  done
done
echo "  warmup complete"
echo ""
echo "[TEST 1] 7-Node Parallel Consensus"
echo "-------------------------------------------------------------------"
for round in $(seq 1 $N); do
  for p in $NODES; do
    curl -s -o /dev/null -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$WM" &
  done
  wait
done
END=$(date +%s%N)
DUR_MS=$(( (END - START) / 1000000 ))
TOTAL=$((N * 7))
NODE_TPS=$(( TOTAL * 1000 / DUR_MS ))
echo "  Iterations:        $N"
echo "  Total requests:    $TOTAL"
echo "  Duration:          ${DUR_MS}ms"
echo "  Consensus TPS:     $NODE_TPS"
echo ""

echo "[TEST 2] Backend /license/register Latency"
echo "-------------------------------------------------------------------"
SIG="5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e"
STAMP=$(date +%s%N)
LIC='{"license_id":"TPS-'$STAMP'","user_id":"bench","product_id":"horizon-core","volume":50000000,"signature":"'$SIG'","signed_data":"{\"license_id\":\"TPS-'$STAMP'\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"'$MR'\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
E=$(date +%s%N)
REG_MS=$(( (E - S) / 1000000 ))
[ $REG_MS -eq 0 ] && REG_MS=1
REG_TPS=$(( 1000 / REG_MS ))
echo "  HTTP status:       $CODE"
echo "  Register latency:  ${REG_MS}ms"
echo "  Single-thread TPS: $REG_TPS"
echo ""
echo "[TEST 3] Switch /tx/batch Raw Ledger Throughput"
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$H_TOK" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null 2>&1

SIZES="1000 5000 10000 25000 50000 100000"
for N in $SIZES; do
  TOK_S=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "$H_TOK" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
  TMP="/tmp/tpsb.json"
  printf '{"session_token":"%s","txs":[' "$TOK_S" > "$TMP"
  while [ $i -le $N ]; do
    if [ $i -gt 1 ]; then printf ',' >> "$TMP"; fi
    i=$((i+1))
  done
  printf ']}' >> "$TMP"
  S=$(date +%s%N)
  curl -s -o /dev/null -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "$H_TOK" --data-binary @"$TMP"
  [ $DUR_MS -eq 0 ] && DUR_MS=1
  TPS=$(( N * 1000 / DUR_MS ))
  MS_LIST="$MS_LIST $DUR_MS"
  printf "  n=%-7s  %sms  %s TPS\n" "$N" "$DUR_MS" "$TPS"
  rm -f "$TMP"
done
echo ""

echo "[TEST 4] End-to-End: Register -> Consensus -> Block"
echo "-------------------------------------------------------------------"
STAMP=$(date +%s%N)
LIC='{"license_id":"E2E-'$STAMP'","user_id":"bench","product_id":"horizon-core","volume":50000000,"signature":"'$SIG'","signed_data":"{\"license_id\":\"E2E-'$STAMP'\",\"user_id\":\"bench\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"'$MR'\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
curl -s -o /dev/null -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$H_KEY" -d "$LIC"
E_REG=$(date +%s%N)
REG_MS=$(( (E_REG - S) / 1000000 ))
A_H=$B_H
for round in 1 2 3 4 5 6; do
  sleep 1
  A_H=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)
done
E_BLK=$(date +%s%N)
BLK_MS=$(( (E_BLK - S) / 1000000 ))
echo "  Register latency:    ${REG_MS}ms"
echo "  Block finality:      ${BLK_MS}ms"
echo ""

echo "==================================================================="
echo "  FINAL RESULTS - Horizon Core v1.4"
echo "==================================================================="
printf "  %-38s | %-12s | %-12s\n" "Component" "Latency" "TPS"
echo "  --------------------------------------|--------------|--------------"
printf "  %-38s | %-12s | %-12s\n" "7-node consensus (parallel)" "${NODE_CYCLE_MS}ms" "$NODE_TPS"
printf "  %-38s | %-12s | %-12s\n" "E2E register -> block" "${BLK_MS}ms" "-"
echo "  --------------------------------------|--------------|--------------"
idx=1
for N in $SIZES; do
  MS=$(echo $MS_LIST | cut -d' ' -f$idx)
  TPS=$(echo $TPS_LIST | cut -d' ' -f$idx)
  printf "    n=%-7s                            | %-12s | %-12s\n" "$N" "${MS}ms" "$TPS"
  idx=$((idx+1))
done
echo ""
echo "  International Comparison (2026):"
echo "  -------------------------------------------------------------------"
printf "  %-25s | %-15s | %-15s\n" "System" "Peak TPS" "Finality"
printf "  %-25s | %-15s | %-15s\n" "Bitcoin" "7" "60 min"
printf "  %-25s | %-15s | %-15s\n" "Solana" "65,000" "~13 sec"
printf "  %-25s | %-15s | %-15s\n" "Horizon Core" "$NODE_TPS" "${BLK_MS}ms"
echo ""
echo "  Report: $OUT"
curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$H_TOK" -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/admin/license/issue" -H "Content-Type: application/json" -H "$H_TOK" -d '{"license_id":"TPS-BENCH","user_id":"bench","volume":999999999,"duration":365}' > /dev/null 2>&1
TPS_LIST=""
MS_LIST=""
  [ -z "$TOK_S" ] && TOK_S="bench"
  printf '{"session_token":"%s","txs":[' "$TOK_S" > "$TMP"
    printf '{"from":"bench","to":"r%s","amount":1}' "$i" >> "$TMP"
  done
  printf ']}' >> "$TMP"
  S=$(date +%s%N)
  curl -s -o /dev/null -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "$H_TOK" --data-binary @"$TMP"
  DUR_MS=$(( (E - S) / 1000000 ))
  [ $DUR_MS -eq 0 ] && DUR_MS=1
  printf "  n=%-7s  %sms  %s TPS\n" "$N" "$DUR_MS" "$TPS"

echo "-------------------------------------------------------------------"
curl -s -o /dev/null -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "$H_KEY" -d "$LIC"
A_H=$B_H
for round in 1 2 3 4 5 6; do
  A_H=$(curl -s "$BE/api/v1/chain/info" | grep -o '"height":[0-9]*' | cut -d: -f2)
E_BLK=$(date +%s%N)
BLK_MS=$(( (E_BLK - S) / 1000000 ))
echo "  Height:              $B_H -> $A_H"

echo "==================================================================="
printf "  %-38s | %-12s | %-12s\n" "E2E register -> block" "${BLK_MS}ms" "-"
idx=1
for N in $SIZES; do
  MS=$(echo $MS_LIST | cut -d' ' -f$idx)
  printf "    n=%-7s                            | %-12s | %-12s\n" "$N" "${MS}ms" "$TPS"
done
echo ""
echo "  International Comparison (2026):"
echo "  -------------------------------------------------------------------"
printf "  %-25s | %-15s | %-15s\n" "System" "Peak TPS" "Finality"
echo "  --------------------------|-----------------|-----------------"
printf "  %-25s | %-15s | %-15s\n" "Bitcoin" "7" "60 min"
printf "  %-25s | %-15s | %-15s\n" "Solana" "65,000" "~13 sec"
printf "  %-25s | %-15s | %-15s\n" "Visa (VisaNet)" "24,000-65,000" "sub-second"
echo ""
echo "  Report: $OUT"
echo "==================================================================="
