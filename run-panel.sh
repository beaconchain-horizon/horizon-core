#!/bin/bash
cd ~/Beaconchain/horizon-core
echo "⏳ Building panel-server..."
go build -o panel-server.exe ./cmd/panel-server
if [ $? -ne 0 ]; then echo "❌ Build failed"; exit 1; fi
taskkill //F //IM panel-server.exe 2>/dev/null || true
PANEL_PORT=9200 nohup ./panel-server.exe > logs/panel-server.log 2>&1 &
sleep 2
echo ""
echo "✅ Panel is UP: http://localhost:9200"
echo "   → Open in browser: http://localhost:9200"
echo ""
curl -s http://localhost:9200/health
echo ""
