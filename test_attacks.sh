#!/usr/bin/env bash
# ==========================================
# Horizon Core - SECURITY ATTACK SUITE
# ==========================================
BE="http://localhost:8080"
TOK="12345678901234567890123456789012"
OK=0; FAIL=0

echo "=========================================="
echo "  🛡️  HORIZON CORE - SECURITY TEST"
echo "=========================================="

# تابع بررسی امنیت
check() {
  local desc="$1"; shift
  local code=$(curl -s -o /dev/null -w "%{http_code}" "$@")
  if [ "$code" = "200" ] || [ "$code" = "201" ]; then
    echo -e "  ❌ \033[1;31m[VULNERABLE]\033[0m $desc (Got HTTP $code)"
    FAIL=$((FAIL+1))
  else
    echo -e "  ✅ \033[1;32m[BLOCKED]\033[0m    $desc (Got HTTP $code)"
    OK=$((OK+1))
  fi
}

# ۱. تست تزریق SQL (SQL Injection)
echo -e "\n[*] Testing SQL Injection..."
check "SQLi in license verify" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/json" -d '{"license_id":"1 OR 1=1","user_id":"x","signature":"x","data":"x"}'
check "SQLi in register" -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "X-Admin-Key: $TOK" -d '{"license_id":"x'"'"'; DROP TABLE--","user_id":"x","product_id":"x","volume":1,"signature":"x","signed_data":"x","public_key":"x"}'

# ۲. تست XSS (Cross-Site Scripting)
echo -e "\n[*] Testing XSS..."
check "XSS in query params" "$BE/api/v1/licenses?q=<script>alert(1)</script>"
check "XSS in packages" -X POST "$BE/api/v1/packages" -H "Content-Type: application/json" -d '{"name":"<img src=x onerror=alert(1)>"}'

# ۳. تست Bypass احراز هویت (Auth Bypass)
echo -e "\n[*] Testing Authentication Bypass..."
check "No Auth Token" "$BE/api/v1/licenses"
check "Fake Auth Token" -H "X-Admin-Key: fake_token_123" "$BE/api/v1/licenses"
check "JWT None Algorithm" -H "Authorization: Bearer eyJhbGciOiJub25lIn0.eyJhZG1pbiI6dHJ1ZX0." "$BE/api/v1/licenses"

# ۴. تست SSRF (دسترسی به متادیتای سرور)
echo -e "\n[*] Testing SSRF..."
check "SSRF to AWS Metadata" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/json" -d '{"data":"http://169.254.169.254/latest/meta-data"}'
check "SSRF to localhost" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/json" -d '{"data":"http://localhost:22"}'

# ۵. تست Path Traversal (دسترسی به فایل‌های سیستم)
echo -e "\n[*] Testing Path Traversal..."
check "Path Traversal to /etc/passwd" "$BE/api/v1/blocks/../../../../etc/passwd"
check "URL-Encoded Path Traversal" "$BE/api/v1/blocks/%2e%2e%2f%2e%2e%2fetc%2fpasswd"

# ۶. تست Buffer Overflow (ارسال حجم زیاد)
echo -e "\n[*] Testing Buffer Overflow (DoS) - This might take a second..."
head -c 9000000 /dev/zero | tr '\0' 'A' > /tmp/big.txt
printf '{"data":"' > /tmp/big.json
cat /tmp/big.txt >> /tmp/big.json
printf '"}' >> /tmp/big.json
check "Oversize Payload (9MB)" -X POST "$BE/api/v1/license/verify" -H "Content-Type: application/json" --data-binary @/tmp/big.json
rm -f /tmp/big.txt /tmp/big.json

echo -e "\n=========================================="
if [ "$FAIL" -eq 0 ]; then
  echo -e "  🏆 RESULT: \033[1;32mALL $OK ATTACKS BLOCKED\033[0m"
else
  echo -e "  ⚠️  RESULT: $OK Blocked | \033[1;31m$FAIL VULNERABLE\033[0m"
fi
echo "=========================================="
