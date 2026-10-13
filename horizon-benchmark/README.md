# Horizon Core — Independent Benchmark

**Public, reproducible benchmark for Horizon Core blockchain.**

This repository allows any independent party to clone, run, and verify the TPS claims of Horizon Core. No source code is included — only compiled binaries and a test runner.

---

## Quick Start

### Linux / macOS

```bash
git clone https://github.com/beaconchain-horizon/horizon-benchmark.git
cd horizon-benchmark
./run.sh
```

### Windows

۱. Download ZIP
۲. Extract
۳. Double-click `run.bat`

---

## Verified Results

| Metric | Value |
|---|---|
| Peak TPS | 31,113 |
| Sustained TPS | 20,000+ |
| Errors | 0 |
| Tampering | 0 |
| ECDSA Sign | < 1 ms |

**Note:** TPS depends on your CPU and system load. On Intel i7-1185G7: 20K–31K.

---

## What This Proves

- ✅ ECDSA P-256 signing on every reading
- ✅ SHA-256 hashing per reading
- ✅ Zero errors across 200,000 transactions
- ✅ Zero tampering (tamper_count = 0)
- ✅ Reproducible on any machine

---

## Files

- `bin/` — compiled binaries (switch, sensortool, tpsbench)
- `config/chain.json` — chain configuration
- `run.sh` — Linux/macOS launcher
- `run.bat` — Windows launcher
- `EVIDENCE.md` — sanitized terminal logs
- `LICENSE` — MIT

---

## No Source Code

This repository contains **only compiled binaries**. The source code is private and available under NDA for qualified partners.

---

## Contact

- **GitHub:** https://github.com/beaconchain-horizon
- **Dashboard:** https://beaconchain-horizon.github.io/

---

**© 2026 Horizon Core — MIT License**
