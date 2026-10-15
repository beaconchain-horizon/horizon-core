# HORIZON ANGEL - NODE LAW v1.0

## 1. Seven Specialty Nodes
| Port | Role   | Equation            | Vote     |
|------|--------|---------------------|----------|
| 9001 | Merkle | SHA256(left||right) | CRITICAL |
| 9002 | Hash   | SHA256(header)      | CRITICAL |
| 9003 | ECDSA  | secp256k1 Verify    | CRITICAL |
| 9004 | Type   | binary or ternary   | CRITICAL |
| 9005 | Volume | used < volume       | MAJORITY |
| 9006 | Time   | issued <= now < exp | MAJORITY |
| 9007 | User   | user_id match       | MAJORITY |

## 2. Consensus
4 CRITICAL (100%) + 2 of 3 MAJORITY = ACCEPT
Otherwise REJECT (fail-closed)

## 3. Rules
- Crash on violation
- Domain separation: 1 equation per node
- Ports: 8080 backend, 8091 central, 9001-9007 nodes, 9101+ rooms
- License single-use, revoked after consumption

Version 1.0 - 2026-10-15
