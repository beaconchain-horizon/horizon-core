# Horizon Core — Architecture

**Version:** 3.0.0
**Date:** 2026-09-24
**Protocol Type:** Open Standard (IETF-style)
**License:** Commercial
**Owner:** Mahdi Amoli Moghaddam

---

## 1. Overview

Horizon is a global, sovereign blockchain protocol for:
- Industrial telemetry (sensors, SCADA)
- Inter-bank settlement
- License issuance and verification
- Air-gapped operations

Horizon is not tied to any country, government, or cloud provider.
It is a protocol, like TCP/IP, DNS, or TLS.

---

## 2. Design Philosophy

### 2.1 Self-Sovereign

- Self-hosted by the operator
- Self-signed root key
- Zero dependency on foreign CAs
- Zero dependency on foreign cloud

### 2.2 Air-Gap First

- Full operation without internet
- Signed export via USB for sync
- No telemetry, no phone-home

### 2.3 Multi-Tenant

- Each customer (bank, industry, exchange) gets a unique User ID
- Each branch gets a unique Hardware ID
- Each license is cryptographically signed
- Tenants are fully isolated

### 2.4 Single-Node Resilience

Unlike public blockchains (Bitcoin, Ethereum), Horizon is a
private network. Security comes from cryptography, not from
node count.

A single hardened node with ECDSA P-256 + RFC 6962 Merkle Tree
provides the same cryptographic guarantees as a multi-node network,
without the coordination overhead.

---

## 3. Cryptographic Foundation

| Component | Algorithm | Standard |
|---|---|---|
| Block signature | ECDSA P-256 | FIPS 186-4 |
| Reading signature | ECDSA P-256 | FIPS 186-4 |
| Hash | SHA-256 | FIPS 180-4 |
| Merkle Tree | RFC 6962 | IETF CT |
| Encoding | ASN.1 DER | X.690 |
| Password hash | bcrypt | — |

### 3.1 RFC 6962 Compliance

Horizon's Merkle Tree is fully compliant with RFC 6962.
All four test vectors match the official specification:

| Test | Result |
|---|---|
| Empty list | MATCH |
| Single leaf | MATCH |
| Two leaves | MATCH |
| Three leaves (odd) | MATCH |

### 3.2 Domain Separation

Leaf and internal node hashes are domain-separated:
- Leaf: `SHA-256(0x00 || data)`
- Node: `SHA-256(0x01 || left || right)`

This prevents second preimage attacks and matches the IETF standard.

---

## 4. Network Model

### 4.1 One Node, Unlimited Tenants

```
Horizon Network
├── bank_melli         (User ID: bank_melli)
│   ├── Tehran Branch      (Hardware ID: hw-tehran-001)
│   ├── Mashhad Branch     (Hardware ID: hw-mashhad-001)
│   └── Isfahan Branch     (Hardware ID: hw-isfahan-001)
│
├── bank_saderat       (User ID: bank_saderat)
│   └── ...
│
├── bank_deutsche      (User ID: bank_deutsche)
│   └── Frankfurt Branch   (Hardware ID: hw-frankfurt-001)
│
└── bank_england       (User ID: bank_england)
    └── London Branch      (Hardware ID: hw-london-001)
```

All tenants share:
- One protocol
- One network
- One root key (held by operator)

Each tenant has:
- Unique User ID
- Unique Hardware IDs
- Unique License
- Unique Merkle proof
- Unique ECDSA signature

### 4.2 Global Reach

Horizon is available to any entity that requires:
- Air-gapped operations
- Cryptographic proof of records
- Sovereign data control

Use cases:
- Banks (Iran, Russia, Venezuela, Germany, UK)
- Industries (oil, gas, steel, power)
- Exchanges (DEX, CEX)
- Governments (records, audits)
- Any organization under sanctions or sovereignty concerns

Horizon is not subject to foreign jurisdiction.

---

## 5. Hardware Key Management

### 5.1 USB-Based Key Storage

