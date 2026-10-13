package crypto

// Cross-verification against openssl.
// Run: go test -tags=crossverify -v ./internal/crypto/...

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const crossVerifyPayload = "horizon-core-cross-verification-payload-v1"

type ecdsaDERSignature struct {
	R *big.Int
	S *big.Int
}

func findOpenSSL(t *testing.T) string {
	t.Helper()
	p, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl not found in PATH; skipping cross-verification")
	}
	return p
}

func requireSecp256k1(t *testing.T, openssl string) {
	t.Helper()
	out, err := exec.Command(openssl, "ecparam", "-list_curves").Output()
	if err != nil {
		t.Skipf("openssl ecparam failed: %v", err)
	}
	if !bytes.Contains(out, []byte("secp256k1")) {
		t.Skip("openssl build does not support secp256k1")
	}
}

func generateOpenSSLKey(t *testing.T, openssl, keyPath, pubPath string) *ecdsa.PrivateKey {
	t.Helper()
	cmd := exec.Command(openssl, "ecparam", "-name", "secp256k1", "-genkey", "-noout", "-out", keyPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("openssl ecparam: %v\n%s", err, out)
	}
	cmd = exec.Command(openssl, "ec", "-in", keyPath, "-pubout", "-out", pubPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("openssl ec -pubout: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatal("failed to PEM-decode openssl key")
	}
	var sec1 struct {
		Version    int
		PrivateKey []byte
	}
	if _, err := asn1.Unmarshal(block.Bytes, &sec1); err != nil {
		t.Fatalf("parse SEC1: %v", err)
	}
	scalar := sec1.PrivateKey
	if len(scalar) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(scalar):], scalar)
		scalar = padded
	}
	if len(scalar) > 32 {
		scalar = scalar[len(scalar)-32:]
	}
	priv, err := Secp256k1PrivateKeyFromBytes(scalar)
	if err != nil {
		t.Fatalf("Secp256k1PrivateKeyFromBytes: %v", err)
	}
	return priv
}

// TestCrossVerify_HorizonSign_OpenSSLVerify proves that a signature
// produced by SignDataRS is accepted by openssl.
func TestCrossVerify_HorizonSign_OpenSSLVerify(t *testing.T) {
	openssl := findOpenSSL(t)
	requireSecp256k1(t, openssl)

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")
	pubPath := filepath.Join(dir, "pub.pem")
	dataPath := filepath.Join(dir, "data.bin")
	sigPath := filepath.Join(dir, "sig.der")

	priv := generateOpenSSLKey(t, openssl, keyPath, pubPath)

	if err := os.WriteFile(dataPath, []byte(crossVerifyPayload), 0644); err != nil {
		t.Fatal(err)
	}

	rsSig, err := SignDataRS([]byte(crossVerifyPayload), priv)
	if err != nil {
		t.Fatalf("SignDataRS: %v", err)
	}
	if len(rsSig) != 64 {
		t.Fatalf("expected 64-byte R||S, got %d", len(rsSig))
	}

	r := new(big.Int).SetBytes(rsSig[:32])
	s := new(big.Int).SetBytes(rsSig[32:])
	der, err := asn1.Marshal(ecdsaDERSignature{R: r, S: s})
	if err != nil {
		t.Fatalf("encode DER: %v", err)
	}
	if err := os.WriteFile(sigPath, der, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(openssl, "dgst", "-sha256", "-verify", pubPath, "-signature", sigPath, dataPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("openssl verify failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Verified OK") {
		t.Fatalf("openssl did not report Verified OK: %s", out)
	}
	t.Logf("Horizon -> openssl: %s", strings.TrimSpace(string(out)))
}

// TestCrossVerify_OpenSSLSign_HorizonVerify proves that a signature
// produced by openssl is accepted by VerifyHashRS.
func TestCrossVerify_OpenSSLSign_HorizonVerify(t *testing.T) {
	openssl := findOpenSSL(t)
	requireSecp256k1(t, openssl)

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")
	pubPath := filepath.Join(dir, "pub.pem")
	dataPath := filepath.Join(dir, "data.bin")
	sigPath := filepath.Join(dir, "sig.der")

	priv := generateOpenSSLKey(t, openssl, keyPath, pubPath)

	if err := os.WriteFile(dataPath, []byte(crossVerifyPayload), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(openssl, "dgst", "-sha256", "-sign", keyPath, "-out", sigPath, dataPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("openssl sign: %v\n%s", err, out)
	}

	der, err := os.ReadFile(sigPath)
	if err != nil {
		t.Fatal(err)
	}
	var parsed ecdsaDERSignature
	if _, err := asn1.Unmarshal(der, &parsed); err != nil {
		t.Fatalf("parse DER: %v", err)
	}

	n := secp256k1.S256().Params().N
	if parsed.S.Cmp(new(big.Int).Rsh(n, 1)) > 0 {
		t.Logf("openssl produced high-S; normalizing to low-S")
		parsed.S = new(big.Int).Sub(n, parsed.S)
	}

	rs := make([]byte, 64)
	parsed.R.FillBytes(rs[:32])
	parsed.S.FillBytes(rs[32:])

	hash := Sha256Hash([]byte(crossVerifyPayload))
	if !VerifyHashRS(&priv.PublicKey, hash, rs) {
		t.Fatal("Horizon rejected a valid openssl signature")
	}
	t.Logf("openssl -> Horizon: verified")
}

// TestCrossVerify_DeterministicSigning proves RFC 6979 determinism.
func TestCrossVerify_DeterministicSigning(t *testing.T) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	data := []byte(crossVerifyPayload)

	sig1, err := SignDataRS(data, priv)
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := SignDataRS(data, priv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sig1, sig2) {
		t.Fatalf("RFC 6979 violated:\n sig1=%s\n sig2=%s",
			hex.EncodeToString(sig1), hex.EncodeToString(sig2))
	}
	t.Logf("RFC 6979 deterministic: %s", hex.EncodeToString(sig1))
}
