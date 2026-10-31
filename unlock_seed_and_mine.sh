#!/usr/bin/env bash
set -e

TOKEN="12345678901234567890123456789012"
SW="http://localhost:8091"
BE="http://localhost:8080"

echo "1) Unlocking Switch Key (Again)..."
curl -s -X POST "$SW/api/v1/key/unlock" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{"password":"bench"}'
echo ""

echo "2) Seeding Accounts (Giving them balance)..."
curl -s -X POST "$SW/api/v1/account/seed" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{"bank_id":"bank_markazi","amount":1000000}' > /dev/null
curl -s -X POST "$SW/api/v1/account/seed" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{"bank_id":"bank_melli","amount":1000000}' > /dev/null
echo "✅ Accounts seeded."

echo "3) Registering License to Backend (Again, because we made a fresh DB)..."
curl -s -X POST "$BE/api/v1/license/register" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: $TOKEN" \
  -d '{
    "license_id": "HORIZON-GOLD-005",
    "user_id": "nicico",
    "product_id": "horizon-core",
    "volume": 1000000,
    "used": 0,
    "duration": 4,
    "merkle_root": "b395f34ddf5da832eb477f3a81bc1e7d03b0011795c8896f18bf6d2382a22c9a",
    "issued_at": 1793298061,
    "expires_at": 1793643661,
    "status": "active",
    "hardware_id": "",
    "signature": "91d98bafb73207445f40d0eb38c0ac66cec226e66e3e41f9b99d17f135476c58057169ab7a8b499f503c1e9e42cbdcbbd38244699c8c1a2abcd9ed255a6cb4d8",
    "signed_data": "{\"license_id\":\"HORIZON-GOLD-005\",\"user_id\":\"nicico\",\"product_id\":\"horizon-core\",\"volume\":1000000,\"duration\":4,\"merkle_root\":\"b395f34ddf5da832eb477f3a81bc1e7d03b0011795c8896f18bf6d2382a22c9a\",\"issued_at\":1793298061,\"expires_at\":1793643661,\"hardware_id\":\"\"}",
    "public_key": "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
  }' > /dev/null
echo "✅ License sent to Backend (queued)."

echo "4) Sending test transaction to Switch..."
SESS=$(curl -s -X POST "$SW/api/v1/signing/session" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

curl -s -X POST "$SW/api/v1/tx/batch" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: $TOKEN" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

echo "5) Waiting 3 seconds for Backend batcher..."
sleep 3

echo "6) Final Check (Height MUST be 5)..."
curl -s "$BE/api/v1/chain/info"
echo ""
