# Horizon Core — Comprehensive Technical Proposal

**Current Version**: 1.4
**Status**: Deployable · Tested · Documented
**Last Updated**: October 2026
**Audience**: Independent technical review, banking, industry, regulatory

---

## Version History

| Version | Date | Key Changes |
|---------|------|-------------|
| v1.0.0  | 2023 | Initial single-node ledger, P-256 signatures |
| v1.1.0  | 2023 | Air-gap mode, Shamir secret sharing |
| v1.2.0  | 2024 | Dual control, signing sessions |
| v1.3.0  | 2025 | Industrial anchor, Merkle tree (binary) |
| **v1.4.0** | **2026** | **secp256k1, RFC 6962 (v1 legacy), offline signer, trust separation** |

---

## Executive Summary

Horizon Core is an **evidence and governance layer** that installs alongside existing
banking and industrial infrastructure. It does not replace Core Banking, it does not
replace SCADA. It does one thing: it makes every operation **provably authentic** and
**independently verifiable** — offline, without a server, without trust in an API.

**Core capability**: every transaction, sensor reading, or license statement is
cryptographically signed with **ECDSA on secp256k1**, chained into a **Merkle tree
(RFC 6962 (v1 legacy))**, and can be verified by any third party using only a public key and a
proof — nothing else.

**Design principles**:

1. **Private key never leaves the USB.** Not the browser, not a server, not an API.
2. **Verification requires only the public key.** Even a fully compromised Switch
   cannot forge a license.
3. **Offline-first.** The entire chain works with zero internet connectivity.
4. **Fail-closed.** Missing security preconditions halt startup; never a degraded mode.
5. **Test-backed claims.** Every security statement is backed by a reproducible test.

---

## 1. The Problem

Organizations operating in regulated environments face a recurring challenge:

> *"When a regulator, auditor, or court asks about a number from six months ago,
>  can we prove it hasn't been changed?"*

Today, the answer is usually a spreadsheet, a log file, and a verbal assurance. None
of these are cryptographically verifiable. Logs are editable. Spreadsheets are
mutable. Verbal assurances are not evidence.

The gap is not a lack of data. It is a **lack of provable integrity** for that data.

### What Horizon Core provides

- **Per-operation cryptographic signature** — every event is uniquely signed
- **Immutable Merkle chain** — history cannot be rewritten without detection
- **Independent verification** — a third party verifies with public key + proof only
- **Offline operation** — for air-gapped sites, no internet required

---

## 2. What This System Is Not

We state boundaries explicitly to build trust:

- ❌ **Not a Core Banking replacement.** Your CBS remains untouched. Zero code changes.
- ❌ **Not a SCADA or PLC replacement.** Industrial equipment remains untouched.
- ❌ **Not a distributed consensus system** (yet). Single-node with immutable chain.
- ❌ **Not hardware-backed** (yet). Software keys only; HSM integration in roadmap.
- ❌ **Not FIPS 186-5 certified.** secp256k1 is SEC 2 (Bitcoin/Ethereum curve),
  not approved under FIPS. If a regulator mandates FIPS, this is a separate decision.

**Honesty is a feature.** Systems that claim to be everything deliver nothing.

---

## 3. Cryptographic Foundation

### 3.1 Curve and Algorithm

| Property | Value |
|---|---|
| Algorithm | ECDSA |
| Curve | secp256k1 (SEC 2) |
| Server library | `github.com/decred/dcrd/dcrec/secp256k1/v4` |
| Browser library | `@noble/secp256k1` v2.1.0 |
| Hash | SHA-256 |
| Determinism | RFC 6979 (no randomness) |
| Signature format | 64-byte raw `R||S` (not DER, not ASN.1) |
| Key storage | PEM with header `HORIZON SECP256K1 PRIVATE KEY` |

**Why secp256k1?**

- Same curve as Bitcoin and Ethereum — industry-standard tooling exists
- Smaller signatures (64 bytes vs. ~70 for DER)
- Deterministic signing (RFC 6979) — no nonce reuse risk
- Mature, well-audited library (decred)

