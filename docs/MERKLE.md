# Merkle Tree

## Current formula (v2)

```
leaf(data)       = SHA256(data)
node(left,right) = SHA256(left || right)
odd level        = SHA256(last || last)
```

No domain separation. Intentional for v2. Used by five call sites, all byte-identical (proven by internal/merkle/conform_test.go):

- internal/chain/append.go
- internal/license/protocol.go
- cmd/switch/license_sign.go
- cmd/regen-license/main.go
- internal/merkle.BinaryRootNoPrefix

## Historical note

v1 used RFC 6962 (0x00 leaf prefix, 0x01 node prefix, largest-power-of-two split). References to RFC 6962 in older docs are v1 legacy and marked as such.

## Security consideration

Without domain separation, a second-preimage attack (CVE-2012-2459 class) is theoretically possible if an adversary can inject a 64-byte payload equal to the concatenation of two leaf hashes. Mitigations:

- Payloads are canonical pipe-separated text, not arbitrary binary.
- License scope limits each root to one customer/branch.
- Any change requires a protocol version bump.
