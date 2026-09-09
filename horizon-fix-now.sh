#!/bin/bash
set -e

cd ~/Beaconchain/horizon-core

echo "=========================================="
echo "   Horizon Angel - Quick Fix"
echo "=========================================="

# ------------------------------------------------------------
# 1. FIX go.mod VERSION (change 1.26 -> 1.24)
# ------------------------------------------------------------
echo ">> Fixing go.mod version..."
sed -i 's/go 1.26/go 1.24/' go.mod

# Also ensure go.sum exists
touch go.sum

# ------------------------------------------------------------
# 2. COMMIT AND PUSH
# ------------------------------------------------------------
echo ">> Pushing changes to GitHub..."
git add go.mod go.sum
git commit -m "fix: downgrade go version to 1.24 for Liara"
git pull origin main --rebase
git push origin main

# ------------------------------------------------------------
# 3. DEPLOY SWITCH (Docker)

cd ~/Beaconchain/horizon-core

# 1. ایجاد Dockerfile استاندارد (اگر وجود ندارد یا خالی است)
cat > Dockerfile << 'EOF'
FROM golang:1.24-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o switch ./cmd/switch
EXPOSE 8080
CMD ["./switch"]
