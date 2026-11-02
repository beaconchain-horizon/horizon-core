package crypto

import (
	"encoding/hex"
	"testing"
)

func TestSecp256k1PublicKeyFromHex_64Bytes(t *testing.T) {
	priv, _ := GenerateSecp256k1ECDSAKey()
	var xb, yb [32]byte
	priv.PublicKey.X.FillBytes(xb[:])
	priv.PublicKey.Y.FillBytes(yb[:])
	hexStr := hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:])

	pub, err := Secp256k1PublicKeyFromHex(hexStr)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if pub.X.Cmp(priv.PublicKey.X) != 0 || pub.Y.Cmp(priv.PublicKey.Y) != 0 {
		t.Fatal("recovered public key mismatch")
	}
}

func TestSecp256k1PublicKeyFromHex_RejectsBadLength(t *testing.T) {
	for _, bad := range []string{"", "00", "abcd", string(make([]byte, 128))} {
		if _, err := Secp256k1PublicKeyFromHex(bad); err == nil {
			t.Fatalf("bad hex %q was accepted", bad)
		}
	}
}
