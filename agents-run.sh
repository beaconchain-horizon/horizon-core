#!/bin/bash
# ==========================================
# HORIZON CORE - AI AGENTS LAUNCHER
# ==========================================

echo "⏳ [1/3] Stopping old AI agent processes..."
taskkill //F //IM node.exe 2>/dev/null || true
sleep 1

echo "⏳ [2/3] Starting Angel Agents Web Server (Node.js)..."
# رفتن به پوشه ایجنت‌ها برای خواندن فایل .env
cd ~/Beaconchain/angel-agents

# اجرای سرور در پس‌زمینه و ذخیره لاگ در پوشه horizon-core
nohup node web/server.js > ~/Beaconchain/horizon-core/logs/agents-web.log 2>&1 &
AGENT_PID=$!

sleep 3

echo "⏳ [3/3] Checking AI Agents Status..."
printf "%-20s %-10s %-15s\n" "Service" "Type" "Status"
printf "%-20s %-10s %-15s\n" "-------" "----" "------"

# چک کردن اینکه پروسه Node.js زنده است یا نه
if tasklist | grep -q "node.exe"; then
    printf "%-20s %-10s %-15s\n" "Angel Web Server" "Node.js" "✅ RUNNING (PID: $AGENT_PID)"
else
    printf "%-20s %-10s %-15s\n" "Angel Web Server" "Node.js" "❌ CRASHED"
    echo "   -> Check log: tail -n 20 ~/Beaconchain/horizon-core/logs/agents-web.log"
fi

echo ""
echo "🏁 Agents Launcher finished."
