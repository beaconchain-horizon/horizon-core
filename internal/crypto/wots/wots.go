// Package wots implements Winternitz One-Time Signature (WOTS+)
// Quantum-resistant hash-based signature. Only uses SHA-256.
package wots

import (
	"bytes"
	"crypto/sha256"
	"errors"
)

const (
	W           = 16                  // Winternitz parameter
	KeySize     = 32                  // element size in bytes
	Len1        = 64                  // message digits (256 bits / 4)
	Len2        = 3                   // checksum digits
	ChainCount  = Len1 + Len2         // 67
	PrivKeySize = ChainCount * KeySize // 2144 bytes
	PubKeySize  = ChainCount * KeySize // 2144 bytes
	SigSize     = ChainCount * KeySize // 2144 bytes
)

var (
	ErrInvalidKeySize = errors.New("wots: invalid key size")
	ErrInvalidMsgSize = errors.New("wots: message hash must be 32 bytes")
)

func chain(x []byte, steps int) []byte {
	h := make([]byte, 32)
	copy(h, x)
	for i := 0; i < steps; i++ {
		sum := sha256.Sum256(h)
		h = sum[:]
	}
	return h
}

func messageToDigits(msg []byte) []byte {
	digits := make([]byte, Len1)
	for i := 0; i < 32; i++ {
		digits[2*i] = msg[i] >> 4
		digits[2*i+1] = msg[i] & 0x0F
	}
	return digits
}

func checksumDigits(digits []byte) []byte {
	var csum uint32
	for _, d := range digits {
		csum += uint32(W - 1 - int(d))
	}
	csum <<= 4 // 12-bit alignment
	return []byte{
		byte((csum >> 8) & 0x0F),
		byte((csum >> 4) & 0x0F),
		byte(csum & 0x0F),
	}
}

// GenerateKeyPair derives a deterministic key pair from a 32-byte seed.
func GenerateKeyPair(seed []byte) (privKey, pubKey []byte, err error) {
	if len(seed) < 32 {
		return nil, nil, errors.New("wots: seed must be at least 32 bytes")
	}
	privKey = make([]byte, PrivKeySize)
	for i := 0; i < ChainCount; i++ {
		sum := sha256.Sum256(append(seed, byte(i)))
		copy(privKey[i*KeySize:], sum[:])
	}
	pubKey = make([]byte, PubKeySize)
	for i := 0; i < ChainCount; i++ {
		elem := privKey[i*KeySize : (i+1)*KeySize]
		copy(pubKey[i*KeySize:], chain(elem, W-1))
	}
	return privKey, pubKey, nil
}

// Sign produces a WOTS+ signature on a 32-byte message hash.
func Sign(privKey, msgHash []byte) ([]byte, error) {
	if len(privKey) != PrivKeySize {
		return nil, ErrInvalidKeySize
	}
	if len(msgHash) != 32 {
		return nil, ErrInvalidMsgSize
	}
	digits := messageToDigits(msgHash)
	allDigits := append(digits, checksumDigits(digits)...)
	sig := make([]byte, SigSize)
	for i, d := range allDigits {
		elem := privKey[i*KeySize : (i+1)*KeySize]
		copy(sig[i*KeySize:], chain(elem, int(d)))
	}
	return sig, nil
}

// Verify checks a WOTS+ signature.
func Verify(pubKey, msgHash, sig []byte) bool {
	if len(pubKey) != PubKeySize || len(sig) != SigSize || len(msgHash) != 32 {
		return false
	}
	digits := messageToDigits(msgHash)
	allDigits := append(digits, checksumDigits(digits)...)
	for i, d := range allDigits {
		elem := sig[i*KeySize : (i+1)*KeySize]
		computed := chain(elem, W-1-int(d))
		expected := pubKey[i*KeySize : (i+1)*KeySize]
		if !bytes.Equal(computed, expected) {
			return false
		}
	}
	return true
}
