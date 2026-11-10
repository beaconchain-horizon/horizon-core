#!/bin/bash
cd ~/Beaconchain/horizon-core

echo "⏳ [1/3] Building Brain..."
go build -o brain.exe ./cmd/brain
if [ $? -ne 0 ]; then echo "❌ Brain build failed"; exit 1; fi

echo "⏳ [2/3] Building Panel Server..."
go build -o panel-server.exe ./cmd/panel-server
if [ $? -ne 0 ]; then echo "❌ Panel build failed"; exit 1; fi

echo "⏳ [3/3] Starting Services..."
taskkill //F //IM brain.exe 2>/dev/null || true
taskkill //F //IM panel-server.exe 2>/dev/null || true
sleep 1

# اجرای مغز
BRAIN_PORT=9300 nohup ./brain.exe > logs/brain.log 2>&1 &
sleep 1

# اجرای سرور پنل (فقط برای سرو کردن HTML و آمار)
PANEL_PORT=9200 nohup ./panel-server.exe > logs/panel.log 2>&1 &
sleep 2

echo ""
echo "✅ Brain Health:"
curl -s http://localhost:9300/api/health
echo ""
echo "✅ Panel: http://localhost:9200"
echo ""
