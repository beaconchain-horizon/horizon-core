package wots

import (
	"crypto/sha256"
	"testing"
)

func TestWOTSSignVerify(t *testing.T) {
	seed := sha256.Sum256([]byte("horizon-authority-seed-2026"))
	priv, pub, err := GenerateKeyPair(seed[:])
	if err != nil {
		t.Fatal(err)
	}

	msg := sha256.Sum256([]byte("license: HORIZON-MASTER-001 | volume: 4000000000"))
	sig, err := Sign(priv, msg[:])
	if err != nil {
		t.Fatal(err)
	}

	if !Verify(pub, msg[:], sig) {
		t.Fatal("valid signature was rejected")
	}

	// tamper with message
	badMsg := sha256.Sum256([]byte("license: HORIZON-MASTER-001 | volume: 100"))
	if Verify(pub, badMsg[:], sig) {
		t.Fatal("forged message was accepted")
	}

	// tamper with signature
	sig[100] ^= 0xFF
	if Verify(pub, msg[:], sig) {
		t.Fatal("tampered signature was accepted")
	}
}
