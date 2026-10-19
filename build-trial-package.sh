#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

VERSION="${1:-trial-$(date +%Y%m%d)}"
OUT="dist/beaconchain-horizon-${VERSION}"
rm -rf "$OUT"
mkdir -p "$OUT"

go build -o backend.exe ./cmd/api/
go build -o switch.exe ./cmd/switch/
go build -o angel-node.exe ./cmd/angel-node/
go build -o pentest.exe ./cmd/pentest/
echo "Building trial package: $VERSION"

for f in backend.exe switch.exe angel-node.exe pentest.exe; do
  if [ -f "$f" ]; then cp "$f" "$OUT/"; echo "  + $f"; fi
done

for f in start-all-silent.bat run-all.sh tps-report.sh; do
  if [ -f "$f" ]; then cp "$f" "$OUT/"; echo "  + $f"; fi
done

cat > "$OUT/START-TRIAL.bat" << 'BATEOF'
@echo off
cd /d "%~dp0"
set HORIZON_TRIAL_MODE=true
set HORIZON_TRIAL_DIR=%~dp0
for /f "delims=" %%a in ('powershell -Command "[guid]::NewGuid().ToString().Replace("-","")"') do set ADMIN_TOKEN=%%a
set SWITCH_URL=http://localhost:8091
set HORIZON_DB_PATH=./horizon-dev.db
set HORIZON_LICENSE_PUBLIC_KEY=cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set SWITCH_PORT=8091

echo Starting Beaconchain Horizon TRIAL...
start "BH-9001" /min cmd /c "set NODE_ROLE=merkle&& set NODE_PORT=9001&& angel-node.exe"
start "BH-9002" /min cmd /c "set NODE_ROLE=hash&& set NODE_PORT=9002&& angel-node.exe"
start "BH-9003" /min cmd /c "set NODE_ROLE=ecdsa&& set NODE_PORT=9003&& angel-node.exe"
start "BH-9004" /min cmd /c "set NODE_ROLE=type&& set NODE_PORT=9004&& angel-node.exe"
start "BH-9005" /min cmd /c "set NODE_ROLE=volume&& set NODE_PORT=9005&& angel-node.exe"
start "BH-9006" /min cmd /c "set NODE_ROLE=time&& set NODE_PORT=9006&& angel-node.exe"
start "BH-9007" /min cmd /c "set NODE_ROLE=user&& set NODE_PORT=9007&& angel-node.exe"
timeout /t 4 /nobreak >nul
start "BH-switch" /min cmd /c "switch.exe"
timeout /t 4 /nobreak >nul
start "BH-backend" /min cmd /c "backend.exe"
timeout /t 6 /nobreak >nul
echo.
echo Trial running. Test: curl -s localhost:8080/api/v1/health
pause
BATEOF

cat > "$OUT/README-TRIAL.txt" << 'RDEOF'
BEACONCHAIN HORIZON — TRIAL PACKAGE
====================================

WHAT IS THIS
An offline-first tamper-evident ledger. Trial is valid for 48
hours from first launch, maximum 1,000,000 transactions.

HOW TO START
1. Extract this folder anywhere
2. Double-click START-TRIAL.bat
3. Wait 15 seconds
4. Verify: curl -s localhost:8080/api/v1/health

INDEPENDENT TESTS
  bash tps-report.sh
  ./pentest.exe -url http://localhost:8080

AFTER 48 HOURS
The backend will refuse to start. To continue, contact the
maintainers for a production license bound to your hardware.

NOT INCLUDED
- Source code
- Production license
- Extended support

(c) 2026 Beaconchain Horizon
RDEOF

cd dist
if command -v zip >/dev/null 2>&1; then
  zip -r "${VERSION}.zip" "$(basename $OUT)" >/dev/null
  echo "Created dist/${VERSION}.zip"
else
  echo "zip not found; folder ready at $OUT"
fi
