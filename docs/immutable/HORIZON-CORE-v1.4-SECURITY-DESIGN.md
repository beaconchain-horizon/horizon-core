# HORIZON CORE — Security Design Specification

> **Classification:** Protocol-level · immutable
> **Scope:** v1.4 · genesis locked 2026-11-03
> **Posture:** fail-closed · air-gap capable · tamper-evident

---

## 1. THREAT MODEL

| Threat class | Adversary goal | Attack surface |
|---|---|---|
| **Forgery** | issue or accept a fake license / block | switch, angel nodes, API |
| **Tampering** | alter an existing block or license in situ | chain store, DB |
| **Replay** | reuse a captured tx or session | switch ingress |
| **Bypass** | reach tx path without a valid license | API, switch |
| **Key theft** | extract private key material | process memory, disk, USB |
| **DoS** | exhaust resources on the switch or angels | HTTP ingress |
| **Supply chain** | inject malicious build / dependency | build pipeline, `go.mod` |
| **Insider** | operator with DB access forges a license | DB write path |

---

## 2. DEFENSE-IN-DEPTH LAYERS

| Layer | Control | Failure mode |
|---|---|---|
| **L1 — Network** | bind explicit ports · CORS whitelist | closed port → unreachable |
| **L2 — Transport** | `MaxBytesReader` 8 MiB · rate limit 60/s burst 120 | over-limit → 429/413 |
| **L3 — Auth** | `X-Admin-Token` ≥ 32 chars · `X-Session-Token` single-use 5 min | missing → 401 |
| **L4 — License** | `X-License-ID` required · signature verify on every request | invalid → 402/401 |
| **L5 — Consensus** | 4 critical unanimous + 2/3 majority | any fail → REJECT |
| **L6 — Crypto** | `secp256k1` · SHA-256 · RFC 6979 · Low-S | verify fail → REJECT |
| **L7 — Storage** | parameterized queries · append-only chain | tamper → chain-verify breaks |
| **L8 — Key custody** | hex only · USB → RAM · never on disk | disk compromise → no key |

> Each layer is **independently sufficient to reject**. No layer trusts the one below it.

---

## 3. ATTACK → CONTROL MATRIX

| # | Vector | Layer | Control | Detection |
|---:|---|---|---|---|
| 1 | Auth bypass | L3 | `X-Admin-Token` constant-time compare | 401 log |
| 2 | Missing license | L4 | `402 before tx` | 402 log |
| 3 | Session replay | L3 | single-use token, 5 min TTL | 401 `invalid_or_expired_session` |
| 4 | SQL injection | L7 | parameterized queries only | query audit |
| 5 | Buffer overflow | L2 | `MaxBytesReader` 8 MiB | 413 |
| 6 | HTTP flood | L2 | rate 60/s · burst 120 | 429 |
| 7 | Forged license | L4+L5 | signature verify + angel consensus | `signature_valid:false` |
| 8 | Forged block | L5 | Merkle + Hash unanimous | `merkle mismatch` / `hash mismatch` |
| 9 | Expired license | L5 | Time node | `license expired` |
| 10 | Volume overrun | L5 | Volume node | `volume exceeded` |
| 11 | User mismatch | L5 | User node | `user mismatch` |
| 12 | Type confusion | L5 | Type node | `bad type` |
| 13 | Signature malleability | L6 | Low-S BIP-62 | `high-S signature` |
| 14 | Key exfiltration | L8 | hex-only · RAM-only | disk audit |
| 15 | Chain rewrite | L7 | `prev_hash` linkage | `chain-verify` fails |

> Full 199-vector suite: `pentest_suite.sh` — all blocked, 0 vulnerable.

---

## 4. CONSENSUS SECURITY — WHY FAIL-CLOSED

