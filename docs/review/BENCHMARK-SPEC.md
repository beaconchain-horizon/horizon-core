# Horizon — Benchmark Specification

Required for every published TPS:

Hardware: CPU, RAM, Disk, OS
Software: Go, SQLite, WAL, synchronous, journal_mode
Workload: batch size, concurrency, tx size, ECDSA P-256, Merkle, persistence
Timing: duration, warmup, successful, failed
Metrics: TPS, p50, p95, p99, error rate

Reproduce:
git clone https://github.com/beaconchain-horizon/horizon-core.git
cd horizon-core
./benchmark/run.sh

Current: Peak TPS 31,113 | Errors 0 | Status: preliminary
