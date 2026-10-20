# Horizon Core — International Audit Report

**Version:** 1.4
**Date:** 2026-10-20
**Standard:** ISO/IEC 25010

## 1. Security (Verified)
- 192 Attack Vectors (OWASP Top 10 + Advanced SCADA/APT)
- SQLi and XSS strictly blocked via Input Sanitization Middleware (HTTP 403).
- SSRF, XXE, JWT, Auth Bypass blocked.
- **Status:** Fail-Closed Security Implemented.

## 2. Performance (Verified)
- **129,366 TPS** (Batch Mode: 100,000 TX in 773ms).
- Environment: Windows 8-core laptop.

## 3. License Logic
- `signed_data` timestamp and hardware_id are strictly enforced.
- Signature verification strictly matches Dashboard generation.

## 4. Reproducibility
- All tests run via `go test ./...`
- Pentest tool exits with failure on any `vuln > 0 || serverErr > 0 || netErr > 0`.
