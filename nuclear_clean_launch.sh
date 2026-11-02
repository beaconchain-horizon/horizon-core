#!/bin/bash

echo "=========================================="
echo " ☢️ NUCLEAR CLEANUP & PERSISTENT LAUNCH"
echo "=========================================="

cd ~/Beaconchain/horizon-core || exit 1

# ۱. کشتن هر چیزی که پورت‌های 8080 تا 8091 را گرفته است (با شناسایی PID)
echo "[1/6] Hunting down processes holding ports 8080-8091..."
for port in {8080..8091}; do
    PID=$(netstat -ano | grep ":$port" | awk '{print $5}' | head -n 1)
    if [ ! -z "$PID" ] && [ "$PID" != "0" ]; then
        echo "Killing PID $PID on port $port..."
        taskkill //F //PID $PID 2>/dev/null || true
    fi
done
taskkill //F //IM api.exe //T 2>/dev/null || true
taskkill //F //IM switch.exe //T 2>/dev/null || true
sleep 3

# ۲. حذف فایل‌های دیتابیس (اگر باز هم قفل بود، اسمشان را تغییر می‌دهیم)
echo "[2/6] Deleting locked databases..."
rm -f horizon-core*.db* 2>/dev/null || true
for f in horizon-core*.db; do
    if [ -f "$f" ]; then
        mv "$f" "${f}.locked_$(date +%s)"
    fi
done

# ۳. بیلد مجدد
echo "[3/6] Rebuilding..."
export GO111MODULE=on CGO_ENABLED=0
go build -o bin/api.exe ./cmd/api || { echo "❌ Build failed!"; exit 1; }

# ۴. ساخت ۷ فایل .bat جداگانه برای جلوگیری از تداخل متغیرها
echo "[4/6] Creating isolated Batch files for each node..."
mkdir -p logs
for i in {1..7}; do
    PORT=$((8080 + i))
    cat > "run_node_$i.bat" << EOF
@echo off
set NODE_ID=node-$i
set PORT=$PORT
cd /d "%~dp0"
.\bin\api.exe > logs\node-$i.log 2>&1
EOF
done

# ۵. اجرای نودها با فایل‌های جداگانه
echo "[5/6] Launching nodes independently..."
for i in {1..7}; do
    start "Node-$i" /MIN cmd /c "run_node_$i.bat"
    timeout /t 1 /nobreak > NUL
done

# ۶. بررسی وضعیت
echo "[6/6] Checking status..."
sleep 3
API_COUNT=$(tasklist //FI "IMAGENAME eq api.exe" 2>/dev/null | grep -c "api.exe")

echo "✅ Active API Nodes: $API_COUNT / 7"

if [ "$API_COUNT" -ge 7 ]; then
    echo "🎉 SUCCESS! Network is fully live and persistent."
else
    echo "⚠️ Warning: Some nodes failed. Let's see why:"
    for i in {1..7}; do
        if [ -f "logs/node-$i.log" ]; then
            echo "--- Node $i Last Lines ---"
            tail -n 3 "logs/node-$i.log"
        fi
    done
fi
echo "=========================================="
