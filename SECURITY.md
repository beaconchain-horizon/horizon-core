# Horizon Core — Security Architecture

## Overview

Horizon Core is a private, Air-Gap capable blockchain for critical industries. This document describes every security layer implemented.

---

## Layer 1 — DoS Protection

| Control | Value | Location |
|---------|-------|----------|
| Rate Limiting | 50 req/s, capacity 100 | `security.go` |
| Connection Limit | 500 concurrent | `switch.go` (netutil.LimitListener) |
| Read Timeout | 30 s | `switch.go` (http.Server) |
| Write Timeout | 30 s | `switch.go` |
| Idle Timeout | 60 s | `switch.go` |
| Max Header Bytes | 1 MB | `switch.go` |
| SYN Cookies | OS-level | Linux kernel |
| DDoS Protection | Liara platform | Liara |

---

## Layer 2 — Replay & Intrusion Prevention

| Control | Value | Location |
|---------|-------|----------|
| Nonce | unique per reading | `internal/industrial/logic.go` |
| Timestamp Window | −300 s to +120 s | `NonceFresh()` |
| Nonce Cache | 300 s window | in-memory map |
| SSRF Protection | 39/40 vectors blocked | `ValidateAgentURL()` |
| Input Validation | regex + sanitize | `ValidateTenantID()`, `SanitizeString()` |
| CORS | whitelist only | `switch.go` |
| Security Headers | nosniff, DENY, XSS-Protection | `SecurityHeadersMiddleware()` |

---

## Layer 3 — Key Protection

| Control | Value | Location |
|---------|-------|----------|
| ECDSA P-256 | FIPS 186-5 | `signData()` |
| Key Rotation | Every 90 days | `keyrotation.go` |
| Shamir Secret Sharing | XOR-based split | `shamir.go` |
| Auto-Lock | 30 min inactivity | `startAutoLock()` |
| Manual Lock | POST /api/v1/key/lock | `lockKeyHandler()` |
| Sign Audit Trail | Every signature logged | `signData()` |
| Hardware ID Binding | license bound to HWID | `license_enforce.go` |

---

## Layer 4 — Supply Chain

| Control | Value | Location |
|---------|-------|----------|
| Dependabot | Go + Docker + Actions | `.github/dependabot.yml` |
| govulncheck | Weekly scan | `.github/workflows/security-scan.yml` |
| gosec | Static analysis | same workflow |
| go.sum verification | `go mod verify` | CI |
| Indirect deps | quic-go, x/net pinned to safe versions | `go.mod` |

---

## Layer 5 — Air-Gap Capability

| Control | Value |
|---------|-------|
| Offline signing | Local ECDSA |
| In-memory DB option | `SWITCH_DB=:memory:` |
| Smart Sync | Only new data on reconnect |
| Zero outbound | No external calls in isolated mode |

---

## Layer 6 — Docker & Deployment

| Control | Value | Location |
|---------|-------|----------|
| Non-root user | UID 10001 | `Dockerfile` |
| Healthcheck | curl /api/v1/health | `Dockerfile` |
| Explicit Dockerfile per app | switch vs backend | `.github/workflows/deploy-liara.yml` |
| Read-only root | Alpine + su-exec | `Dockerfile` |
| Go 1.24 | Stable, matches Liara | `go.mod` |

---

## Verified Metrics

| Metric | Value |
|--------|-------|
| Peak TPS | 31,113 |
| Latency | 6 ms |
| ECDSA Sign | < 1 ms |
| Penetration tests | 160 |
| Security score | 98% |
| Critical bugs | 0 |
| Dependabot alerts | 0 |

---

## Compliance Notes

- ECDSA P-256 is FIPS 186-5 approved
- SHA-256 per transaction
- Merkle Tree for batch proof
- Immutable ledger with tamper detection
- All signatures auditable

---

## Contact

- GitHub: https://github.com/beaconchain-horizon
- Benchmark: https://github.com/beaconchain-horizon/horizon-benchmark
- Air-Gap Demo: https://github.com/beaconchain-horizon/horizon-airgap-demo

---

**© 2026 Horizon Core**
