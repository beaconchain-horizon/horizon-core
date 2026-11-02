#!/bin/bash

echo "=========================================="
echo " 🚀 LAUNCHING PERSISTENT LOCAL NETWORK"
echo "=========================================="

cd ~/Beaconchain/horizon-core || exit 1

# ۱. پاکسازی پروسه‌های قبلی
echo "[1/5] Cleaning up old processes..."
taskkill //F //IM api.exe //T 2>/dev/null || true
taskkill //F //IM switch.exe //T 2>/dev/null || true
sleep 2

# ۲. کامپایل مجدد (برای اطمینان از اجرای آخرین کدهای شاخه main)
echo "[2/5] Building latest binaries..."
export GO111MODULE=on CGO_ENABLED=0
go build -o bin/api.exe ./cmd/api || { echo "❌ Build failed!"; exit 1; }
echo "✅ Build successful."

# ۳. ساخت پوشه لاگ‌ها
mkdir -p logs

# ۴. ساخت فایل اجرایی ویندوز (Batch File)
echo "[3/5] Creating Windows Batch Launcher..."
cat > launch_network.bat << 'EOF'
@echo off
cd /d "%~dp0"
echo Starting Horizon Core Nodes in background...

start "Node-1" /MIN cmd /c "set NODE_ID=node-1&& set PORT=8081&& .\bin\api.exe > logs\node-1.log 2>&1"
start "Node-2" /MIN cmd /c "set NODE_ID=node-2&& set PORT=8082&& .\bin\api.exe > logs\node-2.log 2>&1"
start "Node-3" /MIN cmd /c "set NODE_ID=node-3&& set PORT=8083&& .\bin\api.exe > logs\node-3.log 2>&1"
start "Node-4" /MIN cmd /c "set NODE_ID=node-4&& set PORT=8084&& .\bin\api.exe > logs\node-4.log 2>&1"
start "Node-5" /MIN cmd /c "set NODE_ID=node-5&& set PORT=8085&& .\bin\api.exe > logs\node-5.log 2>&1"
start "Node-6" /MIN cmd /c "set NODE_ID=node-6&& set PORT=8086&& .\bin\api.exe > logs\node-6.log 2>&1"
start "Node-7" /MIN cmd /c "set NODE_ID=node-7&& set PORT=8087&& .\bin\api.exe > logs\node-7.log 2>&1"

echo Starting Horizon Switch...
if exist .\horizon-bank-package\switch.exe (
    start "Horizon-Switch" /MIN cmd /c ".\horizon-bank-package\switch.exe > logs\switch.log 2>&1"
) else (
    echo Switch.exe not found in horizon-bank-package.
)

exit
EOF

# ۵. اجرای فایل Batch به صورت مخفی و مستقل
echo "[4/5] Launching network independently of this terminal..."
cmd //c start /MIN launch_network.bat

# ۶. بررسی وضعیت
echo "[5/5] Verifying network status..."
sleep 4

# چک کردن تعداد پروسه‌های api.exe
API_COUNT=$(tasklist //FI "IMAGENAME eq api.exe" 2>/dev/null | grep -c "api.exe")
SWITCH_COUNT=$(tasklist //FI "IMAGENAME eq switch.exe" 2>/dev/null | grep -c "switch.exe")

echo "✅ Active API Nodes: $API_COUNT / 7"
echo "✅ Active Switches: $SWITCH_COUNT"

if [ "$API_COUNT" -ge 7 ]; then
    echo "🎉 NETWORK IS LIVE AND PERSISTENT!"
    echo "👉 You can now safely close this Git Bash terminal. The nodes will keep running in the background."
    echo "📁 Logs are saved in the 'logs/' directory."
else
    echo "⚠️ Warning: Not all nodes started. Check the logs."
fi

echo "=========================================="
