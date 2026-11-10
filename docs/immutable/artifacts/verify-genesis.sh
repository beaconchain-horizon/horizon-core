#!/usr/bin/env bash
# verify-genesis.sh — مستقل، بدون نیاز به سرویس‌های در حال اجرا
# بررسی می‌کند GENESIS-public.json با GENESIS.json محلی مطابقت دارد
set -eu
cd "$(dirname "$0")"

RED=$'\e[31m'; GREEN=$'\e[32m'; YELLOW=$'\e[33m'; RESET=$'\e[0m'

echo "=== Horizon Core v1.4 — Genesis Public Verification ==="

# ۱) اثر انگشت فایل آرتیفکت
EXPECTED=$(awk '{print $1}' GENESIS-public.json.sha256)
ACTUAL=$(sha256sum GENESIS-public.json | awk '{print $1}')
if [ "$EXPECTED" = "$ACTUAL" ]; then
  echo "${GREEN}[OK]${RESET}   GENESIS-public.json checksum matches"
else
  echo "${RED}[FAIL]${RESET} checksum mismatch: expected=$EXPECTED got=$ACTUAL"
  exit 1
fi

# ۲) استخراج مقادیر و مقایسه با GENESIS.json در ریشهٔ پروژه
GENESIS_SRC="$(cd ../../../ && pwd)/GENESIS.json"
if [ ! -f "$GENESIS_SRC" ]; then
  echo "${YELLOW}[WARN]${RESET} GENESIS.json not found at $GENESIS_SRC — skipping cross-check"
  exit 0
fi

PUB_FROM_ARTIFACT=$(grep -o '"authority_pubkey_hex": *"[^"]*"' GENESIS-public.json | sed 's/.*"\([^"]*\)"$/\1/')
PUB_FROM_SOURCE=$(grep -o '"authority_pubkey"[^,]*' "$GENESIS_SRC" | grep -o '[0-9a-fA-F]\{130\}' | head -1)

if [ "$PUB_FROM_ARTIFACT" = "$PUB_FROM_SOURCE" ]; then
  echo "${GREEN}[OK]${RESET}   authority_pubkey matches GENESIS.json"
else
  echo "${RED}[FAIL]${RESET} authority_pubkey mismatch"
  echo "        artifact: $PUB_FROM_ARTIFACT"
  echo "        source:   $PUB_FROM_SOURCE"
  exit 1
fi

GEN_FROM_ARTIFACT=$(grep -o '"genesis_hash": *"[^"]*"' GENESIS-public.json | sed 's/.*"\([^"]*\)"$/\1/')
GEN_FROM_SOURCE=$(grep -o '"genesis_hash"[^,]*' "$GENESIS_SRC" | grep -o '[0-9a-fA-F]\{64\}' | head -1)

if [ "$GEN_FROM_ARTIFACT" = "$GEN_FROM_SOURCE" ]; then
  echo "${GREEN}[OK]${RESET}   genesis_hash matches GENESIS.json"
else
  echo "${RED}[FAIL]${RESET} genesis_hash mismatch"
  exit 1
fi

echo ""
echo "${GREEN}=== ALL CHECKS PASSED ===${RESET}"
echo "Protocol artifacts verified. See ../HORIZON-CORE-v1.4-CANONICAL.md §1"
