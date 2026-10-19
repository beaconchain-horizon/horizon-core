#!/bin/bash

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

echo "=============================================="
echo " Horizon Core v3.0 - Prepare & Run"
echo "=============================================="

echo "[1/7] Checking Go..."
go version

echo
echo "[2/7] Setting Go module to 1.24.0..."
go mod edit -go=1.24.0

if ! grep -q '^go 1.24.0$' go.mod; then
    echo "ERROR: go.mod was not set to Go 1.24.0"
    exit 1
fi

echo "OK: $(grep '^go ' go.mod)"

echo
echo "[3/7] Checking license public key..."

export HORIZON_LICENSE_PUBLIC_KEY="048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a"

KEY_LEN="${#HORIZON_LICENSE_PUBLIC_KEY}"

if [ "$KEY_LEN" -ne 130 ]; then
    echo "ERROR: License public key length is $KEY_LEN"
    echo "Expected: 130 hex characters"
    exit 1
fi

echo "OK: secp256k1 public key configured (${KEY_LEN} hex chars)"

echo
echo "[4/7] Checking license key implementation..."

if ! grep -q 'HORIZON_LICENSE_PUBLIC_KEY' cmd/switch/license_enforce.go; then
    echo "ERROR: license_enforce.go does not reference HORIZON_LICENSE_PUBLIC_KEY"
    exit 1
fi

echo "OK: license enforcement code found"

echo
echo "[5/7] Building Horizon Switch..."

go build -o switch.exe ./cmd/switch

if [ ! -f switch.exe ]; then
    echo "ERROR: switch.exe was not created"
    exit 1
fi

echo "OK: switch.exe built"

echo
echo "[6/7] Stopping old Switch processes..."

taskkill //F //IM switch.exe 2>/dev/null || true
sleep 2

echo
echo "[7/7] Starting Horizon Switch v3.0..."
echo "=============================================="
echo "Port: 8080"
echo "Go module: 1.24.0"
echo "License key: configured"
echo "=============================================="

export SWITCH_PORT=8080
export SWITCH_DB="$ROOT/data/horizon-switch.db"
export CHAIN_CONFIG="$ROOT/config/chain.json"
export CUSTOMERS_CONFIG="$ROOT/config/customers.json"

export BACKEND_URL="http://localhost:8080"
export CORE_ENGINE_URL="http://localhost:8080"
export LICENSE_SERVER_URL="http://localhost:8080"

# Test credentials only.
# Replace/rotate before production.
export ADMIN_TOKEN="REPLACE_WITH_ADMIN_TOKEN"
export ADMIN_PASSWORD_HASH='$2a$10$BOJ9T6ZC9fbG5n1y4yeLDORlccFh3iLEZJ.axiW7IG.tSWep2nBd6'

echo
printf 'HORIZON_LICENSE_PUBLIC_KEY length: %s\n' "${#HORIZON_LICENSE_PUBLIC_KEY}"
printf 'SWITCH_DB: %s\n' "$SWITCH_DB"
printf 'SWITCH_PORT: %s\n' "$SWITCH_PORT"
echo

exec ./switch.exe