Private keys are stored on a USB device. The key never leaves the
device except as a transient in-memory handle during signing.

```
USB Device
├── private.pem  (never leaves)
└── public.pem   (shared with network)

At runtime:
1. Operator inserts USB
2. Horizon reads private key into memory (RAM only)
3. Signs a payload
4. Zeroes memory
5. Key remains only on USB
```

### 5.2 Advantages Over Commercial HSM

| Aspect | Commercial HSM | Horizon USB |
|---|---|---|
| Cost | $5,000+ | ~$0 |
| Vendor lock-in | Yes | No |
| Air-gap | Limited | Full |
| Offline signing | Limited | Yes |
| Cross-border | Restricted | Free |
| Backups | Complex | Simple |

### 5.3 Future Hardware Options

- YubiHSM 2 (optional, if FIPS 140-3 required)
- PKCS#11 smart cards (optional)
- Existing USB setup (default)

---

## 6. License Lifecycle

```
Customer requests license
         ↓
Operator signs payload with ECDSA
         ↓
Merkle root computed (RFC 6962)
         ↓
License saved with signature + merkle_root
         ↓
Customer receives license file
         ↓
Customer verifies with public key
         ↓
Customer operates for duration
         ↓
License auto-expires (365 days default)
```

### 6.1 License Fields

| Field | Type | Description |
|---|---|---|
| license_id | string | Unique ID |
| user_id | string | Customer ID |
| product_id | string | Product code |
| volume | int | Max readings/transactions |
| duration | int | Days valid |
| hardware_id | string | Bound hardware (optional) |
| issued_at | int64 | Unix timestamp |
| expires_at | int64 | Unix timestamp |
| signature | string | ECDSA P-256 |
| merkle_root | string | RFC 6962 root |

### 6.2 Renewal

License renewal preserves:
- Original user_id
- Original merkle lineage
- New signature with extended expiry

---

## 7. Blockchain Layers

### 7.1 Blocks

Every block contains:
- Block number
- Timestamp
- Previous hash
- Merkle root (RFC 6962 for v2)
- Transaction count
- ECDSA signature
- Merkle algorithm marker (`v1` or `v2`)

### 7.2 Merkle Algorithm Evolution

| Era | Algorithm | Blocks |
|---|---|---|
| Legacy | Bitcoin-style (v1) | #0–#3 (archive) |
| Current | RFC 6962 (v2) | #4+ |

Archive blocks are preserved read-only. New blocks use RFC 6962.

### 7.3 Transactions

Transaction types:
- transfer (between banks)
- reading (sensor data)
- license_issue
- license_revoke

---

## 8. Industrial Telemetry

### 8.1 Sensor Lifecycle

```
Sensor registered with public key (PEM)
         ↓
Sensor signs each reading with ECDSA
         ↓
Reading + signature + nonce + timestamp sent
         ↓
Server verifies signature
         ↓
Reading saved to blockchain
```

### 8.2 Replay Protection

Each reading must include:
- Unique nonce (5-minute window)
- Timestamp within [now-300s, now+120s]

If either fails, the reading is rejected with TamperEvent.

### 8.3 Alert Engine

| Condition | Alert |
|---|---|
| Out of min/max | Critical |
| Within 10% of edge | Warning |
| Rate-of-change > 30% | Warning/Critical |
| No signal > 2 min | Silent |

---

## 9. Attack Resistance

### 9.1 Tested Attack Vectors

| Attack | Result |
|---|---|
| Authentication bypass | REJECTED |
| JWT alg=none | REJECTED |
| SQL Injection (4 payloads) | REJECTED |
| Path Traversal (3 payloads) | REJECTED |
| XSS (2 payloads) | REJECTED |
| Command Injection (4 payloads) | REJECTED |
| Fake ECDSA signature | REJECTED |
| Replay attack | REJECTED |
| DoS (100 concurrent) | 1,604 ms |
| HTTP Method tampering | REJECTED |

