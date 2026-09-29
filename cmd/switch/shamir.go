package main

import (
	"crypto/rand"
	"fmt"
)

// shamirSplit performs a simple XOR-based secret sharing
// For production, use github.com/hashicorp/vault/shamir
func shamirSplit(secret []byte, total, threshold int) ([][]byte, error) {
	if threshold > total || threshold < 2 {
		return nil, fmt.Errorf("invalid threshold %d/%d", threshold, total)
	}

	shares := make([][]byte, total)
	for i := range shares {
		shares[i] = make([]byte, len(secret))
		if _, err := rand.Read(shares[i]); err != nil {
			return nil, err
		}
	}

	// Last share = XOR of secret with all previous shares
	shares[total-1] = make([]byte, len(secret))
	copy(shares[total-1], secret)
	for i := 0; i < total-1; i++ {
		for j := 0; j < len(secret); j++ {
			shares[total-1][j] ^= shares[i][j]
		}
	}

	return shares, nil
}

// shamirCombine reconstructs a secret from shares
func shamirCombine(shares [][]byte) ([]byte, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("no shares")
	}
	secret := make([]byte, len(shares[0]))
	for _, share := range shares {
		if len(share) != len(secret) {
			return nil, fmt.Errorf("share length mismatch")
		}
		for i, b := range share {
			secret[i] ^= b
		}
	}
	return secret, nil
}
