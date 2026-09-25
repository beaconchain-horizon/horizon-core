#!/usr/bin/env bash
set -u

REPO="$HOME/Beaconchain/horizon-core"
FRONTEND="$REPO/frontend"

echo
echo "=============================================="
echo " HORIZON STORE - CURRENT FRONTEND TEST"
echo " NO COMMIT / NO PUSH"
echo "=============================================="
echo

cd "$REPO" || exit 1

echo "[1/7] Checking frontend..."

if [ ! -f "$FRONTEND/package.json" ]; then
  echo "ERROR: frontend/package.json not found."
  exit 1
fi

echo "OK: frontend/package.json"
echo

echo "[2/7] Checking Next.js..."

if ! grep -q '"next"' "$FRONTEND/package.json"; then
  echo "ERROR: Next.js dependency not found."
  exit 1
fi

echo "OK: Next.js detected."
echo

echo "[3/7] Checking project structure..."

for ITEM in \
  "$FRONTEND/app" \
  "$FRONTEND/components" \
  "$FRONTEND/lib" \
  "$FRONTEND/next.config.ts" \
  "$FRONTEND/tsconfig.json"
do
  if [ ! -e "$ITEM" ]; then
    echo "ERROR: Missing:"
    echo "$ITEM"
    exit 1
  fi
done

echo "OK: required Next.js structure found."
echo

echo "[4/7] Removing temporary/accidental files..."

rm -rf "$REPO/.store-install-temp"
rm -rf "$REPO/.store-test-temp"
rm -rf "$REPO/.frontend-store-test"
rm -rf "$FRONTEND/.store-install-temp"

# ZIPs accidentally left in repository
find "$REPO" -maxdepth 2 -type f \
  \( -name "horizon-store-next-ready*.zip" -o -name "*.zip" \) \
  -print -delete 2>/dev/null || true

echo "Cleanup complete."
echo

echo "[5/7] Installing frontend dependencies..."

cd "$FRONTEND" || exit 1

npm install

if [ $? -ne 0 ]; then
  echo
  echo "=============================================="
  echo "NPM INSTALL: FAIL"
  echo "=============================================="
  exit 1
fi

echo
echo "NPM INSTALL: PASS"
echo

echo "[6/7] Running lint..."

npm run lint

if [ $? -ne 0 ]; then
  echo
  echo "=============================================="
  echo "LINT: FAIL"
  echo "=============================================="
  exit 1
fi

echo
echo "LINT: PASS"
echo

echo "[7/7] Running production build..."

npm run build

if [ $? -ne 0 ]; then
  echo
  echo "=============================================="
  echo "BUILD: FAIL"
  echo "=============================================="
  exit 1
fi

cd "$REPO" || exit 1

echo
echo "=============================================="
echo " ALL TESTS PASSED"
echo "=============================================="
echo
echo "frontend/package.json : OK"
echo "Next.js               : OK"
echo "npm install           : PASS"
echo "npm run lint          : PASS"
echo "npm run build         : PASS"
echo
echo "NO COMMIT"
echo "NO PUSH"
echo
echo "Git status:"
git status --short
echo
echo "=============================================="
echo " READY FOR FINAL REVIEW"
echo "=============================================="

read -r -p "Press ENTER to close..."
