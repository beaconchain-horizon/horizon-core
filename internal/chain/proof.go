package chain

import (
	"crypto/sha256"
	"encoding/hex"
)

type ProofStep struct {
	Hash   string
	IsLeft bool
}

type MerkleProof struct {
	Leaf  string
	Steps []ProofStep
}

// BuildProof builds a proof.
// IMPORTANT: leaves are assumed to be already hashed (like in ComputeMerkleRoot).
func BuildProof(leaves [][]byte, index int) (*MerkleProof, error) {
	level := make([][]byte, len(leaves))
	copy(level, leaves)

	proof := &MerkleProof{
		Leaf:  hex.EncodeToString(level[index]),
		Steps: []ProofStep{},
	}

	idx := index
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			var right []byte
			if i+1 < len(level) {
				right = level[i+1]
			} else {
				right = level[i]
			}

			var buf [64]byte
			copy(buf[0:32], left)
			copy(buf[32:64], right)
			h := sha256.Sum256(buf[:])
			next = append(next, h[:])

			if idx == i {
				proof.Steps = append(proof.Steps, ProofStep{Hash: hex.EncodeToString(right), IsLeft: false})
			} else if idx == i+1 {
				proof.Steps = append(proof.Steps, ProofStep{Hash: hex.EncodeToString(left), IsLeft: true})
			}
		}
		level = next
		idx = idx / 2
	}
	return proof, nil
}

func VerifyProof(proof *MerkleProof, expectedRoot string) bool {
	current, _ := hex.DecodeString(proof.Leaf)
	for _, step := range proof.Steps {
		sib, _ := hex.DecodeString(step.Hash)
		var buf [64]byte
		if step.IsLeft {
			copy(buf[0:32], sib)
			copy(buf[32:64], current)
		} else {
			copy(buf[0:32], current)
			copy(buf[32:64], sib)
		}
		h := sha256.Sum256(buf[:])
		current = h[:]
	}
	return hex.EncodeToString(current) == expectedRoot
}
