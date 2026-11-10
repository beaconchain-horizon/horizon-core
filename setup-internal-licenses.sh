#!/bin/bash
# این اسکریپت لایسنس‌های داخلی را (اگر وجود نداشته باشند) در سوییچ آفلاین می‌سازد
PORT=$1
TOKEN="bench0000000000000000000000000000000000000000000000000000000bench"

if [ -z "$PORT" ]; then
    echo "Usage: ./setup-internal-licenses.sh <port>"
    exit 1
fi

# صدور لایسنس برای API
curl -s -X POST http://localhost:$PORT/api/v1/admin/license/issue \
  -H "Content-Type: application/json" -H "X-Admin-Token: $TOKEN" \
  -d '{"license_id":"INTERNAL-API-LIC","user_id":"system-api","product_id":"horizon-core","volume":1000000000,"duration":3650,"license_type":"internal"}' > /dev/null

# صدور لایسنس برای نودها
curl -s -X POST http://localhost:$PORT/api/v1/admin/license/issue \
  -H "Content-Type: application/json" -H "X-Admin-Token: $TOKEN" \
  -d '{"license_id":"INTERNAL-NODE-LIC","user_id":"system-nodes","product_id":"horizon-core","volume":1000000000,"duration":3650,"license_type":"internal"}' > /dev/null

echo "✅ لایسنس‌های داخلی (API و Nodes) صادر شدند (هر کدام ۱ میلیارد حجم)."
