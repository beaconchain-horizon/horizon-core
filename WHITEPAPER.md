# Horizon Core v1.4 — Sovereign Federated Ledger
### Technical Whitepaper

## 1. Executive Summary
Horizon Core is a sovereign, air-gapped, industrial-grade distributed ledger engineered for regulated industries — banking, energy, and critical infrastructure — where data residency, Byzantine resilience, and tamper-evident auditability are non-negotiable. Unlike first-generation consortium ledgers whose default ordering services are crash-fault-tolerant (CFT) rather than Byzantine-fault-tolerant (BFT), Horizon Core treats adversarial — not merely fail-stop — behavior as its baseline assumption.

The platform is defined by three architectural pillars:
- **Federated Practical Byzantine Fault Tolerance (F-PBFT).** Consensus begins with eight core nodes (4 Critical, 3 Majority, 1 Agent Guardian). As enterprise clients onboard, each deploys its own consensus node on its own physical infrastructure, distributing the validator set across organizations. The commit quorum scales dynamically with membership, preserving a ⅓ Byzantine fault tolerance at every scale.
- **Air-Gap Security Boundary.** Key material never persists on networked storage. Consensus signing uses secp256k1 keys provisioned exclusively through a documented air-gap transfer protocol. Node-to-node communication is mutually authenticated via mTLS.
- **Verifiable Audit Compression.** Horizon Core's dual Merkle structure yields audit proofs of 512 bytes against 348,890 bytes of raw record data — a 99.85% reduction in audit-storage footprint — without sacrificing cryptographic verifiability.

A distinguishing commercial property: **license enforcement is a consensus-level invariant.** Client transaction volume is bounded by the validator set itself. A license is valid only once committed into a block ratified by the active quorum.

**Validated performance (single-host development benchmark):** 26,344 TPS ingestion with zero data loss, and 4,150 TPS end-to-end commit across 1,000,000 transactions.

## 2. Architecture & Dynamic Node Scaling

### 2.1 Core Topology
Horizon Core separates edge from consensus:
- **Edge tier — Cloudflare Worker (API Gateway).** Terminates client TLS, enforces authentication and rate-limiting, validates request shape, and routes authorized transactions to the consensus tier. It holds no ledger state.
- **Consensus tier — Go backend cluster.** Validator nodes run the PBFT commit protocol, maintain the PostgreSQL online store and the SQLite air-gapped offline store, and produce ratified blocks.

### 2.2 The Eight Core Nodes
| Role | Count | Responsibility |
|---|---|---|
| Critical | 4 | Merkle construction, hash commitment, ECDSA verification, type validation |
| Majority | 3 | Volume accounting, timestamp ordering, user attestation |
| Agent Guardian | 1 | Offline statistical auditor; flags anomalies; **non-voting** |

### 2.3 Federated Membership & Dynamic Quorum
Membership is not fixed. Each new enterprise client is issued a license and deploys a consensus node on its own physical infrastructure. The validator set grows with the consortium, eliminating single-host failure correlation.
The commit quorum scales with active membership. For *N* active voting nodes, the protocol tolerates: **f = ⌊(N − 1) / 3⌋** Byzantine faults, with a commit quorum of **2f + 1**.

### 2.4 Communication Complexity
PBFT's all-to-all voting carries O(n²) per-round message overhead. Horizon Core targets an operating range of **8–~40 federated nodes**. For larger federations, the roadmap specifies **threshold-signature aggregation (BLS)** for 100+ nodes.

## 3. Fault Tolerance Matrix

**Exact formulation:** f = ⌊(N−1)/3⌋ Byzantine faults; commit quorum = 2f + 1.

| N (voting) | f (Byzantine) | Quorum | Fault Fraction |
|---:|---:|---:|---:|
| 7 (initial) | 2 | 5 | 28.6% |
| 10 | 3 | 7 | 30.0% |
| 20 | 6 | 13 | 30.0% |
| 40 | 13 | 27 | 32.5% |
| → ∞ | → N/3 | → 2N/3 | → 33.3% |

- **O(n²) ceiling:** ~2n²+n messages per consensus instance; pure PBFT practical ceiling at ~40 nodes. BLS aggregation collapses this to O(n) for 100+ node federations.
- **Epoch-transition liveness:** Membership changes are epoch-bound with a configurable grace window (default 100 blocks). The departing node continues participating under the old membership M_E until the boundary, so the live set never drops below quorum during churn.

## 4. Security & Threat Model

- **STRIDE mapping** — all six categories with specific controls (Zero-PEM for Information Disclosure, dual Merkle for Tampering, quorum-ratified licenses for Elevation of Privilege).
- **MITRE ATT&CK** — representative techniques mapped (T1552 credentials-in-files → Zero-PEM; T1091 removable media → read-only USB + locked appliance; T1190 public app → Worker edge + WAF).
- **Three-zone air-gap model:**
  - **Zone 1 (air-gapped):** Key provisioning — secp256k1 keys loaded via read-only USB into node RAM at boot; Zero-PEM invariant (never on disk).
  - **Zone 2 (networked, NOT air-gapped):** Online consensus — Go nodes + PostgreSQL + Worker edge; keys in RAM, mTLS between nodes.
  - **Zone 3 (air-gapped):** Offline SQLite audit store; one-way data flow from Zone 2 via read-only media.

## 5. Edge / Backend Topology

- **Cloudflare Worker (`horizon-core`):** TLS termination, auth, rate-limiting, request validation, routing. Stateless, horizontally scaled, no ledger state.
- **Go Backend Cluster:** PBFT consensus, dual Merkle construction, license enforcement, fsync commit to PostgreSQL, one-way audit emission to Zone 3.
- **Failure isolation:** edge failure → anycast reroute; node failure → quorum absorbs; PostgreSQL failure → rebuild from block log; audit zone decoupled (async, no throughput impact).

## Appendix A (Forthcoming)
- Security Evidence Register (192 vectors STRIDE mapping)
- Benchmark Methodology (Hardware specs, fsync settings, multi-site pending)
