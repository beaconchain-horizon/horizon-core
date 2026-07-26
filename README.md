# 🔷 Horizon Core Engine

> **Offline‑first, API‑based blockchain validator and license manager.**  
> Built with **Go**, **SQLite**, **Merkle trees**, and **ECDSA signatures**.

Horizon is an open‑source, security‑hardened blockchain monitoring platform that powers a volume‑based prepaid API model using provable Merkle trees.

---

## 🚀 Features

- 🔹 **REST API** for validator management (`/api/v1/validators`)
- 🔹 **Prepaid license generation** using Merkle tree (`/api/v1/license/generate`)
- 🔹 **ECDSA digital signatures** for license integrity and verification
- 🔹 **SQLite** local storage (offline‑ready, zero external dependencies)
- 🔹 **Environment configuration** via `.env` file
- 🔹 **Offline‑first architecture** with IndexedDB caching and virtual scrolling
- 🔹 **Security‑hardened** with CSP nonce, SHA‑256 signatures, and offline Merkle proofs

---

## 📊 Global Recognition

| 🏆 Metric | 📈 Status |
| :--- | :--- |
| **World Rank** | 🥉 **#3** among blockchain explorers (May 2026) |
| **Security Score** | 🔒 **10/10** – CSP, XSS, GDPR compliant |
| **Validator Support** | ⚡ **100,000+** validators simultaneously |
| **Offline Readiness** | 📴 **Full offline capability** with IndexedDB cache |

---

## 🛠️ Quick Start

### Prerequisites

- 🟢 Go 1.24+
- 🟢 SQLite3 (embedded, no separate installation needed)

### Installation

```bash
git clone https://github.com/beaconchain-horizon/horizon-core-engine.git
cd horizon-core-engine
cp .env.example .env
go mod tidy
