#!/usr/bin/env bash
set -euo pipefail
cd "$HOME/Beaconchain/horizon-core" || exit 1

# ==========================================
# 🛡️ EMERGENCY TROJAN: SAFE KILL & RESTORE
# This script performs a backup, wipes the DB,
# and immediately restores it to prove the backup.
# ==========================================

BACKUP_DIR="$HOME/Beaconchain/horizon_safe_backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/horizon-core_$TIMESTAMP.db"

echo "═══════════════════════════════════════════════════════════════════"
echo " 🚀 STARTING EMERGENCY TROJAN PROTOCOL"
echo "═══════════════════════════════════════════════════════════════════"

# 1) CREATE EMERGENCY BACKUP
echo
echo "▌ [1] CREATING EMERGENCY BACKUP..."
mkdir -p "$BACKUP_DIR"

# Stop API to ensure clean DB backup
taskkill //F //IM api.exe //T 2>/dev/null || true
taskkill //F //IM agent-node.exe //T 2>/dev/null || true
sleep 2

if [ ! -f "horizon-core.db" ]; then
    echo "❌ DB not found! Aborting."
    exit 1
fi

cp horizon-core.db "$BACKUP_FILE"
[ -f horizon-core.db-wal ] && cp horizon-core.db-wal "${BACKUP_FILE}-wal" || true
[ -f horizon-core.db-shm ] && cp horizon-core.db-shm "${BACKUP_FILE}-shm" || true

SIZE=$(stat -c%s "$BACKUP_FILE" 2>/dev/null || echo 0)
echo "✅ Backup created: $BACKUP_FILE ($SIZE bytes)"

if [ "$SIZE" -lt 1000 ]; then
    echo "❌ Backup file is too small (invalid). Aborting."
    exit 1
fi

# 2) KILL PATCH (TROJAN PAYLOAD)
echo
echo "▌ [2] EXECUTING KILL PATCH..."
rm -f horizon-core.db horizon-core.db-wal horizon-core.db-shm

if [ -f "horizon-core.db" ]; then
    echo "❌ Kill Patch failed to delete DB."
    exit 1
fi
echo "✅ SYSTEM WIPED OUT."

# 3) RESTORE PROTOCOL
echo
echo "▌ [3] INITIATING RESTORE PROTOCOL..."
cp "$BACKUP_FILE" horizon-core.db
[ -f "${BACKUP_FILE}-wal" ] && cp "${BACKUP_FILE}-wal" horizon-core.db-wal || true
[ -f "${BACKUP_FILE}-shm" ] && cp "${BACKUP_FILE}-shm" horizon-core.db-shm || true

echo "✅ DB file restored from backup."

# 4) REBOOT & VERIFY
echo
echo "▌ [4] RESTARTING SERVICES..."
./bin/api.exe > api.log 2>&1 &
sleep 3

echo
echo "▌ [5] VERIFICATION..."
INFO=$(curl -s --max-time 3 http://localhost:8080/api/v1/chain/info)
if echo "$INFO" | grep -q "height"; then
    HEIGHT=$(echo "$INFO" | jq -r .height 2>/dev/null || echo "?")
    echo "    🏆 SUCCESS! System restored. Height=$HEIGHT"
else
    echo "    ❌ FAILED! API is not responding."
    exit 1
fi

echo "═══════════════════════════════════════════════════════════════════"