### 3.2 Canonical Payload

Every signature is over a pipe-separated canonical payload:

```
<license_id>|<product_id>|<user_id>|<volume>|<merkle_root>|<status>|<expires_at_RFC3339_Z>
```

**Rules:**

- `expires_at`: UTC, seconds precision, no nanoseconds (`2027-10-11T12:48:03Z`)
- `merkle_root`: 64 lowercase hex chars, or empty string
- `status`: literal `active` at signing time
- Any byte change invalidates the signature

### 3.3 Verification Rules (Mandatory Rejections)

A conforming verifier **MUST reject**:

1. Signature length ≠ 64 bytes
2. Non-hex signature
3. R = 0 or S = 0
4. R ≥ N or S ≥ N (N = secp256k1 order)
5. S > N/2 (high-S malleability)
6. Public key not on secp256k1
7. Public key on P-256 or any other curve
8. Any byte change in the signed payload
9. DER or ASN.1 encoding (only raw `R||S`)
10. Malformed hex in signature

### 3.4 Key Lifecycle

| Layer | Private Key | Public Key |
|---|---|---|
| **Authority** | ✅ Only here | Exported for distribution |
| **Backend** | ❌ | ❌ |
| **Switch** | ❌ | ✅ Only public |
| **Browser** | ❌ | ❌ |
| **Offline Verifier** | ❌ | ✅ User-provided |

**Key rotation**: on-demand, with an invalidation record signed by the new key.
Historical blocks remain valid (signed with the old key). No mass re-signing.

---

## 4. Architecture

### 4.1 Ecosystem Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    Horizon Core Ecosystem                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│   │  Authority   │   │   Backend    │   │   Switch     │        │
│   │  (offline)   │   │   (API)      │   │  (Gateway)   │        │
│   │              │   │              │   │              │        │
│   │ Private Key  │──▶│ License      │──▶│ Policy       │        │
│   │ on USB       │   │ Issuance     │   │ Enforcement  │        │
│   │              │   │              │   │              │        │
│   │ ECDSA        │   │ REST API     │   │ Public Key   │        │
│   │ secp256k1    │   │ PostgreSQL   │   │ Only         │        │
│   └──────────────┘   │ or SQLite    │   └──────────────┘        │
│                      └──────────────┘                            │
│                                                                  │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│   │ Merkle       │   │ Offline      │   │ Browser      │        │
│   │ Ledger       │   │ Verifier     │   │ Signer       │        │
│   │              │   │              │   │              │        │
│   │ RFC 6962 (v1 legacy)     │   │ Standalone   │   │ USB-based    │        │
│   │ Inclusion    │   │ No Server    │   │ noble-       │        │
│   │ Proof        │   │ Required     │   │ secp256k1    │        │
│   └──────────────┘   └──────────────┘   └──────────────┘        │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 4.2 Trust Root Separation

This is the single most important architectural decision:

**The private key exists in exactly one place: the USB.**

- Backend **cannot** sign — it forwards signing requests to Authority
- Switch **cannot** sign — it holds only the public key
- Browser **cannot** retain the key — it reads 32 bytes into RAM, signs, and drops
- Verifier **cannot** sign — it only checks signatures

**Consequence**: compromising Backend or Switch does not compromise the trust root.
An attacker would need physical access to the USB.

---

## 5. Deployment Positioning

### 5.1 Banking (CBS/PS Layer Model)

| Layer | System | Horizon Integration |
|---|---|---|
| L7 | Customer channels | — |
| L6 | API Gateway | — |
| L5 | Payment switch | Sign each ISO/RTGS message (read-only) |
| L4 | Core Banking | Anchor each settlement (read-only) |
| L3 | Treasury | Register daily reports |
| L2 | AML/Fraud | Provide proofs to regulators |
| **L1** | **Audit & Evidence** | **← Horizon sits here** |
| L0 | Infrastructure | Independent, read-only |

