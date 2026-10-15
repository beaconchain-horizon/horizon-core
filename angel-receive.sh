#!/usr/bin/env bash
set -e

LIC="${1:?usage: $0 <license.json from Dashboard>}"
TOKEN="${ADMIN_TOKEN:?ADMIN_TOKEN required}"

echo "=== 1. Extract license fields ==="
LIC_ID=$(grep -o "\"license_id\":\"[^\"]*\"" "$LIC" | cut -d"\"" -f4)
SIG=$(grep -o "\"signature\":\"[^\"]*\"" "$LIC" | cut -d"\"" -f4)
EXPIRES=$(grep -o "\"expires_at\":[0-9]*" "$LIC" | cut -d":" -f2)

PAYLOAD=$(printf "{\"left\":\"%s\",\"right\":\"%s\",\"header\":\"%s\",\"message\":%s,\"signature\":\"%s\",\"merkle_type\":\"binary\",\"used\":0,\"volume\":%s,\"issued_at\":%s,\"expires_at\":%s,\"user_id\":\"%s\",\"product_id\":\"%s\",\"expected_user\":\"%s\",\"expected_product\":\"%s\"}" "$LIC_ID" "$USER_ID" "$LIC_ID" "$(printf %s "$MSG" | python -c "import sys,json;print(json.dumps(sys.stdin.read()))")" "$SIG" "$VOL" "$ISSUED" "$EXPIRES" "$USER_ID" "$PROD_ID" "$USER_ID" "$PROD_ID")
    --data-binary @"$LIC"
  echo ""
else
  exit 1
fi
