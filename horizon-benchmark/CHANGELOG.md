# Changelog

All notable changes to the Horizon Benchmark package.

## [1.0.1] - 2026-10-01

### Added

- SECURITY.md - links to main security policy.
- CHANGELOG.md - this file.

### Changed

- Documented actual TPS range measured across hardware:
  - Laptop (8 cores): ~15,000 TPS
  - Dedicated server: 19,000-31,000 TPS
  - Theoretical on 32+ cores: up to 35,000+ TPS

## [1.0.0] - 2026-09-27

### Added

- Initial public benchmark package.
- TPS benchmark: 200,000 readings, c=50, bs=500.
- run.sh / run.bat launchers.
- Precompiled switch and tpsbench binaries.

### Verified

- 19,000-31,000 TPS on dedicated server.
- ~15,000 TPS on laptop (8-core CPU).
- Latency: ~6 ms.
- ECDSA signing: less than 1 ms.
- Zero errors in 200,000-reading run.
