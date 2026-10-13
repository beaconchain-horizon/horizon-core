package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"testing"
)

func TestSecp256k1SignVerifyAndPEMRoundTrip(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("generate secp256k1 key: %v", err)
	}

	data := []byte(`{"sensor_id":"test-001","value":42.5,"nonce":"abc123","timestamp":1791540000}`)

	sig, err := SignDataRS(data, priv)
	if err != nil {
		t.Fatalf("sign data: %v", err)
	}
	if len(sig) != Secp256k1SignatureSize {
		t.Fatalf("signature length = %d; want %d", len(sig), Secp256k1SignatureSize)
	}

	if !VerifyHashRS(&priv.PublicKey, Sha256Hash(data), sig) {
		t.Fatal("signature did not verify")
	}

	ok, err := VerifyData(&priv.PublicKey, data, hex.EncodeToString(sig))
	if err != nil || !ok {
		t.Fatalf("VerifyData failed: valid=%v err=%v", ok, err)
	}

	if VerifySignature(&priv.PublicKey, []byte("different message"), hex.EncodeToString(sig)) {
		t.Fatal("signature unexpectedly verified for different data")
	}

	privatePEM, err := PrivateKeyPEM(priv)
	if err != nil {
		t.Fatalf("encode private key: %v", err)
	}

	loaded, err := ParsePrivateKeyPEM(privatePEM)
	if err != nil {
		t.Fatalf("parse private key PEM: %v", err)
	}

	sig2, err := SignDataRS(data, loaded)
	if err != nil {
		t.Fatalf("sign with reloaded key: %v", err)
	}
	if !VerifyHashRS(&loaded.PublicKey, Sha256Hash(data), sig2) {
		t.Fatal("signature from reloaded key did not verify")
	}
}

func TestRejectP256ForSecp256k1Signing(t *testing.T) {
	legacyKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate P-256 test key: %v", err)
	}

	if _, err := SignDataRS([]byte("test"), legacyKey); err == nil {
		t.Fatal("P-256 key was unexpectedly accepted by secp256k1 signer")
	}

	der, err := x509.MarshalECPrivateKey(legacyKey)
	if err != nil {
		t.Fatalf("marshal legacy P-256 test key: %v", err)
	}
	legacyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	})

	if _, err := ParsePrivateKeyPEM(string(legacyPEM)); err == nil {
		t.Fatal("legacy EC PRIVATE KEY was unexpectedly accepted")
	}
}

func TestVerifyDataRejectsP256DERAndPublicPEM(t *testing.T) {
	legacyKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate P-256 test key: %v", err)
	}
	data := []byte("legacy P-256 DER must not pass secp256k1 verification")
	r, s, err := ecdsa.Sign(rand.Reader, legacyKey, Sha256Hash(data))
	if err != nil {
		t.Fatalf("create legacy DER signature: %v", err)
	}
	der, err := asn1.Marshal(legacyECDSASignature{R: r, S: s})
	if err != nil {
		t.Fatalf("encode legacy DER signature: %v", err)
	}
	valid, err := VerifyData(&legacyKey.PublicKey, data, hex.EncodeToString(der))
	if valid || err == nil {
		t.Fatal("P-256 DER signature was not rejected")
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&legacyKey.PublicKey)
	if err != nil {
		t.Fatalf("encode P-256 public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	if _, err := ParsePublicKeyPEM(string(publicPEM)); err == nil {
		t.Fatal("P-256 public key PEM was not rejected")
	}
}

// TestVerifyDataRejectsSecp256k1DER proves DER/ASN.1 is rejected
// outright per CRYPTO-CONTRACT.md. Accepting DER alongside raw
// R||S would re-open ECDSA malleability via encoding differences.
func TestVerifyDataRejectsSecp256k1DER(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("generate secp256k1 key: %v", err)
	}
	data := []byte("DER must be rejected even for secp256k1")
	r, s, err := ecdsa.Sign(rand.Reader, priv, Sha256Hash(data))
	if err != nil {
		t.Fatalf("create secp256k1 DER signature: %v", err)
	}
	der, err := asn1.Marshal(legacyECDSASignature{R: r, S: s})
	if err != nil {
		t.Fatalf("encode secp256k1 DER signature: %v", err)
	}
	valid, err := VerifyData(&priv.PublicKey, data, hex.EncodeToString(der))
	if valid || err == nil {
		t.Fatalf("secp256k1 DER signature should be rejected: valid=%v err=%v", valid, err)
	}
}

func TestLegacyDERSignatureRejectsTrailingData(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("generate secp256k1 key: %v", err)
	}
	data := []byte("reject trailing bytes after legacy DER signature")
	r, s, err := ecdsa.Sign(rand.Reader, priv, Sha256Hash(data))
	if err != nil {
		t.Fatalf("sign test data: %v", err)
	}
	der, err := asn1.Marshal(legacyECDSASignature{R: r, S: s})
	if err != nil {
		t.Fatalf("marshal DER signature: %v", err)
	}
	der = append(der, 0x00)
	valid, err := VerifyData(&priv.PublicKey, data, hex.EncodeToString(der))
	if valid || err == nil {
		t.Fatal("DER signature with trailing bytes was not rejected")
	}
}
