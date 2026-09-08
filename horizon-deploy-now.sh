#!/bin/bash
set -e

cd ~/Beaconchain/horizon-core

echo "=========================================="
echo "   Horizon Angel - Final Deploy"
echo "=========================================="

# ------------------------------------------------------------
# 1. REMOVE BROKEN FILES & FIX IMPORTS
# ------------------------------------------------------------
echo ">> Removing broken switchclient..."
rm -f internal/switchclient/client.go

echo ">> Removing unused import in webhook..."
sed -i '/"encoding\/json"/d' cmd/webhook/main.go

# ------------------------------------------------------------
# 2. CREATE DOCKERFILE FOR SWITCH (if empty or missing)
# ------------------------------------------------------------
echo ">> Creating Dockerfile for Switch..."
cat > Dockerfile << 'DOCKER'
FROM golang:1.24-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o switch ./cmd/switch
EXPOSE 8080
CMD ["./switch"]
DOCKER

# ------------------------------------------------------------
# 3. FIX go.mod (clean version)
# ------------------------------------------------------------
echo ">> Fixing go.mod..."
cat > go.mod << 'GOMOD'
module horizon-core

go 1.24

require (
    github.com/gin-contrib/cors v1.4.0
    github.com/gin-gonic/gin v1.9.1
    gorm.io/driver/postgres v1.5.2
    gorm.io/gorm v1.25.4
    github.com/glebarez/sqlite v1.8.0
)
GOMOD

touch go.sum

# ------------------------------------------------------------
# 4. CLEAN CACHE & BUILD
# ------------------------------------------------------------
echo ">> Cleaning cache and building..."
go clean -cache
go mod tidy
go build ./...
go test ./...

# ------------------------------------------------------------
# 5. COMMIT & PUSH
# ------------------------------------------------------------
echo ">> Committing and pushing to GitHub..."
git add .
git commit -m "fix: final cleanup, Dockerfile, go.mod and build fixes" || echo "No changes to commit"
git pull origin main --rebase
git push origin main

# ------------------------------------------------------------
# 6. DEPLOY BACKEND (Go platform)
# ------------------------------------------------------------
echo ">> Deploying Backend..."
liara deploy --app horizon-backend --platform go --port 8080 --path ./cmd/api || echo "Backend deploy failed"

# ------------------------------------------------------------
# 7. DEPLOY SWITCH (Docker)
# ------------------------------------------------------------
echo ">> Deploying Switch..."
liara deploy --app horizon-switch --platform docker --port 8080 --path . || echo "Switch deploy failed"

# ------------------------------------------------------------
# 8. DEPLOY FRONTEND (Static - via Liara web panel or CLI)
# ------------------------------------------------------------
echo ">> Deploying Frontend (Static)..."
# Try with --platform static
if liara deploy --app horizon-frontend --platform static --path ./frontend; then
    echo "✅ Frontend deployed via --platform static"
else
    echo "⚠️ --platform static failed. Trying with --static..."
    liara deploy --app horizon-frontend --static --path ./frontend || {
        echo "❌ CLI deploy failed. Please upload manually via Liara web panel:"
        echo "   1. Go to https://console.liara.ir"
        echo "   2. Create a new Static app named 'horizon-frontend'"
        echo "   3. Upload the contents of ./frontend folder"
    }
fi

# ------------------------------------------------------------
# 9. TEST SERVICES
# ------------------------------------------------------------
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
echo "✅ Deployment process completed."
echo "If frontend shows 404/502, deploy it manually via Liara web panel."
