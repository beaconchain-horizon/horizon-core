#!/usr/bin/env bash

cd ~/Beaconchain/horizon-core

echo "1) Stopping old API/Switch processes..."
taskkill //F //IM horizon-api.exe //T 2>/dev/null || true
taskkill //F //IM horizon-switch.exe //T 2>/dev/null || true
taskkill //F //IM go.exe //T 2>/dev/null || true
sleep 2

echo "2) Starting Backend (API) and Switch..."
export ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export HORIZON_DB_PATH="$HOME/Beaconchain/horizon-core/horizon-dev.db"
export SERVER_PORT=8080
export SWITCH_PORT=8091

nohup go run ./cmd/api > api.log 2>&1 &
nohup go run ./cmd/switch > switch.log 2>&1 &

echo "3) Waiting 6 seconds for services to boot..."
sleep 6

echo "4) Fetching REAL Chain Info..."
RESPONSE=$(curl -s http://localhost:8080/api/v1/chain/info)

if [ -z "$RESPONSE" ]; then
    echo "❌ API is still not responding. Check api.log for errors."
    tail -n 5 api.log
else
    echo "✅ REAL CHAIN INFO:"
    echo "$RESPONSE"
fi
