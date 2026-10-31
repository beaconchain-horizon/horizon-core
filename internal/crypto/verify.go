package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Secp256k1PublicKeyFromHex decodes the public key from the offline dashboard
func Secp256k1PublicKeyFromHex(hexKey string) (*ecdsa.PublicKey, error) {
	if len(hexKey) == 0 {
		return nil, fmt.Errorf("empty key")
	}
	if len(hexKey) > 2 && hexKey[0:2] == "0x" {
		hexKey = hexKey[2:]
	}
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	if len(b) != 65 || b[0] != 0x04 {
		return nil, fmt.Errorf("invalid key format")
	}
	x := new(big.Int).SetBytes(b[1:33])
	y := new(big.Int).SetBytes(b[33:65])
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// VerifyData strictly verifies the ECDSA signature
func VerifyData(pub *ecdsa.PublicKey, data []byte, sigHex string) (bool, error) {
	if pub == nil {
		return false, fmt.Errorf("nil pub")
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != 64 {
		return false, fmt.Errorf("invalid sig")
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	hash := sha256.Sum256(data)
	return ecdsa.Verify(pub, hash[:], r, s), nil
}
