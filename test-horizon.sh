#!/bin/bash

echo "=========================================="
echo "   HORIZON COMPREHENSIVE TEST SUITE"
echo "=========================================="
echo ""

# 1. HEALTH CHECK
echo "1. HEALTH CHECK"
echo "------------------------------------------"
echo -n "Backend:  "
curl -s -o /dev/null -w "%{http_code}" https://horizon-backend.liara.run/api/v1/health
echo ""
echo -n "Switch:   "
curl -s -o /dev/null -w "%{http_code}" https://horizon-switch.liara.run/health
echo ""
echo -n "Frontend: "
curl -s -o /dev/null -w "%{http_code}" https://horizon-frontend.liara.run
echo ""
echo ""

# 2. LICENSE GENERATION
echo "2. LICENSE GENERATION"
echo "------------------------------------------"
curl -s -X POST https://horizon-backend.liara.run/api/v1/license/generate -H "Content-Type: application/json" -d '{"product_id":"horizon-core","user_id":"bank_melli","volume":100,"duration":8760}'
echo ""
echo ""

# 3. TPS BENCHMARK - BACKEND
echo "3. TPS BENCHMARK - BACKEND (30s, 200 concurrent)"
echo "------------------------------------------"
hey -z 30s -c 200 https://horizon-backend.liara.run/api/v1/health | grep -E "Requests/sec|Average|Slowest|Fastest|Total:|200"
echo ""

# 4. TPS BENCHMARK - SWITCH
echo "4. TPS BENCHMARK - SWITCH (30s, 200 concurrent)"
echo "------------------------------------------"
hey -z 30s -c 200 https://horizon-switch.liara.run/health | grep -E "Requests/sec|Average|Slowest|Fastest|Total:|200"
echo ""

# 5. STABILITY TEST
echo "5. STABILITY TEST (60s, 300 concurrent)"
echo "------------------------------------------"
hey -z 60s -c 300 https://horizon-switch.liara.run/health | grep -E "Requests/sec|Average|Slowest|200"
echo ""

# 6. BLOCKCHAIN STATS
echo "6. BLOCKCHAIN STATS"
echo "------------------------------------------"
curl -s https://horizon-switch.liara.run/stats | head -c 300
echo ""
echo ""

echo "=========================================="
echo "   TEST SUITE COMPLETED"
echo "=========================================="
