#!/usr/bin/env bash
set -u
R=$'\e[31m'; G=$'\e[32m'; Y=$'\e[33m'; C=$'\e[36m'; X=$'\e[0m'
ok()   { printf "%s[OK]%s   %s\n" "$G" "$X" "$1"; }
fail() { printf "%s[FAIL]%s %s\n" "$R" "$X" "$1"; }
info() { printf "%s[..]%s   %s\n" "$C" "$X" "$1"; }
sec()  { printf "\n%s═══ %s ═══%s\n" "$C" "$1" "$X"; }

cd "$(dirname "$0")"
# گرفتن مسیر ویندوز به شکل C:\Users\...
WIN_ROOT=$(pwd -W | sed 's|/|\\|g')
mkdir -p logs

mk_bat() {
  local file="$1"; shift
  : > "$file"
  for line in "$@"; do
    printf '%s\r\n' "$line" >> "$file"
  done
}

# ═══ ۰) توقف پردازه‌های قبلی ═══
sec "0. توقف سرویس‌های قبلی"
powershell.exe -NoProfile -Command "Get-Process horizon-switch,horizon-angel,horizon-agent,horizon-api,brain -ErrorAction SilentlyContinue | Stop-Process -Force" 2>/dev/null || true
sleep 2
ok "متوقف شدند"

# ═══ ۱) ساخت launcher ها ═══
sec "1. ساخت launcher ها (با مسیر مطلق ویندوز)"
ADMIN_TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"
PUBKEY="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
DSN="host=localhost port=5432 user=postgres password=postgres dbname=horizon_switch sslmode=disable"

# --- Switch ---
mk_bat run-switch.bat \
  '@echo off' \
  "cd /d \"$WIN_ROOT\"" \
  'if not exist logs mkdir logs' \
  'set "DB_ENGINE=postgres"' \
  "set \"DB_DSN=$DSN\"" \
  "set \"HORIZON_LICENSE_PUBLIC_KEY=$PUBKEY\"" \
  "set \"ADMIN_TOKEN=$ADMIN_TOKEN\"" \
  'set "SWITCH_PORT=8091"' \
  'horizon-switch.exe > logs\switch.log 2>&1'
ok "run-switch.bat"

# --- API (اضافه کردن توکن برای رفع خطای CN_TOKEN) ---
mk_bat run-api.bat \
  '@echo off' \
  "cd /d \"$WIN_ROOT\"" \
  'if not exist logs mkdir logs' \
  'set "PORT=8080"' \
  'set "DB_ENGINE=postgres"' \
  "set \"DB_DSN=$DSN\"" \
  "set \"HORIZON_LICENSE_PUBLIC_KEY=$PUBKEY\"" \
  "set \"CN_TOKEN=$ADMIN_TOKEN\"" \
  "set \"API_KEY=$ADMIN_TOKEN\"" \
  'horizon-api.exe > logs\api.log 2>&1'
ok "run-api.bat"

# --- 8 Angels ---
for pair in "Merkle:9001:horizon-angel" "Hash:9002:horizon-angel" "ECDSA:9003:horizon-angel" "Type:9004:horizon-angel" "Volume:9005:horizon-angel" "Time:9006:horizon-angel" "User:9007:horizon-angel" "AgentGuardian:9008:horizon-agent"; do
  role="${pair%%:*}"
  rest="${pair#*:}"
  port="${rest%%:*}"
  exe="${rest##*:}"
  
  mk_bat "run-angel-$port.bat" \
    '@echo off' \
    "cd /d \"$WIN_ROOT\"" \
    'if not exist logs mkdir logs' \
    "set \"NODE_ROLE=$role\"" \
    "set \"NODE_PORT=$port\"" \
    "$exe.exe > logs\\angel-$port.log 2>&1"
  ok "run-angel-$port.bat ($role)"
done

# --- Brain ---
mk_bat run-brain.bat \
  '@echo off' \
  "cd /d \"$WIN_ROOT\"" \
  'if not exist logs mkdir logs' \
  'brain.exe > logs\brain.log 2>&1'
ok "run-brain.bat"

# ═══ ۲) اجرای سرویس‌ها ═══
sec "2. اجرای سوئیچ (8091)"
cmd //c start "" /B "run-switch.bat" >/dev/null 2>&1
sleep 3
if curl -s --max-time 2 http://localhost:8091/health 2>/dev/null | grep -q "horizon-switch"; then
  ok "سوئیچ روشن شد ✅"
else
  fail "سوئیچ روشن نشد"
fi

sec "3. اجرای API (8080)"
cmd //c start "" /B "run-api.bat" >/dev/null 2>&1
sleep 2
if curl -s --max-time 2 http://localhost:8080/api/v1/health 2>/dev/null | grep -q "status"; then
  ok "API روشن شد ✅"
else
  info "API پاسخ نداد (لاگ را چک کنید)"
fi

sec "4. اجرای ۸ آنجل"
for port in 9001 9002 9003 9004 9005 9006 9007 9008; do
  cmd //c start "" /B "run-angel-$port.bat" >/dev/null 2>&1
done
sleep 4

sec "5. اجرای Brain (9300)"
cmd //c start "" /B "run-brain.bat" >/dev/null 2>&1
sleep 2

# ═══ ۳) تأیید نهایی ═══
sec "6. تأیید سلامت همه سرویس‌ها"
FAILED=0

printf "  :8091 switch         "
H=$(curl -s --max-time 2 http://127.0.0.1:8091/health 2>/dev/null)
if echo "$H" | grep -q "horizon-switch"; then ok "✅"; else fail "❌"; FAILED=$((FAILED+1)); fi

printf "  :8080 api            "
H=$(curl -s --max-time 2 http://127.0.0.1:8080/api/v1/health 2>/dev/null)
if [ -n "$H" ]; then ok "✅"; else fail "❌"; FAILED=$((FAILED+1)); fi

for port in 9001 9002 9003 9004 9005 9006 9007 9008; do
  printf "  :%s                  " "$port"
  H=$(curl -s --max-time 2 "http://127.0.0.1:$port/health" 2>/dev/null)
  if echo "$H" | grep -q "role"; then ok "✅"; else fail "❌"; FAILED=$((FAILED+1)); fi
done

printf "  :9300 brain          "
H=$(curl -s --max-time 2 http://127.0.0.1:9300/health 2>/dev/null)
if [ -n "$H" ]; then ok "✅"; else info "—"; fi

sec "خلاصه"
if [ "$FAILED" -eq 0 ]; then
  printf "%s✅ همه سرویس‌ها بالا آمدند%s\n" "$G" "$X"
else
  printf "%s⚠️  %d سرویس مشکل دارد%s\n" "$Y" "$FAILED" "$X"
  echo "برای دیدن خطا: tail -20 logs/api.log"
  echo "برای دیدن خطا: tail -20 logs/switch.log"
fi

echo ""
read -r -p "برای بستن Enter بزنید..." _
