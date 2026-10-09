package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestE2E_LicenseSigningRoundTrip(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	pubHex := hex.EncodeToString(priv.PublicKey.X.Bytes()) + hex.EncodeToString(priv.PublicKey.Y.Bytes())
	t.Logf("public key: %s...", pubHex[:16])

	payload := []byte(`{"license_id":"lic_test_001","user_id":"acme","product_id":"horizon-core","volume":100}`)

	sig, err := SignDataRS(payload, priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	sigHex := hex.EncodeToString(sig)
	t.Logf("signature: %s...", sigHex[:16])

	if len(sigHex) != 128 {
		t.Fatalf("signature hex = %d chars; want 128", len(sigHex))
	}

	ok, err := VerifyData(&priv.PublicKey, payload, sigHex)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if !ok {
		t.Fatal("valid signature was rejected")
	}

	tampered := []byte(strings.Replace(string(payload), "volume\":100", "volume\":999", 1))
	ok, _ = VerifyData(&priv.PublicKey, tampered, sigHex)
	if ok {
		t.Fatal("tampered payload was accepted")
	}

	tamperedSig := []byte(sigHex)
	if tamperedSig[0] == '0' {
		tamperedSig[0] = '1'
	} else {
		tamperedSig[0] = '0'
	}
	ok, _ = VerifyData(&priv.PublicKey, payload, string(tamperedSig))
	if ok {
		t.Fatal("tampered signature was accepted")
	}

	otherPriv, _ := GenerateSecp256k1ECDSAKey()
	ok, _ = VerifyData(&otherPriv.PublicKey, payload, sigHex)
	if ok {
		t.Fatal("wrong public key verified")
	}

	sig2, _ := SignDataRS(payload, priv)
	if hex.EncodeToString(sig2) != sigHex {
		t.Fatal("RFC6979 violated: signatures differ")
	}

	t.Log("E2E round-trip OK")
}
