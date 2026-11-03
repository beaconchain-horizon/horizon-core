package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// HexToPrivateKey parses a 64-char hex string (optionally 0x-prefixed) into an ecdsa.PrivateKey.
func HexToPrivateKey(s string) (*ecdsa.PrivateKey, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "0x"), "0X")
	if len(s) != 64 {
		return nil, fmt.Errorf("hex private key must be 64 chars, got %d", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.D = new(big.Int).SetBytes(b)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(b)
	return priv, nil
}

// HexToPublicKey parses a 65-byte uncompressed hex pubkey (04-prefixed) into an ecdsa.PublicKey.
func HexToPublicKey(s string) (*ecdsa.PublicKey, error) {
	s = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "0x"), "0X")
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != 65 || b[0] != 0x04 {
		return nil, fmt.Errorf("pubkey must be 65-byte uncompressed (04 prefix), got len=%d", len(b))
	}
	curve := secp256k1.S256()
	x := new(big.Int).SetBytes(b[1:33])
	y := new(big.Int).SetBytes(b[33:65])
	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("point not on secp256k1 curve")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
