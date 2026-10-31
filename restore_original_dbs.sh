#!/usr/bin/env bash
set -e

echo "1) Stopping API and Switch (Nodes are NOT touched)..."
taskkill //F //IM horizon-api.exe //T 2>/dev/null || true
taskkill //F //IM horizon-switch.exe //T 2>/dev/null || true
sleep 2

cd ~/Beaconchain/horizon-core

echo "2) Restoring the ORIGINAL backend database (Block 1 to 5)..."
# کپی کردن فایل بکاپ به جای فایل خالی فعلی
cp backup_dbs/horizon-dev.db.1793367182 horizon-dev.db
echo "✅ Backend DB restored (112 KB)."

echo "3) Restoring the ORIGINAL switch database (4.6 GB)..."
cp backup_dbs/horizon-switch.db horizon-switch.db
echo "✅ Switch DB restored."

echo "4) Starting Backend (Reading the restored DB)..."
export ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export SERVER_PORT=8080
export HORIZON_DB_PATH="$HOME/Beaconchain/horizon-core/horizon-dev.db"
nohup go run ./cmd/api > api.log 2>&1 &
sleep 6

echo "5) Starting Switch (Reading the restored DB)..."
export SWITCH_PORT=8091
nohup go run ./cmd/switch > switch.log 2>&1 &
sleep 6

echo "6) Final Check: Is Block 5 back in the chain?"
curl -s http://localhost:8080/api/v1/chain/info
echo ""
