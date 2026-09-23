# Horizon — Response to Red-Team Review

Date: 2026-09-23

Claims:
- Consensus undefined: addressed (CONSENSUS.md)
- Air-Gap/Sync ambiguous: addressed (AIRGAP.md)
- TPS needs spec: addressed (BENCHMARK-SPEC.md)
- Go 1.27 not released: reviewer error (released 2026-08-19)
- SQLite low TPS: overstated (config-dependent)
- HSM important: addressed (SECURITY-ROADMAP.md)
- FIPS 140-2 L3 mandatory: incorrect (FIPS 140-3 current)
- 270 agents not security: valid
- Needs security testing: valid
- Score 3/10: no rubric, dismissed

P0: fix tpsbench invalid_signature + unify canonical JSON/SHA256/ECDSA
