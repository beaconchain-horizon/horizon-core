#!/bin/bash
# ═══════════════════════════════════════════════
#  Horizon Core — Benchmark (inside container)
# ═══════════════════════════════════════════════
set -uo pipefail

echo ""
echo "════════════════════════════════════════════" 
echo "  Horizon Core — Benchmark"
echo "  Go 1.24 · Alpine · Localhost only"
echo "════════════════════════════════════════════"
echo ""

# ---------- ۱. تولید کلید ----------
echo "[1/5] Generating fresh ECDSA P-256 signing key..."
./bin/sensortool genkey --out=/tmp/test-keys 2>&1 | head -2
echo ""

# ---------- ۲. آماده‌سازی دیتابیس ----------
echo "[2/5] Preparing fresh database..."
rm -f /tmp/bench.db
echo ""

# ---------- ۳. اجرای سرور ----------
echo "[3/5] Starting switch server (127.0.0.1:8080)..."
export SWITCH_DB=/tmp/bench.db
export CHAIN_CONFIG=/app/config/chain.json
export ADMIN_TOKEN=$(head -c 32 /dev/urandom | base64 | tr -d "/+=" | head -c 32)

./bin/switch > /tmp/switch.log 2>&1 &
SERVER_PID=$!
sleep 3

# بررسی اینکه سرور بالا اومده
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "ERROR: Server failed to start"
    cat /tmp/switch.log
    exit 1
fi
echo "  Server PID: $SERVER_PID"
echo "  Server listening on 127.0.0.1:8080"
echo ""

# ---------- ۴. ساخت Site و Sensor ----------
echo "[4/5] Creating test site and sensor..."
API="http://127.0.0.1:8080/api/v1"
TOK="X-Admin-Token: $ADMIN_TOKEN"

curl -s -X POST "$API/industrial/sites" -H "$TOK" -H "Content-Type: application/json" \
  -d "{\"site_id\":\"bench\",\"name\":\"Benchmark\",\"type\":\"test\"}" > /dev/null

PUB=$(cat /tmp/test-keys/public.pem | sed ":a;N;$!ba;s/\n/\\n/g")
curl -s -X POST "$API/industrial/sensors" -H "$TOK" -H "Content-Type: application/json" \
  -d "{\"sensor_id\":\"bench-001\",\"site_id\":\"bench\",\"name\":\"Bench\",\"type\":\"temperature\",\"unit\":\"C\",\"min_value\":0,\"max_value\":100,\"public_key\":\"$PUB\"}" > /dev/null

echo "  Site: bench"
echo "  Sensor: bench-001"
echo ""

# ---------- ۵. اجرای Benchmark ----------
echo "[5/5] Running TPS benchmark..."
echo ""
echo "────────────────────────────────────────────"
echo "────────────────────────────────────────────"
./bin/tpsbench \
  -url=$API/industrial/reading/batch \
  -key=/tmp/test-keys/private.pem \
  -sensor=bench-001 \
  -n=5000 \
  -c=50 \
  -bs=100

echo ""
echo "────────────────────────────────────────────"
echo "  TEST 2: Batch Size 500"
echo "────────────────────────────────────────────"
./bin/tpsbench \
  -url=$API/industrial/reading/batch \
  -key=/tmp/test-keys/private.pem \
  -sensor=bench-001 \
  -n=5000 \
  -c=50 \
  -bs=500

echo ""
echo "────────────────────────────────────────────"
echo "  INTEGRITY CHECK"
echo "────────────────────────────────────────────"
curl -s "$API/industrial/dashboard" -H "$TOK"
echo ""

# ---------- پاک‌سازی ----------
kill $SERVER_PID 2>/dev/null
echo ""
echo "════════════════════════════════════════════"
echo "  Benchmark complete"
echo "════════════════════════════════════════════"
