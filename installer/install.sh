#!/bin/bash
# ============================================================
#  HORIZON SWITCH — Bank-Side Installer (Plan B, Token-Based)
# ============================================================
set -e

if [ "$EUID" -ne 0 ]; then
  echo "❌ نیاز به root: sudo ./install.sh"
  exit 1
fi

INSTALL_DIR="/opt/horizon"
DATA_DIR="$INSTALL_DIR/data"
LOG_DIR="$INSTALL_DIR/logs"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "════════════════════════════════════════════════════════════════"
echo "  🛡️  HORIZON SWITCH — Installation"
echo "════════════════════════════════════════════════════════════════"

# ─── چک‌ها ───
[ "$(uname -m)" != "x86_64" ] && { echo "❌ فقط x86_64"; exit 1; }

# ─── دایرکتوری ───
mkdir -p "$INSTALL_DIR" "$DATA_DIR" "$LOG_DIR"

# ─── کپی فایل‌ها ───
cp "$SCRIPT_DIR/horizon-switch" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/horizon-agent" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/horizon-switch" "$INSTALL_DIR/horizon-agent"

[ -f "$SCRIPT_DIR/.env" ] && cp "$SCRIPT_DIR/.env" "$INSTALL_DIR/.env" && chmod 600 "$INSTALL_DIR/.env"
[ -f "$SCRIPT_DIR/agent.json" ] && cp "$SCRIPT_DIR/agent.json" "$INSTALL_DIR/agent.json" && chmod 600 "$INSTALL_DIR/agent.json"
[ -d "$SCRIPT_DIR/config" ] && cp -r "$SCRIPT_DIR/config/"* "$DATA_DIR/" 2>/dev/null || true

# ─── سرویس systemd ───
cat > /etc/systemd/system/horizon-switch.service <<SVCEOF
[Unit]
Description=Horizon Switch
After=network.target

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/horizon-switch
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/switch.log
StandardError=append:$LOG_DIR/switch.error.log

[Install]
WantedBy=multi-user.target
SVCEOF

cat > /etc/systemd/system/horizon-agent.service <<SVCEOF
[Unit]
Description=Horizon Client Agent
After=network.target horizon-switch.service
Requires=horizon-switch.service

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/horizon-agent --config=$INSTALL_DIR/agent.json
Restart=always
RestartSec=30
StandardOutput=append:$LOG_DIR/agent.log
StandardError=append:$LOG_DIR/agent.error.log

[Install]
WantedBy=multi-user.target
SVCEOF

# ─── راه‌اندازی ───
systemctl daemon-reload
systemctl enable --now horizon-switch.service
sleep 3
systemctl enable --now horizon-agent.service
sleep 2

# ─── تست ───
HTTP=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health 2>/dev/null || echo "000")

if [ "$HTTP" = "200" ]; then
  echo ""
  echo "✅ نصب موفق"
  echo "   Switch: http://localhost:8080"
  echo "   Logs:   journalctl -u horizon-switch -f"
else
  echo "❌ خطا (HTTP: $HTTP)"
  tail -20 "$LOG_DIR/switch.log" 2>/dev/null
  exit 1
fi
