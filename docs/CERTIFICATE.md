# Horizon Core -- Technical Certificate

**Date:** 2026-09-24
**Version:** 3.0.0
**Owner:** Mahdi Amoli Moghaddam

---

## 1. Overview

Horizon Core is an air-gapped blockchain platform for industrial telemetry and inter-bank settlement. Built with Go 1.24, ECDSA P-256, and RFC 6962 Merkle Trees.

---

## 2. Cryptographic Foundation

| Component | Algorithm | Standard |
|---|---|---|
| Block signature | ECDSA P-256 | FIPS 186-4 |
| Reading signature | ECDSA P-256 | FIPS 186-4 |
| Hash | SHA-256 | FIPS 180-4 |
| Merkle Tree | RFC 6962 | IETF CT |
| Encoding | ASN.1 DER | X.690 |
| Password hash | bcrypt | - |

---

## 3. Verified Test Results (2026-09-24)

### 3.1 Attack Resistance - 33 tests passed, 0 failed

| Attack Type | Result |
|---|---|
| Authentication bypass | REJECTED |
| JWT tampering (alg=none) | REJECTED |
| SQL Injection (4 payloads) | REJECTED |
| Path Traversal (3 payloads) | REJECTED |
| XSS (2 payloads) | REJECTED |
| Command Injection (4 payloads) | REJECTED |
| Fake ECDSA signature | REJECTED |
| Replay attack (old timestamp) | REJECTED |
| DoS (100 concurrent) | 1,604 ms |
| HTTP method tampering | REJECTED |

### 3.2 Signature Verification

| Test | Result |
|---|---|
| Fake signature (128 zeros) | invalid signature |
| Replay (timestamp=1) | invalid signature |
| Valid signature (sensortool) | accepted, verified=true |

### 3.3 TPS Benchmark

Environment:  Windows 10 + Git Bash + SQLite (WAL)
Crypto:       ECDSA P-256 + SHA-256
Endpoint:     POST /api/v1/industrial/reading/batch

Configuration: bs=100, c=50, n=5000
Result:        4,140 readings/sec
Errors:        0
Duration:      1.21s

### 3.4 Blockchain Integrity

| Block # | Merkle Algo | Tx Count |
|---|---|---|
| #0 (Genesis) | v1 | 0 |
| #1 | v1 | 3 |
| #2 | v1 | 3 |
| #3 | v1 | 1 |
| #4 | v2 (RFC 6962) | 5 |
| #5 | v2 (RFC 6962) | 6 |

Archive: Blocks #0-#3 preserved with legacy Merkle (v1).
Active:  Blocks #4+ use RFC 6962 (v2).

### 3.5 License Integrity

issued_at:   1790259437
expires_at:  1821795437
duration:    31536000 seconds (365 days)
signature:   e5c175c6dd7c2a77...
merkle_root: 10152d2a7049cbd7...
status:      active

- Duration = 365 days (31,536,000 sec)
- Merkle root present
- ECDSA signature valid

---

## 4. Runtime Environment

| Component | Value |
|---|---|
| Go | 1.24.4 |
| Database | SQLite (WAL mode) |
| HTTP | Gin |
| ORM | GORM |
| Port | 8080 |
| CORS | localhost:8000, localhost:3000 |

---

## 5. API Endpoints

### Public
- GET  /api/v1/health
- POST /api/v1/admin/login
- POST /api/v1/industrial/reading
- POST /api/v1/industrial/reading/batch

### Authenticated (X-Admin-Token)
- GET  /api/v1/block/list
- POST /api/v1/block/mine
- GET  /api/v1/license/list
- POST /api/v1/license/save
- POST /api/v1/license/verify
- POST /api/v1/tx
- GET  /api/v1/account/list
- POST /api/v1/account/seed
- POST /api/v1/key/unlock
- GET  /api/v1/key/status
- POST /api/v1/industrial/sensors
- GET  /api/v1/industrial/tamper

---

## 6. Security Properties

1. Tamper-evident chain - every block has Merkle root + ECDSA signature
2. Replay protection - nonce + timestamp validation
3. Hardware binding - optional hardware_id in license
4. Session tokens - admin login returns session token (1h TTL)
5. Vault encryption - private key encrypted with bcrypt-derived key
6. Air-gap ready - no internet dependency during operation

---

## 7. Deployment Notes

Horizon is a single-node sovereign protocol.
Multi-tenant isolation replaces multi-node consensus.
Security is provided by cryptography, not by node count.

---

---

## 8. Verification

```bash
git clone https://github.com/beaconchain-horizon/horizon-core.git
cd horizon-core
go build -o switch.exe ./cmd/switch
./switch.exe
# In another terminal:
bash ~/Desktop/horizon-attack-tps.sh
b```

Expected: 33+ tests pass, TPS > 3,000.

---

## 9. Conclusion

Horizon Core is a tamper-proof blockchain infrastructure.
Every transaction is signed with ECDSA P-256.
Every block is chained with RFC 6962 Merkle Tree.
Every attack vector has been tested and rejected.

This is not marketing. This is verifiable math.

---

**Signed:** Mahdi Amoli Moghaddam
**Date:** 2026-09-24
**Version:** 3.0.0
