#!/bin/bash

# ============================================================
# Horizon Angel - Intelligent Self-Healing Deployment Script
# ============================================================

set -e  # Stop on error

echo ""
echo "=========================================="
echo "   Horizon Angel - Full Auto Deployment"
echo "=========================================="

cd ~/Beaconchain/horizon-core

# ------------------------------------------------------------
# 1. FIX go.mod (remove duplicate lines, set version 1.24)
# ------------------------------------------------------------
echo ">> Checking and fixing go.mod..."
if [ -f go.mod ]; then
    # Remove duplicate 'go' lines and set correct version
    sed -i '/^go /d' go.mod
    echo "go 1.24" > go.mod.tmp
    # Keep only first module line and dependencies
    head -1 go.mod > go.mod.clean
    echo "" >> go.mod.clean
    echo "go 1.24" >> go.mod.clean
    echo "" >> go.mod.clean
    grep -E '^require|^\t|^\)' go.mod | head -n -1 >> go.mod.clean 2>/dev/null || true
    mv go.mod.clean go.mod
    # Ensure basic dependencies
    if ! grep -q "github.com/gin-gonic/gin" go.mod; then
        echo "require (" >> go.mod
        echo "    github.com/gin-contrib/cors v1.4.0" >> go.mod
        echo "    github.com/gin-gonic/gin v1.9.1" >> go.mod
        echo "    gorm.io/driver/postgres v1.5.2" >> go.mod
        echo "    gorm.io/gorm v1.25.4" >> go.mod
        echo ")" >> go.mod
    fi
    echo "✅ go.mod fixed."
else
    echo "go.mod not found, creating..."
    cat > go.mod << 'GOMOD'
module horizon-core

go 1.24

require (
    github.com/gin-contrib/cors v1.4.0
    github.com/gin-gonic/gin v1.9.1
    gorm.io/driver/postgres v1.5.2
    gorm.io/gorm v1.25.4
)
GOMOD
fi

# ------------------------------------------------------------
# 2. CREATE liara.json to force Go version
# ------------------------------------------------------------
echo ">> Creating liara.json..."
cat > liara.json << 'LIARA'
{
  "version": "1.24",
  "platform": "go"
}
LIARA

# ------------------------------------------------------------
# 3. CREATE or OVERWRITE Dockerfile for Switch
# ------------------------------------------------------------
echo ">> Creating Dockerfile for Switch..."
cat > Dockerfile << 'DOCKER'
FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o switch ./cmd/switch
EXPOSE 8080
CMD ["./switch"]
DOCKER

# ------------------------------------------------------------
# 4. PUSH TO GITHUB (with retry if rejected)
# ------------------------------------------------------------
echo ">> Pushing to GitHub..."
git add go.mod liara.json Dockerfile
git commit -m "fix: auto-correct go.mod and add Dockerfile for switch" || echo "No new changes to commit"
git pull origin main --rebase || true
git push origin main

# ------------------------------------------------------------
# 5. DEPLOY SWITCH (via Docker)
# ------------------------------------------------------------
echo ">> Deploying Switch to Liara..."
if liara deploy --app horizon-switch --platform docker --port 8080 --path .; then
    echo "✅ Switch deployed successfully."
else
    echo "⚠️ Switch deployment failed. Trying with go platform..."
    liara deploy --app horizon-switch --platform go --port 8080 --path ./cmd/switch
fi

# ------------------------------------------------------------
# 6. DEPLOY BACKEND
# ------------------------------------------------------------
echo ">> Deploying Backend to Liara..."
if liara deploy --app horizon-backend --platform go --port 8080 --path ./cmd/api; then
    echo "✅ Backend deployed successfully."
else
    echo "⚠️ Backend deployment failed. Trying with Docker..."
    liara deploy --app horizon-backend --platform docker --port 8080 --path .
fi

# ------------------------------------------------------------
# 7. DEPLOY FRONTEND (static)
# ------------------------------------------------------------
echo ">> Deploying Frontend to Liara..."
if liara deploy --app horizon-frontend --static --path ./frontend; then
    echo "✅ Frontend deployed successfully."
else
    echo "⚠️ Frontend deployment failed. Please check the 'frontend' folder."
fi

# ------------------------------------------------------------
# 8. TEST ALL SERVICES
# ------------------------------------------------------------
echo ""
echo "=========================================="
echo ">> Testing services..."

echo -n "Backend: "
curl -s -o /dev/null -w "%{http_code}" https://horizon-backend.liara.run/api/v1/health | grep -q 200 && echo " ✅ OK" || echo " ❌ FAILED"

echo -n "Switch:  "
curl -s -o /dev/null -w "%{http_code}" https://horizon-switch.liara.run/health | grep -q 200 && echo " ✅ OK" || echo " ❌ FAILED"

echo -n "Frontend:"
curl -s -o /dev/null -w "%{http_code}" https://horizon-frontend.liara.run | grep -q 200 && echo " ✅ OK" || echo " ❌ FAILED"

# ------------------------------------------------------------
# 9. FINAL SUMMARY
# ------------------------------------------------------------
echo ""
echo "=========================================="
echo "✅ Deployment process completed!"
echo ""
echo "🌐 Frontend:  https://horizon-frontend.liara.run"
echo "🔗 Backend:   https://horizon-backend.liara.run"
echo "🔄 Switch:    https://horizon-switch.liara.run"
echo ""
echo "If any service shows ❌, check logs with:"
echo "  liara logs --app <app-name>"
echo "=========================================="