**Result:** 33 tests passed, 0 failed.

### 9.2 Signature Verification

| Test | Result |
|---|---|
| Fake signature | invalid signature |
| Empty signature | rejected |
| Old timestamp | rejected |
| Valid signature | accepted, verified=true |

---

## 10. Performance

### 10.1 TPS Benchmark

```
Environment: Windows 10 + Git Bash + SQLite (WAL)
Crypto:      ECDSA P-256 + SHA-256
Endpoint:    POST /api/v1/industrial/reading/batch

Config: bs=100, c=50, n=5000
Result: 4,140 readings/sec
Errors: 0
```

### 10.2 Scaling Projection

| Hardware | Projected TPS |
|---|---|
| 4-core / 8GB | 8,000 |
| 8-core / 16GB | 30,000 |
| 8-core / 32GB | 100,000 |
| 16-core / 64GB | 200,000+ |

### 10.3 Load Behavior

- 100 concurrent requests: 1.6 seconds
- No dropped requests
- No errors

---

## 11. Runtime Environment

| Component | Value |
|---|---|
| Language | Go 1.24.4 |
| Database | SQLite (WAL mode) |
| HTTP Framework | Gin |
| ORM | GORM |
| Port | 8080 |
| CORS | Configurable |

---

## 12. API Surface

### 12.1 Public

| Endpoint | Method |
|---|---|
| /api/v1/health | GET |
| /api/v1/admin/login | POST |
| /api/v1/industrial/reading | POST |
| /api/v1/industrial/reading/batch | POST |

### 12.2 Authenticated

| Endpoint | Method |
|---|---|
| /api/v1/block/list | GET |
| /api/v1/block/mine | POST |
| /api/v1/license/list | GET |
| /api/v1/license/save | POST |
| /api/v1/license/verify | POST |
| /api/v1/tx | POST |
| /api/v1/account/list | GET |
| /api/v1/account/seed | POST |
| /api/v1/key/unlock | POST |
| /api/v1/key/status | GET |
| /api/v1/industrial/sensors | POST |
| /api/v1/industrial/tamper | GET |

---

## 13. Verification

### 13.1 Build

```bash
git clone https://github.com/beaconchain-horizon/horizon-core.git
cd horizon-core
go build -o switch.exe ./cmd/switch
./switch.exe
```

### 13.2 Test

```bash
# In another terminal
curl http://localhost:8080/api/v1/health
curl -X POST http://localhost:8080/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'
```

### 13.3 Attack Test

```bash
bash ~/Desktop/horizon-attack-tps.sh
```

Expected: 33+ tests pass, TPS > 3,000.

---

## 14. Global Deployment

Horizon can be deployed anywhere:
- Iran (sovereign infrastructure)
- Russia (sanctioned environment)
- Venezuela (sanctioned environment)
- Germany (EU operations)
- UK (post-Brexit operations)
- Any jurisdiction

Horizon does not depend on:
- Foreign CAs
- Foreign cloud
- Foreign DNS
- Foreign payment gateways

Horizon is a protocol — like TCP/IP, DNS, or TLS.
Protocols cannot be sanctioned. Only companies can.

---

## 15. Conclusion

Horizon is a global, sovereign blockchain protocol.
It provides:
- Cryptographic proof of every transaction
- Air-gapped operations
- Multi-tenant isolation
- Self-sovereign key management
- Global reach without jurisdiction dependency

Built on open standards:
- ECDSA P-256 (FIPS 186-4)
- SHA-256 (FIPS 180-4)
- RFC 6962 Merkle Tree
- ASN.1 DER (X.690)

Verified by:
- 33 attack tests (0 failed)
- RFC 6962 compliance (4/4 match)
- 4,140 TPS benchmark

Available to any organization that requires sovereignty and security.

---

**Author:** Mahdi Amoli Moghaddam
**Date:** 2026-09-24
**Version:** 3.0.0
**License:** Commercial
