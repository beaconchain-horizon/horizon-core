#!/usr/bin/env bash

echo "=========================================="
echo "   DIAGNOSTIC REPORT"
echo "=========================================="

echo "1) Checking API Logs (Last 10 lines)..."
if [ -f api.log ]; then
    tail -n 10 api.log
else
    echo "⚠️ api.log not found!"
fi

echo ""
echo "2) Finding all .db files in the project..."
find . -name "*.db" -type f

echo ""
echo "3) Checking if API is running on port 8080..."
curl -s -o /dev/null -w "API Status Code: %{http_code}\n" http://localhost:8080/api/v1/health || echo "❌ API is NOT responding."

echo ""
echo "4) Checking if Switch is running on port 8091..."
curl -s -o /dev/null -w "Switch Status Code: %{http_code}\n" http://localhost:8091/api/v1/health || echo "❌ Switch is NOT responding."

echo "=========================================="
