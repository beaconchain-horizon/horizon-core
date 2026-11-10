# Horizon Core v1.4 — Sovereign Federated Ledger
### Technical Whitepaper (Draft — Sections 1 & 2)

## 1. Executive Summary

Horizon Core is a sovereign, air-gapped, industrial-grade distributed ledger engineered for regulated industries — banking, energy, and critical infrastructure — where data residency, Byzantine resilience, and tamper-evident auditability are non-negotiable. Unlike first-generation consortium ledgers whose default ordering services are crash-fault-tolerant (CFT) rather than Byzantine-fault-tolerant (BFT), Horizon Core treats adversarial — not merely fail-stop — behavior as its baseline assumption.

The platform is defined by three architectural pillars:

**Federated Practical Byzantine Fault Tolerance (F-PBFT).** Consensus begins with eight core nodes — four Critical nodes (Merkle construction, hashing, ECDSA verification, type validation), three Majority nodes (volume accounting, timestamp ordering, user attestation), and one Agent Guardian operating as a non-voting offline auditor. As enterprise clients onboard, each is issued a license and deploys its own consensus node on its own physical infrastructure, distributing the validator set across organizations, cities, and jurisdictions. The commit quorum scales dynamically with membership, preserving a ⅓ Byzantine fault tolerance at every scale.

**Air-Gap Security Boundary.** Key material never persists on networked storage. Consensus signing uses secp256k1 keys provisioned exclusively through a documented air-gap transfer protocol — read-only media, locked-down signing appliances, and no auto-execution of external content. Node-to-node communication is mutually authenticated via mTLS, with the certificate lifecycle managed independently from the license-signing key hierarchy.

**Verifiable Audit Compression.** Horizon Core's dual Merkle structure (binary for online commit, ternary for offline audit reconstruction) yields audit proofs of 512 bytes against 348,890 bytes of raw record data — a 99.85% reduction in audit-storage footprint — without sacrificing cryptographic verifiability.

A distinguishing commercial property: **license enforcement is a consensus-level invariant.** Client transaction volume is bounded by the validator set itself, not by a centralized licensing database. A license is valid only once committed into a block ratified by the active quorum, making issuance, renewal, and revocation tamper-evident and externally auditable.

**Validated performance (single-host development benchmark):** 26,344 TPS ingestion with zero data loss, and 4,150 TPS end-to-end commit — inclusive of PBFT messaging and fsync-grade durability — across 1,000,000 transactions. Multi-site distributed benchmarks are pending and will be published upon federated deployment.

## 2. Architecture & Dynamic Node Scaling

### 2.1 Core Topology

Horizon Core separates edge from consensus:

- **Edge tier — Cloudflare Worker (API Gateway).** The `horizon-core` Worker terminates client TLS, enforces authentication and rate-limiting, validates request shape, and routes authorized transactions to the consensus tier. It holds no ledger state; it is the public, horizontally-scaled entry point.
- **Consensus tier — Go backend cluster.** Validator nodes run the PBFT commit protocol, maintain the PostgreSQL online store and the SQLite air-gapped offline store, and produce ratified blocks. The dual Merkle trees are constructed here, and license invariants are enforced here.

This separation ensures the consensus core is never directly exposed to the public internet, and that edge scaling (Workers) is decoupled from consensus scaling (federated nodes).

### 2.2 The Eight Core Nodes

| Role | Count | Responsibility |
|---|---|---|
| Critical | 4 | Merkle construction, hash commitment, ECDSA verification, type validation |
| Majority | 3 | Volume accounting, timestamp ordering, user attestation |
| Agent Guardian | 1 | Offline statistical auditor; flags anomalies; **non-voting** |

The Agent Guardian is deliberately **non-voting**: it observes committed blocks and statistical deviations but cannot block or forge consensus. This preserves a symmetric PBFT fault model among the seven voting nodes while retaining an independent detection capability.

### 2.3 Federated Membership & Dynamic Quorum

Membership is not fixed. Each new enterprise client is issued a license and deploys a consensus node on its own physical infrastructure — a bank hosts Node 9 in its data center; an energy operator hosts Node 10 in another jurisdiction. The validator set therefore grows with the consortium, and the physical distribution of nodes across organizations and geographies eliminates the single-host failure correlation that a co-located deployment would introduce.

The commit quorum scales with active membership. For *N* active voting nodes, the protocol tolerates:

> **f = ⌊(N − 1) / 3⌋** Byzantine faults, with a commit quorum of **2f + 1**.

As N grows, the tolerated fault fraction converges to ⅓ — the theoretical optimum for partially-synchronous Byzantine consensus. With the initial seven voting nodes (f = 2), the minimum quorum is five; the platform's conservative default of six provides an additional safety margin.

### 2.4 Membership Reconfiguration

Nodes are not added mid-commit. Admission is **license-gated and epoch-bound**: a new client's node is proposed in a license block, ratified by the current quorum, and activated at a defined epoch boundary, avoiding consensus churn during ongoing rounds. Departure and revocation follow the same epoch-bound protocol, with a grace window to prevent liveness disruption when a client node goes offline.

### 2.5 Communication Complexity — Operating Range & Roadmap

PBFT's all-to-all voting carries O(n²) per-round message overhead, which bounds the practical validator count for the pure PBFT tier. Horizon Core targets an operating range of **8–~40 federated nodes** for the core consensus tier — consistent with real-world consortium BFT deployments. For larger federations, the roadmap specifies **threshold-signature aggregation (BLS)**, which collapses per-round communication to O(n) without altering the fault-tolerance bound, enabling federations of 100+ nodes at sustained throughput.
