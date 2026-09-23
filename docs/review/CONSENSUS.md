# Horizon — Consensus Model

Status: Single-node tamper-evident ledger (not distributed consensus)

## Current
- Nodes: 1
- Ledger: append-only hash chain
- Integrity: Merkle root + ECDSA P-256 per block
- Storage: SQLite WAL (local)
- Distribution: none

## Not Yet
- Not distributed consensus
- Not BFT/PBFT
- Not multi-node replication

## Terminology
- Use: Tamper-evident ledger, Single-writer, Offline-first runtime
- Avoid: Distributed blockchain, Consensus network, Offline consensus

## Roadmap
P0: hash chain + Merkle + ECDSA + SQLite (done)
P1: ordering + finality + consensus + fork test (planned)
P2: HSM + mTLS + audit log + rotation (planned)
