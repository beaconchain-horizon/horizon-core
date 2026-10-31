#!/usr/bin/env bash
set -e

TOKEN="12345678901234567890123456789012"
SW="http://localhost:8091"
BE="http://localhost:8080"

echo "1) Unlock switch key..."
curl -s -X POST "$SW/api/v1/key/setup" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null || true

curl -s -X POST "$SW/api/v1/key/unlock" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{"password":"bench"}' > /dev/null || true

echo "2) Verify key status..."
curl -s "$SW/api/v1/key/status" -H "X-Admin-Token: $TOKEN" || true
echo ""

echo "3) Create a batch transaction (this must return processed:1 and chain:batched:1)..."
SESS=$(curl -s -X POST "$SW/api/v1/signing/session" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS" ]; then
  echo "❌ No session token. Check switch admin auth header (X-Admin-Token vs X-Customer-Token)."
  exit 1
fi

curl -s -X POST "$SW/api/v1/tx/batch" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

sleep 3

echo "4) Check chain height..."
curl -s "$BE/api/v1/chain/info"
echo ""

echo "5) Check latest blocks..."
curl -s "$BE/api/v1/blocks?limit=6" | head -c 500
echo ""
