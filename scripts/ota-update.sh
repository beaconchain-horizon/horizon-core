#!/bin/bash
# ============================================================
#  HORIZON OTA UPDATE
#  بروزرسانی سوئیچ مشتری از راه دور
# ============================================================
set -e

INSTALL_DIR="${HORIZON_HOME:-/opt/horizon}"
BACKUP_DIR="$INSTALL_DIR/.backups/ota-$(date +%s)"

echo "═══ OTA UPDATE ═══"

mkdir -p "$BACKUP_DIR"
cp -r "$INSTALL_DIR/data" "$BACKUP_DIR/" 2>/dev/null || true
cp "$INSTALL_DIR/docker-compose.yml" "$BACKUP_DIR/" 2>/dev/null || true

cd "$INSTALL_DIR"

# ۱) توقف سرویس
echo "🛑 Stopping switch..."
docker-compose stop switch backend 2>/dev/null || docker compose stop switch backend 2>/dev/null

# ۲) بک‌آپ دیتابیس
echo "💾 Backing up database..."
cp -r "$INSTALL_DIR/data" "$BACKUP_DIR/data-final" 2>/dev/null || true

# ۳) دریافت نسخه جدید (از git یا URL)
if [ -d ".git" ]; then
    echo "📥 Pulling latest..."
    git pull origin main
fi

# ۴) بیلد مجدد
echo "🔨 Rebuilding..."
docker-compose build --no-cache switch 2>/dev/null || docker compose build --no-cache switch

# ۵) راه‌اندازی
echo "🚀 Starting..."
docker-compose up -d 2>/dev/null || docker compose up -d

sleep 5

# ۶) تست
if curl -sf http://localhost:8080/api/v1/health &> /dev/null; then
    echo "✅ OTA UPDATE SUCCESSFUL"
    echo "   Backup: $BACKUP_DIR"
    echo "   Rollback: scripts/ota-rollback.sh $BACKUP_DIR"
else
    echo "❌ OTA FAILED — rolling back..."
    cd "$INSTALL_DIR"
    cp "$BACKUP_DIR/data-final" -r "$INSTALL_DIR/data" 2>/dev/null || true
    docker-compose restart switch 2>/dev/null || docker compose restart switch
    exit 1
fi
