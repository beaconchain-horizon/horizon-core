package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"

	"fmt"
	"math/big"
	"strings"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpECDSA "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

const (
	horizonSecpPrivatePEMType = "HORIZON SECP256K1 PRIVATE KEY"
	horizonSecpPublicPEMType  = "HORIZON SECP256K1 PUBLIC KEY"
	Secp256k1SignatureSize    = 64
)

func GenerateSecp256k1ECDSAKey() (*ecdsa.PrivateKey, error) {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	return priv.ToECDSA(), nil
}

func Secp256k1PrivateKeyFromBytes(b []byte) (*ecdsa.PrivateKey, error) {
	if len(b) != secp256k1.PrivKeyBytesLen {
		return nil, fmt.Errorf("secp256k1 private key must be 32 bytes, got %d", len(b))
	}

	n := secp256k1.S256().Params().N
	d := new(big.Int).SetBytes(b)

	if d.Sign() <= 0 || d.Cmp(n) >= 0 {
		return nil, errors.New("secp256k1 private key is out of range")
	}

	priv := secp256k1.PrivKeyFromBytes(b)
	return priv.ToECDSA(), nil
}

func Secp256k1PrivateKeyFromHex(value string) (*ecdsa.PrivateKey, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")

	raw, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("private key hex: %w", err)
	}

	return Secp256k1PrivateKeyFromBytes(raw)
}

func signKeyBytes(priv *ecdsa.PrivateKey) ([]byte, error) {
	if priv == nil || priv.D == nil || priv.Curve == nil {
		return nil, errors.New("private key or curve is nil")
	}

	curve := secp256k1.S256()
	params := priv.Curve.Params()
	want := curve.Params()
	if params == nil || params.Name != want.Name || params.N == nil || params.N.Cmp(want.N) != 0 {
		return nil, errors.New("private key curve is not secp256k1")
	}
	if priv.D.Sign() <= 0 || priv.D.Cmp(want.N) >= 0 {
		return nil, errors.New("private key scalar is invalid for secp256k1")
	}

	out := make([]byte, 32)
	priv.D.FillBytes(out)
	expected := secp256k1.PrivKeyFromBytes(out).ToECDSA()
	if priv.X == nil || priv.Y == nil || priv.X.Cmp(expected.X) != 0 || priv.Y.Cmp(expected.Y) != 0 {
		return nil, errors.New("private/public key mismatch for secp256k1")
	}
	return out, nil
}

func SignHashRS(priv *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
	if len(hash) != 32 {
		return nil, errors.New("hash must be exactly 32 bytes")
	}

	keyBytes, err := signKeyBytes(priv)
	if err != nil {
		return nil, err
	}

	key := secp256k1.PrivKeyFromBytes(keyBytes)
	sig := secpECDSA.Sign(key, hash)

	var out [Secp256k1SignatureSize]byte

	r := sig.R()
	s := sig.S()

	r.PutBytesUnchecked(out[:32])
	s.PutBytesUnchecked(out[32:])

	return out[:], nil
}

func SignDataRS(data []byte, priv *ecdsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key is nil")
	}

	digest := Sha256Hash(data)

	return SignHashRS(priv, digest)
}

func parseRS(sig []byte) (*big.Int, *big.Int, error) {
	if len(sig) != Secp256k1SignatureSize {
		return nil, nil, fmt.Errorf(
			"R||S signature must be exactly 64 bytes, got %d",
			len(sig),
		)
	}

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])

	n := secp256k1.S256().Params().N

	if r.Sign() <= 0 || s.Sign() <= 0 {
		return nil, nil, errors.New("R or S is zero")
	}

	if r.Cmp(n) >= 0 || s.Cmp(n) >= 0 {
		return nil, nil, errors.New("R or S is outside secp256k1 range")
	}

	// Reject high-S signatures. ECDSA allows both (r, s) and (r, n-s)
	// to verify for the same message and key, which enables signature
	// malleability: a third party can duplicate a valid transaction
	// with a different signature and produce two distinct Merkle leaves.
	// Low-S only is enforced by RFC 6979 and BIP-62.
	halfN := new(big.Int).Rsh(n, 1)
	if s.Cmp(halfN) > 0 {
		return nil, nil, errors.New("S is over half order (high-S rejected)")
	}

	return r, s, nil
}

func validatedSecp256k1PublicKey(pub *ecdsa.PublicKey) (*ecdsa.PublicKey, bool) {
	if pub == nil || pub.Curve == nil || pub.X == nil || pub.Y == nil {
		return nil, false
	}

	curve := secp256k1.S256()
	params := pub.Curve.Params()
	want := curve.Params()
	if params == nil || want == nil || params.Name != want.Name || params.BitSize != want.BitSize ||
		params.P == nil || params.N == nil || params.B == nil || params.Gx == nil || params.Gy == nil ||
		params.P.Cmp(want.P) != 0 || params.N.Cmp(want.N) != 0 || params.B.Cmp(want.B) != 0 ||
		params.Gx.Cmp(want.Gx) != 0 || params.Gy.Cmp(want.Gy) != 0 {
		return nil, false
	}
	if !curve.IsOnCurve(pub.X, pub.Y) {
		return nil, false
	}
	return &ecdsa.PublicKey{Curve: curve, X: new(big.Int).Set(pub.X), Y: new(big.Int).Set(pub.Y)}, true
}

func VerifyHashRS(pub *ecdsa.PublicKey, hash []byte, sig []byte) bool {
	canonicalPub, valid := validatedSecp256k1PublicKey(pub)
	if !valid || len(hash) != 32 {
		return false
	}

	r, s, err := parseRS(sig)
	if err != nil {
		return false
	}
	return ecdsa.Verify(canonicalPub, hash, r, s)
}

