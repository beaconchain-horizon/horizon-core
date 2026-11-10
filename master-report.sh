#!/usr/bin/env bash
# master-report.sh — گزارش کار کامل شبکه Horizon (هر بلوک، هر نود)
set -u
R=$'\e[31m'; G=$'\e[32m'; Y=$'\e[33m'; C=$'\e[36m'; X=$'\e[0m'

REPORT_FILE="report-$(date +%Y%m%d-%H%M%S).txt"
log() { echo "$1" | tee -a "$REPORT_FILE"; }
sec() { log ""; log "═══ $1 ═══"; }

PORTS=(9001 9002 9003 9004 9005 9006 9007 9008)
declare -A ROLE_OF
NOW=$(date +%s); PAST=$((NOW-86400)); FUTURE=$((NOW+86400))
A64=$(printf 'a%.0s' {1..64})
Z64=$(printf '0%.0s' {1..64})
D128=$(printf 'd%.0s' {1..128})
TX1='{"from":"alice","to":"bob","amount":10,"license_id":"LIC-A","type":"transfer","payload":""}'
TX_HASH=$(printf "%s" "$TX1" | sha256sum | awk '{print $1}')

build_block() {
  local sig=$1 merkle=$2 bh=$3 mt=$4 exp=$5 used=$6 uid=$7 lic=$8 tx=$9
  cat <<JSON
{"block_number":1,"block_hash":"$bh","prev_hash":"$A64","merkle_root":"$merkle","merkle_type":"$mt","timestamp":$NOW,"tx_count":1,"signature":"$sig","signer_addr":"$A64","licenses":[{"license_id":"$lic","user_id":"$uid","volume":1000,"used":$used,"issued_at":$PAST,"expires_at":$exp,"status":"active"}],"txs":[$tx]}
JSON
}

extract_vote()   { echo "$1" | sed -n 's/.*"vote":"\([^"]*\)".*/\1/p' | head -1; }
extract_reason() { echo "$1" | sed -n 's/.*"reason":"\([^"]*\)".*/\1/p' | head -1; }

sec "STEP 1 — شناسایی نقش نودها"
for p in "${PORTS[@]}"; do
  h=$(curl -s --max-time 2 "http://127.0.0.1:$p/health" 2>/dev/null)
  role=$(echo "$h" | sed -n 's/.*"role":"\([^"]*\)".*/\1/p')
  ROLE_OF[$p]="$role"
  log "  :$p → ${role:-(no response)}"
done

run_case() {
  local name=$1 payload=$2
  sec "CASE: $name"
  printf "  %-6s  %-14s  %-8s  %s\n" "PORT" "ROLE" "VOTE" "REASON" | tee -a "$REPORT_FILE"
  printf "  %-6s  %-14s  %-8s  %s\n" "─────" "────" "────" "──────" | tee -a "$REPORT_FILE"
  for p in "${PORTS[@]}"; do
    resp=$(curl -s --max-time 5 -X POST "http://127.0.0.1:$p/consensus" \
      -H 'Content-Type: application/json' -d "$payload" 2>/dev/null)
    v=$(extract_vote "$resp"); r=$(extract_reason "$resp"); role="${ROLE_OF[$p]}"
    if [ -z "$v" ]; then
      printf "  %-6s  %-14s  %-8s  %s\n" ":$p" "$role" "NO-RESP" "$(echo "$resp" | head -c 60)" | tee -a "$REPORT_FILE"
    else
      printf "  %-6s  %-14s  %-8s  %s\n" ":$p" "$role" "$v" "$r" | tee -a "$REPORT_FILE"
    fi
  done
}

# ============ CASE MATRIX ============
run_case "01_VALID_BASELINE" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$FUTURE" 0 alice LIC-A "$TX1")"

run_case "02_BAD_SIGNATURE_all_zeros" \
  "$(build_block "$Z64$Z64" "$TX_HASH" "$A64" binary "$FUTURE" 0 alice LIC-A "$TX1")"

run_case "03_LICENSE_EMPTY_USER_ID" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$FUTURE" 0 "" LIC-A "$TX1")"

run_case "04_USER_MISMATCH_tx.from=alice_license.user=bob" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$FUTURE" 0 bob LIC-A "$TX1")"

run_case "05_LICENSE_EXPIRED" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$PAST" 0 alice LIC-A "$TX1")"

run_case "06_VOLUME_EXCEEDED" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$FUTURE" 999999 alice LIC-A "$TX1")"

run_case "07_BAD_MERKLE_ROOT" \
  "$(build_block "$D128" "$Z64" "$A64" binary "$FUTURE" 0 alice LIC-A "$TX1")"

run_case "08_BAD_BLOCK_HASH" \
  "$(build_block "$D128" "$TX_HASH" "$Z64" binary "$FUTURE" 0 alice LIC-A "$TX1")"

run_case "09_BAD_MERKLE_TYPE_quad" \
  "$(build_block "$D128" "$TX_HASH" "$A64" quad "$FUTURE" 0 alice LIC-A "$TX1")"

run_case "10_TX_LICENSE_NOT_IN_BLOCK" \
  "$(build_block "$D128" "$TX_HASH" "$A64" binary "$FUTURE" 0 alice LIC-B "$TX1")"

# ============ SWITCH / API ============
sec "STEP — license/verify روی سوئیچ و API (لایسنس جعلی)"
sw=$(curl -s -w "\n%{http_code}" --max-time 3 -X POST "http://127.0.0.1:8091/license/verify" \
  -H 'Content-Type: application/json' -H 'X-Session-Token: probe' \
  -d '{"license_id":"FAKE-DOES-NOT-EXIST"}' 2>/dev/null)
log "  switch :8091 → HTTP $(echo "$sw" | tail -n1)  $(echo "$sw" | head -n -1)"

ap=$(curl -s -w "\n%{http_code}" --max-time 3 -X POST "http://127.0.0.1:8080/api/v1/license/verify" \
  -H 'Content-Type: application/json' -H 'X-License-ID: FAKE' \
  -d '{"license_id":"FAKE-DOES-NOT-EXIST"}' 2>/dev/null)
log "  api :8080    → HTTP $(echo "$ap" | tail -n1)  $(echo "$ap" | head -n -1)"

# ============ TPS ============
sec "STEP — TPS load"
TPS_SCRIPT=""
for c in "$HOME/beaconchain-horizon/horizon-core/tps-test.sh" \
         "$HOME/horizon-test/tps-test.sh" "./tps-test.sh"; do
  [ -f "$c" ] && { TPS_SCRIPT="$c"; break; }
done
if [ -n "$TPS_SCRIPT" ]; then
  log "  running $TPS_SCRIPT (max 180s)..."
  if timeout 180 bash "$TPS_SCRIPT" > /tmp/tps_run.log 2>&1; then
    log "  ✅ completed"
    tail -n 15 /tmp/tps_run.log | sed 's/^/    /' | tee -a "$REPORT_FILE"
  else
    log "  ⚠️  failed/timeout"
    tail -n 15 /tmp/tps_run.log | sed 's/^/    /' | tee -a "$REPORT_FILE"
  fi
else
  log "  ⚠️  tps-test.sh پیدا نشد"
fi

sec "پایان — گزارش کامل: $REPORT_FILE"
read -r -p "Enter..." _
