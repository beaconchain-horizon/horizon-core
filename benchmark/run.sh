#!/bin/bash
echo "=== Horizon Benchmark ==="
echo "OS: $(uname -s -r -m)"
echo "Go: $(go version | awk '{print $3}')"
./tpsbench --n=200000 --c=50 --bs=500
