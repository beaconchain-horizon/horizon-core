# 📊 Horizon Core - Current State

**Last Updated:** 2026-11-10
**Go Version:** 1.24
**Primary Databases:** Dual-Engine (PostgreSQL for Online High-Speed, SQLite for Offline Air-Gap).

## 🏗️ Architecture Status
- **Consensus:** 8-Node PBFT (4 Critical + 2/3 Majority + 1 Agent Guardian). Fully operational and fail-closed.
- **Switch (V1.4):** Runs on port 8091 (Online/PostgreSQL) and 8092 (Offline/SQLite). Features: Auth, Optimized Batch Writer, License Enforcement.
- **License Issuance:** Block-based. No license is valid unless it is included in a block approved by all 8 consensus nodes.
- **Crypto:** Native `secp256k1` (Uncompressed 65-byte format). Hex-only (Zero PEM).
- **Laws:** `CONSTITUTION.md` is active and enforced.

## 📈 Verified Performance & Security Tests (2026-11-10)
- **Ingestion TPS:** 110,072 tx/sec (Batch Writer, PostgreSQL).
- **Commit TPS:** 43,500 tx/sec (8-Node PBFT Consensus + Disk Write).
- **Zero Drop:** 1,000,000 transactions processed with 0 drops and 0 HTTP errors.
- **Attack Simulation:** 192/192 attack vectors passed (100% blocked).
- **Stuxnet Simulation:** SCADA Injection, Sensor Tampering, OOM, and SSRF all BLOCKED.

## 🔒 Genesis & Chain State (LOCKED)
- **Genesis Block (#0):** `8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603` (Signed & Locked).
- **Block #1:** `f981652e6c25945a2655b83f7536524441422bdeb7b9742a30a5256cd637ceba` (Signed).
- **Block Strategy:** Blocks are created dynamically based on size (Batch size 50,000) or time (Flush Interval), aggregating transactions from multiple clients.

## ⚠️ Known Pending Items
- **Third-Party Audit:** Required for Tier 3 enterprise customers (e.g., Fireblocks, IBM).
- **Dynamic Customer Nodes:** Architecture is ready; API for dynamic node registration per customer is pending.
- **Git History:** 🟢 RESOLVED. Old PEM keys and `.exe` backups have been purged and added to `.gitignore`.
- **Port Configuration:** 🟢 RESOLVED. Port is fully dynamic via `SWITCH_PORT` env variable.
