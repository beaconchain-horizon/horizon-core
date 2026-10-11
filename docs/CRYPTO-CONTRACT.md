# Horizon v1.4 — Crypto Contract

Canonical specification for license signing and verification.

## Curve

- **Algorithm**: ECDSA
- **Curve**: secp256k1 (SEC 2, same as Bitcoin/Ethereum)
- **Library**: `github.com/decred/dcrd/dcrec/secp256k1/v4` (server), `@noble/secp256k1` v2.1.0 (browser)

## Signature format

- **Wire format**: raw R||S, exactly 64 bytes
- **Hex encoding**: 128 lowercase hex chars
- **No DER, no ASN.1** — legacy DER is rejected
- **High-S rejected** (malleability protection)

## Hashing

- **Hash**: SHA-256 over the canonical payload
- **Determinism**: RFC 6979 (no randomness)
- **Domain**: none at v1.4 (payload already namespaced by field order)

## Canonical payload (pipe-separated)

```
<license_id>|<product_id>|<user_id>|<volume>|<merkle_root>|<status>|<expires_at_RFC3339_Z>
```

- `expires_at_RFC3339_Z`: UTC, seconds precision, no nanoseconds (e.g. `2027-10-11T11:31:12Z`)
- `merkle_root`: 64 lowercase hex chars, or empty string
- `status`: literal `active` at signing time

Any change to any byte of the payload invalidates the signature.

## Key storage

| Layer | Key |
|---|---|
| Authority / USB | **private key only here** |
| Backend | private key (currently; will move to Authority) |
| Switch | **public key only** |
| Browser | public key only, never private |
| Repo | never |

## PEM format

- Header: `HORIZON SECP256K1 PRIVATE KEY`
- Body: raw 32-byte private scalar, PEM-base64 encoded
- Alternative accepted: PKCS#8 `PRIVATE KEY` with secp256k1 curve
- Rejected: `EC PRIVATE KEY` (P-256 legacy)

## Public key encoding

- 128 hex chars = X (32 bytes) || Y (32 bytes), uncompressed, no prefix
- Accepted variants on input: 33-byte compressed, 65-byte uncompressed with 0x04 prefix, 128-hex

## Verification rules

The verifier MUST reject:

1. Signature length ≠ 64 bytes
2. Non-hex signature
3. R = 0 or S = 0
4. R ≥ N or S ≥ N (N = secp256k1 order)
5. S > N/2 (high-S)
6. Public key not on secp256k1
7. Public key on P-256 or any other curve
8. Any byte change in the payload

## Test coverage

See `internal/crypto/secp256k1_test.go` and `internal/crypto/e2e_test.go`:

- `TestSign_IsDeterministic`
- `TestSign_FormatIsExactly64Bytes`
- `TestSignVerify_RoundTrip`
- `TestVerify_RejectsModifiedPayload`
- `TestVerify_RejectsModifiedSignature`
- `TestVerify_RejectsWrongPublicKey`
- `TestVerify_RejectsMalformedHex`
- `TestVerify_RejectsWrongSignatureLength`
- `TestVerify_RejectsZeroRorS`
- `TestVerify_RejectsOutOfRangeRS`
- `TestVerifyDataRejectsP256PublicKey`
- `TestVerify_RejectsLegacyDER`
