#!/bin/bash

echo ""
echo "=========================================="
echo "   HORIZON ANGEL - TPS BENCHMARK SUITE"
echo "=========================================="
echo ""

# Test 1: Health Check
echo "▶ [1/6] Health Check..."
echo ""
echo -n "  Backend:  "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-backend.liara.run/api/v1/health
echo -n "  Switch:   "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-switch.liara.run/health
echo -n "  Frontend: "
curl -s -o /dev/null -w "%{http_code}\n" https://horizon-frontend.liara.run
echo ""
sleep 2

# Test 2: Light Load (100 concurrent)
echo "▶ [2/6] Light Load - 100 concurrent, 20s..."
echo ""
hey -z 20s -c 100 https://horizon-switch.liara.run/health 2>&1 | grep -E "Requests/sec|Average|Total:|\[200\]"
echo ""
sleep 5

# Test 3: Medium Load (200 concurrent)
echo "▶ [3/6] Medium Load - 200 concurrent, 30s..."
echo ""
hey -z 30s -c 200 https://horizon-switch.liara.run/health 2>&1 | grep -E "Requests/sec|Average|Total:|\[200\]"
echo ""
sleep 5

# Test 4: Heavy Load (500 concurrent)
echo "▶ [4/6] Heavy Load - 500 concurrent, 30s..."
echo ""
hey -z 30s -c 500 https://horizon-switch.liara.run/health 2>&1 | grep -E "Requests/sec|Average|Total:|\[200\]|\[503\]|\[502\]"
echo ""
sleep 5

# Test 5: Extreme Load (1000 concurrent)
echo "▶ [5/6] Extreme Load - 1000 concurrent, 30s..."
echo ""
hey -z 30s -c 1000 https://horizon-switch.liara.run/health 2>&1 | grep -E "Requests/sec|Average|Total:|\[200\]|\[503\]|\[502\]"
echo ""
sleep 5

# Test 6: Backend TPS
echo "▶ [6/6] Backend TPS - 200 concurrent, 30s..."
echo ""
hey -z 30s -c 200 https://horizon-backend.liara.run/api/v1/health 2>&1 | grep -E "Requests/sec|Average|Total:|\[200\]"
echo ""

echo "=========================================="
echo "   ✅ ALL TESTS COMPLETE"
echo "=========================================="
echo ""
echo "📋 Copy the entire output above and send it."
echo ""
