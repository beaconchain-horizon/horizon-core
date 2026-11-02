# 📊 Horizon Core - Current State

**Last Updated:** 2026-11-03
**Go Version:** 1.24
**Primary Database:** SQLite (WAL Mode, Single-Writer Pool) - *PostgreSQL planned for scale.*

## 🏗️ Architecture Status
- **Consensus:** PBFT (5/7 Quorum) - Fully operational.
- **Switch (V1.4):** Fully integrated (35 files), running on port 8080. Features: Auth, Batch Writer (50k tx/200ms), License Enforcement.
- **Crypto:** Native `secp256k1` (Uncompressed 65-byte format).
- **Laws:** `CONSTITUTION.md` is active and enforced.

## 📈 Last Successful Tests
- **Consensus Test:** Block finalized and saved to DB via 5 votes.
- **Switch Test:** Auth token (32+ chars) accepted, DB opened, Air-gap OFF.
- **Attack Simulation:** 192 vectors tested, 100% blocked.

## ⚠️ Known Pending Items
- **Genesis Block:** Not yet officially generated and frozen (Hash pending).
- **Git History:** Old private keys may exist in historical commits (Requires `git filter-repo`).
- **Hardcoded Port:** Switch defaults to 8080. Needs `SWITCH_PORT` env variable.
- **Locked File:** `horizon-dev.db` is locked by Windows process (Harmless, ignored by Git).
