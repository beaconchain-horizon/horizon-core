package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func SHA256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func SignData(data []byte, privKey *ecdsa.PrivateKey) (string, error) {
	if privKey == nil {
		return "", fmt.Errorf("private key is nil")
	}
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", err
	}
	sig := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(sig), nil
}
