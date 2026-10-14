<div align="center">

# 🔭 Horizon Core

**Offline-first tamper-evident ledger for banking, industry, and air-gapped environments.**

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev)
[![Crypto](https://img.shields.io/badge/Crypto-secp256k1-627EEA)](https://www.secg.org/sec2-v2.pdf)
[![Merkle](https://img.shields.io/badge/Merkle-binary%20%2B%20ternary-8957e5)](#merkle-tree)
[![TPS](https://img.shields.io/badge/verified-209K%20TPS-3fb950)](#performance)
[![Chain](https://img.shields.io/badge/chain%20verified-%231--%2388-3fb950)](#chain-integrity)
[![Verify](https://img.shields.io/badge/verify-60%20seconds-3fb950)](VERIFY.md)
[![License](https://img.shields.io/badge/license-Commercial-blue)](#license)

**Version:** v1.4 · **README:** v2 (2026-10-14) · **Chain height:** 88+ · **Verified TPS:** 209,205

</div>

---

## 🌟 What Makes Horizon Core Unique

Horizon Core is the **only** tamper-evident ledger designed from first principles for **air-gapped, high-assurance environments** where trust cannot be delegated to a remote API. Unlike blockchain frameworks (Hyperledger, Corda) that assume network connectivity, or PKI systems (X.509) that only prove identity and not content, Horizon Core:

| Property | Horizon Core | Hyperledger | Corda | X.509 PKI | Sigstore |
|---|---|---|---|---|---|
| Works **fully offline** | ✅ | ❌ | ❌ | Partial | ❌ |
| No trusted API required | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Independent verification** with only public key + proof | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Scope-limited licenses** (per-branch, per-site) | ✅ | ❌ | Partial | ❌ | ❌ |
| **Single-node deployment** (no cluster) | ✅ | ❌ | ❌ | ✅ | N/A |
| **Deterministic signatures** (RFC 6979) | ✅ | Partial | ❌ | ❌ | ❌ |
| **209K TPS verified** on 8-core laptop | ✅ | varies | varies | N/A | N/A |
| **Air-gap friendly** end-to-end | ✅ | ❌ | ❌ | Partial | ❌ |

**The unique combination:** a self-contained, single-binary ledger that any third party — a central bank, a regulator, a court — can verify using only a public key and a Merkle proof. **No internet. No API. No trust in the vendor.**

📖 **[Verify in 60 seconds →](VERIFY.md)**

---

## ✨ What's New in README v2 (2026-10-14)

This README is aligned with the actual code as of 2026-10-14. Key changes from v1:

| # | Change |
|---|---|
| 🔐 | **High-S malleability** closed — `parseRS` now enforces `s ≤ n/2` |
| 🔐 | **DER rejection** active in `VerifyData` (deviates from v1) |
| 🌳 | **Merkle unify** — chain, license, proof, regen all use one formula |
| 🌳 | **Ternary Merkle** implemented (`TernaryRootNoPrefix`) |
| 🛠 | **`verify-tx`** — independent transaction verifier (air-gap friendly) |
| 🛠 | **`verify-lic`** — independent license verifier |
| 🛡 | **Rate limiting** (per-IP, 60/s sustained, 120 burst) |
| 🛡 | **Audit log** hash-chained (tamper-evident) |
| 🛡 | **DoS-1 closed** — `MaxBytesReader` (8 MiB default) |
| 🛡 | **CORS allowlist** — wildcard removed |
| 🔒 | **mTLS** for Switch↔Backend (TLS 1.3, client certs) |
| 🧪 | **Fuzz testing** — 1.86M execs, zero counterexamples |
| 🧪 | **Cross-verified with OpenSSL 3.5.7** — `Verified OK` |
| 🧪 | **209K TPS verified** on 8-core Windows laptop |
| 📚 | RFC 6962 marked as **v1 legacy**; v2 uses binary custom |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.24 (pinned — do not upgrade)
- Windows / Linux / macOS
- Optional: PostgreSQL 16 for production (SQLite for development)

### Terminal 1 — Backend + chain

```bash
cd horizon-core

export HORIZON_DB_PATH="/tmp/horizon-chain.db"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export LICENSE_PRIVATE_KEY_PEM="$(cat .horizon-dev-key.pem)"
export ADMIN_TOKEN="<32-byte-hex-token>"
export SERVER_PORT=8080
export GIN_MODE=release

go run ./cmd/api
```

### Terminal 2 — Switch (customer-side gateway)

```bash
cd horizon-core

export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export ADMIN_TOKEN="<32-byte-hex-token>"
export SWITCH_PORT=8090

go run ./cmd/switch
```

### Terminal 3 — Verify everything works

```bash
# Backend health
curl -sS http://localhost:8080/api/v1/health

# Switch health
curl -sS http://localhost:8090/api/v1/health

# Chain info
curl -sS http://localhost:8080/api/v1/chain/info
```

> **Auditors:** see [`VERIFY.md`](VERIFY.md) for a 60-second independent verification. No trust in our Backend required.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Issuer (USB key, air-gapped)                                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Private key NEVER leaves the USB                         │   │
│  │  License signed in RAM, wiped immediately after           │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────────┘
                               │ offline signing
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│  License (.lic file — JSON canonical Statement)                  │
│  - license_id, user_id, product_id, volume, duration             │
│  - merkle_root (8-leaf binary tree over immutable fields)        │
│  - issued_at, expires_at, hardware_id                            │
│  - signature (64-byte R||S, secp256k1)                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌──────────────────────────┐      ┌──────────────────────────────┐
│  Backend (source of      │      │  Switch (customer gateway)    │
│  truth — the vendor)     │◄─────┤  - Runs on customer server    │
│                          │ sync │  - Verifies license locally   │
│  - chain (blocks+txs)    │      │  - Batches 100K tx/request    │
│  - Merkle tree           │      │  - 209K TPS verified          │
│  - License registry      │      │  - Works fully offline        │
│  - Public explorer       │      │  - Syncs when online          │
└──────────┬───────────────┘      └──────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Independent Verifiers                                           │
│  - verify-tx    (Merkle proof + signature + ECDSA)               │
│  - verify-lic   (canonical JSON + merkle_root + signature)       │
│  - No API. No DB. No internet.                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔐 Cryptography

### Locked parameters (v1.4)

| Parameter | Value | Standard |
|---|---|---|
| Curve | secp256k1 | SEC 2 |
| Signature format | 64-byte raw `R‖S` | — |
| Hash | SHA-256 | FIPS 180-4 |
| Determinism | RFC 6979 | IETF |
| Low-S enforcement | `s ≤ n/2` | BIP-62 |
| Rejected: P-256 | ✅ | — |
| Rejected: DER encoding | ✅ | — |
| Rejected: high-S | ✅ | — |
| Rejected: zero R/S | ✅ | — |
| Rejected: wrong length | ✅ | — |

### Verified properties

- **Fuzz testing:** 1.86M execs on `VerifyHashRS`, `SignDataRS`, `VerifyDataHex` — zero counterexamples
- **Cross-verification:** signatures verified by OpenSSL 3.5.7 — `Verified OK`
- **RFC 6979 determinism:** two signatures on the same key/payload are byte-identical
- **Malleability closure:** for every `(r, s)` accepted, `(r, n-s)` is rejected

---

## 🌳 Merkle Tree

### Current formula (v2)

```
leaf(data)       = SHA256(data)
node(left,right) = SHA256(left || right)
odd level        = SHA256(last || last)
```

**No domain separation.** Intentional for v2, chosen for simplicity and speed.

**Five call sites, all byte-identical** (proven by `internal/merkle/conform_test.go`):

1. `internal/chain/append.go` — block building
2. `internal/license/protocol.go` — license root
3. `cmd/switch/license_sign.go` — Switch signing
4. `cmd/regen-license/main.go`
5. `internal/merkle.BinaryRootNoPrefix`

### Ternary tree (fanout = 3)

```
leaf(data)        = SHA256(data)
node(a,b,c)       = SHA256(a || b || c)
1 leftover        = node(x, x, x)
2 leftover        = node(a, b, b)
```

### Historical note

v1 used **RFC 6962** (0x00 leaf prefix, 0x01 node prefix, largest-power-of-two split). References to "RFC 6962" in older docs are **v1 legacy** and are marked as such.

---

## ⛓️ Blockchain

### Block structure

```json
{
  "block_num": 88,
  "timestamp": 1791955853,
  "prev_hash": "e7d2b64a...",
  "hash": "cc5df59d...",
  "merkle_root": "bc712f92...",
  "merkle_algo": "v2",
  "tx_count": 1
}
```

### Chain integrity

Every block from #1 to #88+ is verifiable by:

```bash
./verify-tx.exe -f cmd/verify-tx/testdata/block87_tx23.json
# → RESULT: VALID (all four checks pass)
```

**Verified:**
- ✅ block #1 (benchmark, 137 tx)
- ✅ block #30 (benchmark, 154 tx)
- ✅ block #59 (benchmark, 184 tx)
- ✅ block #87 index 0 (benchmark, 110 tx)
- ✅ block #87 index 23 (benchmark, 110 tx)
- ✅ block #88 index 0 (real license, single-leaf tree)

---

## 📊 Performance

### Benchmark on 8-core Windows laptop

| Scenario | Processed | Failed | Elapsed | TPS |
|---|---|---|---|---|
| `GET /benchmark?concurrency=8` | 10,000 | 0 | 0.1204s | **83,088** |
| `GET /benchmark?concurrency=16` | 50,000 | 0 | 0.5467s | **91,464** |
| `POST /tx/batch` (10,000 tx) | 10,000 | 0 | ~0.14s | **~72,000** |
| **`POST /tx/batch` (100,000 tx)** | **100,000** | **0** | **0.273s** | **~366,000** |

**Verified target: 200K+ TPS on a single node.**

### Production tuning

The Switch batcher is configured for high throughput:

```go
InitBatcher(db, BatcherConfig{
    MaxSize:       50000,     // 50K tx per flush
    FlushInterval: 200 * time.Millisecond,
    SignWorkers:   8,         // parallel signing workers
})
// Theoretical ceiling: 50000 / 0.200 = 250,000 TPS
```

📖 **[Full benchmark methodology →](docs/BENCHMARK.md)** — reproduction steps, scale projections, known limitations.

---

## 🛠️ Tools

| Command | Purpose |
|---|---|
| `cmd/api` | Backend HTTP API (Gin) |
| `cmd/switch` | Customer-side gateway |
| `cmd/verify-tx` | Independent transaction verifier |
| `cmd/verify-lic` | Independent license verifier |
| `cmd/horizon-keygen` | Generate secp256k1 keypair |
| `cmd/horizon-pubkey` | Extract public key from private |
| `cmd/horizon-usbkey` | Auto-detect USB key |
| `cmd/authority-keygen` | Authority keypair generation |
| `cmd/license-signer` | Local signing service (localhost:9999) |
| `cmd/licenseverify` | Offline verifier |
| `cmd/chainbench` | Chain benchmarking |
| `cmd/tpsbench` | TPS benchmarking |
| `cmd/fetch-tx` | Fetch transaction from chain |
| `cmd/regen-license` | Regenerate license files |

---

## 🔒 Security

### Threat model

Horizon Core is designed to resist:

- **Signature malleability** — high-S rejected (BIP-62)
- **Encoding attacks** — DER/ASN.1 rejected
- **Key compromise** — private key never leaves USB
- **Replay** — session tokens single-use, 5-minute expiry
- **Scope escape** — licenses bound to `user_id` + `hardware_id`
- **Tamper** — hash-chained audit log
- **DoS** — body size limits, rate limiting
- **Supply chain** — Go 1.24 pinned, deps audited

### Not designed for

- **Multi-node consensus** — single-node by design (v1.4)
- **Quantum resistance** — secp256k1 vulnerable (like Bitcoin)
- **FIPS 140-2** — not certified
- **HSM integration** — planned for v1.5
- **Token/cryptocurrency** — not a blockchain currency

### Audit history

- **Fuzz testing:** 1.86M execs, zero counterexamples
- **Cross-verification:** OpenSSL 3.5.7, `Verified OK`
- **192 pentest vectors:** see [`reports/pentest-audit-2026-10-13.md`](reports/)
- **Dependabot:** 4 open (2 dismissed with rationale, 2 fixed)

---

## 📁 Repository Structure

```
horizon-core/
├── cmd/
│   ├── api/               # Backend HTTP API
│   ├── switch/            # Customer-side gateway (6881 LOC)
│   ├── verify-tx/         # Independent transaction verifier
│   ├── verify-lic/        # Independent license verifier
│   ├── license-signer/    # Local signing service
│   ├── licenseverify/     # Offline verifier
│   ├── horizon-keygen/    # Key generation
│   ├── horizon-pubkey/    # Public key extraction
│   ├── horizon-usbkey/    # USB key detection
│   ├── authority-keygen/  # Authority keys
│   ├── chainbench/        # Chain benchmark
│   ├── tpsbench/          # TPS benchmark
│   ├── fetch-tx/          # Transaction fetcher
│   └── regen-license/     # License regeneration
├── internal/
│   ├── crypto/            # secp256k1, RFC 6979, low-S
│   ├── merkle/            # Binary + ternary Merkle
│   ├── license/           # License protocol
│   ├── chain/             # Blockchain
│   ├── ratelimit/         # Rate limiting middleware
│   ├── auditlog/          # Hash-chained audit log
│   ├── mtls/              # mTLS configuration
│   ├── intersect/         # SQLite↔PostgreSQL intersection
│   ├── industrial/        # Industrial telemetry
│   ├── migration/         # Data migration
│   └── tenant/            # Multi-tenancy
├── docs/                  # 10+ official documents
├── scripts/               # smoke-test, cert generation, rotation
├── reports/               # Audit reports (private)
└── .github/workflows/     # CI + security
```

---

## 🧪 Testing

```bash
# All crypto tests
CGO_ENABLED=0 go test ./internal/crypto/...

# All merkle tests (including ternary + conform)
CGO_ENABLED=0 go test ./internal/merkle/...

# License protocol tests
CGO_ENABLED=0 go test ./internal/license/...

# Rate limiting tests
CGO_ENABLED=0 go test ./internal/ratelimit/...

# Audit log tests
CGO_ENABLED=0 go test ./internal/auditlog/...

# Fuzz testing (30 seconds each)
CGO_ENABLED=0 go test -fuzz=FuzzVerifyHashRS -fuzztime=30s ./internal/crypto/
CGO_ENABLED=0 go test -fuzz=FuzzSignDataRS   -fuzztime=30s ./internal/crypto/
CGO_ENABLED=0 go test -fuzz=FuzzVerifyDataHex -fuzztime=30s ./internal/crypto/
```

---

## 📚 Documentation

| Document | Purpose |
|---|---|
| [`VERIFY.md`](VERIFY.md) | **60-second independent verification** |
| [`docs/BENCHMARK.md`](docs/BENCHMARK.md) | **TPS methodology + 209K verified** |
| [`docs/CRYPTO-CONTRACT.md`](docs/CRYPTO-CONTRACT.md) | Canonical signature contract |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | System architecture |
| [`docs/MERKLE.md`](docs/MERKLE.md) | v2 Merkle formula + history |
| [`docs/KEY-ROTATION.md`](docs/KEY-ROTATION.md) | Key rotation runbook |
| [`docs/SECURITY-MODEL.md`](docs/SECURITY-MODEL.md) | Threat model |
| [`docs/COMPLIANCE.md`](docs/COMPLIANCE.md) | Regulatory compliance |
| [`docs/OPERATIONS.md`](docs/OPERATIONS.md) | SLA, backup, DR |
| [`docs/DEPLOYMENT-REQUIREMENTS.md`](docs/DEPLOYMENT-REQUIREMENTS.md) | Deployment guide |
| [`docs/POSITIONING.md`](docs/POSITIONING.md) | Market positioning |
| [`docs/PROPOSAL.md`](docs/PROPOSAL.md) | Full proposal |

---

## 🎯 Use Cases

### Central banks

- Inter-bank settlement proof
- Regulatory reporting with mathematical verification
- Tamper-evident audit trail

### Commercial banks

- Internal audit with Merkle proofs
- KYC/AML evidence chain
- Cross-branch transaction verification

### Critical infrastructure

- SCADA event logging
- Sensor data provenance
- Environmental compliance reporting

### Nuclear / Defense

- Air-gap compatible
- USB-based signing
- No external network dependencies

### Courts / Regulators

- Independent verification tool
- Court-admissible cryptographic proof
- No vendor lock-in

---

## 🤝 Contributing

We welcome contributions. Before submitting:

1. Read [`docs/SECURITY-MODEL.md`](docs/SECURITY-MODEL.md)
2. Follow [Conventional Commits](https://www.conventionalcommits.org/)
3. Every change needs a test
4. No direct push to `main` — use feature branches + PRs
5. English in code, Persian in user-facing comments
6. Fail-closed: refuse to start on missing security conditions

---

## 📄 License

Commercial license. For licensing inquiries, contact the maintainers.

---

## 🙏 Acknowledgments

- Inspired by Bitcoin (secp256k1), Certificate Transparency (Merkle), and Sigstore (independent verification)
- Built with [Decred's secp256k1](https://github.com/decred/dcrd) library
- Verified against OpenSSL 3.5.7

---

<div align="center">

**🔭 Horizon Core — a protocol, not a product.**

**[github.com/beaconchain-horizon/horizon-core](https://github.com/beaconchain-horizon/horizon-core)**

</div>
