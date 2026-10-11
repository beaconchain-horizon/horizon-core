package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"

	hcrypto "horizon-core/internal/crypto"
)

// Sha256Hash is retained for compatibility with existing package callers.
func Sha256Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// SignData delegates to the shared secp256k1 signing implementation.
func SignData(data []byte, privKey *ecdsa.PrivateKey) (string, error) {
	return hcrypto.SignData(data, privKey)
}

// GenerateKeyPair delegates to the shared secp256k1 key format.
func GenerateKeyPair() (string, error) {
	return hcrypto.GenerateKeyPair()
}
