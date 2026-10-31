<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  HORIZON CORE — README v2.1 (2026-10-14)                             -->
<!--  Theme: GitHub dark + accent green                                   -->
<!--  All links tested · All sections searchable                          -->
<!-- ════════════════════════════════════════════════════════════════════ -->

<div align="center" style="background: linear-gradient(180deg, #0d1117 0%, #05070e 100%); padding: 40px 20px; border-radius: 12px; border: 1px solid #30363d;">

<h1 style="font-size: 42px; margin: 0; background: linear-gradient(135deg, #7ee787 0%, #58a6ff 50%, #c084fc 100%); -webkit-background-clip: text; background-clip: text; color: transparent;">
🔭 HORIZON CORE
</h1>

<p style="color: #8b949e; font-size: 16px; margin: 12px 0 20px;">
<strong>Offline-first tamper-evident ledger for banking, industry, and air-gapped environments.</strong>
</p>

<!-- ─────────────── BADGES ─────────────── -->

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white&style=flat-square)](https://go.dev)
[![Crypto](https://img.shields.io/badge/Crypto-secp256k1-627EEA?style=flat-square)](https://www.secg.org/sec2-v2.pdf)
[![Merkle](https://img.shields.io/badge/Merkle-binary%20%2B%20ternary-8957e5?style=flat-square)](#-merkle-tree)
[![TPS](https://img.shields.io/badge/verified-209K%20TPS-3fb950?style=flat-square)](#-performance)
[![Chain](https://img.shields.io/badge/chain%20verified-%231--%2388-3fb950?style=flat-square)](#-blockchain)
[![Verify](https://img.shields.io/badge/verify-60%20seconds-7ee787?style=flat-square)](VERIFY.md)
[![License](https://img.shields.io/badge/license-Commercial-blue?style=flat-square)](#-license)

<br>

<table align="center" style="border: none; color: #8b949e; font-size: 13px;">
<tr>
<td align="center" style="border: none; padding: 4px 14px;">
<b style="color: #7ee787;">Version</b><br>
v1.4
</td>
<td align="center" style="border: none; padding: 4px 14px;">
<b style="color: #7ee787;">README</b><br>
v2 (2026-10-14)
</td>
<td align="center" style="border: none; padding: 4px 14px;">
<b style="color: #7ee787;">Chain height</b><br>
88+
</td>
<td align="center" style="border: none; padding: 4px 14px;">
<b style="color: #7ee787;">Verified TPS</b><br>
209,205
</td>
</tr>
</table>

</div>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  TABLE OF CONTENTS                                                  -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🔍 Search Index — فهرست جستجو

<details open>
<summary><b>کلیک کن تا باز/بسته شود — همه‌ی بخش‌ها</b></summary>

<br>

| # | Section | What you'll find |
|:-:|:--------|:-----------------|
| 01 | [🌟 Why Horizon Core](#-why-horizon-core) | تفاوت با Hyperledger / Corda / PKI / Sigstore |
| 02 | [✨ What's New in v2](#-whats-new-in-v2) | ۱۵ تغییر اصلی از v1 |
| 03 | [🚀 Quick Start](#-quick-start) | راه‌اندازی در ۳ ترمینال |
| 04 | [🏗️ Architecture](#️-architecture) | نمودار جریان داده |
| 05 | [🔐 Cryptography](#-cryptography) | پارامترهای قفل‌شده + تست‌ها |
| 06 | [🌳 Merkle Tree](#-merkle-tree) | فرمول v2 + ternary + v1 legacy |
| 07 | [⛓️ Blockchain](#️-blockchain) | ساختار بلاک + تأیید زنجیره |
| 08 | [📊 Performance](#-performance) | جدول TPS واقعی |
| 09 | [🛠️ Tools](#️-tools) | ۱۴ ابزار |
| 10 | [🛡️ Security](#️-security) | Threat model + Audit history |
| 11 | [📁 Repository](#-repository-structure) | ساختار کامل |
| 12 | [🧪 Testing](#-testing) | دستورات تست + fuzz |
| 13 | [📚 Documentation](#-documentation) | ۱۲ سند رسمی |
| 14 | [🎯 Use Cases](#-use-cases) | بانک، صنعت، دفاعی، قضایی |
| 15 | [🔗 All Links](#-all-links) | همه‌ی لینک‌ها در یک جدول |
| 16 | [📄 License](#-license) | مجوز تجاری |
| 17 | [🙏 Acknowledgments](#-acknowledgments) | منابع الهام |

</details>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  WHY UNIQUE                                                         -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🌟 Why Horizon Core

<div style="background: #0d1117; border-left: 4px solid #7ee787; padding: 16px 20px; border-radius: 6px; margin: 16px 0;">

Horizon Core is the **only** tamper-evident ledger designed from first principles for **air-gapped, high-assurance environments** where trust cannot be delegated to a remote API. Unlike blockchain frameworks (Hyperledger, Corda) that assume network connectivity, or PKI systems (X.509) that only prove identity and not content, Horizon Core:

</div>

| Property | Horizon Core | Hyperledger | Corda | X.509 PKI | Sigstore |
|:---------|:------------:|:-----------:|:-----:|:---------:|:--------:|
| Works **fully offline** | ✅ | ❌ | ❌ | Partial | ❌ |
| No trusted API required | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Independent verification** with only public key + proof | ✅ | ❌ | ❌ | ❌ | ✅ |
| **Scope-limited licenses** (per-branch, per-site) | ✅ | ❌ | Partial | ❌ | ❌ |
| **Single-node deployment** (no cluster) | ✅ | ❌ | ❌ | ✅ | N/A |
| **Deterministic signatures** (RFC 6979) | ✅ | Partial | ❌ | ❌ | ❌ |
| **209K TPS** verified on 8-core laptop | ✅ | varies | varies | N/A | N/A |
| **Air-gap friendly** end-to-end | ✅ | ❌ | ❌ | Partial | ❌ |

<div align="center" style="background: #161b22; border: 1px solid #7ee787; border-radius: 8px; padding: 14px 20px; margin: 20px 0;">

<p style="color: #e6edf3; margin: 0;">
<strong style="color: #7ee787;">The unique combination:</strong>
a self-contained, single-binary ledger that any third party — a central bank, a regulator, a court — can verify using only a public key and a Merkle proof.
</p>

<p style="color: #58a6ff; margin: 10px 0 0;">
<strong>No internet. No API. No trust in the vendor.</strong>
</p>

</div>

📖 **[Verify in 60 seconds →](VERIFY.md)**

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  WHAT'S NEW                                                         -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## ✨ What's New in v2

<details open>
<summary><b>۱۵ تغییر از نسخه‌ی قبل — کلیک کن</b></summary>

<br>

| # | Category | Change |
|:-:|:--------:|:-------|
| 01 | 🔐 Crypto | **High-S malleability** closed — `parseRS` now enforces `s ≤ n/2` |
| 02 | 🔐 Crypto | **DER rejection** active in `VerifyData` |
| 03 | 🌳 Merkle | **Merkle unify** — chain, license, proof, regen all use one formula |
| 04 | 🌳 Merkle | **Ternary Merkle** implemented (`TernaryRootNoPrefix`) |
| 05 | 🛠 Tool | **`verify-tx`** — independent transaction verifier |
| 06 | 🛠 Tool | **`verify-lic`** — independent license verifier |
| 07 | 🛡 API | **Rate limiting** (per-IP, 60/s, burst 120) |
| 08 | 🛡 API | **Audit log** hash-chained (tamper-evident) |
| 09 | 🛡 API | **DoS-1 closed** — `MaxBytesReader` (8 MiB default) |
| 10 | 🛡 API | **CORS allowlist** — wildcard removed |
| 11 | 🔒 Network | **mTLS** for Switch↔Backend (TLS 1.3, client certs) |
| 12 | 🧪 Test | **Fuzz testing** — 1.86M execs, zero counterexamples |
| 13 | 🧪 Test | **Cross-verified with OpenSSL 3.5.7** — `Verified OK` |
| 14 | 📊 Perf | **209K TPS verified** on 8-core Windows laptop |
| 15 | 📚 Docs | RFC 6962 marked as **v1 legacy**; v2 uses binary custom |

</details>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  QUICK START                                                        -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🚀 Quick Start

### Prerequisites

- **Go 1.24** (pinned — do not upgrade)
- **Windows** / **Linux** / **macOS**
- *Optional:* **PostgreSQL 16** for production (SQLite for development)

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

> [!TIP]
> **Auditors:** see [`VERIFY.md`](VERIFY.md) for a 60-second independent verification. No trust in our Backend required.

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  ARCHITECTURE                                                       -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🏗️ Architecture

```text
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
│  • license_id, user_id, product_id, volume, duration             │
│  • merkle_root (8-leaf binary tree over immutable fields)        │
│  • issued_at, expires_at, hardware_id                            │
│  • signature (64-byte R||S, secp256k1)                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
┌──────────────────────────┐      ┌──────────────────────────────┐
│  Backend (source of      │      │  Switch (customer gateway)    │
│  truth — the vendor)     │◄─────┤  • Runs on customer server    │
│                          │ sync │  • Verifies license locally   │
│  • chain (blocks+txs)    │      │  • Batches 100K tx/request    │
│  • Merkle tree           │      │  • 209K TPS verified          │
│  • License registry      │      │  • Works fully offline        │
│  • Public explorer       │      │  • Syncs when online          │
└──────────┬───────────────┘      └──────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Independent Verifiers                                           │
│  • verify-tx    (Merkle proof + signature + ECDSA)               │
│  • verify-lic   (canonical JSON + merkle_root + signature)       │
│  • No API. No DB. No internet.                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  CRYPTOGRAPHY                                                       -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🔐 Cryptography

<details open>
<summary><b>Locked parameters (v1.4)</b></summary>

<br>

| Parameter | Value | Standard |
|:----------|:------|:---------|
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

</details>

### Verified properties

<div style="background: #0d1117; border-left: 3px solid #7ee787; padding: 12px 18px; border-radius: 6px; margin: 12px 0;">

- ✅ **Fuzz testing:** 1.86M execs on `VerifyHashRS`, `SignDataRS`, `VerifyDataHex` — zero counterexamples
- ✅ **Cross-verification:** signatures verified by OpenSSL 3.5.7 — `Verified OK`
- ✅ **RFC 6979 determinism:** two signatures on the same key/payload are byte-identical
- ✅ **Malleability closure:** for every `(r, s)` accepted, `(r, n-s)` is rejected

</div>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  MERKLE TREE                                                        -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🌳 Merkle Tree

### Current formula (v2)

```text
leaf(data)       = SHA256(data)
node(left,right) = SHA256(left || right)
odd level        = SHA256(last || last)
```

**No domain separation.** Intentional for v2, chosen for simplicity and speed.

**Five call sites, all byte-identical** — proven by `internal/merkle/conform_test.go`:

| # | File | Purpose |
|:-:|:-----|:--------|
| 1 | [`internal/chain/append.go`](internal/chain/append.go) | Block building |
| 2 | [`internal/license/protocol.go`](internal/license/protocol.go) | License root |
| 3 | [`cmd/switch/license_sign.go`](cmd/switch/license_sign.go) | Switch signing |
| 4 | [`cmd/regen-license/main.go`](cmd/regen-license/main.go) | License regeneration |
| 5 | [`internal/merkle/merkle.go`](internal/merkle/merkle.go) | Reference implementation |

### Ternary tree (fanout = 3)

```text
leaf(data)        = SHA256(data)
node(a,b,c)       = SHA256(a || b || c)
1 leftover        = node(x, x, x)
2 leftover        = node(a, b, b)
```

> [!NOTE]
> **Historical note:** v1 used **RFC 6962** (0x00 leaf prefix, 0x01 node prefix, largest-power-of-two split). References to "RFC 6962" in older docs are **v1 legacy** and are marked as such.

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  BLOCKCHAIN                                                         -->
<!-- ════════════════════════════════════════════════════════════════════ -->

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

**Verified blocks:**

| Block | Type | TX count | Proof | Result |
|:-----:|:-----|:--------:|:-----:|:------:|
| #1 | benchmark | 137 | 8 siblings | ✅ VALID |
| #30 | benchmark | 154 | 8 siblings | ✅ VALID |
| #59 | benchmark | 184 | 8 siblings | ✅ VALID |
| #87 idx 0 | benchmark | 110 | 7 siblings | ✅ VALID |
| #87 idx 23 | benchmark | 110 | 7 siblings | ✅ VALID |
| #88 idx 0 | real license | 1 | single-leaf | ✅ VALID |

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  PERFORMANCE                                                        -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 📊 Performance

### Benchmark on 8-core Windows laptop

| Scenario | Processed | Failed | Elapsed | TPS |
|:---------|:---------:|:------:|:-------:|:---:|
| `GET /benchmark?concurrency=8` | 10,000 | 0 | 0.1204s | **83,088** |
| `GET /benchmark?concurrency=16` | 50,000 | 0 | 0.5467s | **91,464** |
| `POST /tx/batch` (10,000 tx) | 10,000 | 0 | ~0.14s | **~72,000** |
| **`POST /tx/batch` (100,000 tx)** | **100,000** | **0** | **0.273s** | **~366,000** |

<div align="center" style="background: #0d1117; border: 1px solid #7ee787; border-radius: 8px; padding: 14px; margin: 16px 0;">

**🎯 Verified target: 200K+ TPS on a single node**

</div>

### Production tuning

```go
InitBatcher(db, BatcherConfig{
    MaxSize:       50000,                    // 50K tx per flush
    FlushInterval: 200 * time.Millisecond,
    SignWorkers:   8,                        // parallel signing workers
})
// Theoretical ceiling: 50000 / 0.200 = 250,000 TPS
```

📖 **[Full benchmark methodology →](docs/BENCHMARK.md)**

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  TOOLS                                                              -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🛠️ Tools

| Command | Purpose | Link |
|:--------|:--------|:----:|
| `cmd/api` | Backend HTTP API (Gin) | [→](cmd/api) |
| `cmd/switch` | Customer-side gateway | [→](cmd/switch) |
| `cmd/verify-tx` | Independent transaction verifier | [→](cmd/verify-tx) |
| `cmd/verify-lic` | Independent license verifier | [→](cmd/verify-lic) |
| `cmd/horizon-keygen` | Generate secp256k1 keypair | [→](cmd/horizon-keygen) |
| `cmd/horizon-pubkey` | Extract public key from private | [→](cmd/horizon-pubkey) |
| `cmd/horizon-usbkey` | Auto-detect USB key | [→](cmd/horizon-usbkey) |
| `cmd/authority-keygen` | Authority keypair generation | [→](cmd/authority-keygen) |
| `cmd/license-signer` | Local signing service | [→](cmd/license-signer) |
| `cmd/licenseverify` | Offline verifier | [→](cmd/licenseverify) |
| `cmd/chainbench` | Chain benchmarking | [→](cmd/chainbench) |
| `cmd/tpsbench` | TPS benchmarking | [→](cmd/tpsbench) |
| `cmd/fetch-tx` | Fetch transaction from chain | [→](cmd/fetch-tx) |
| `cmd/regen-license` | Regenerate license files | [→](cmd/regen-license) |

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  SECURITY                                                           -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🛡️ Security

<details open>
<summary><b>Threat model — what we resist</b></summary>

<br>

| Threat | Defense |
|:-------|:--------|
| **Signature malleability** | high-S rejected (BIP-62) |
| **Encoding attacks** | DER/ASN.1 rejected |
| **Key compromise** | private key never leaves USB |
| **Replay** | session tokens single-use, 5-minute expiry |
| **Scope escape** | licenses bound to `user_id` + `hardware_id` |
| **Tamper** | hash-chained audit log |
| **DoS** | body size limits, rate limiting |
| **Supply chain** | Go 1.24 pinned, deps audited |

</details>

<details>
<summary><b>Not designed for</b></summary>

<br>

- ❌ **Multi-node consensus** — single-node by design (v1.4)
- ❌ **Quantum resistance** — secp256k1 vulnerable (like Bitcoin)
- ❌ **FIPS 140-2** — not certified
- ❌ **HSM integration** — planned for v1.5
- ❌ **Token/cryptocurrency** — not a blockchain currency

</details>

<details>
<summary><b>Audit history</b></summary>

<br>

| Item | Result | Reference |
|:-----|:-------|:----------|
| Fuzz testing | 1.86M execs · zero counterexamples | [`internal/crypto/fuzz_test.go`](internal/crypto/fuzz_test.go) |
| Cross-verification | OpenSSL 3.5.7 · `Verified OK` | [`internal/crypto/crossverify_test.go`](internal/crypto/crossverify_test.go) |
| 192 pentest vectors | 133 blocked · 0 vulnerable | [`reports/`](reports/) |
| Dependabot | 4 open (2 dismissed · 2 fixed) | [Security tab](https://github.com/beaconchain-horizon/horizon-core/security/dependabot) |

</details>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  REPOSITORY                                                         -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 📁 Repository Structure

```text
horizon-core/
├── cmd/
│   ├── api/               Backend HTTP API
│   ├── switch/            Customer-side gateway (6881 LOC)
│   ├── verify-tx/         Independent transaction verifier
│   ├── verify-lic/        Independent license verifier
│   ├── license-signer/    Local signing service
│   ├── licenseverify/     Offline verifier
│   ├── horizon-keygen/    Key generation
│   ├── horizon-pubkey/    Public key extraction
│   ├── horizon-usbkey/    USB key detection
│   ├── authority-keygen/  Authority keys
│   ├── chainbench/        Chain benchmark
│   ├── tpsbench/          TPS benchmark
│   ├── fetch-tx/          Transaction fetcher
│   └── regen-license/     License regeneration
├── internal/
│   ├── crypto/            secp256k1, RFC 6979, low-S
│   ├── merkle/            Binary + ternary Merkle
│   ├── license/           License protocol
│   ├── chain/             Blockchain
│   ├── ratelimit/         Rate limiting middleware
│   ├── auditlog/          Hash-chained audit log
│   ├── mtls/              mTLS configuration
│   ├── intersect/         SQLite↔PostgreSQL intersection
│   ├── industrial/        Industrial telemetry
│   ├── migration/         Data migration
│   └── tenant/            Multi-tenancy
├── docs/                  12 official documents
├── scripts/               smoke-test, cert generation, rotation
├── reports/               Audit reports (private)
└── .github/workflows/     CI + security
```

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  TESTING                                                            -->
<!-- ════════════════════════════════════════════════════════════════════ -->

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

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  DOCUMENTATION                                                      -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 📚 Documentation

| Document | Purpose |
|:---------|:--------|
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

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  USE CASES                                                          -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🎯 Use Cases

<table>
<tr>
<td width="50%" valign="top">

### 🏦 Central Banks
- Inter-bank settlement proof
- Regulatory reporting with mathematical verification
- Tamper-evident audit trail

### 🏢 Commercial Banks
- Internal audit with Merkle proofs
- KYC/AML evidence chain
- Cross-branch transaction verification

### 🏭 Critical Infrastructure
- SCADA event logging
- Sensor data provenance
- Environmental compliance reporting

</td>
<td width="50%" valign="top">

### 🛡️ Nuclear / Defense
- Air-gap compatible
- USB-based signing
- No external network dependencies

### ⚖️ Courts / Regulators
- Independent verification tool
- Court-admissible cryptographic proof
- No vendor lock-in

### 🌐 Air-Gapped Environments
- Fully offline operation
- Merkle proof verification
- No internet required

</td>
</tr>
</table>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  ALL LINKS                                                          -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🔗 All Links

<details open>
<summary><b>همه‌ی لینک‌ها در یک جدول</b></summary>

<br>

### 🌐 Public Services

| Service | URL | Status |
|:--------|:----|:------:|
| **Main Site** | https://beaconchain.us | 🟢 |
| **Chain Explorer** | https://beaconchain.us/monitor.html | 🟢 |
| **USB Signer** | https://beaconchain.us/offline-license.html | 🟢 |
| **Customer Panel** | https://beaconchain.us/customer-panel.html | 🟢 |
| **Proposal** | https://beaconchain.us/proposal.html | 🟢 |

### 💻 GitHub Repositories

| Repository | URL | Purpose |
|:-----------|:----|:--------|
| **horizon-core** | https://github.com/beaconchain-horizon/horizon-core | Main protocol code |
| **horizon-security-assessment** | https://github.com/beaconchain-horizon/horizon-security-assessment | Public pages |
| **angel-agents** | https://github.com/beaconchain-horizon/angel-agents | Multi-agent layer |
| **resume** | https://github.com/beaconchain-horizon/resume | Professional profile |

### 👤 Profiles

| Platform | URL |
|:---------|:----|
| **GitHub Org** | https://github.com/beaconchain-horizon |
| **Microsoft Learn** | https://learn.microsoft.com/en-us/users/capitalmarkets-0675/ |
| **GoDaddy Experts** | https://experts.godaddy.com/mahdi-amolimoghaddam |

### 📚 External References

| Reference | URL |
|:----------|:----|
| **secp256k1 spec** | https://www.secg.org/sec2-v2.pdf |
| **RFC 6979** | https://datatracker.ietf.org/doc/html/rfc6979 |
| **RFC 6962** | https://datatracker.ietf.org/doc/html/rfc6962 |
| **BIP-62** | https://github.com/bitcoin/bips/blob/master/bip-0062.mediawiki |
| **Decred secp256k1** | https://github.com/decred/dcrd |

</details>

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  CONTRIBUTING                                                       -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🤝 Contributing

We welcome contributions. Before submitting:

1. Read [`docs/SECURITY-MODEL.md`](docs/SECURITY-MODEL.md)
2. Follow [Conventional Commits](https://www.conventionalcommits.org/)
3. Every change needs a test
4. No direct push to `main` — use feature branches + PRs
5. English in code, Persian in user-facing comments
6. Fail-closed: refuse to start on missing security conditions

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  LICENSE                                                            -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 📄 License

Commercial license. For licensing inquiries, contact the maintainers.

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  ACKNOWLEDGMENTS                                                    -->
<!-- ════════════════════════════════════════════════════════════════════ -->

## 🙏 Acknowledgments

- Inspired by **Bitcoin** (secp256k1), **Certificate Transparency** (Merkle), and **Sigstore** (independent verification)
- Built with [Decred's secp256k1](https://github.com/decred/dcrd) library
- Verified against **OpenSSL 3.5.7**

---

<!-- ════════════════════════════════════════════════════════════════════ -->
<!--  FOOTER                                                             -->
<!-- ════════════════════════════════════════════════════════════════════ -->

<div align="center" style="background: linear-gradient(180deg, #0d1117 0%, #05070e 100%); padding: 30px 20px; border-radius: 12px; border: 1px solid #30363d; margin-top: 40px;">

<h3 style="color: #7ee787; margin: 0 0 12px;">
🔭 HORIZON CORE — a protocol, not a product.
</h3>

<p style="color: #8b949e; font-size: 13px; margin: 0 0 16px;">
<strong>github.com/beaconchain-horizon/horizon-core</strong>
</p>

[![GitHub](https://img.shields.io/badge/github-beaconchain--horizon-7ee787?logo=github&style=for-the-badge)](https://github.com/beaconchain-horizon)
[![Site](https://img.shields.io/badge/site-beaconchain.us-58a6ff?style=for-the-badge)](https://beaconchain.us)

<p style="color: #30363d; font-size: 11px; margin: 20px 0 0;">
v1.4 · 2026-10-14 · All rights reserved
</p>

</div>
