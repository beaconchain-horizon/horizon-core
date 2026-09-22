#!/bin/bash
set -e
echo "=== Horizon Benchmark ==="
./bin/sensortool genkey --out=/tmp/k
export SWITCH_DB=/tmp/b.db
export CHAIN_CONFIG=/app/config/chain.json
export ADMIN_TOKEN=bench
export GIN_MODE=release
./bin/switch > /tmp/s.log 2>&1 &
sleep 4
curl -s -X POST http://127.0.0.1:8080/api/v1/industrial/sites -H "X-Admin-Token: bench" -H "Content-Type: application/json" -d "{\"site_id\":\"b\",\"name\":\"B\",\"type\":\"test\"}"
P=$(cat /tmp/k/public.pem | sed ":a;N;\$!ba;s/\n/\\\\n/g")
curl -s -X POST http://127.0.0.1:8080/api/v1/industrial/sensors -H "X-Admin-Token: bench" -H "Content-Type: application/json" -d "{\"sensor_id\":\"b1\",\"site_id\":\"b\",\"name\":\"B1\",\"type\":\"temperature\",\"unit\":\"C\",\"min_value\":0,\"max_value\":100,\"public_key\":\"$P\"}"
./bin/tpsbench -url=http://127.0.0.1:8080/api/v1/industrial/reading/batch -key=/tmp/k/private.pem -sensor=b1 -n=5000 -c=50 -bs=100
curl -s http://127.0.0.1:8080/api/v1/industrial/dashboard -H "X-Admin-Token: bench"
