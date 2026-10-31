#!/usr/bin/env bash

echo "=========================================="
echo "   BACKEND LOGS (Last 20 lines)"
echo "=========================================="
if [ -f api.log ]; then
    tail -n 20 api.log
else
    echo "❌ api.log not found."
fi

echo ""
echo "=========================================="
echo "   SWITCH LOGS (Last 20 lines)"
echo "=========================================="
if [ -f switch.log ]; then
    tail -n 20 switch.log
else
    echo "❌ switch.log not found."
fi
