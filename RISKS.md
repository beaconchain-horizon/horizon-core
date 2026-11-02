# ⚠️ Horizon Core - Risks & Technical Debt

This document tracks known risks, tech debt, and contradictions to be resolved.

## 1. TPS Claims (Clarification Needed)
- **Risk:** Claiming "Millions of TPS" is misleading for end-to-end operations.
- **Reality:** Millions TPS achieved in *Raw Batch Writer* (no consensus/disk). End-to-end PBFT + SQLite is lower.
- **Action:** Publish accurate end-to-end benchmark.

## 2. Git History Security (High Risk)
- **Risk:** Private keys (`.pem`, `.env`) were committed in the past.
- **Action:** Run `git log --all --full-history -- "*.pem"`. If found, use `git filter-repo` or BFG to purge history.

## 3. Genesis Block Missing (Critical)
- **Risk:** Network is running without a frozen, immutable Genesis block.
- **Action:** Generate Genesis, save hash to `GENESIS.json`, and lock it.

## 4. Zero-Storage Contradiction
- **Risk:** `license-signer` and `authority-keygen` generate keys.
- **Clarification:** These tools MUST operate strictly in-memory. They must NEVER write private keys to disk.

## 5. Key Format Limitation
- **Risk:** System only accepts uncompressed (65-byte, starts with `04`).
- **Action:** Document this. If integration with legacy systems requires 33-byte compressed keys, update `internal/crypto`.

## 6. Attack Vectors Scope
- **Risk:** "192 Vectors" sounds final, but it's just the baseline.
- **Action:** Document that 192 is the *current* coverage. Continuous expansion is required.

## 7. Hardcoded Port (Tech Debt)
- **Risk:** Switch is hardcoded to port 8080.
- **Action:** Refactor to read `SWITCH_PORT` from env in the next sprint.
