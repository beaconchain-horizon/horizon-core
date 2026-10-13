#!/usr/bin/env bash
# Horizon Core - signing key rotation planner
set -euo pipefail
NEW_KEY="${1:-}"
if [ -z "$NEW_KEY" ]; then echo usage: $0 new-key.pem >&2; exit 2; fi
if [ ! -f "$NEW_KEY" ]; then echo key not found >&2; exit 2; fi
echo validating
head -1 "$NEW_KEY" | grep -q HORIZON || { echo not a HORIZON secp256k1 key >&2; exit 2; }
echo extracting public key
NEW_PUB=$(go run ./cmd/horizon-pubkey "$NEW_KEY" 2>/dev/null | tr -d "\r\n")
if [ -z "$NEW_PUB" ]; then echo failed >&2; exit 2; fi
echo
echo rotation plan:
echo  1. stop backend
echo  2. replace .horizon-dev-key.pem with $NEW_KEY
echo  3. export LICENSE_PRIVATE_KEY_PEM
echo  4. export HORIZON_LICENSE_PUBLIC_KEY=$NEW_PUB
echo  5. start backend
echo
echo NOTE: keep old key for 90-day overlap
