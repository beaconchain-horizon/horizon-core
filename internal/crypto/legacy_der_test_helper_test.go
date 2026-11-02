package crypto

import "math/big"

// legacyECDSASignature is a test-only helper used by the crypto test
// suite to build ASN.1 DER encoded signatures. Production code no
// longer parses or accepts DER (see secp256k1.go VerifyData); these
// tests exist to prove DER is rejected.
type legacyECDSASignature struct {
	R *big.Int
	S *big.Int
}
