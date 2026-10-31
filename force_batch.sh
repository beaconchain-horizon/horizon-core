#!/usr/bin/env bash
echo "1. Obtaining session token..."
SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session -H "Content-Type: application/json" -H "X-Admin-Key: 12345678901234567890123456789012" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$SESS" ]; then
    echo "❌ Failed to get session token. Is the switch running?"
    exit 1
fi
echo "   Session: ${SESS:0:16}..."

echo "2. Sending a test transaction to trigger batch creation..."
curl -s -X POST http://localhost:8091/api/v1/tx/batch \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: 12345678901234567890123456789012" \
  -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1}]}"
echo ""

echo "3. Waiting for batcher (2 seconds)..."
sleep 2

echo "4. Checking Chain Height again..."
HEIGHT=$(curl -s http://localhost:8080/api/v1/chain/info | grep -o '"height":[0-9]*' | cut -d: -f2)
echo "   Current Height: $HEIGHT"

if [ "$HEIGHT" -ge 5 ]; then
    echo "✅ SUCCESS: Block 5 (NICICO) has been created!"
else
    echo "⚠️ Height is still $HEIGHT. Trying one more batch..."
    SESS=$(curl -s -X POST http://localhost:8091/api/v1/signing/session -H "Content-Type: application/json" -H "X-Admin-Key: 12345678901234567890123456789012" -d '{}' | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
    curl -s -X POST http://localhost:8091/api/v1/tx/batch -H "Content-Type: application/json" -H "X-Admin-Key: 12345678901234567890123456789012" -d "{\"session_token\":\"$SESS\",\"txs\":[{\"from\":\"bank_markazi\",\"to\":\"bank_melli\",\"amount\":1},{\"from\":\"bank_melli\",\"to\":\"bank_markazi\",\"amount\":1}]}"
    sleep 2
    HEIGHT=$(curl -s http://localhost:8080/api/v1/chain/info | grep -o '"height":[0-9]*' | cut -d: -f2)
    echo "   Height after retry: $HEIGHT"
fi
