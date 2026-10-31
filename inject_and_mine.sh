#!/usr/bin/env bash
set -e

echo "1) Injecting License directly into Switch (port 8091)..."
# استفاده از هدر درست (X-Admin-Token) و ارسال لایسنس دقیق
curl -s -X POST http://localhost:8091/api/v1/license/save \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
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
  }'
echo ""

echo "2) Verifying if Switch sees the license..."
curl -s http://localhost:8091/api/v1/license/check -H "X-Admin-Token: 12345678901234567890123456789012"
echo ""

echo "3) Sending test transaction to Switch to trigger block creation..."
SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

curl -s -X POST http://localhost:8091/api/v1/tx/batch \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

echo "4) Waiting 3 seconds for Backend to mine the block..."
sleep 3

echo "5) Final Check (Height MUST be 5)..."
curl -s http://localhost:8080/api/v1/chain/info
echo ""
