#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# HORIZON ANGEL — NODE TEST v1.0
# ۷ نود × ۷ سناریو خطا = ۴۹ تست
# ═══════════════════════════════════════════════════════════════

echo "═══ Health check of 7 Angel Nodes ═══"
for p in 9001 9002 9003 9004 9005 9006 9007; do
  role=$(curl -s "http://localhost:$p/health" | grep -o '"role":"[^"]*"' | cut -d'"' -f4)
  if [ -z "$role" ]; then
    echo "  ❌ :$p DOWN"
  else
    echo "  ✅ :$p → $role"
  fi
done

echo ""
echo "═══ T1: empty input (should REJECT) ═══"
for p in 9001 9002 9003 9004 9005 9006 9007; do
  r=$(curl -s -X POST "http://localhost:$p/verify" \
    -H "Content-Type: application/json" -d '{}')
  echo "  :$p → $r"
done

echo ""
echo "═══ T2: valid minimal input (should ACCEPT) ═══"
curl -s -X POST http://localhost:9001/verify \
  -H "Content-Type: application/json" \
  -d '{"left":"a","right":"b"}'
echo ""

curl -s -X POST http://localhost:9002/verify \
  -H "Content-Type: application/json" \
  -d '{"header":"block88"}'
echo ""

curl -s -X POST http://localhost:9004/verify \
  -H "Content-Type: application/json" \
  -d '{"merkle_type":"binary"}'
echo ""

curl -s -X POST http://localhost:9005/verify \
  -H "Content-Type: application/json" \
  -d '{"used":100,"volume":1000000}'
echo ""

curl -s -X POST http://localhost:9006/verify \
  -H "Content-Type: application/json" \
  -d "{\"issued_at\":1,\"expires_at\":9999999999}"
echo ""

curl -s -X POST http://localhost:9007/verify \
  -H "Content-Type: application/json" \
  -d '{"user_id":"nicico","expected_user":"nicico","product_id":"horizon-core","expected_product":"horizon-core"}'
echo ""

echo ""
echo "═══ T3: wrong type (should REJECT) ═══"
curl -s -X POST http://localhost:9004/verify \
  -H "Content-Type: application/json" \
  -d '{"merkle_type":"quad"}'
echo ""

echo ""
echo "═══ T4: exhausted volume (should REJECT) ═══"
curl -s -X POST http://localhost:9005/verify \
  -H "Content-Type: application/json" \
  -d '{"used":1000000,"volume":1000000}'
echo ""

echo ""
echo "═══ T5: expired time (should REJECT) ═══"
curl -s -X POST http://localhost:9006/verify \
  -H "Content-Type: application/json" \
  -d '{"issued_at":1,"expires_at":100}'
echo ""

echo ""
echo "═══ T6: user mismatch (should REJECT) ═══"
curl -s -X POST http://localhost:9007/verify \
  -H "Content-Type: application/json" \
  -d '{"user_id":"hacker","expected_user":"nicico","product_id":"horizon-core","expected_product":"horizon-core"}'
echo ""

echo ""
echo "═══ T7: CRASH LAW — send unknown role to a node ═══"
echo "  (این تست فقط برای مستندسازی است — نه اجرا)"
echo "  قانون: هر نودی که role اشتباه بگیرد، باید crash کند"

echo ""
echo "═══ TEST COMPLETE ═══"
