#!/bin/bash
BACKUP_DIR="backups"
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="horizon_backup_$TIMESTAMP.tar.gz"

echo "📦 Creating backup: $BACKUP_NAME"
# بک‌آپ گرفتن از کدها، دیتابیس‌ها و فایل‌های تنظیمات
tar -czf "$BACKUP_DIR/$BACKUP_NAME" cmd internal go.mod go.sum *.db *.yml *.sh 2>/dev/null

echo "✅ Backup created successfully at $BACKUP_DIR/$BACKUP_NAME"
