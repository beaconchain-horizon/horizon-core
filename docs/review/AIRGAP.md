# Horizon — Air-Gap Architecture

Air-Gap and Sync are two phases, not contradictory.

## Phase 1: Air-Gapped Runtime
Sensor -> Switch -> SQLite (no internet)

## Phase 2: Controlled Transfer
SQLite -> Signed Export -> USB -> Gateway -> Deferred Sync

## Terminology
- Air-Gap Runtime
- Offline Verification
- Controlled Data Transfer
- Deferred Synchronization
