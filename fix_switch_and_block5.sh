#!/usr/bin/env bash
echo "1. Checking Switch health on port 8091..."
if ! curl -s --max-time 2 http://localhost:8091/api/v1/health | grep -q "online"; then
    echo "⚠️ Switch is down. Restarting it in background..."
    # اجرای سوییچ در پس‌زمینه
    export ADMIN_TOKEN="12345678901234567890123456789012"
    export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
    export SWITCH_PORT=8091
    nohup go run ./cmd/switch > switch_restart.log 2>&1 &
    echo "   Waiting 3 seconds for switch to boot..."
    sleep 3
else
    echo "✅ Switch is running."
fi

echo "2. Getting session token with correct header (X-Admin-Token)..."
SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS" ]; then
    echo "❌ Still failed to get session. Trying with X-Customer-Token..."
    SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session \
      -H "Content-Type: application/json" \
      -H "X-Customer-Token: 12345678901234567890123456789012" \
      -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$SESS" ]; then
    echo "❌ FATAL: Could not get session token. Check switch logs."
    exit 1
fi
echo "   Session: ${SESS:0:16}..."

echo "3. Sending test transaction to trigger batch creation..."
curl -s -X POST http://localhost:8091/api/v1/tx/batch \
  -H "Content-Type: application/json" \
  -H "X-Admin-Token: 12345678901234567890123456789012" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

echo "4. Waiting for batcher (3 seconds)..."
sleep 3

echo "5. Checking Chain Height again..."
HEIGHT=$(curl -s http://localhost:8080/api/v1/chain/info | grep -o '"height":[0-9]*' | cut -d: -f2)
echo "   Current Height: $HEIGHT"

if [ "$HEIGHT" -ge 5 ]; then
    echo "✅ SUCCESS: Block 5 (NICICO) has been created!"
else
    echo "⚠️ Height is $HEIGHT. The batch might need more time or the switch needs checking."
fi
