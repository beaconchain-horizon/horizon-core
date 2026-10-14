<div align="center">

# 🔭 Horizon Core

**Offline-first tamper-evident ledger for banking, industry, and air-gapped environments.**

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev)
[![Crypto](https://img.shields.io/badge/Crypto-secp256k1-627EEA)](https://www.secg.org/sec2-v2.pdf)
[![Merkle](https://img.shields.io/badge/Merkle-binary%20%2B%20ternary-8957e5)](#merkle-tree)
[![Status](https://img.shields.io/badge/chain%20verified-%231--%2388-3fb950)](#chain-integrity)

**Version:** v1.4 · **README:** v2 (2026-10-14) · **Chain height:** 88 · **TX total:** 12,857

</div>

---

## ✨ What's new in v2 (README update)

این README با واقعیت کد امروز (۲۰۲۶-۱۰-۱۴) هم‌راست شده. تغییرات کلیدی نسبت به v1:

| # | تغییر |
|---|---|
| 🔐 | **High-S malleability** بسته شد — `parseRS` حالا `s ≤ n/2` را اجباری می‌کند |
| 🔐 | **DER rejection** در `VerifyData` فعال شد (خلاف v1) |
| 🌳 | **Merkle unify** — chain، license، proof، regen همه یک فرمول دارند |
| 🌳 | **Ternary Merkle** پیاده شد (`TernaryRootNoPrefix`) |
| 🛠 | **`verify-tx`** — ابزار مستقل تأیید تراکنش (بدون Backend، air-gap friendly) |
| 🛠 | **`verify-lic`** — ابزار مستقل تأیید لایسنس |
| 🛡 | **Rate limiting** (per-IP، 60/s، burst 120) |
| 🛡 | **Audit log** hash-chained (tamper-evident) |
| 🛡 | **DoS-1 بسته** — `MaxBytesReader` (8 MiB default) |
| 🛡 | **CORS allowlist** — wildcard حذف شد |
| 🔒 | **mTLS** برای Switch↔Backend (TLS 1.3، client cert) |
| 🧪 | **Fuzz testing** — ۱.۸۶M اجرا، صفر counterexample |
| 🧪 | **Cross-verify با OpenSSL 3.5.7** — `Verified OK` |
| 📚 | RFC 6962 به‌عنوان **v1 legacy** علامت خورد؛ v2 binary custom |

---

## 🚀 Quick start

```bash
# Terminal 1 — Backend + chain
export HORIZON_DB_PATH="/tmp/horizon-chain.db"
export HORIZON_LICENSE_PUBLIC_KEY="cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
export LICENSE_PRIVATE_KEY_PEM="$(cat .horizon-dev-key.pem)"
export ADMIN_TOKEN="<your-admin-token>"
export SERVER_PORT=8080
export GIN_MODE=release
./backend.exe
