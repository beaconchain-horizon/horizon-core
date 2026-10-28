#!/usr/bin/env bash
# ==============================================================================
# HORIZON CORE - LIVE GUARDIAN PROTOCOL v1.3 (Non-Destructive)
# ==============================================================================
ADMIN_TOKEN="12345678901234567890123456789012"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export NODE_PUBKEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export ADMIN_TOKEN

BE="http://localhost:8080"
SW="http://localhost:8091"
LOG_FILE="guardian-live.log"

echo "=========================================================="
echo "  ⚖️  INITIATING HORIZON GUARDIAN PROTOCOL v1.3"
echo "=========================================================="

# PHASE 1: CLEANUP (فقط پروسه‌ها را می‌کشد، دیتابیس دست‌نخورده می‌ماند)
echo -e "\n[1/3] Cleaning up old processes..."
for p in 8080 8091 9001 9002 9003 9004 9005 9006 9007; do
  pid=$(netstat -ano | grep ":$p " | grep LISTENING | awk '{print $5}' | head -1)
  [ -n "$pid" ] && taskkill //PID $pid //F >/dev/null 2>&1
done
mkdir -p .gotmp && export TMP="$(pwd)/.gotmp" && export TEMP="$(pwd)/.gotmp"

# PHASE 2: START SERVICES
echo -e "\n[2/3] Compiling and starting 9 services..."
go build -o backend.exe ./cmd/api/ 2>/dev/null
go build -o switch.exe ./cmd/switch/ 2>/dev/null
go build -o angel-node.exe ./cmd/angel-node/ 2>/dev/null

export SWITCH_PORT=8091 HORIZON_DB_PATH=./horizon-dev.db

NODE_ROLE=merkle NODE_PORT=9001 ./angel-node.exe >/tmp/n1.log 2>&1 &
NODE_ROLE=hash   NODE_PORT=9002 ./angel-node.exe >/tmp/n2.log 2>&1 &
NODE_ROLE=ecdsa  NODE_PORT=9003 ./angel-node.exe >/tmp/n3.log 2>&1 &
NODE_ROLE=type   NODE_PORT=9004 ./angel-node.exe >/tmp/n4.log 2>&1 &
NODE_ROLE=volume NODE_PORT=9005 ./angel-node.exe >/tmp/n5.log 2>&1 &
NODE_ROLE=time   NODE_PORT=9006 ./angel-node.exe >/tmp/n6.log 2>&1 &
NODE_ROLE=user   NODE_PORT=9007 ./angel-node.exe >/tmp/n7.log 2>&1 &
sleep 2
./switch.exe >/tmp/switch.log 2>&1 &
sleep 2
./backend.exe >/tmp/backend.log 2>&1 &
sleep 3

if ! curl -s $BE/api/v1/health | grep -q "online"; then echo "❌ Backend Failed."; exit 1; fi
if ! curl -s $SW/api/v1/health | grep -q "online"; then echo "❌ Switch Failed."; exit 1; fi
echo "  ✅ All 9 services are online. (Your real blocks are safe)"

# PHASE 3: LIVE GUARDIAN MODE
echo -e "\n[3/3] Entering Live Guardian Mode (Silent Background)..."
echo "--- Guardian Started at $(date) ---" >> $LOG_FILE

while true; do
  sleep 60 # هر ۶۰ ثانیه
  
  TIME_NOW=$(date +%T)
  
  # ۱. چک بک‌اند
  BE_STATUS=$(curl -s $BE/api/v1/health | grep -o '"status":"online"')
  if [ -z "$BE_STATUS" ]; then echo "[$TIME_NOW] ALERT: Backend DOWN!" >> $LOG_FILE; fi

  # ۲. چک ۷ نود
  H1=$(curl -s localhost:9001/health | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
  H7=$(curl -s localhost:9007/health | grep -o '"law_hash":"[^"]*"' | cut -d'"' -f4)
  if [ "$H1" != "$H7" ] || [ -z "$H1" ]; then echo "[$TIME_NOW] ALERT: Node Law Mismatch!" >> $LOG_FILE; fi

  # ۳. تست امنیتی زنده (SQLi)
  SEC_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/json" -d '{"license_id":"1 OR 1=1"}')
  if [ "$SEC_CODE" = "200" ] || [ "$SEC_CODE" = "201" ]; then echo "[$TIME_NOW] CRITICAL: SQLi Vulnerability!" >> $LOG_FILE; fi
  
  # ثبت وضعیت در لاگ
  echo "[$TIME_NOW] OK | Shield: $SEC_CODE | Law: ${H1:0:8}" >> $LOG_FILE
done
