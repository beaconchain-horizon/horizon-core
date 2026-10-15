#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

TX="${1:?usage: $0 <tx.json>}"
BODY=$(cat "$TX")

CRITICAL="9001 9002 9003 9004"
IMPORTANT="9005 9006 9007"

cy=0; ct=0; iy=0; it=0

for p in $CRITICAL; do
  ct=$((ct+1))
  R=$(curl -s --max-time 2 -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$BODY")
  V=$(echo "$R" | grep -o "\"vote\":true" || true)
  if [ -n "$V" ]; then cy=$((cy+1)); echo "  [$p] YES"; else echo "  [$p] NO"; fi
done

for p in $IMPORTANT; do
  it=$((it+1))
  R=$(curl -s --max-time 2 -X POST "http://localhost:$p/verify" -H "Content-Type: application/json" -d "$BODY")
  V=$(echo "$R" | grep -o "\"vote\":true" || true)
  if [ -n "$V" ]; then iy=$((iy+1)); echo "  [$p] YES"; else echo "  [$p] NO"; fi
done

echo ""
echo "Critical: $cy/$ct (need 100%)"
echo "Important: $iy/$it (need 2)"

if [ "$cy" -eq "$ct" ] && [ "$iy" -ge 2 ]; then
  echo "CONSENSUS REACHED - ACCEPT"
  exit 0
else
  echo "NO CONSENSUS - REJECT (fail-closed)"
  exit 1
fi
