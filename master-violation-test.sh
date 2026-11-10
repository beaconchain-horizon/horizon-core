#!/usr/bin/env bash
# master-violation-test.sh — تست تخلف روی ۸ آنجل و مسیر verify
set -u
R=$'\e[31m'; G=$'\e[32m'; Y=$'\e[33m'; C=$'\e[36m'; X=$'\e[0m'
ok()   { printf "%s[OK]%s   %s\n" "$G" "$X" "$1"; }
vio()  { printf "%s[VIOLATION]%s %s\n" "$R" "$X" "$1"; VIOLATORS+=("$1"); }
warn() { printf "%s[WARN]%s %s\n" "$Y" "$X" "$1"; }
sec()  { printf "\n%s== %s ==%s\n" "$C" "$1" "$X"; }

PORTS=(9001 9002 9003 9004 9005 9006 9007 9008)
VIOLATORS=()
NOW=$(date +%s); PAST=$((NOW-86400)); FUTURE=$((NOW+86400))
A64=$(printf 'a%.0s' {1..64})
Z64=$(printf '0%.0s' {1..64})
D128=$(printf 'd%.0s' {1..128})

sec "1. Discover roles"
declare -A ROLE_OF
for p in "${PORTS[@]}"; do
  h=$(curl -s --max-time 2 "http://127.0.0.1:$p/health" 2>/dev/null)
  role=$(echo "$h" | sed -n 's/.*"role":"\([^"]*\)".*/\1/p')
  ROLE_OF[$p]="$role"
  [ -n "$role" ] && ok ":$p → $role" || warn ":$p no response ($h)"
done

sec "2. Compute valid merkle root for sample tx"
TX='{"from":"A","to":"B","amount":10,"license_id":"L1","type":"transfer","payload":""}'
GOOD_MERKLE=$(printf "%s" "$TX" | sha256sum | awk '{print $1}')
echo "sha256(tx) = $GOOD_MERKLE"

PAYLOAD_GOOD=$(cat <<JSON
{"block_number":1,"block_hash":"$A64","prev_hash":"$A64","merkle_root":"$GOOD_MERKLE","merkle_type":"binary","timestamp":$NOW,"tx_count":1,"signature":"$D128","signer_addr":"addr1","licenses":[{"license_id":"L1","user_id":"u1","volume":1000,"used":0,"issued_at":$PAST,"expires_at":$FUTURE,"status":"active"}],"txs":[$TX]}
JSON
)

PAYLOAD_BAD_MERKLE=$(echo "$PAYLOAD_GOOD" | sed "s/\"merkle_root\":\"$GOOD_MERKLE\"/\"merkle_root\":\"$Z64\"/")
PAYLOAD_BAD_HASH=$(echo "$PAYLOAD_GOOD" | sed "s/\"block_hash\":\"$A64\"/\"block_hash\":\"$Z64\"/")
PAYLOAD_BAD_SIG=$(echo "$PAYLOAD_GOOD" | sed "s/\"signature\":\"$D128\"/\"signature\":\"$Z64$Z64\"/")
PAYLOAD_BAD_TYPE=$(echo "$PAYLOAD_GOOD" | sed 's/"merkle_type":"binary"/"merkle_type":"weird"/')
PAYLOAD_EXPIRED=$(echo "$PAYLOAD_GOOD" | sed "s/\"expires_at\":$FUTURE/\"expires_at\":$PAST/")
PAYLOAD_OVER_VOL=$(echo "$PAYLOAD_GOOD" | sed 's/"used":0/"used":999999/')
PAYLOAD_WRONG_USER=$(echo "$PAYLOAD_GOOD" | sed 's/"user_id":"u1"/"user_id":"attacker"/;s/"from":"A"/"from":"victim"/')

test_angel() {
  local port=$1 label=$2 expect=$3 payload=$4
  local role="${ROLE_OF[$port]}"
  local resp=$(curl -s --max-time 4 -X POST "http://127.0.0.1:$port/consensus" \
    -H 'Content-Type: application/json' -d "$payload" 2>/dev/null)
  local vote=$(echo "$resp" | sed -n 's/.*"vote":"\([^"]*\)".*/\1/p')
  local reason=$(echo "$resp" | sed -n 's/.*"reason":"\([^"]*\)".*/\1/p')
  if [ -z "$vote" ]; then
    warn ":$port ($role) $label → NO VOTE"
  elif [ "$vote" = "$expect" ]; then
    ok ":$port ($role) $label → $vote"
  elif [ "$expect" = "REJECT" ]; then
    vio ":$port ($role) ACCEPTED $label (should REJECT) reason=[$reason]"
  else
    warn ":$port ($role) REJECTED $label (expected ACCEPT) reason=[$reason]"
  fi
}

