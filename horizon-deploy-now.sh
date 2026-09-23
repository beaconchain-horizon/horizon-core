#!/bin/bash
set -e

cd ~/Beaconchain/horizon-core

echo "=========================================="
echo "   Horizon Angel - Deploy All Services"
echo "=========================================="

# 1. Deploy Backend
echo ">> Deploying Backend..."
liara deploy --app horizon-backend --platform go --port 8080 --path ./cmd/api

# 2. Deploy Switch (Docker)
echo ">> Deploying Switch..."
liara deploy --app horizon-switch --platform docker --port 8080 --path .

# 3. Deploy Frontend (Static) - Manual via Liara CLI (if --static works)
echo ">> Deploying Frontend..."
if liara deploy --app horizon-frontend --platform static --path ./frontend; then
    echo "✅ Frontend deployed via --platform static"
else
    echo "⚠️ --platform static failed, trying manual upload..."
    echo "Please upload the contents of ./frontend manually at:"
    echo "https://console.liara.ir/apps/horizon-frontend/deploy"
fi

# 4. Test all services
echo ""
echo "=========================================="
echo "Testing services..."
echo -n "Backend:  "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-backend.liara.run/api/v1/health
echo -n "Switch:   "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-switch.liara.run/health
echo -n "Frontend: "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-frontend.liara.run
echo "=========================================="
echo "✅ Done."
echo "🌐 Frontend: https://horizon-frontend.liara.run"
echo "🔗 Backend:  https://horizon-backend.liara.run"
echo "🔄 Switch:   https://horizon-switch.liara.run"
