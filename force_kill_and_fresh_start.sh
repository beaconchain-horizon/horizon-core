#!/usr/bin/env bash
set -e

echo "1) Force killing processes on ports 8080 and 8091 (PowerShell)..."
powershell -Command "Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id \$_.OwningProcess -Force }"
powershell -Command "Get-NetTCPConnection -LocalPort 8091 -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id \$_.OwningProcess -Force }"
sleep 2

cd ~/Beaconchain/horizon-core

echo "2) Backing up old databases (forcing move)..."
mkdir -p backup_dbs
# اگر فایل قفل باشد، این خطا می‌دهد، اما حالا که پروسه‌ها بسته شده‌اند، باید موفق شود
rm -f backup_dbs/horizon-dev.db.old
rm -f backup_dbs/horizon-switch.db.old
mv horizon-dev.db backup_dbs/horizon-dev.db.old 2>/dev/null || echo "Note: horizon-dev.db not found in root."
mv horizon-switch.db backup_dbs/horizon-switch.db.old 2>/dev/null || echo "Note: horizon-switch.db not found in root."
mv data/*.db backup_dbs/ 2>/dev/null || true
echo "✅ Old databases moved."

echo "3) Starting Backend (Creating FRESH DB)..."
export ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export SERVER_PORT=8080
nohup go run ./cmd/api > api.log 2>&1 &
echo "Waiting for backend to create new DB and block 0..."
sleep 6

echo "4) Starting Switch..."
export SWITCH_PORT=8091
nohup go run ./cmd/switch > switch.log 2>&1 &
sleep 6

echo "5) Unlocking Switch Key..."
curl -s -X POST http://localhost:8091/api/v1/key/setup \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d '{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}' > /dev/null || true

curl -s -X POST http://localhost:8091/api/v1/key/unlock \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d '{"password":"bench"}' > /dev/null || true

echo "6) Registering EXACT Original Block 5 (user_id: nicico)..."
curl -s -X POST http://localhost:8080/api/v1/license/register \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: 12345678901234567890123456789012" \
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

echo "7) Sending test transaction to trigger batcher..."
SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

curl -s -X POST http://localhost:8091/api/v1/tx/batch \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

sleep 3

echo "8) Final Check..."
curl -s http://localhost:8080/api/v1/chain/info
echo ""
