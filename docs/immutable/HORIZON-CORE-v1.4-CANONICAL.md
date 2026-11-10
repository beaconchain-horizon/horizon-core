<div align="center">
  <img src="assets/logo.png" width="300" alt="Horizon Core Logo"/>
  <br/>
  <h1>HORIZON CORE</h1>
  <p><strong>v1.4 — TRUST · VERIFY · MOVE FORWARD</strong></p>
  <hr/>
</div>

<div align="center">
<h1>HORIZON CORE</h1>
<p><strong>Offline-first tamper-evident ledger — v1.4</strong></p>
<p>Genesis LOCKED · Block #1 SIGNED · 7-Node Consensus</p>
</div>

## Genesis — locked 2026-11-03

- Chain ID: horizon-core-v1.4
- Genesis Hash: 8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603
- Block #1 Hash: f981652e6c25945a2655b83f7536524441422bdeb7b9742a30a5256cd637ceba
- Authority Pubkey: 04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
- Genesis Sig: afac005b...d236face (R||S, 64 bytes)
- Block #1 Sig: 19f1b567...94bd335 (R||S, 64 bytes)
- Key format: hex only (no PEM)

Verify locally:

go run ./cmd/chain-verify

## Architecture — 7-Node Specialty Consensus

Backend (8080) → Switch (8091) → 7 Specialty Nodes (9001-9007)
                                  ↓
                              Agent Node (9008) — inspector, separate

### Port Map

- 8080 — Backend / API
- 8091 — Switch (operational gateway)
- 9001 — Angel Merkle: SHA256(left||right)
- 9002 — Angel Hash: SHA256(header)
- 9003 — Angel ECDSA: secp256k1 verify
- 9004 — Angel Type: binary or ternary
- 9005 — Angel Volume: used < volume
- 9006 — Angel Time: issued <= now < exp
- 9007 — Angel User: user_id match
- 9008 — Agent Node (AI agents inspector, NOT consensus)
- 9101+ — Rooms

### Consensus (ANGEL-NODE-LAW.md)

- 4 CRITICAL (Merkle, Hash, ECDSA, Type) — 100% required
- 2 of 3 MAJORITY (Volume, Time, User) — required
- Otherwise: REJECT (fail-closed)

### Dual Merkle

- Binary: node(a,b) = SHA256(a||b) — internal/merkle/merkle.go
- Ternary: node(a,b,c) = SHA256(a||b||c) — internal/merkle/ternary.go (unique to Horizon)

## Cryptography

- Curve: secp256k1 (SEC 2)
- Hash: SHA-256 (FIPS 180-4)
- Determinism: RFC 6979
- Low-S: s <= n/2 (BIP-62)
- Key format: hex only (64 chars, 0x optional)

### Zero PEM

Private keys never stored as PEM. Only hex strings:

- On USB: /path/to/mykey.txt (64 hex chars, 0x optional)
- Loaded into RAM by cmd/block-sign, cmd/license-signer, cmd/switch key/setup
- Never written to disk by the server

- Fuzz testing: 1.86M execs — zero counterexamples
- Cross-verification: OpenSSL 3.5.7 — Verified OK

## Security — Fail-Closed

- Auth Bypass → X-Admin-Token (>=32 chars) → BLOCKED
- No License → 402 before tx → BLOCKED
- No Signing Session → X-Session-Token single-use 5min → BLOCKED
- SQL Injection → parameterized queries → BLOCKED
- Buffer Overflow → MaxBytesReader 8MiB → BLOCKED
- DoS → rate limiting 60/s burst 120 → BLOCKED
- Replay → single-use session tokens → BLOCKED

199 Attack Vectors — reproducible via pentest_suite.sh. All blocked, 0 vulnerable.

## Tools

- cmd/genesis-lock — lock genesis with authority sig
- cmd/genesis-verify — verify genesis against GENESIS.json
- cmd/chain-verify — walk entire chain, verify prev_hash links + signed status
- cmd/block-produce — append new block to chain
- cmd/block-sign — sign block hash with hex key from stdin
- cmd/api — backend HTTP API
- cmd/switch — operational gateway (license, tx, mine)
- cmd/angel-node — 7 specialty nodes
- cmd/agent-node — AI agent inspector
- cmd/verify-tx — independent tx verifier
- cmd/license-signer — hex-based offline license signer

## Documentation

- CONSTITUTION.md — 7 laws, immutable
- ANGEL-NODE-LAW.md — 7-node consensus rules
- GENESIS.json — locked genesis artifact
- CHAIN_STATE.json — pinned chain state
- STATE.md — current build status
- RISKS.md — known technical debt
- THREAT_MODEL.md — attack surface

---

<div align="center">
HORIZON CORE — a protocol, not a product.<br>
v1.4 · genesis locked 2026-11-03 · all rights reserved
</div>
