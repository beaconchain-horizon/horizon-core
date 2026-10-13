#!/bin/bash
set -e
cd "$(dirname "$0")"

echo "=== Horizon Core Benchmark ==="
echo ""

taskkill //F //IM switch.exe 2>/dev/null || true
sleep 3

rm -f test-keys/private.pem test-keys/public.pem
./bin/sensortool genkey --out=test-keys

CFG="$(pwd)/config/chain.json"
(SWITCH_DB=:memory: CHAIN_CONFIG="$CFG" ADMIN_TOKEN=bench GIN_MODE=release ./bin/switch > /tmp/horizon.log 2>&1 &)
sleep 5

API="http://127.0.0.1:8080/api/v1"

echo ""
echo "Test: 200K readings, c=50, bs=500"
./bin/tpsbench -url=$API/industrial/reading/batch -key=test-keys/private.pem -sensor=tps-001 -n=200000 -c=50 -bs=500

echo ""
echo "=== Integrity Check ==="
curl -s "$API/industrial/dashboard" -H "X-Admin-Token: bench"
echo ""
echo ""
taskkill //F //IM switch.exe 2>/dev/null || true
echo "=== Done ==="
