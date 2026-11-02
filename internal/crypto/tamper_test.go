package crypto

import (
	"encoding/hex"
	"testing"
)

// TestTamper_RejectsAllSingleByteFlips exhaustively flips each of the
// 64 signature bytes and asserts that every variant is rejected.
func TestTamper_RejectsAllSingleByteFlips(t *testing.T) {
	priv, _ := GenerateSecp256k1ECDSAKey()
	payload := []byte("HORIZON-GOLD-001|horizon-core|acme|100||active|2027-10-11T00:00:00Z")
	hash := Sha256Hash(payload)

	sig, err := SignHashRS(priv, hash)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	for i := 0; i < len(sig); i++ {
		tampered := make([]byte, len(sig))
		copy(tampered, sig)
		tampered[i] ^= 0x01

		if VerifyHashRS(&priv.PublicKey, hash, tampered) {
			t.Fatalf("byte %d flip was accepted", i)
		}
	}
}

// TestTamper_RejectsPayloadChange asserts that changing any of the
// signed fields invalidates the signature.
func TestTamper_RejectsPayloadChange(t *testing.T) {
	priv, _ := GenerateSecp256k1ECDSAKey()

	goodPayload := []byte("HORIZON-GOLD-001|horizon-core|acme|100||active|2027-10-11T00:00:00Z")
	hash := Sha256Hash(goodPayload)

	sig, _ := SignHashRS(priv, hash)
	sigHex := hex.EncodeToString(sig)

	badPayloads := [][]byte{
		[]byte("HORIZON-GOLD-002|horizon-core|acme|100||active|2027-10-11T00:00:00Z"),
		[]byte("HORIZON-GOLD-001|horizon-core-evil|acme|100||active|2027-10-11T00:00:00Z"),
		[]byte("HORIZON-GOLD-001|horizon-core|acme|999||active|2027-10-11T00:00:00Z"),
		[]byte("HORIZON-GOLD-001|horizon-core|acme|100||revoked|2027-10-11T00:00:00Z"),
		[]byte("HORIZON-GOLD-001|horizon-core|acme|100||active|2027-10-12T00:00:00Z"),
	}

	for _, bp := range badPayloads {
		ok, _ := VerifyData(&priv.PublicKey, bp, sigHex)
		if ok {
			t.Fatalf("tampered payload accepted: %s", bp)
		}
	}
}

// TestTamper_RejectsCrossKey asserts that a signature from one key
// does not verify under another.
func TestTamper_RejectsCrossKey(t *testing.T) {
	privA, _ := GenerateSecp256k1ECDSAKey()
	privB, _ := GenerateSecp256k1ECDSAKey()

	payload := []byte("cross-key test")
	sig, _ := SignDataRS(payload, privA)

	if VerifyHashRS(&privB.PublicKey, Sha256Hash(payload), sig) {
		t.Fatal("signature from A verified under B")
	}
}
