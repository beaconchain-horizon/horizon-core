#!/usr/bin/env bash
set -e

echo "========================================"
echo " Horizon Core - Commit & Push"
echo "========================================"

echo
echo "[1/5] Current status"
git status --short

echo
echo "[2/5] Removing temporary install files"
rm -rf .store-install-temp

echo
echo "[3/5] Staging ALL changes"
git add -A

echo
echo "[4/5] Creating commit"
git commit -m "feat: integrate Horizon Next.js Store frontend"

echo
echo "[5/5] Pushing to origin/main"
git push origin main

echo
echo "========================================"
echo " DONE"
echo "========================================"
git status
git log -1 --oneline
