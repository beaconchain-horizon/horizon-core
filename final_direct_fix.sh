#!/usr/bin/env bash
set -e

echo "1) Stopping all services..."
taskkill //F //IM horizon-api.exe //T 2>/dev/null || true
taskkill //F //IM horizon-switch.exe //T 2>/dev/null || true
taskkill //F //IM go.exe //T 2>/dev/null || true
sleep 2

echo "2) Directly cleaning horizon-dev.db (with Busy Timeout)..."
cat > direct_clean.go << 'GO_CODE'
package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "horizon-dev.db?_busy_timeout=5000")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM license_records WHERE id = 1555")
	if err != nil {
		fmt.Println("Error deleting ID 1555:", err)
	} else {
		rows, _ := res.RowsAffected()
		fmt.Printf("✅ Deleted stuck record ID 1555 (rows: %d)\n", rows)
	}

	db.Exec("DELETE FROM license_records WHERE license_id = 'HORIZON-GOLD-005'")
	fmt.Println("✅ Cleared old HORIZON-GOLD-005 (if existed).")
}
GO_CODE

go run direct_clean.go
rm direct_clean.go

echo "3) Starting Backend..."
export ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export SERVER_PORT=8080
nohup go run ./cmd/api > api.log 2>&1 &
sleep 5

echo "4) Starting Switch..."
export SWITCH_PORT=8091
nohup go run ./cmd/switch > switch.log 2>&1 &
sleep 5

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
