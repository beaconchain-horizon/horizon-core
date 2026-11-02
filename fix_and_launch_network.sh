#!/bin/bash

echo "=========================================="
echo " 🚀 FIXING DATABASE LOCKS & LAUNCHING NETWORK"
echo "=========================================="

cd ~/Beaconchain/horizon-core || exit 1

# ۱. پاکسازی
echo "[1/5] Cleaning up old processes and DBs..."
taskkill //F //IM api.exe //T 2>/dev/null || true
taskkill //F //IM switch.exe //T 2>/dev/null || true
sleep 2
rm -f horizon-core*.db*

# ۲. اصلاح init.go برای ساخت دیتابیس اختصاصی برای هر نود
echo "[2/5] Modifying DB Init for Node Isolation..."
cat > internal/db/init.go << 'GOEOF'
package db

import (
	"fmt"
	"log"
	"os"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	
	// ساخت نام دیتابیس بر اساس NODE_ID (پیش‌فرض node-default)
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "default"
	}
	dbName := fmt.Sprintf("horizon-core-%s.db", nodeID)

	DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil { log.Fatalf("DB Init Failed: %v", err) }
	
	err = DB.AutoMigrate(&Block{}, &Transaction{}, &Agent{}, &License{}, &SystemLaw{})
	if err != nil { log.Fatalf("Migration Failed: %v", err) }
	
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-001", Title: "Zero-Storage Private Key", Status: "MANDATORY"})
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-002", Title: "Hardware-Bound License", Status: "MANDATORY"})
	DB.FirstOrCreate(&SystemLaw{LawID: "LAW-003", Title: "All Internal Comms Require License", Status: "MANDATORY"})
	
	log.Printf("✅ Database initialized: %s", dbName)
}
GOEOF

# ۳. بیلد مجدد
echo "[3/5] Building latest binaries..."
export GO111MODULE=on CGO_ENABLED=0
go build -o bin/api.exe ./cmd/api || { echo "❌ Build failed!"; exit 1; }
echo "✅ Build successful."

# ۴. ساخت فایل اجرایی Batch با تاخیر
echo "[4/5] Creating robust Windows Batch Launcher..."
mkdir -p logs
cat > launch_network.bat << 'EOF'
@echo off
cd /d "%~dp0"
echo Starting Horizon Core Nodes in background...

start "Node-1" /MIN cmd /c "set NODE_ID=node-1&& set PORT=8081&& .\bin\api.exe > logs\node-1.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-2" /MIN cmd /c "set NODE_ID=node-2&& set PORT=8082&& .\bin\api.exe > logs\node-2.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-3" /MIN cmd /c "set NODE_ID=node-3&& set PORT=8083&& .\bin\api.exe > logs\node-3.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-4" /MIN cmd /c "set NODE_ID=node-4&& set PORT=8084&& .\bin\api.exe > logs\node-4.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-5" /MIN cmd /c "set NODE_ID=node-5&& set PORT=8085&& .\bin\api.exe > logs\node-5.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-6" /MIN cmd /c "set NODE_ID=node-6&& set PORT=8086&& .\bin\api.exe > logs\node-6.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-7" /MIN cmd /c "set NODE_ID=node-7&& set PORT=8087&& .\bin\api.exe > logs\node-7.log 2>&1"
timeout /t 1 /nobreak > NUL

echo Starting Horizon Switch (if exists)...
if exist .\horizon-bank-package\switch.exe (
    start "Horizon-Switch" /MIN cmd /c ".\horizon-bank-package\switch.exe > logs\switch.log 2>&1"
)

exit
EOF

# ۵. اجرای شبکه
echo "[5/5] Launching network..."
cmd //c start /MIN launch_network.bat
sleep 6

# چک کردن وضعیت
API_COUNT=$(tasklist //FI "IMAGENAME eq api.exe" 2>/dev/null | grep -c "api.exe")
SWITCH_COUNT=$(tasklist //FI "IMAGENAME eq switch.exe" 2>/dev/null | grep -c "switch.exe")

echo "✅ Active API Nodes: $API_COUNT / 7"
echo "✅ Active Switches: $SWITCH_COUNT"

if [ "$API_COUNT" -ge 7 ]; then
    echo "🎉 NETWORK IS LIVE AND PERSISTENT!"
    echo "👉 Check the 'logs/' directory for individual node logs."
else
    echo "⚠️ Warning: Some nodes might have failed. Check logs."
fi
echo "=========================================="
