#!/bin/bash
echo "Starting 9 Services (1 Main API + 1 Switch + 7 Nodes)..."
# اجرای ۷ نود
for i in {1..7}; do
    PORT=$((9000 + i))
    # اجرای نود در پس‌زمینه با پورت اختصاصی
    NODE_ID="node-$i" PORT=$PORT ./bin/api.exe > logs/node-$i.log 2>&1 &
done
# اجرای سوییچ (اگر وجود داشت)
if [ -f "./horizon-bank-package/switch.exe" ]; then
    PORT=9000 ./horizon-bank-package/switch.exe > logs/switch.log 2>&1 &
fi
echo "All 9 services started in background."
