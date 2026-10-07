# Horizon Core — Security Model

Version: v1.1
Date: 2026-10-07
Owner: Mahdi Amoli Moghaddam

## 1. Signature Algorithm

ECDSA on secp256k1 (SEC 2, Ethereum-compatible).

- Algorithm: ECDSA
- Curve: secp256k1
- Hash: SHA-256
- Signature format: hex(r||s) — 128 hex chars
- Address format: Ethereum 0x...
- Private key source: env HORIZON_ISSUER_KEY (hex, 64 chars)

Why: same curve as Bitcoin and Ethereum. Any Ethereum tool can verify our signatures. Migrated from P-256 on 2026-10-07.

## 2. Key Management

Issuer key production:
- Env: HORIZON_ISSUER_KEY
- Format: 64 hex chars (32 bytes)
- Location: Liara env ONLY
- Never in git, frontend, README, chat

Dev fallback: chaindata/issuer.pem

Load: env first, then file. Production is fail-closed if env missing.

Backup: 3 copies, Shamir (3,5), separate physical locations.

Rotation: 90 days. All old signatures become invalid.

## 3. Signature Flow

Canonical message:
txID|planID|providerID|personnelID|merchantID|amount|installments|nonce

Sign: SHA-256 then ECDSA. Return 128 hex chars.
Verify: load pub PEM, decode sig (64 bytes), recompute, ecdsa.Verify.

Endpoints:
- POST /api/v1/tx/create
- GET /api/v1/tx/{id}
- GET /api/v1/tx/verify/{id}
- GET /api/v1/issuer/pubkey
- GET /api/v1/issuer/address

## 4. Key / Layer / Env

| Key | Env | Public |
|-----|-----|--------|
| Issuer Private | HORIZON_ISSUER_KEY | No |
| Issuer Public | served via endpoint | Yes |
| Issuer Address | served via endpoint | Yes |
| Admin Token | ADMIN_TOKEN | No |
| API Key | API_KEY | No |
| Switch Admin | SWITCH_ADMIN_TOKEN | No |
| DB Password | POSTGRES_PASSWORD | No |
| DB URL | DATABASE_URL | No |
| Liara Token | LIARA_API_TOKEN | No |
| AI Key | AI_API_KEY | No |
| Admin Pass Hash | ADMIN_PASSWORD_HASH | No |
| Session Secret | ADMIN_SESSION_SECRET | No |

Rule: never prefix a secret with NEXT_PUBLIC_.

## 5. Merkle Tree

Two modes: binary (fanout=2, RFC 6962) and ternary (fanout=3, Horizon custom).

Block contains: fanout, merkle_root, tx_ids, hash, signature, signer_addr.

Proof endpoint: GET /api/v1/proof/{tx_id}?fanout=2|3

## 6. Air-Gap Mode

No outbound HTTP. All signing local. Manual sync only. Software diode.

## 7. HWID Binding

Linux: /sys/class/dmi/id/product_uuid or /etc/machine-id
Windows: PowerShell CIM Win32_ComputerSystemProduct.UUID

Rule: HWID only from server-side. Fail-closed.

## 8. Session and Auth

Store admin: bcrypt cost 12, HTTP-only cookie, 24h expiry.
Switch admin: X-Admin-Token, min 32 chars.
Customer: OTP via SMS, HTTP-only cookie.

## 9. Rate Limiting

- tx/create: 100/min per IP
- anchor: 10/min per IP
- verify: 60/min per IP
- auth/login: 5/min per IP
- AI chat: 20/min per user
- Public: 200/min per IP

## 10. Audit Log

Hash-chained. Immutable. Daily Merkle root anchored externally.

## 11. Compliance

- secp256k1: SEC 2
- SHA-256: FIPS 180-4
- Merkle: RFC 6962 (binary) + Horizon custom (ternary)
- TLS: 1.2+
- bcrypt: cost 12+
- Randomness: crypto/rand

## 12. Migration Note

Before: ECDSA P-256 (NIST)
Current: ECDSA secp256k1 (SEC 2)
Date: 2026-10-07
All older signatures invalid under new scheme.

This document is the authoritative security reference.

-- Horizon Angel - under command of Mahdi Amoli Moghaddam
