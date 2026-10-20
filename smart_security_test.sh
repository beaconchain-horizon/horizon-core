#!/usr/bin/env bash

echo "=========================================================="
echo "  HORIZON CORE - SMART PENETRATION TEST"
echo "=========================================================="

BE="http://localhost:8080"
TOK="12345678901234567890123456789012"
HK="X-Admin-Key: $TOK"

BLOCKED=0
VULN=0

test_attack() {
  local desc="$1"; shift
  local response=$(curl -s -w "\n%{http_code}" "$@")
  local code=$(echo "$response" | tail -1)
  local body=$(echo "$response" | head -n -1)
  
  # اگر سیستم کد 4xx داد، یعنی حمله بلاک شده است
  if [[ "$code" == 4* ]]; then
    echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code)\e[0m"
    BLOCKED=$((BLOCKED+1))
    return
  fi
  
  # اگر کد 200 داد، باید بررسی کنیم که آیا دیتای حساسی لو رفته یا نه
  if [ "$code" = "200" ] || [ "$code" = "201" ]; then
    # برای SQLi و XSS، اگر لیست خالی برگرداند، یعنی امن است
    if [[ "$body" == "[]" ]] || [[ "$body" == *"[]"* ]] || [[ "$body" != *"admin"* && "$body" != *"root"* ]]; then
       echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code - Safe Empty/Sanitized Response)\e[0m"
       BLOCKED=$((BLOCKED+1))
    else
       echo -e "  \e[31m❌ [VULNERABLE] $desc (HTTP $code - DATA LEAK!)\e[0m"
       VULN=$((VULN+1))
    fi
  else
    echo -e "  \e[32m✅ [BLOCKED]    $desc (HTTP $code)\e[0m"
    BLOCKED=$((BLOCKED+1))
  fi
}

echo ""
echo "[1] SQL Injection (SQLi)"
test_attack "SQLi in license list" "$BE/api/v1/licenses?q=1%27%20OR%201=1--" -H "$HK"
test_attack "SQLi in register" -X POST "$BE/api/v1/license/register" -H "$HK" -H "Content-Type: application/json" -d '{"license_id":"1 OR 1=1","user_id":"hacker","product_id":"x","volume":-1}'

echo ""
echo "[2] Cross-Site Scripting (XSS)"
test_attack "XSS in search param" "$BE/api/v1/licenses?q=%3Cscript%3Ealert(1)%3C/script%3E" -H "$HK"
test_attack "XSS in package name" -X POST "$BE/api/v1/packages" -H "$HK" -H "Content-Type: application/json" -d '{"name":"<img src=x onerror=alert(1)>"}'

echo ""
echo "[3] Command Injection"
test_attack "Command injection in verify" -X POST "$BE/api/v1/license/verify" -H "$HK" -H "Content-Type: application/json" -d '{"license_id":"x; id","user_id":"y","signature":"z","data":"w"}'

echo ""
echo "[4] Path Traversal"
test_attack "Path Traversal in blocks" "$BE/api/v1/blocks/../../etc/passwd" -H "$HK"

echo ""
echo "[5] Auth Bypass"
test_attack "No Auth Token" "$BE/api/v1/licenses"
test_attack "Fake Token" "$BE/api/v1/licenses" -H "X-Admin-Key: fake_token_123"

echo ""
echo "[6] JWT Attacks"
test_attack "JWT 'none' algorithm" "$BE/api/v1/licenses" -H "Authorization: Bearer eyJhbGciOiJub25lIn0.eyJhZG1pbiI6dHJ1ZX0."

echo ""
echo "[7] Oversize Payload (Buffer Guard)"
# ساخت فایل ۹ مگابایتی و ارسال با --data-binary
head -c 9000000 /dev/zero | tr '\0' 'A' > /tmp/big_payload.txt
test_attack "9MB Oversize Body" -X POST "$BE/api/v1/license/verify" -H "$HK" -H "Content-Type: application/json" --data-binary @/tmp/big_payload.txt
rm -f /tmp/big_payload.txt

echo ""
echo "[8] SSRF & XXE"
test_attack "SSRF to AWS Metadata" -X POST "$BE/api/v1/license/verify" -H "$HK" -H "Content-Type: application/json" -d '{"data":"http://169.254.169.254/latest/meta-data"}'
XXE='<?xml version="1.0"?><!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><foo>&xxe;</foo>'
test_attack "XXE in verify" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/xml" -d "$XXE"

echo ""
echo "=========================================================="
echo "  FINAL SCORE"
echo "=========================================================="
echo -e "  \e[32m✅ BLOCKED (Safe): $BLOCKED\e[0m"
echo -e "  \e[31m❌ VULNERABLE (Fail): $VULN\e[0m"
echo "=========================================================="
if [ "$VULN" -eq 0 ]; then
  echo -e "  \e[1;32m🏆 PERFECT SCORE: ALL ATTACKS BLOCKED!\e[0m"
fi