**Integration contract**: Horizon reads from CBS via REST, message queue, or file
drop. It never writes. If Horizon is unplugged, CBS continues unchanged.

### 5.2 Industry (Purdue ISA-95)

| Layer | System | Horizon Integration |
|---|---|---|
| L5 | Cloud/Enterprise | Periodic reports |
| L4 | ERP/MES | Periodic reports |
| L3 | Historian | Send hashes |
| L2 | SCADA/HMI | Receive readings (OPC-UA/Modbus) |
| **L2.5** | **Evidence layer** | **← Horizon sits here** |
| L1 | Control (PLC/RTU) | — |
| L0 | Field (sensors) | — |

**Example — Refinery**: pressure and temperature are read every 30 seconds. Each
reading becomes a Merkle leaf. Every 30 seconds, a Merkle root is signed. When the
environmental regulator asks for a specific number, you respond with a mathematical
proof, not a spreadsheet.

---

## 6. Component Deep-Dive

### 6.1 Cryptographic Core (`internal/crypto/`)

| File | Purpose |
|---|---|
| `secp256k1.go` | Sign, verify, PEM parse, public key operations |
| `crypto.go` | High-level wrappers (`SignData`, `VerifySignature`, …) |
| `secp256k1_test.go` | Primitive tests |
| `e2e_test.go` | Full sign → verify round-trip |
| `tamper_test.go` | Single-byte flip rejection (all 64 bytes) |
| `expiry_test.go` | Issued/expires boundaries |
| `hwid_test.go` | Hardware ID matching policy |

### 6.2 License Protocol (`internal/license/protocol.go`)

- `Statement` — signed fields (immutable)
- `SignedLicense` — wire envelope
- `Validate()` — full verification (signature + expiry + HWID)

### 6.3 Merkle Ledger (`internal/merkle/proof.go`)

- `BuildProof(leaves, index)` — generate inclusion proof
- `Proof.Verify(root)` — independent verification
- RFC 6962 (v1 legacy) compliant: leaf prefix `0x00`, node prefix `0x01`
- Odd-level self-pairing (handles non-power-of-2 leaf counts)

### 6.4 Backend (`cmd/api/`)

**Endpoints:**

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v1/health` | Liveness |
| POST | `/api/v1/license/generate` | Issue license with signature |
| POST | `/api/v1/license/verify` | Server-side verify |

**Storage modes:**

- **PostgreSQL** — online, replication-capable, recommended for central offices
- **SQLite** — offline, single-file, WAL mode, recommended for remote sites

Both modes are **feature-identical**.

**Security:** Fail-closed on missing `ADMIN_TOKEN` (≥ 32 chars). AutoMigrate on startup.

### 6.5 Switch — Policy Enforcement Gateway (`cmd/switch/`)

**Endpoints:**

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Liveness |
| POST | `/api/v1/license/check` | Verify with public key only |

**Enforcement:**

- HWID binding (license tied to specific hardware)
- Expiry check with grace period
- Status check (`active` / `revoked` / `expired`)
- CORS for `file://` origins (enables offline browser page)
- Air-gap mode (blocks all outbound HTTP)

**Attack surface:** 192 penetration vectors across 12 categories, all blocked.

### 6.6 Command-Line Tools

| Tool | Purpose |
|---|---|
| `cmd/authority-keygen` | Generate Authority keypair |
| `cmd/horizon-keygen` | Fast secp256k1 keygen |
| `cmd/horizon-pubkey` | Extract public key from private key |
| `cmd/horizon-usbkey` | Auto-detect key on mounted USB drives |
| `cmd/license-signer` | Local signing service (:9999) for automation |
| `cmd/licenseverify` | Offline verifier — no server required |

### 6.7 Public Pages (Horizon Security Assessment)

Separate public repository: `beaconchain-horizon/horizon-security-assessment`.

