package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Secp256k1PublicKeyFromHex decodes the uncompressed 65-byte public key (04...)
func Secp256k1PublicKeyFromHex(hexKey string) (*secp256k1.PublicKey, error) {
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
	return secp256k1.ParsePubKey(b)
}

// VerifyData strictly verifies the ECDSA signature using secp256k1 curve
func VerifyData(pub *secp256k1.PublicKey, data []byte, sigHex string) (bool, error) {
	if pub == nil {
		return false, fmt.Errorf("nil pub")
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature hex length")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Construct the secp256k1 signature from R and S
	sig := secp256k1.NewSignature(r, s)

	hash := sha256.Sum256(data)
	return sig.Verify(hash[:], pub), nil
}