func VerifyData(pub *ecdsa.PublicKey, data []byte, signatureHex string) (bool, error) {
	canonicalPub, valid := validatedSecp256k1PublicKey(pub)
	if !valid {
		return false, errors.New("public key must be a valid secp256k1 public key")
	}

	signatureHex = strings.TrimSpace(signatureHex)
	signatureHex = strings.TrimPrefix(signatureHex, "0x")
	signatureHex = strings.TrimPrefix(signatureHex, "0X")
	sig, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, fmt.Errorf("bad signature hex: %w", err)
	}

	hash := Sha256Hash(data)
	if len(sig) != Secp256k1SignatureSize {
		return false, fmt.Errorf(
			"signature must be exactly %d bytes (raw R||S); got %d bytes — DER/ASN.1 is not accepted",
			Secp256k1SignatureSize, len(sig),
		)
	}
	return VerifyHashRS(canonicalPub, hash, sig), nil
}

func VerifySignature(pub *ecdsa.PublicKey, data []byte, signatureHex string) bool {
	ok, _ := VerifyData(pub, data, signatureHex)
	return ok
}

// Secp256k1PublicKeyFromHex parses a secp256k1 public key from hex.
// Accepts 33-byte compressed, 64-byte X||Y, or 65-byte uncompressed.
func Secp256k1PublicKeyFromHex(value string) (*ecdsa.PublicKey, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")

	raw, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("public key hex: %w", err)
	}

	switch len(raw) {
	case 33:
		pub, err := secp256k1.ParsePubKey(raw)
		if err != nil {
			return nil, err
		}
		return pub.ToECDSA(), nil
	case 64:
		x := new(big.Int).SetBytes(raw[:32])
		y := new(big.Int).SetBytes(raw[32:])
		return validatedOrError(x, y)
	case 65:
		if raw[0] != 0x04 {
			return nil, errors.New("invalid uncompressed secp256k1 public key prefix")
		}
		x := new(big.Int).SetBytes(raw[1:33])
		y := new(big.Int).SetBytes(raw[33:65])
		return validatedOrError(x, y)
	default:
		return nil, fmt.Errorf("unsupported secp256k1 public key length: %d bytes", len(raw))
	}
}

func validatedOrError(x, y *big.Int) (*ecdsa.PublicKey, error) {
	curve := secp256k1.S256()
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("public key is not on secp256k1 curve")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
func PrivateKeyPEM(priv *ecdsa.PrivateKey) (string, error) {
	keyBytes, err := signKeyBytes(priv)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  horizonSecpPrivatePEMType,
		Bytes: keyBytes,
	})), nil
}

func ParsePrivateKeyPEM(value string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}

	switch block.Type {
	case horizonSecpPrivatePEMType:
		return Secp256k1PrivateKeyFromBytes(block.Bytes)
	case "EC PRIVATE KEY":
		return nil, errors.New("legacy EC PRIVATE KEY format is disabled; generate a fresh secp256k1 key and rotate the signing identity")
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ec, ok := key.(*ecdsa.PrivateKey)
		if !ok || ec.D == nil || ec.Curve == nil || ec.Curve.Params() == nil {
			return nil, errors.New("PKCS8 key is not a valid ECDSA private key")
		}
		curve := secp256k1.S256()
		params, want := ec.Curve.Params(), curve.Params()
		if params.Name != want.Name || params.N == nil || params.N.Cmp(want.N) != 0 {
			return nil, errors.New("PKCS8 key is not secp256k1; refusing to reinterpret its scalar")
		}
		if ec.D.Sign() <= 0 || ec.D.Cmp(want.N) >= 0 {
			return nil, errors.New("PKCS8 secp256k1 private key scalar is out of range")
		}
		raw := make([]byte, 32)
		ec.D.FillBytes(raw)
		expected := secp256k1.PrivKeyFromBytes(raw).ToECDSA()
		if ec.X == nil || ec.Y == nil || ec.X.Cmp(expected.X) != 0 || ec.Y.Cmp(expected.Y) != 0 {
			return nil, errors.New("PKCS8 key public/private mismatch for secp256k1")
		}
		return expected, nil
	default:
		return nil, fmt.Errorf("unsupported private key PEM type: %s", block.Type)
	}
}

func secp256k1PublicKeyPEM(priv *ecdsa.PrivateKey) (string, error) {
	keyBytes, err := signKeyBytes(priv)
	if err != nil {
		return "", err
	}

	key := secp256k1.PrivKeyFromBytes(keyBytes)

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  horizonSecpPublicPEMType,
		Bytes: key.PubKey().SerializeCompressed(),
	})), nil
}

func ParsePublicKeyPEM(value string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}

	switch block.Type {
	case horizonSecpPublicPEMType:
		pub, err := secp256k1.ParsePubKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return pub.ToECDSA(), nil

	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ec, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not ECDSA")
		}
		canonical, valid := validatedSecp256k1PublicKey(ec)
		if !valid {
			return nil, errors.New("public key is not a valid secp256k1 key")
		}
		return canonical, nil

	default:
		return nil, fmt.Errorf("unsupported public key PEM type: %s", block.Type)
	}
}

func PublicAddressFromKey(priv *ecdsa.PrivateKey) string {
	keyBytes, err := signKeyBytes(priv)
	if err != nil {
		return ""
	}
	key := secp256k1.PrivKeyFromBytes(keyBytes)
	pub := key.PubKey().SerializeCompressed()
	h := Sha256Hash(pub)
	return "0x" + hex.EncodeToString(h[:20])
}
