# 🛡️ Horizon Core - Threat Model

**Target:** SCADA (Industrial), Banks, Critical Infrastructure.
**Architecture:** Air-gap capable, Sovereign-Neutral.

## 🎯 Adversaries (Who are we defending against?)
1. **Nation-State Actors:** Highly funded, capable of Quantum computing and 0-days (e.g., Stuxnet-level).
2. **Malicious Insiders:** Employees with access to the Switch or Nodes.
3. **Cyber Criminals:** Ransomware, Data theft, DDoS.
4. **Malicious Nodes:** Byzantine nodes in the consensus network.

## 🛡️ Defenses in Place
- **Zero-Storage:** No private keys on disk. (Defeats physical theft).
- **PBFT 5/7:** Tolerates up to 2 malicious/failed nodes.
- **Native secp256k1:** High-performance, battle-tested cryptography.
- **WAF & Rate Limit:** Blocks Layer 7 attacks (SQLi, XSS, DDoS).
- **mTLS (V1.4):** Encrypted communication between nodes.
- **Fail-Closed Switch:** Refuses to start without `ADMIN_TOKEN` (32+ chars).

## ⚠️ Out of Scope (Current Phase)
- **Quantum Attacks:** `secp256k1` is vulnerable to Shor's algorithm. PQC is optional but not active by default.
- **Physical SCADA Attacks:** Horizon Core secures the data layer, not the physical PLCs.
