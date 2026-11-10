# Horizon Core v1.4 — Enterprise Sales Pitch

## One-Page Executive Brief

### The Problem
Banks and critical-infrastructure operators need a tamper-evident transaction log that survives internal compromise. Traditional databases cannot provide this. Public blockchains fail on throughput and data exposure. Consortium ledgers (Fabric, Corda) are not Byzantine-hard by default.

### The Solution
Horizon Core is a sovereign distributed ledger engineered for regulated industries where adversarial behavior is the baseline threat.
1. **Federated PBFT:** Tolerates 1/3 Byzantine nodes. Clients deploy their own nodes.
2. **Air-Gap Security:** Keys never touch network storage. Supervised boot ceremony.
3. **Verifiable Audit Compression:** 99.85% storage reduction using dual Merkle.
4. **Commercial Differentiator:** License enforcement is a consensus-level invariant.

### Verified Performance
- Ingestion: 26,344 TPS
- Commit (PBFT + fsync): 4,150 TPS
- 1,000,000 tx test: 0.00% drop
- 192 attack vectors blocked (STRIDE methodology)

### Engagement
1. Technical deep-dive
2. Pilot deployment
3. Consortium membership

