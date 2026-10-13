# Evidence — Reference Benchmark Run

**Date:** 2026-09-27
**System:** Windows 11 + Git Bash
**CPU:** Intel i7-1185G7 @ 3.00GHz

---

## Reference Run — 200K Readings

```
$ ./run.sh
=== Horizon Benchmark ===
keys written to test-keys

Test: 200K readings, c=50, bs=500
Pre-signing 400 batches (200000 readings)...

>>> 200000 readings | c=50 bs=500 | 6.30s | OK=400 ERR=0 | ★TPS=31113

=== Integrity Check ===
{"readings_count":200000,"tamper_count":0}
```

---

## What Each Number Means

| Field | Value | Meaning |
|---|---|---|
| readings | 200,000 | Total transactions signed |
| c | 50 | Concurrency (parallel requests) |
| bs | 500 | Batch size (readings per request) |
| OK | 400 | Successful batches |
| ERR | 0 | Failed batches |
| TPS | 31,113 | Transactions per second |
| tamper_count | 0 | Integrity violations |

---

## Reproducibility

To reproduce on your machine:

```bash
git clone https://github.com/beaconchain-horizon/horizon-benchmark.git
cd horizon-benchmark
./run.sh
```

Expected output: `TPS=20000+` with `ERR=0` and `tamper_count=0`.

---

## Security Verification

| Check | Status |
|---|---|
| Server binding | 127.0.0.1 only |
| Authentication | X-Admin-Token required |
| Signing | ECDSA P-256 per reading |
| Hashing | SHA-256 per reading |
| Tamper detection | 0 violations |
| Network exposure | None (localhost) |
| Credentials in logs | None |

---

**Signed:** ECDSA P-256
**Date:** 2026-09-27
