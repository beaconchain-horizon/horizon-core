# Response to Red-Team Review

| Claim | Response | Status |
|---|---|---|
| Consensus needed | See CONSENSUS.md | addressed |
| Air-Gap/Sync ambiguity | See AIRGAP.md | addressed |
| TPS needs spec | See BENCHMARK-SPEC.md | addressed |
| Go 1.27 not released | Invalid — released 2026-08-19 | dismissed |
| SQLite cannot do TPS | Overstated — depends on config | partial |
| HSM for banking | See SECURITY-ROADMAP.md | addressed |
| FIPS 140-2 L3 mandatory | Invalid — FIPS 140-3 current | dismissed |
| 270 Agents not security | Valid | accepted |
| Needs security testing | Valid | accepted |
| Score 3/10 | No rubric → not actionable | dismissed |

## P0 (open)
- [ ] Fix tpsbench invalid_signature
- [ ] Unify SignableReading
- [ ] Unify canonical JSON
- [ ] Unify ECDSA encoding
