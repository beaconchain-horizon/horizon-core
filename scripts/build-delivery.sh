#!/bin/bash
# ============================================================
#  HORIZON PLAN B — Delivery Package Builder (Token-Based)
#  بدون URL hardcode — همه چیز با توکن
# ============================================================
set -e

# ─── پارس آرگومان‌ها ───
TYPE=""
NAME=""
TENANT_ID=""
AGENT_TOKEN=""
LIARA_API_TOKEN=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --type=*)              TYPE="${1#*=}"; shift ;;
    --name=*)              NAME="${1#*=}"; shift ;;
    --id=*)                TENANT_ID="${1#*=}"; shift ;;
    --agent-token=*)       AGENT_TOKEN="${1#*=}"; shift ;;
    --liara-api-token=*)   LIARA_API_TOKEN="${1#*=}"; shift ;;
    *) echo "unknown option: $1"; exit 1 ;;
  esac
done

if [ -z "$TYPE" ] || [ -z "$NAME" ] || [ -z "$TENANT_ID" ] || [ -z "$AGENT_TOKEN" ] || [ -z "$LIARA_API_TOKEN" ]; then
  echo "❌ پارامترهای اجباری:"
  echo "  --type=bank|refinery|powerplant|gas|generic"
  echo "  --name=\"نام مشتری\""
  echo "  --id=tenant-id"
  echo "  --agent-token=<64-char-hex>"
  echo "  --liara-api-token=<JWT>"
  exit 1
fi

OUT_DIR="deliveries/${TENANT_ID}"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR/config"

echo "📦 ساخت پکیج برای: $NAME ($TENANT_ID)"
echo ""

# ─── [1] Build باینری‌ها ───
echo "🔨 [1/5] Build باینری‌ها..."
GOOS=linux GOARCH=amd64 go build -o "$OUT_DIR/horizon-switch" ./cmd/switch
GOOS=linux GOARCH=amd64 go build -o "$OUT_DIR/horizon-agent" ./cmd/client-agent
echo "   ✓ horizon-switch"
echo "   ✓ horizon-agent"
echo ""

# ─── [2] قالب صنفی ───
echo "📋 [2/5] قالب صنفی..."
if [ -f "templates/${TYPE}.json" ]; then
  sed "s/TENANT_PLACEHOLDER/$TENANT_ID/g; s/TENANT_NAME_PLACEHOLDER/$NAME/g" \
    "templates/${TYPE}.json" > "$OUT_DIR/config/template.json"
  echo "   ✓ $TYPE"
else
  sed "s/TENANT_PLACEHOLDER/$TENANT_ID/g; s/TENANT_NAME_PLACEHOLDER/$NAME/g" \
    "templates/generic.json" > "$OUT_DIR/config/template.json"
  echo "   ✓ generic"
fi

# chain.json اگه هست
[ -f "config/chain.json" ] && cp config/chain.json "$OUT_DIR/config/chain.json"
echo ""

# ─── [3] .env با توکن‌ها ───
echo "⚙️  [3/5] .env..."

cat > "$OUT_DIR/.env" <<ENVEOF
# ═══════════════════════════════════════════════════════════
#  HORIZON SWITCH — $NAME
#  ⚠️ محرمانه — این فایل را به کسی ندهید
# ═══════════════════════════════════════════════════════════

TENANT_ID=$TENANT_ID
AGENT_TOKEN=$AGENT_TOKEN
LIARA_API_TOKEN=$LIARA_API_TOKEN

SWITCH_PORT=8080
SWITCH_DB=/opt/horizon/data/horizon-switch.db
CHAIN_CONFIG=/opt/horizon/data/chain.json
CUSTOMERS_CONFIG=/opt/horizon/data/customers.json
ADMIN_TOKEN=$(openssl rand -hex 32 2>/dev/null || echo "change_me_$(date +%s)")

GIN_MODE=release
HEARTBEAT_SEC=60
ENVEOF

chmod 600 "$OUT_DIR/.env"
echo "   ✓ .env (chmod 600)"
echo ""

# ─── [4] agent.json بدون URL ───
echo "🛰️  [4/5] agent.json (توکن‌محور)..."

cat > "$OUT_DIR/agent.json" <<AGENTEOF
{
  "tenant_id": "$TENANT_ID",
  "liara_region": "api.liara.ir",
  "liara_project_id": "horizon-switch",
  "liara_api_token": "$LIARA_API_TOKEN",
  "agent_token": "$AGENT_TOKEN",
  "heartbeat_sec": 60
}
AGENTEOF

chmod 600 "$OUT_DIR/agent.json"
echo "   ✓ agent.json (chmod 600)"
echo ""

# ─── [5] install.sh ───
echo "📜 [5/5] install.sh..."
cp installer/install.sh "$OUT_DIR/install.sh"
chmod +x "$OUT_DIR/install.sh"

# README
cat > "$OUT_DIR/README.txt" <<READMEEOF
════════════════════════════════════════════════════════════════
  HORIZON SWITCH — Installation
  $NAME
════════════════════════════════════════════════════════════════

روی سرور لینوکس (root):
  sudo ./install.sh

تست:
  curl http://localhost:8080/api/v1/health

لاگ:
  journalctl -u horizon-switch -f
  journalctl -u horizon-agent -f

⚠️ این پکیج حاوی توکن است — محرمانه بماند.
════════════════════════════════════════════════════════════════
READMEEOF

echo "   ✓ install.sh"
echo "   ✓ README.txt"
echo ""
echo "✅ پکیج آماده: $OUT_DIR"
