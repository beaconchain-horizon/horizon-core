#!/bin/bash
# ============================================================
#  HORIZON SWITCH — INSTALLER (Linux / macOS)
#  نصب روی سرور مشتری
# ============================================================
set -e

INSTALL_DIR="${HORIZON_HOME:-/opt/horizon}"
DATA_DIR="${INSTALL_DIR}/data"
LOG_DIR="${INSTALL_DIR}/logs"

echo "════════════════════════════════════════════"
echo "  🛡️  HORIZON SWITCH — Installation"
echo "  Target: $INSTALL_DIR"
echo "════════════════════════════════════════════"

# چک پیش‌نیازها
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed."
    echo "   Install: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed."
    exit 1
fi

# ساخت دایرکتوری‌ها
echo "📁 Creating directories..."
mkdir -p "$INSTALL_DIR" "$DATA_DIR" "$LOG_DIR"

# کپی فایل‌ها
echo "📦 Copying files..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cp -r "$SCRIPT_DIR"/* "$INSTALL_DIR/" 2>/dev/null || true

# ساخت .env از نمونه
if [ ! -f "$INSTALL_DIR/.env" ]; then
    echo "⚙️  Creating .env from template..."
    cat > "$INSTALL_DIR/.env" <<EOF
POSTGRES_PASSWORD=$(openssl rand -hex 16 2>/dev/null || echo "horizon_secret")
ADMIN_TOKEN=$(openssl rand -hex 32 2>/dev/null || echo "admin_token_change_me")
API_KEY=$(openssl rand -hex 32 2>/dev/null || echo "api_key_change_me")
EOF
    echo "  ⚠️  Edit $INSTALL_DIR/.env to customize"
fi

# Build و راه‌اندازی
cd "$INSTALL_DIR"
echo "🔨 Building images..."
if command -v docker-compose &> /dev/null; then
    docker-compose build
else
    docker compose build
fi

echo "🚀 Starting services..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

sleep 5

# تست
echo "🧪 Testing health..."
if curl -sf http://localhost:8080/api/v1/health &> /dev/null; then
    echo ""
    echo "════════════════════════════════════════════"
    echo "  ✅ INSTALLATION SUCCESSFUL"
    echo "════════════════════════════════════════════"
    echo "  Switch:  http://localhost:8080"
    echo "  Backend: http://localhost:8081"
    echo "  Data:    $DATA_DIR"
    echo "  Logs:    $LOG_DIR"
    echo ""
    echo "  To view logs:  docker logs horizon-switch -f"
    echo "  To stop:       cd $INSTALL_DIR && docker-compose down"
    echo "════════════════════════════════════════════"
else
    echo "❌ Health check failed. Check logs:"
    docker logs horizon-switch 2>&1 | tail -20
    exit 1
fi
