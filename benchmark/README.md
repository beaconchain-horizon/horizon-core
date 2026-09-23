# Horizon Core — Independent Benchmark

**Public, reproducible benchmark for Horizon Core blockchain.**

## Quick Start

### Linux / macOS
```bash
git clone https://github.com/beaconchain-horizon/beaconchain-horizon.github.io.git
cd beaconchain-horizon.github.io/horizon-benchmark
./run.sh
```

### Windows
۱. Download ZIP
۲. Extract
۳. Double-click `run.bat`

## Verified Results

| Metric | Value |
|---|---|
| Peak TPS | 31,113 |
| Errors | 0 |
| Tampering | 0 |
| ECDSA Sign | < 1 ms |

## Files

- `bin/` — compiled binaries (switch, sensortool, tpsbench)
- `config/chain.json` — chain configuration
- `run.sh` — Linux/macOS launcher
- `run.bat` — Windows launcher
- `EVIDENCE.md` — sanitized terminal logs
- `LICENSE` — MIT

## Reproduce

```bash
./run.sh
```

Expected output: TPS=30000+ with ERR=0

© 2026 Horizon Core — MIT License