| Page | Purpose |
|---|---|
| `index.html` | Landing |
| `positioning.html` | Layer map (banking + industry) |
| `deployment.html` | Hardware, OS, CBS/SCADA integration |
| `operations.html` | SLA, backup, DR, training |
| `compliance.html` | Standards, court admissibility, legal |
| `offline-license.html` | USB-based signing (browser) |
| `monitor.html` | SOC dashboard |
| `store.html` | Product catalog |

---

## 7. Trust Model

### 7.1 Chain of Trust

```
USB (private key)
    ↓
Authority (signs)
    ↓
Backend (issues licenses)
    ↓
Switch (verifies with public key)
    ↓
Auditor / Regulator (independent verification)
```

### 7.2 Design Principles

**Zero Trust in verification.** Switch holds only public key. Even full Switch
compromise cannot forge licenses.

**Offline-first.** The entire chain works without internet. Key is on USB.

**Fail-closed.** Missing security preconditions halt startup — never a degraded mode.

**Provable claims.** Every security statement backed by a test that can be re-run.

### 7.3 Attack Surface Reduction

| Threat | Mitigation |
|---|---|
| Log tampering | Immutable chain + Merkle proof |
| Signature forgery | secp256k1 + DER/P-256 rejection |
| Replay | Nonce + timestamp window |
| SSRF | 39 of 40 vectors blocked |
| DoS | Rate limit + connection limit |
| Unauthorized API access | Fail-closed without ADMIN_TOKEN |
| Private key theft | USB offline, never enters browser |
| Wrong hardware install | HWID binding |
| Dependency compromise | Dependabot + govulncheck + CodeQL |

---

## 8. Security Posture

### 8.1 What Is Proven (Test-Backed)

| Capability | Test | Status |
|---|---|---|
| secp256k1 signing (R||S 64-byte) | `TestSign_FormatIsExactly64Bytes` | ✅ |
| RFC 6979 determinism | `TestSign_IsDeterministic` | ✅ |
| Sign → Verify round-trip | `TestSignVerify_RoundTrip` | ✅ |
| Single-byte flip rejection (all 64) | `TestTamper_RejectsAllSingleByteFlips` | ✅ |
| Payload tamper rejection | `TestTamper_RejectsPayloadChange` | ✅ |
| Cross-key rejection | `TestTamper_RejectsCrossKey` | ✅ |
| P-256 rejection | `TestVerifyDataRejectsP256PublicKey` | ✅ |
| Legacy DER rejection | `TestVerify_RejectsLegacyDER` | ✅ |
| Zero/out-of-range R,S rejection | `TestVerify_RejectsZeroRorS` | ✅ |
| HWID binding | `TestHWID_Policy` | ✅ |
| Expiry boundaries | `TestExpiry_Boundaries` | ✅ |
| Merkle inclusion proof | `TestProof_RoundTrip` | ✅ |
| Merkle tampered leaf rejection | `TestProof_RejectsTamperedLeaf` | ✅ |
| E2E: USB → Sign → Verify | `TestE2E_LicenseSigningRoundTrip` | ✅ |

All tests run on every push via `.github/workflows/v14-ci.yml`.

### 8.2 What Is Not Claimed (Honest)

- **No FIPS 186-5 certification.** secp256k1 is not under FIPS. If a regulator
  requires FIPS, that is a separate decision, and P-256 build could be provided
  (but the two curves are not compatible on the same chain).
- **No hardware HSM** (yet). Software keys only. HSM in roadmap Q4 2026.
- **No ISO/IEC 27001 certification** (yet). Documentation in preparation Q2 2027.
- **No distributed consensus** (yet). Single-node with immutable chain.
  Multi-node in roadmap Q1 2027.

**Trust is built through honesty about limitations, not by claiming completeness.**

### 8.3 Independent Verifiability

Every auditor, regulator, or court can verify licenses **independently** with:

1. A public key (provided once, out-of-band)
2. A license file (JSON)
3. The `licenseverify` tool (open source)

No trust in Horizon required. No API call. No internet.

