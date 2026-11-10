#!/bin/bash
cd ~/Beaconchain/horizon-core
taskkill //F //IM brain.exe 2>/dev/null || true
go build -o brain.exe ./cmd/brain
if [ $? -ne 0 ]; then echo "❌ build failed"; exit 1; fi
BRAIN_PORT=9300 nohup ./brain.exe > logs/brain.log 2>&1 &
sleep 2
echo "✅ Brain on :9300"
curl -s http://localhost:9300/api/health
echo ""
echo ""
echo "=== تست ==="
curl -s -X POST http://localhost:9300/api/chat -H "Content-Type: application/json" -d '{"message":"سلام"}' | head -c 300
echo ""
