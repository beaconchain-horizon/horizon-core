# Horizon Core — Explorer & Monitoring Protocol Rules

## Status: MANDATORY & IMMUTABLE

### 1. Red Line (Forbidden Data)
- **Hardware ID:** Under NO circumstances should the Hardware ID (or any substring/prefix of it) be displayed, searchable, or exposed in the Explorer or any public monitoring dashboard. Organizational identities (Banks, Defense, etc.) must remain completely anonymous.

### 2. Allowed Public Data (Searchable & Displayable)
The following data points MUST be available for search and visualization in the Explorer to prove mathematical integrity without compromising privacy:
1. **TxID** (Transaction ID)
2. **Timestamp / Block Number** (Unix 10-digit Time)
3. **Block Hash**
4. **Public Key** (Anonymous)
5. **Signed Merkle Root**

### 3. Mathematical Transparency
The Explorer must demonstrate the validity of the ledger via cryptographic proofs (Signatures, Merkle Roots) while strictly adhering to the Data Sovereignty model (zero personal data stored on-chain).
