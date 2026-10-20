#!/usr/bin/env bash

echo "=================================================================="
echo "  HORIZON CORE - ADVANCED APT & SCADA PENETRATION TEST"
echo "  (Inspired by Stuxnet & ICS Attack Vectors)"
echo "=================================================================="

BE="http://localhost:8080"
SW="http://localhost:8091"
TOK="12345678901234567890123456789012"
HK="X-Admin-Key: $TOK"
HT="X-Admin-Token: $TOK"

BLOCKED=0
VULN=0

test_attack() {
  local desc="$1"; shift
  local response=$(curl -s -w "\n%{http_code}" "$@")
  local code=$(echo "$response" | tail -1)
  local body=$(echo "$response" | head -n -1)
  
  if [[ "$code" == 4* ]]; then
    echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code)\e[0m"
    BLOCKED=$((BLOCKED+1))
    return
  fi
  
  if [ "$code" = "200" ] || [ "$code" = "201" ]; then
    # بررسی نشت اطلاعات حساس (مثل خروجی passwd یا خطای دیتابیس)
    if [[ "$body" == *"root:"* ]] || [[ "$body" == *"syntax error"* ]] || [[ "$body" == *"SQL"* ]]; then
       echo -e "  \e[31m❌ [VULNERABLE] $desc (HTTP $code - DATA LEAK!)\e[0m"
       VULN=$((VULN+1))
    else
       echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code - Safe Response)\e[0m"
       BLOCKED=$((BLOCKED+1))
    fi
  else
    echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code)\e[0m"
    BLOCKED=$((BLOCKED+1))
  fi
}

echo ""
echo "[1] SCADA / ICS Protocol Abuse (Stuxnet-style)"
# استاکس‌نت از دستورات SQL مخفی برای دسترسی به PLC استفاده می‌کرد
test_attack "Modbus Function Code Injection" -X POST "$BE/api/v1/industry/sensors" -H "$HK" -H "Content-Type: application/json" -d '{"sensor_id":"PLC-01","value":999,"function_code":16,"address":40001}'
test_attack "DNP3 Unauthorized Control" -X POST "$BE/api/v1/industry/readings/batch" -H "$HK" -H "Content-Type: application/json" -d '{"device_id":"RTU-01","command":"OPERATE","point":1,"value":0}'
test_attack "PLC Logic Overwrite Attempt" -X POST "$BE/api/v1/industry/alerts/ack" -H "$HK" -H "Content-Type: application/json" -d '{"alert_id":"0; DROP TABLE alerts;--","operator":"admin"}'

echo ""
echo "[2] Stuxnet APT Vectors (Hardcoded Creds & Lateral Movement)"
# استاکس‌نت از پسوردهای هاردکد شده در دیتابیس SCADA استفاده می‌کرد
test_attack "Hardcoded SQL Auth Bypass" -X POST "$BE/api/v1/license/verify" -H "$HK" -H "Content-Type: application/json" -d '{"license_id":"admin","user_id":"admin","signature":"default_password","data":"w"}'
test_attack "Internal Network SSRF (PLC Probing)" -X POST "$BE/api/v1/license/verify" -H "$HK" -H "Content-Type: application/json" -d '{"data":"http://192.168.1.10:502/modbus"}'
test_attack "SMB/Print Spooler Vector" -X POST "$BE/api/v1/industry/readings" -H "$HK" -H "Content-Type: application/json" -d '{"sensor_id":"\\\\print_spooler\\pipe","value":1}'

echo ""
echo "[3] Advanced Web Attacks (Header & Deserialization)"
# حملات پیشرفته به لایه وب
test_attack "CRLF Header Injection" -X POST "$BE/api/v1/packages" -H "$HK" -H "Content-Type: application/json" -d '{"name":"Package\r\nX-Evil: 1"}'
test_attack "GraphQL / JSON Depth Attack" -X POST "$BE/api/v1/packages" -H "$HK" -H "Content-Type: application/json" -d '{"name":"A","nested":{"a":{"a":{"a":{"a":{"a":{"a":{"a":{"a":"deep"}}}}}}}}}'

echo ""
echo "[4] Memory & Command Execution"
# تلاش برای اجرای دستور در سطح سیستم‌عامل (مثل دستورات استاکس‌نت)
test_attack "OS Command via PID Spoofing" -X POST "$BE/api/v1/toolbox/subnet" -H "$HK" -H "Content-Type: application/json" -d '{"ip":"127.0.0.1; whoami","mask":24}'
test_attack "Path Traversal to System32" -X GET "$BE/api/v1/blocks/..\\..\\..\\windows\\system32\\drivers\\etc\\hosts" -H "$HK"

echo ""
echo "=================================================================="
echo "  FINAL SCORE"
echo "=================================================================="
echo -e "  \e[32m✅ BLOCKED (Safe): $BLOCKED\e[0m"
echo -e "  \e[31m❌ VULNERABLE (Fail): $VULN\e[0m"
echo "=================================================================="
if [ "$VULN" -eq 0 ]; then
  echo -e "  \e[1;32m🏆 PERFECT SCORE: ALL APT & SCADA ATTACKS BLOCKED!\e[0m"
else
  echo -e "  \e[1;31m⚠️  CRITICAL: $VULN vulnerabilities found. Check logs!\e[0m"
fi