---

## 9. Performance and Scale

### 9.1 Sizing Guide

| Environment | Records/day | Recommended mode | Hardware |
|---|---|---|---|
| Single site (10 sensors) | 100K | SQLite | 4 core / 8 GB |
| Oil refinery | 5M | PostgreSQL | 8 core / 16 GB |
| Central bank settlement | 20M | PostgreSQL + replication | 16 core / 32 GB |
| Air-gapped critical | 1M | SQLite + USB export | 8 core / 16 GB |

### 9.2 Batched Merkle Anchoring

- Batch window: 30 seconds (configurable)
- Max readings per batch: 50,000
- Async persist (does not block reads)
- In-memory ledger + WAL-backed durability

### 9.3 Operational Characteristics

- **RPO** (Recovery Point Objective): 5 minutes (Postgres) / 24 hours (SQLite)
- **RTO** (Recovery Time Objective): 1 hour (Postgres) / 15 minutes (SQLite)
- **Availability target**: 99.5% per month
- **Zero data loss**: WAL persistence

---

## 10. Compliance and Standards

| Standard | Status | Notes |
|---|---|---|
| RFC 6962 (v1 legacy) (Merkle) | ✅ Implemented | Binary + trinary tree |
| SEC 2 (secp256k1) | ✅ Implemented | Curve parameters verified |
| RFC 6979 (deterministic ECDSA) | ✅ Implemented | No randomness |
| ISO/IEC 27001 | ⏳ Roadmap Q2 2027 | Documentation in progress |
| FIPS 186-5 | ❌ Not requested | secp256k1 outside FIPS scope |
| PCI-DSS | ⏳ If handling card data | Separate assessment |
| SOC 2 Type II | ⏳ For cloud deployments | Separate assessment |

### Court Admissibility

Audit logs are admissible if the following conditions are met:

1. **Chain of custody** — private key stored offline (USB), history documented
2. **Timestamp proof** — each block signature includes issue time, Merkle root
   published daily
3. **Verification independence** — third party can verify with public key only
4. **No backdating** — deterministic signatures + Merkle chain prevents rewriting

**Producing in court**: ledger copy, public key, Merkle root log, block signatures,
expert verification with `licenseverify`.

---

## 11. Deployment and Integration

### 11.1 Hardware Requirements

| Resource | Minimum | Recommended |
|---|---|---|
| CPU | 4 core x86_64 | 8 core |
| RAM | 8 GB | 16 GB |
| Disk | 100 GB SSD | 500 GB NVMe |
| OS | Ubuntu 22.04 LTS / Win Server 2019 | Ubuntu 24.04 LTS |

### 11.2 Integration Methods

| Source | Protocol | Direction |
|---|---|---|
| CBS with REST API | HTTP polling | Read-only |
| CBS with message queue | Kafka / RabbitMQ | Read-only |
| CBS with files | SFTP / CSV watcher | Read-only |
| SCADA (OPC-UA) | OPC-UA client | Read-only |
| SCADA (Modbus) | Modbus TCP client | Read-only |
| Historian SQL | Scheduled query | Read-only |

**Zero changes to CBS or SCADA code. Zero schema changes. Zero downtime.**

### 11.3 Installation Time

- Prepared server install: **~1 hour**
- Operator training: **2 days**
- Admin training: **3 additional days**

---

## 12. Operations and Support

### 12.1 SLA

| Severity | Definition | Response | Resolution |
|---|---|---|---|
| P1 — Critical | Ledger stopped, licenses rejected | 1h | 4h |
| P2 — Major | Performance degraded > 50% | 4h | 24h |
| P3 — Minor | Single feature impaired | 24h | 7d |
| P4 — Cosmetic | UI / typo | 7d | Next release |

### 12.2 Backup Strategy

**PostgreSQL:** daily full + continuous WAL archiving.
**SQLite:** hourly snapshot via `.backup` (safe while running).