| Scenario | Critical votes | Majority votes | Outcome |
|---|---:|---:|---|
| All honest | 4/4 | 3/3 | ACCEPT |
| One critical offline | 3/4 | — | **REJECT** |
| Two critical offline | 2/4 | — | **REJECT** |
| One critical malicious (veto) | 3/4 | — | **REJECT** |
| One majority offline | 4/4 | 2/3 | ACCEPT |
| Two majority offline | 4/4 | 1/3 | **REJECT** |
| AgentGuardian veto | — | — | **REJECT** |

> **Property:** a single compromised critical node **cannot** cause an accept. A single offline critical node **cannot** cause an accept. Silence = REJECT.

---

## 5. CRYPTOGRAPHIC INVARIANTS

| Invariant | Enforcement | Evidence |
|---|---|---|
| Signature determinism | RFC 6979 (no RNG in signing) | fuzz 1.86M, 0 counterexamples |
| No signature malleability | `s ≤ n/2` enforced pre-verify | BIP-62 test suite |
| Cross-implementation compatibility | OpenSSL 3.5.7 verify | `crossverify_test.go` |
| Hash collision resistance | SHA-256 (FIPS 180-4) | standard |
| Curve safety | `secp256k1` (SEC 2) | standard |

---

## 6. KEY LIFECYCLE

| Phase | Custodian | Medium | Duration |
|---|---|---|---|
| Generation | offline machine | ephemeral | one-shot |
| Storage | operator | USB `/path/to/mykey.txt` | until rotation |
| Load | `cmd/switch key/setup` | RAM only | process lifetime |
| Use | signing ops | RAM | per-op |
| Unload | process exit | — | zeroed by OS |
| Disk write | **forbidden** | — | never |

> **Invariant:** no server process ever writes a private key to disk. If a `.pem` or `.key` file appears under the repo, that is a **P0 incident**.

---

## 7. AIR-GAP POSTURE

| Capability | Online switch (8091) | Offline switch (8092) |
|---|---|---|
| Internet reachable | yes | **no** |
| Accepts external tx | yes | via removable media |
| Signs blocks | yes | yes |
| Consensus scope | full 8 nodes | full 8 nodes |
| Recovery path | — | signed block import |

> The offline switch exists so that **signing authority can operate without network exposure**. It is not a warm standby — it is an **independent signer**.

---

## 8. RESIDUAL RISKS

| Risk | Severity | Mitigation | Status |
|---|---|---|---|
| `signer_addr` is an address, not pubkey → requires ecrecover | medium | out of v1.4 scope | **documented** |
| ECDSA node validates format, not full signature | medium | full verify lives in switch | **by design** |
| AgentGuardian never vetoes in current build | low | inspector role only | **by design** |
| Offline switch (8092) not deployed in test env | low | separate deployment | **open** |
| Legacy DBs coexisting with `horizon_switch` | medium | planned purge | **open** |

---

## 9. AUDIT CHECKLIST

| # | Check | Command | Pass criteria |
|---:|---|---|---|
| 1 | Genesis intact | `go run ./cmd/chain-verify` | exit 0 |
| 2 | Genesis hash matches | `go run ./cmd/genesis-verify` | `OK` |
| 3 | No PEM in repo | `find . -name '*.pem' -o -name '*.key'` | empty |
| 4 | All 8 angels live | `curl :900{1..8}/health` | 8 responses |
| 5 | Switch signature verifies | `POST /license/verify` | `signature_valid:true` |
| 6 | Fake license rejected | `POST /license/verify` (fake) | `signature_valid:false` or 404 |
| 7 | Rate limit enforced | burst > 120 | 429 returned |
| 8 | Session replay blocked | reuse `X-Session-Token` | 401 |
| 9 | Pentest suite | `bash pentest_suite.sh` | 0 vulnerable |
| 10 | Fuzz corpus | `go test -fuzz` | 0 counterexamples |

---

<div align="center">
HORIZON CORE — a protocol, not a product.<br/>
Security design · v1.4 · immutable · genesis locked 2026-11-03
</div>
