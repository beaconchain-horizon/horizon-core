#!/bin/bash
# ============================================================
#  HORIZON ONBOARD — افزودن مشتری جدید
#  Usage: ./onboard.sh --type=bank --name="بانک ملی" --id=bank-melli
# ============================================================
set -e

CENTRAL="${CENTRAL_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

TYPE=""
NAME=""
TENANT_ID=""
CONTACT_EMAIL=""
CONTACT_PHONE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --type=*)          TYPE="${1#*=}"; shift ;;
        --name=*)          NAME="${1#*=}"; shift ;;
        --id=*)            TENANT_ID="${1#*=}"; shift ;;
        --email=*)         CONTACT_EMAIL="${1#*=}"; shift ;;
        --phone=*)         CONTACT_PHONE="${1#*=}"; shift ;;
        --token=*)         ADMIN_TOKEN="${1#*=}"; shift ;;
        --central=*)       CENTRAL="${1#*=}"; shift ;;
        *) echo "unknown option: $1"; exit 1 ;;
    esac
done

if [ -z "$TYPE" ] || [ -z "$NAME" ] || [ -z "$TENANT_ID" ]; then
    echo "Usage: $0 --type=bank|refinery|powerplant|gas --name=\"اسم\" --id=tenant-id"
    exit 1
fi

# چک قالب
TEMPLATE="templates/${TYPE}.json"
if [ ! -f "$TEMPLATE" ]; then
    echo "⚠️  Template $TYPE not found, using generic"
    TEMPLATE="templates/generic.json"
fi

echo "════════════════════════════════════════════"
echo "  🚀 ONBOARDING: $NAME ($TENANT_ID)"
echo "  Type: $TYPE"
echo "  Template: $TEMPLATE"
echo "════════════════════════════════════════════"

# ۱) ثبت tenant در مرکز
echo ""
echo "── [1/3] ثبت tenant در مرکز ──"

GEN_TOKEN=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-f0-9' | fold -w 64 | head -n 1)

RESP=$(curl -s -X POST "$CENTRAL/api/v1/tenant/create" \
    -H "Content-Type: application/json" \
    -H "X-Admin-Token: $ADMIN_TOKEN" \
    -d "{\"tenant_id\":\"$TENANT_ID\",\"name\":\"$NAME\",\"type\":\"$TYPE\",\"contact_email\":\"$CONTACT_EMAIL\",\"contact_phone\":\"$CONTACT_PHONE\",\"agent_token\":\"$GEN_TOKEN\"}")

echo "$RESP"

# ۲) ساخت config مشتری از قالب
echo ""
echo "── [2/3] ساخت config از قالب ──"

OUT_DIR="customers/$TENANT_ID"
mkdir -p "$OUT_DIR"

sed "s/TENANT_PLACEHOLDER/$TENANT_ID/g; s/TENANT_NAME_PLACEHOLDER/$NAME/g" \
    "$TEMPLATE" > "$OUT_DIR/template.json"

cat > "$OUT_DIR/agent.json" <<EOF
{
  "tenant_id": "$TENANT_ID",
  "central_url": "$CENTRAL",
  "agent_token": "$GEN_TOKEN",
  "switch_url": "http://localhost:8080",
  "heartbeat_sec": 60
}
EOF

echo "  ✓ $OUT_DIR/template.json"
echo "  ✓ $OUT_DIR/agent.json"

# ۳) خلاصه
echo ""
echo "── [3/3] خلاصه ──"
echo ""
echo "════════════════════════════════════════════"
echo "  ✅ TENANT CREATED: $TENANT_ID"
echo "════════════════════════════════════════════"
echo ""
echo "📋 تحویل به انفورماتیک مشتری:"
echo "  ۱. بسته نصبی:    ./install.sh"
echo "  ۲. کانفیگ agent: $OUT_DIR/agent.json"
echo "  ۳. توکن agent:   $GEN_TOKEN"
echo ""
echo "🔐 مراحل بعدی:"
echo "  ۱. در سرور مشتری: ./install.sh"
echo "  ۲. فایل agent.json رو در سرور کپی کن"
echo "  ۳. agent رو راه‌اندازی کن: ./horizon-agent --config agent.json"
echo "  ۴. هفته‌ای یک بار لاگ heartbeat رو در پنل مرکزی چک کن"
echo "════════════════════════════════════════════"