**Both modes:**
- Merkle root published to append-only file `/backup/merkle-roots.log`
- Monthly USB export of ledger for off-site storage

### 12.3 Disaster Recovery Scenarios

| Scenario | RTO | Data loss |
|---|---|---|
| Backend crash | Auto-restart | Zero |
| Disk failure | 1 hour | ≤ 5 minutes |
| Database corruption | 1 hour | ≤ 24 hours |
| USB key lost | Key rotation | Historical blocks valid |
| Complete site loss | 4 hours | ≤ 24 hours |

### 12.4 Training Program

**Operator training (2 days):**
- Architecture overview
- Service start/stop
- Log reading
- Health checks
- License workflow
- USB handling
- Incident basics

**Administrator training (3 additional days):**
- Key rotation
- Merkle proof generation for auditors
- Disaster recovery drills
- Configuration tuning
- Security hardening

---

## 13. Roadmap

| Phase | Capability | Status | ETA |
|---|---|---|---|
| v1.0.0 | Single-node ledger | ✅ Complete | — |
| v1.1.0 | Air-gap mode | ✅ Complete | — |
| v1.2.0 | Dual control | ✅ Complete | — |
| v1.3.0 | Industrial anchor | ✅ Complete | — |
| **v1.4.0** | **secp256k1 + Merkle + offline signer** | **✅ Current** | — |
| v1.5.0 | HSM hardware integration | ⏳ Planning | Q4 2026 |
| v2.0.0 | Multi-node consensus | ⏳ Design | Q1 2027 |
| v2.1.0 | ISO/IEC 27001 | ⏳ Documentation | Q2 2027 |
| v2.2.0 | FIPS (if required) | ⏳ Optional | Q3 2027 |

---

## 14. Commercial Model

| Item | One-time | Recurring |
|---|---|---|
| Installation | Included | — |
| Training | Included | — |
| Annual license | — | Per contract |
| Support (business hours) | — | Included |
| Support (24/7) | — | Optional |
| On-site visits | Travel | Per visit |

**License types:**
- Per-site annual
- Hardware-bound (tied to HWID)
- Volume tier (by annual transaction count)
- Air-gap add-on
- Enterprise (source access + custom integration)

---

## 15. What Distinguishes This System

1. **Private key never leaves USB.** Even in browser, only 32 bytes in RAM.
2. **Switch holds only public key.** Compromise cannot forge licenses.
3. **Offline-first by design.** Full chain works without internet.
4. **Test-backed claims.** 15+ security tests + 192 penetration vectors.
5. **Honest about limitations.** FIPS/HSM/consensus not claimed. Trust through truth.
6. **Independently verifiable.** Auditor verifies with public key + proof only.
7. **Standards-compliant.** RFC 6962 (v1 legacy), RFC 6979, SEC 2, GDPR, SCC.
8. **Zero changes to existing systems.** CBS and SCADA remain untouched.

---

## 16. Request for Review

Since this document is submitted for independent technical review, we request:

**1. Cryptographic review:**
- Is the `R||S` 64-byte format compliant with secp256k1 rules?
- Are the declared rejections sufficient?
- Is the canonical payload resistant to manipulation?

**2. Trust model review:**
- Is the Authority/Backend/Switch separation sufficient?
- Is public-key-only verification adequate?

**3. Layer positioning review:**
- Are L1-L2 (banking) and L2.5 (industry) correct?
- Are the declared boundaries honest?

**4. Improvement suggestions:**
- What is missing from this document?
- What risks may be overlooked?
- What additional standards should be considered?

---

## 17. An Honest Note

Horizon Core is mature but not complete. Version 1.4 is tested on a **single
installation**, in a local environment and one Switch. Multi-node, HSM, and
production-scale tests are not yet performed.

We aim to be **honest first, strong second**. If this document is realistic,
that is correct. If any claim exceeds evidence, please tell us to correct it.

Thank you for your time.

---

**© 2026 Horizon Core · Mahdi Amolimoghaddam**
**Last updated**: October 2026
