#!/bin/bash
cd ~/Beaconchain/horizon-core
echo "⏳ Building Runner..."
go build -o agent-runner.exe ./cmd/agent-runner
if [ $? -ne 0 ]; then echo "❌ Build failed!"; exit 1; fi

echo "⏳ Checking Guardian (9008)..."
if ! curl -s -o /dev/null http://localhost:9008/health; then
    echo "❌ Guardian offline. Run ./consensus-run.sh first."
    exit 1
fi

taskkill //F //IM agent-runner.exe 2>/dev/null || true
AGENT_RUNNER_PORT=9101 nohup ./agent-runner.exe > logs/agent-runner.log 2>&1 &
sleep 2

echo "✅ Runner is UP."
curl -s http://localhost:9101/health
echo ""
