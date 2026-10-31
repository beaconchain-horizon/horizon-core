package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

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
		return nil, fmt.Errorf("invalid key format (need 65 bytes with 0x04 prefix)")
	}
	return secp256k1.ParsePubKey(b)
}

func VerifyData(pub *secp256k1.PublicKey, data []byte, sigHex string) (bool, error) {
	if pub == nil {
		return false, fmt.Errorf("nil pub")
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature hex length")
	}

	var rScalar, sScalar secp256k1.ModNScalar
	if overflow := rScalar.SetByteSlice(sigBytes[:32]); overflow {
		return false, fmt.Errorf("r overflow")
	}
	if overflow := sScalar.SetByteSlice(sigBytes[32:]); overflow {
		return false, fmt.Errorf("s overflow")
	}

	sig := secp256k1ecdsa.NewSignature(&rScalar, &sScalar)
	hash := sha256.Sum256(data)
	return sig.Verify(hash[:], pub), nil
}
