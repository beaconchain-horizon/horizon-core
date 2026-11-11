#!/usr/bin/env bash
# report-real.sh — فقط مشاهده. هیچ داده‌ای ساخته نمی‌شود.
set -u
C=$'\e[36m'; G=$'\e[32m'; R=$'\e[31m'; Y=$'\e[33m'; X=$'\e[0m'
sec() { printf "\n%s═══ %s ═══%s\n" "$C" "$1" "$X"; }

sec "1) لایسنس‌های واقعی در سوئیچ"
LICENSES=$(curl -s -H "X-Admin-Token: $ADMIN_TOKEN" http://localhost:8091/api/v1/license/list)
echo "$LICENSES" | head -c 400; echo

# یک لایسنس واقعی نمونه
SAMPLE_ID=$(echo "$LICENSES" | grep -o '"license_id":"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
echo ""
echo "نمونه: $SAMPLE_ID"

sec "2) verify همان لایسنس واقعی روی سوئیچ (۸۰۹۱)"
curl -s -X POST http://localhost:8091/api/v1/license/verify \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $ADMIN_TOKEN" \
  -d "{\"license_id\":\"$SAMPLE_ID\"}"
echo

sec "3) وضعیت ۸ آنجل"
for p in 9001 9002 9003 9004 9005 9006 9007 9008; do
  echo -n "  :$p  "
  curl -s --max-time 2 "http://127.0.0.1:$p/health"
  echo
done

sec "4) آمار هر آنجل (چقدر کار کرده)"
for p in 9001 9002 9003 9004 9005 9006 9007 9008; do
  echo -n "  :$p  "
  curl -s --max-time 2 "http://127.0.0.1:$p/stats"
  echo
done

sec "5) chain-verify (واقعی)"
go run ./cmd/chain-verify 2>&1 | tail -10

sec "6) آمار سوئیچ"
curl -s http://localhost:8091/stats 2>/dev/null | head -c 500
echo
