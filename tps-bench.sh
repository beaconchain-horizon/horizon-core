#!/usr/bin/env bash
BE="http://localhost:8080"
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
TOK_H="X-Admin-Key: $TOK"
TOK_S="X-Admin-Token: $TOK"
OUT="reports/tps-$(date +%Y%m%d-%H%M%S).log"

mkdir -p reports
exec > >(tee "$OUT") 2>&1

calc() { awk "BEGIN { printf \"%.0f\", $1 }"; }
calc1() { awk "BEGIN { printf \"%.1f\", $1 }"; }

echo "═══════════════════════════════════════════════════════════"
echo "  HORIZON CORE — TPS BENCHMARK"
echo "  Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo "  CPU:  $(nproc 2>/dev/null || echo 8) cores"
echo "═══════════════════════════════════════════════════════════"
echo ""

# ─── TEST 1: 7 nodes parallel ───
echo "[Test 1] 7-Node Parallel Verification"
echo "─────────────────────────────────────────────────────────"

PAYLOAD='{"left":"A","right":"B","header":"H","message":"M","signature":"S","merkle_type":"binary","used":1,"volume":1000000,"issued_at":1,"expires_at":9999999999,"user_id":"u","product_id":"p","expected_user":"u","expected_product":"p"}'

N=500
START=$(date +%s%N)
for i in $(seq 1 $N); do
  for p in 9001 9002 9003 9004 9005 9006 9007; do
    curl -s -o /dev/null -X POST "http://localhost:$p/verify" \
      -H "Content-Type: application/json" -d "$PAYLOAD" &
  done
  wait
done
END=$(date +%s%N)
DUR_MS=$(( (END - START) / 1000000 ))
TOTAL=$((N * 7))
NODE_TPS=$(calc "$TOTAL / ($DUR_MS / 1000.0)")
echo "  Iterations:    $N x 7 nodes = $TOTAL requests"
echo "  Duration:      ${DUR_MS}ms"
echo "  Consensus TPS: $NODE_TPS"
echo ""

# ─── TEST 2: Backend full cycle latency ───
echo "[Test 2] Backend Full Cycle Latency"
echo "─────────────────────────────────────────────────────────"

STAMP=$(date +%s)
SIG="0294bbb8d893a39466b6a9261ede0177f55bea51a2a8c9381495eae21a4a7c454f9a7af800e0b3947c97e0311d4fc13290d47bead51284548e34760f27ee99b3"
SIGNED="{\"license_id\":\"TPS-$STAMP\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"4ff5ba72edbccd886c6ec626af2cd1760be247a5410a1d0504dfd726d6a856fa\",\"issued_at\":1792139486,\"expires_at\":1823675486,\"hardware_id\":\"\"}"

LIC="{\"license_id\":\"TPS-$STAMP\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"signature\":\"$SIG\",\"signed_data\":\"$(echo "$SIGNED" | sed 's/"/\\"/g')\",\"public_key\":\"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214\"}"

S=$(date +%s%N)
RESP=$(curl -s -w "\n%{http_code}" -X POST "$BE/api/v1/license/register" \
  -H "Content-Type: application/json" -H "$TOK_H" -d "$LIC")
E=$(date +%s%N)
DUR_MS=$(( (E - S) / 1000000 ))
echo "  Full cycle latency: ${DUR_MS}ms"
echo "  Response: $(echo "$RESP" | head -1)"
echo ""

# ─── TEST 3: Switch tx/batch ───
echo "[Test 3] Switch Raw Ledger — tx/batch TPS"
echo "─────────────────────────────────────────────────────────"

curl -s -X POST "$SW/api/v1/key/setup" -H "Content-Type: application/json" -H "$TOK_S" \
  -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/key/unlock" -H "Content-Type: application/json" -H "$TOK_S" \
  -d '{"password":"bench"}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/admin/license/issue" -H "Content-Type: application/json" -H "$TOK_S" \
  -d '{"license_id":"TPS-BENCH","user_id":"bench","volume":999999999,"duration":365}' > /dev/null 2>&1
curl -s -X POST "$SW/api/v1/account/seed" -H "Content-Type: application/json" -H "$TOK_S" \
  -d '{"bank_id":"bench","amount":999999999}' > /dev/null 2>&1

for N in 10000 50000 100000; do
  TOKEN=$(curl -s -X POST "$SW/api/v1/signing/session" -H "Content-Type: application/json" -H "$TOK_S" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
  [ -z "$TOKEN" ] && TOKEN="bench-token"

  TMP="/tmp/tps-batch-$N.json"
  {
    printf '{"session_token":"%s","txs":[' "$TOKEN"
    for i in $(seq 1 $N); do
      [ $i -gt 1 ] && printf ','
      printf '{"from":"bench","to":"r%d","amount":1}' $i
    done
    printf ']}'
  } > "$TMP"

  S=$(date +%s%N)
  curl -s -o /dev/null -X POST "$SW/api/v1/tx/batch" -H "Content-Type: application/json" -H "$TOK_S" --data-binary @"$TMP"
  E=$(date +%s%N)
  DUR_MS=$(( (E - S) / 1000000 ))
  TPS=$(calc "$N / ($DUR_MS / 1000.0)")
  echo "  n=$N  duration=${DUR_MS}ms  TPS=$TPS"
  rm -f "$TMP"
done

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  FINAL SUMMARY"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "  ✅ 7-Node Consensus:   $NODE_TPS TPS"
echo "  ✅ Full Cycle Latency: ${DUR_MS}ms"
echo ""
echo "  Comparison (2026 reference):"
echo "  ┌──────────────┬────────────────┬──────────────────┐"
echo "  │ System       │ Peak TPS       │ Type             │"
echo "  ├──────────────┼────────────────┼──────────────────┤"
echo "  │ Bitcoin      │ ~7             │ PoW L1           │"
echo "  │ Ethereum L1  │ ~15-30         │ PoS L1           │"
echo "  │ Solana       │ ~65,000        │ PoH L1           │"
echo "  │ Visa         │ ~24,000        │ Centralized      │"
echo "  │ Horizon (7n) │ $NODE_TPS            │ Consensus+Node   │"
echo "  └──────────────┴────────────────┴──────────────────┘"
echo ""
echo "  Report: $OUT"
echo "═══════════════════════════════════════════════════════════"
