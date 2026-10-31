#!/usr/bin/env bash
set -e

DIR="horizon-trial-nicico-real"
ZIP="horizon-trial-nicico-real.zip"

echo "1) Cleaning old files..."
rm -rf "$DIR" "$ZIP"
mkdir -p "$DIR/bin" "$DIR/config" "$DIR/static"

echo "2) Creating the REAL signed license (user_id: nicico)..."
cat > "$DIR/config/trial.lic" << 'LIC'
{
  "license_id": "HORIZON-GOLD-005",
  "user_id": "nicico",
  "product_id": "horizon-core",
  "volume": 1000000,
  "used": 0,
  "duration": 4,
  "merkle_root": "b395f34ddf5da832eb477f3a81bc1e7d03b0011795c8896f18bf6d2382a22c9a",
  "issued_at": 1793298061,
  "expires_at": 1793643661,
  "status": "active",
  "hardware_id": "",
  "signature": "91d98bafb73207445f40d0eb38c0ac66cec226e66e3e41f9b99d17f135476c58057169ab7a8b499f503c1e9e42cbdcbbd38244699c8c1a2abcd9ed255a6cb4d8",
  "signed_data": "{\"license_id\":\"HORIZON-GOLD-005\",\"user_id\":\"nicico\",\"product_id\":\"horizon-core\",\"volume\":1000000,\"duration\":4,\"merkle_root\":\"b395f34ddf5da832eb477f3a81bc1e7d03b0011795c8896f18bf6d2382a22c9a\",\"issued_at\":1793298061,\"expires_at\":1793643661,\"hardware_id\":\"\"}",
  "public_key": "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
}
LIC

echo "3) Building binaries (API + Switch, NO NODE)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/horizon-api.exe" ./cmd/api
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$DIR/bin/horizon-switch.exe" ./cmd/switch

echo "4) Copying static explorer files..."
cp -r cmd/api/static/* "$DIR/static/" 2>/dev/null || echo "WARNING: static folder not found, explorer might not load."

echo "5) Creating start.bat (Auto-registers license)..."
cat > "$DIR/start.bat" << 'BAT'
@echo off
echo Starting Horizon Core Trial (Real License)...
set HORIZON_TRIAL_MODE=true
set ADMIN_TOKEN=12345678901234567890123456789012
set HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set SERVER_PORT=8080
set SWITCH_PORT=8091

start "Horizon-API" bin\horizon-api.exe
start "Horizon-Switch" bin\horizon-switch.exe

echo Waiting for services to start...
timeout /t 5 /nobreak >nul

echo Registering REAL license to Backend...
curl.exe -s -X POST http://localhost:8080/api/v1/license/register -H "Content-Type: application/json" -H "X-Admin-Key: %ADMIN_TOKEN%" -d @config/trial.lic
echo.
echo System Ready! Open http://localhost:8080 in your browser.
pause
BAT

echo "6) Creating README..."
cat > "$DIR/README.md" << 'README'
# Horizon Core - Trial Edition (Real License)

This package contains a valid, ECDSA-signed license for **1,000,000 transactions** and **4 days**.

## How to run:
1. Double-click `start.bat`.
2. Open your browser at `http://localhost:8080`.
3. You will see the live blockchain explorer with the genesis block and your license already registered.

## Features:
- API (Backend) on port 8080
- Switch on port 8091
- Fully offline (Air-Gap ready)
- Real signature verification

*Trust, Verify, Move Forward.*
README

echo "7) Zipping with PowerShell..."
powershell -Command "Compress-Archive -Path '$DIR' -DestinationPath '$ZIP' -Force"

echo "8) Calculating SHA-256..."
HASH=$(sha256sum "$ZIP" | cut -d' ' -f1)
echo "✅ Done! File: $ZIP"
echo "🔑 SHA-256: $HASH"
echo ""
echo "You can now upload this ZIP to GitHub Releases."
