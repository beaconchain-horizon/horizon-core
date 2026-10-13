package crypto

import (
	"encoding/hex"
	"math/big"
	"testing"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestVerify_AcceptsValidLowS is the baseline: a signature produced by
// SignDataRS must be accepted by VerifyHashRS.
func TestVerify_AcceptsValidLowS(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	data := []byte("horizon-malleability-test-payload")
	sig, err := SignDataRS(data, priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !VerifyHashRS(&priv.PublicKey, Sha256Hash(data), sig) {
		t.Fatal("valid low-S signature was rejected")
	}
}

// TestVerify_RejectsHighS proves that ECDSA malleability is closed:
// flipping S to (n - S) produces a mathematically valid signature that
// MUST be rejected, otherwise a duplicate transaction can enter the
// Merkle tree.
func TestVerify_RejectsHighS(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	data := []byte("horizon-malleability-test-payload")
	sig, err := SignDataRS(data, priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(sig) != Secp256k1SignatureSize {
		t.Fatalf("unexpected signature length: %d", len(sig))
	}

	n := secp256k1.S256().Params().N
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	highS := new(big.Int).Sub(n, s)

	flipped := make([]byte, Secp256k1SignatureSize)
	r.FillBytes(flipped[:32])
	highS.FillBytes(flipped[32:])

	halfN := new(big.Int).Rsh(n, 1)
	if highS.Cmp(halfN) <= 0 {
		t.Fatalf("test setup error: flipped S is not over half order")
	}

	hash := Sha256Hash(data)
	if VerifyHashRS(&priv.PublicKey, hash, flipped) {
		t.Fatal("high-S signature was accepted — malleability vulnerability")
	}
}

// TestVerify_RejectsLegacyDER proves that DER/ASN.1 encoded signatures
// are rejected outright, per CRYPTO-CONTRACT.md. Accepting DER on top
// of raw R||S would re-open the same malleability surface through
// encoding differences.
func TestVerify_RejectsLegacyDER(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	// Minimal syntactically valid ASN.1 SEQUENCE { INTEGER, INTEGER }.
	// Values are dummies; the point is that the *format* must be rejected
	// before any parsing occurs.
	der := []byte{
		0x30, 0x06,
		0x02, 0x01, 0x01,
		0x02, 0x01, 0x02,
	}

	ok, err := VerifyData(&priv.PublicKey, []byte("payload"), hex.EncodeToString(der))
	if err == nil {
		t.Fatalf("DER signature should produce an error, got ok=%v err=nil", ok)
	}
	if ok {
		t.Fatal("DER signature was accepted")
	}
}

// TestVerify_RejectsWrongLength proves that anything that is not exactly
// 64 bytes (raw R||S) is rejected with an explicit error.
func TestVerify_RejectsWrongLength(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	for _, length := range []int{0, 32, 63, 65, 128} {
		buf := make([]byte, length)
		ok, err := VerifyData(&priv.PublicKey, []byte("payload"), hex.EncodeToString(buf))
		if err == nil {
			t.Fatalf("length %d: expected error, got ok=%v err=nil", length, ok)
		}
		if ok {
			t.Fatalf("length %d: signature was accepted", length)
		}
	}
}
