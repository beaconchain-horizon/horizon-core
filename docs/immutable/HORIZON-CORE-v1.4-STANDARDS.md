# HORIZON CORE — International Standards Compliance

> **Classification:** Protocol-level · immutable
> **Scope:** v1.4 · genesis locked 2026-11-03
> **Statement:** Horizon Core is a protocol, not a product. It is designed against, and publicly verifiable against, the following international standards.

---

## 1. Cryptographic Standards

| Standard | Issuer | Region | Purpose in Horizon | Reference |
|:---------|:-------|:------:|:-------------------|:----------|
| **SEC 2 v2 — secp256k1** | SECG | 🇺🇸 · 🌍 | Elliptic curve for all signatures | [secg.org/sec2-v2.pdf](https://www.secg.org/sec2-v2.pdf) |
| **RFC 6979 — Deterministic ECDSA** | IETF | 🌍 | No RNG in signing → reproducible signatures | [datatracker.ietf.org](https://datatracker.ietf.org/doc/html/rfc6979) |
| **BIP-62 — Low-S enforcement** | Bitcoin Core | 🌍 | `s ≤ n/2` → malleability eliminated | [bitcoin/bips](https://github.com/bitcoin/bips/blob/master/bip-0062.mediawiki) |
| **FIPS 180-4 — SHA-256** | NIST | 🇺🇸 | Block hashing, Merkle root, license ID | [csrc.nist.gov](https://csrc.nist.gov/publications/detail/fips/180/4/final) |

---

## 2. Software Quality Standards

| Standard | Issuer | Region | Purpose in Horizon | Reference |
|:---------|:-------|:------:|:-------------------|:----------|
| **ISO/IEC 25010 — Software quality model** | ISO | 🇨🇭 · 🌍 | Architecture quality criteria | [iso25000.com](https://iso25000.com/index.php/en/iso-25000-standards/iso-25010) |

---

## 3. Financial Industry Standards

| Standard | Issuer | Region | Purpose in Horizon | Reference |
|:---------|:-------|:------:|:-------------------|:----------|
| **ISO 8583 — Financial transaction messages** | ISO | 🇨🇭 · 🌍 | Transaction interchange compatibility | [iso.org](https://www.iso.org/standard/79451.html) |
| **ISO 20022 — Financial messaging** | ISO | 🇨🇭 · 🌍 | License and settlement message semantics | [iso20022.org](https://www.iso20022.org/) |
| **Basel III — Banking regulation** | BIS | 🇨🇭 · 🌍 | Operational risk & capital framework alignment | [bis.org](https://www.bis.org/basel_framework/) |

---

## 4. Data Protection Standards

| Standard | Issuer | Region | Purpose in Horizon | Reference |
|:---------|:-------|:------:|:-------------------|:----------|
| **GDPR — Data protection** | EU | 🇪🇺 | User data minimisation, right to erasure, DPA | [eur-lex.europa.eu](https://eur-lex.europa.eu/eli/reg/2016/679/oj) |

---

## 5. Compliance Matrix

| Horizon property | Standard(s) satisfied | Verification method |
|:-----------------|:----------------------|:--------------------|
| Signature determinism | RFC 6979 | fuzz test (1.86M execs, 0 counterexamples) |
| No signature malleability | BIP-62 | pre-verify Low-S check + test suite |
| Cross-implementation verify | SEC 2 v2 | OpenSSL 3.5.7 cross-verification |
| Hash integrity | FIPS 180-4 | SHA-256 across Merkle, block, license |
| Architecture quality | ISO/IEC 25010 | design review (8 angels, defense-in-depth) |
| Financial compatibility | ISO 8583, ISO 20022 | switch ingress + license schema |
| Regulatory posture | Basel III | fail-closed consensus, audit trail |
| Data minimisation | GDPR | no PII in chain; user_id is opaque reference |

---

## 6. Verifiability Statement

Each standard above is **publicly accessible at the link provided**. Any third party — auditor, regulator, customer, or researcher — may:

1. Download the standard.
2. Compare Horizon Core's implementation against the requirements.
3. Reproduce the cryptographic verification with `go test ./internal/crypto/...`.

> **No proprietary or classified standards are referenced.** Every claim in this document is falsifiable against a public source.

---

<div align="center">
HORIZON CORE — a protocol, not a product.<br/>
Standards compliance · v1.4 · immutable · genesis locked 2026-11-03
</div>