sec "3. Role-targeted tests"
for p in "${PORTS[@]}"; do
  role="${ROLE_OF[$p]}"
  case "$role" in
    Merkle)
      test_angel $p "empty+zero-root ACCEPT"  "ACCEPT" "{\"block_number\":1,\"block_hash\":\"$A64\",\"prev_hash\":\"$A64\",\"merkle_root\":\"$Z64\",\"merkle_type\":\"binary\",\"timestamp\":$NOW,\"tx_count\":0,\"signature\":\"$D128\",\"signer_addr\":\"addr1\",\"licenses\":[],\"txs\":[]}"
      test_angel $p "non-empty+wrong-root REJECT" "REJECT" "$PAYLOAD_BAD_MERKLE"
      test_angel $p "non-empty+good-root ACCEPT"  "ACCEPT" "$PAYLOAD_GOOD"
      ;;
    Hash)    test_angel $p "bad-hash REJECT"      "REJECT" "$PAYLOAD_BAD_HASH" ;;
    ECDSA)   test_angel $p "bad-sig REJECT"       "REJECT" "$PAYLOAD_BAD_SIG" ;;
    Type)    test_angel $p "bad-type REJECT"      "REJECT" "$PAYLOAD_BAD_TYPE" ;;
    Volume)  test_angel $p "over-volume REJECT"   "REJECT" "$PAYLOAD_OVER_VOL" ;;
    Time)    test_angel $p "expired REJECT"       "REJECT" "$PAYLOAD_EXPIRED" ;;
    User)
      test_angel $p "user-mismatch REJECT"        "REJECT" "$PAYLOAD_WRONG_USER"
      test_angel $p "good ACCEPT"                 "ACCEPT" "$PAYLOAD_GOOD"
      ;;
    AgentGuardian) test_angel $p "bad-merkle REJECT" "REJECT" "$PAYLOAD_BAD_MERKLE" ;;
    *)        test_angel $p "generic-bad REJECT" "REJECT" "$PAYLOAD_BAD_MERKLE" ;;
  esac
done

sec "4. Switch verify endpoint (fake license)"
resp=$(curl -s -w "\n%{http_code}" --max-time 3 -X POST "http://127.0.0.1:8091/license/verify" \
  -H 'Content-Type: application/json' -d '{"license_id":"FAKE-DOES-NOT-EXIST"}' 2>/dev/null)
code=$(echo "$resp" | tail -n1); body=$(echo "$resp" | head -n -1)
if [ "$code" = "404" ]; then ok "switch rejects fake license (404)"
elif [ "$code" = "200" ] && echo "$body" | grep -q '"valid":true'; then
  vio "switch ACCEPTED fake license: $body"
else warn "switch response $code: $body"; fi

sec "5. API verify endpoint (fake license)"
resp=$(curl -s -w "\n%{http_code}" --max-time 3 -X POST "http://127.0.0.1:8080/api/v1/license/verify" \
  -H 'Content-Type: application/json' -d '{"license_id":"FAKE-DOES-NOT-EXIST"}' 2>/dev/null)
code=$(echo "$resp" | tail -n1); body=$(echo "$resp" | head -n -1)
if [ "$code" = "404" ]; then ok "api rejects fake license (404)"
elif [ "$code" = "200" ] && echo "$body" | grep -q '"valid":true'; then
  vio "api ACCEPTED fake license: $body"
else warn "api response $code: $body"; fi

sec "6. TPS load"
if [ -f tps-test.sh ]; then
  echo "running tps-test.sh (timeout 120s)..."
  if timeout 120 bash tps-test.sh > /tmp/tps_out.log 2>&1; then
    ok "tps-test passed"; tail -n 10 /tmp/tps_out.log | sed 's/^/    /'
  else
    warn "tps-test failed/timeout"; tail -n 10 /tmp/tps_out.log | sed 's/^/    /'
  fi
else warn "tps-test.sh not found"; fi

sec "Summary"
if [ ${#VIOLATORS[@]} -eq 0 ]; then
  echo "${G}No violations detected.${X}"
else
  echo "${R}VIOLATORS (${#VIOLATORS[@]}):${X}"
  for v in "${VIOLATORS[@]}"; do echo "  - $v"; done
fi
read -r -p "Enter..." _
