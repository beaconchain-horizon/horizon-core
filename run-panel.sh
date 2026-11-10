#!/bin/bash
cd ~/Beaconchain/horizon-core
echo "⏳ در حال کامپایل بک‌اند..."
go build -o panel-server.exe ./cmd/panel-server
if [ $? -ne 0 ]; then echo "❌ خطا در کامپایل"; exit 1; fi

taskkill //F //IM panel-server.exe 2>/dev/null || true
PANEL_PORT=9200 nohup ./panel-server.exe > logs/panel.log 2>&1 &
sleep 2
echo "✅ پنل روی http://localhost:9200 بالا آمد."
echo "⚠️ برای اینکه چت کار کند، مطمئن شوید Ollama نصب است (ollama serve)."
