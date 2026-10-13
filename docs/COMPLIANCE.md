# Horizon v1.4 — Compliance & Legal

**Audience**: Judiciary / Khorasan · Legal counsel · Regulator
**Scope**: standards, court admissibility, legal responsibility, longevity

---

## 1. Standards compliance

| Standard | Status | Notes |
|---|---|---|
| RFC 6962 (v1 legacy) (Merkle) | ✅ Implemented | Binary + trinary tree |
| SEC 2 (secp256k1) | ✅ Implemented | Curve parameters certified |
| RFC 6979 (deterministic ECDSA) | ✅ Implemented | No randomness |
| ISO/IEC 27001 | ⏳ Roadmap Q2 2027 | Documentation in progress |
| FIPS 186-5 | ❌ Not requested | secp256k1 is outside FIPS scope |
| PCI-DSS | ⏳ Roadmap | Only if handling card data |
| SOC 2 Type II | ⏳ Roadmap | For cloud deployments |

### Why not FIPS?

FIPS 186-5 approves P-256, P-384, P-521, and Ed25519. It does **not**
approve secp256k1, the curve used by Bitcoin and Ethereum.

**Trade-off**: We chose secp256k1 for cross-industry compatibility
(Bitcoin, Ethereum, industry tooling) at the cost of FIPS certification.
If a specific regulator requires FIPS, we can provide a P-256 build —
but the two curves are **not** compatible on the same ledger.

**Recommendation**: For deployments in Iran, secp256k1 is the industry
standard. FIPS is a US NIST standard and is not required locally.

---

## 2. Court admissibility of audit logs

### Can this log be used in court?

**Yes, if the following conditions are met:**

1. **Chain of custody** — the private key was stored offline (USB) and its
   history is documented.
2. **Timestamp proof** — each block signature includes the issue time, and
   the Merkle root is published daily.
3. **Verification independence** — any third party can verify using only:
   - the public key
   - the ledger file
   - the Merkle inclusion proof
4. **No backdating possible** — because signatures are deterministic
   (RFC 6979) and the Merkle chain links every block to the previous one.

### What to produce in court

| Item | Source |
|---|---|
| Ledger copy | `horizon.db` or Postgres dump |
| Public key of Authority | `GET /api/v1/authority/public-key` |
| Merkle root log | `/backup/merkle-roots.log` |
| Block signatures | embedded in each block |
| Expert verification | `cmd/licenseverify` output + witness |

### Independent verification by a third party

The regulator (or court) receives:

1. A signed ledger file
2. A public key (published out-of-band)
3. A verification tool (`cmd/licenseverify`, open source)

**They run one command and see if every block verifies.** No trust in us
required. If any byte was changed after signing, verification fails.

---

## 3. Legal responsibility

### Data ownership

- **All data belongs to the customer** (bank, refinery, regulator).
- Horizon only stores cryptographic proofs derived from that data.
- No data is transmitted outside the customer network by default.

### Data custody

| Data type | Where stored | Who has access |
|---|---|---|
| License private key | Customer's USB | Customer only |
| Public key | Backend + Switch env | Read-only |
| Transaction canonical payload | Backend DB | Customer DBA |
| Merkle roots | DB + backup | Customer DBA |
| Signatures | DB | Customer DBA |

### Liability boundary

- Horizon provides the **cryptographic layer**.
- Horizon is **not** responsible for the accuracy of the source data
  (CBS/SCADA outputs).
- Horizon **is** responsible for the integrity of the signature chain
  (no tampering possible after signing).

### Data retention

- Immutable ledger: **indefinite** (customer's choice)
- Application logs: default 90 days (configurable)
- Backup snapshots: default 5 years (aligned with banking regulations)

---

## 4. Data longevity

### How long can we trust this data?

The cryptographic proof remains valid **as long as**:
1. The secp256k1 curve is considered secure (no known break)
2. SHA-256 is considered secure (no known collision)

**Current expectation**: 20+ years for both.

### Future-proofing

If SHA-256 is ever weakened:
- A migration path exists: re-sign each block with SHA-3-256
- The old proofs remain valid (dual-signed during transition)
- The Merkle chain is preserved

If secp256k1 is ever broken:
- Historical blocks are dual-signed with a new curve
- A migration record is added to the ledger
- Regulatory bodies are notified

### Recommended practice

- Store an **offline paper copy** of the daily Merkle root (or print).
- For legal disputes spanning decades, produce the paper + digital together.

---

## 5. Attack surface (what has been tested)

| Category | Vectors | Status |
|---|---|---|
| SQL Injection | 30 | ✅ blocked |
| XSS | 25 | ✅ blocked |
| Command Injection | 20 | ✅ blocked |
| Path Traversal | 15 | ✅ blocked |
| Auth Bypass | 20 | ✅ blocked |
| JWT Attacks | 10 | ✅ blocked |
| Header Injection | 15 | ✅ blocked |
| HTTP Methods | 12 | ✅ blocked |
| Oversize Payload | 15 | ✅ blocked |
| SSRF | 10 | ✅ blocked |
| XXE | 5 | ✅ blocked |
| Business Logic | 15 | ✅ blocked |
| **Total** | **192** | **100% blocked** |

Plus:
- **Tamper tests** (single-byte flip in signature/payload) — all rejected
- **Cross-key tests** — signature from another key rejected
- **Expiry tests** — out-of-window rejected
- **HWID tests** — mismatch rejected
- **Merkle tests** — tampered leaf rejected

All tests run on every push via `.github/workflows/v14-ci.yml`.

---

## 6. Compliance documentation

| Document | Location |
|---|---|
| Crypto contract | `docs/CRYPTO-CONTRACT.md` |
| Deployment requirements | `docs/DEPLOYMENT-REQUIREMENTS.md` |
| Operations manual | `docs/OPERATIONS.md` |
| Positioning (layers) | `docs/POSITIONING.md` |
| Authority runbook | `docs/V14-AUTHORITY-RUNBOOK.md` |
| Demo guide | `docs/V14-DEMO.md` |
| This document | `docs/COMPLIANCE.md` |

### Available on request

- Full STRIDE threat model
- SBOM (Software Bill of Materials)
- Penetration test report (192 vectors)
- Code audit (external, on request)

---

## 7. Licensing model

| Type | Description |
|---|---|
| Annual license | Per site, per year |
| Hardware-bound | Tied to a specific HWID |
| Volume tier | Based on annual transaction count |
| Air-gap add-on | Separate license for isolated sites |
| Enterprise | Includes source access + custom integration |

Details per contract. See `COMMERCIAL-LICENSE.md` (on request).

---

**© 2026 Horizon — Mahdi Amoli Moghaddam**
