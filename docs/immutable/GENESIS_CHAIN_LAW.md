# Horizon Core — The Immutable Genesis & Chain Law

## Status: MANDATORY, ABSOLUTE, AND IMMUTABLE

### 1. The One and Only Genesis
- There is ONLY ONE Genesis Block (Block 0). It is absolute, singular, and its hash (`8c0763...`) is the Root of Trust for the entire ecosystem.
- No new Genesis blocks shall ever be created. The chain always extends from this single root.

### 2. Absolute Chain Linkage (No Orphans)
- Every block MUST contain the cryptographic hash of the IMMEDIATELY PRECEDING block (`prev_hash`).
- No transaction, signature, license, or node can exist in isolation. Every new block must recursively link back to the Genesis Block.
- If a block's `prev_hash` does not match the current head of the chain (or any historical block), the block is mathematically VOID and MUST be rejected by the 7 Consensus Nodes.

### 3. Consequence of Violation
- Any attempt to create a rogue node, a detached block, or a "Solid" license without chain linkage will be instantly rejected by the `Hash` and `Merkle` validators.
- The chain is a continuous, unbroken mathematical rope starting from the Genesis. Any break in this rope results in automatic failure (Fail-Closed).
