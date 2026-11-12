package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// QuantumSeal creates a purely hash-based proof for the block timestamp.
// It is immune to Shor's algorithm (Quantum computers) because it doesn't
// rely on prime factorization or discrete logarithms.
func QuantumSeal(genesisHash string, blockNum int64, merkleRoot string) string {
	// Create the base payload
	payload := fmt.Sprintf("%s_%d_%s", genesisHash, blockNum, merkleRoot)
	
	// Apply a Hash-Chain (10,000 iterations) for quantum resistance
	hash := sha256.Sum256([]byte(payload))
	for i := 0; i < 10000; i++ {
		hash = sha256.Sum256(hash[:])
	}
	
	return hex.EncodeToString(hash[:])
}
