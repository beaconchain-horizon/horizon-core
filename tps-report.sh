#!/bin/bash
echo "Running TPS Benchmark (Audit Mode)..."
# شبیه‌سازی تست با tpsbench
./tpsbench.exe -c 100 -r 10 -batch 10000 -token "AUDIT-TOKEN" > tps_output.txt 2>&1
echo "✅ TPS report generated in tps_output.txt"
