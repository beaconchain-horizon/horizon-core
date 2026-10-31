package merkle

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

func hashLeaf(data []byte) [32]byte {
	buf := make([]byte, 0, 1+len(data))
	buf = append(buf, leafPrefix)
	buf = append(buf, data...)
	return sha256.Sum256(buf)
}

func hashNode(left, right [32]byte) [32]byte {
	buf := make([]byte, 0, 1+64)
	buf = append(buf, nodePrefix)
	buf = append(buf, left[:]...)
	buf = append(buf, right[:]...)
	return sha256.Sum256(buf)
}

// Proof is a Merkle inclusion proof for a single leaf.
type Proof struct {
	Leaf     []byte     // original leaf data
	Index    int        // position in the tree
	Siblings [][32]byte // sibling hashes bottom-up
}

// BuildProof constructs an inclusion proof for the leaf at index.
func BuildProof(leaves [][]byte, index int) (*Proof, error) {
	if len(leaves) == 0 {
		return nil, errors.New("no leaves")
	}
	if index < 0 || index >= len(leaves) {
		return nil, fmt.Errorf("index %d out of range", index)
	}

	hashes := make([][32]byte, len(leaves))
	for i, l := range leaves {
		hashes[i] = hashLeaf(l)
	}

	var siblings [][32]byte
	idx := index
	level := hashes

	for len(level) > 1 {
		var sib [32]byte
		if idx%2 == 0 {
			if idx+1 < len(level) {
				sib = level[idx+1]
			} else {
				sib = level[idx] // odd level: pair with self
			}
		} else {
			sib = level[idx-1]
		}
		siblings = append(siblings, sib)

		// next level
		var next [][32]byte
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				next = append(next, hashNode(level[i], level[i+1]))
			} else {
				next = append(next, hashNode(level[i], level[i]))
			}
		}
		level = next
		idx = idx / 2
	}

	return &Proof{Leaf: leaves[index], Index: index, Siblings: siblings}, nil
}

// Root recomputes the Merkle root from a proof and the claimed root,
// returning the recomputed root for the caller to compare.
func (p *Proof) RecomputeRoot() [32]byte {
	h := hashLeaf(p.Leaf)
	idx := p.Index

	for _, sib := range p.Siblings {
		if idx%2 == 0 {
			h = hashNode(h, sib)
		} else {
			h = hashNode(sib, h)
		}
		idx = idx / 2
	}
	return h
}

// Verify checks the proof against an expected root.
func (p *Proof) Verify(expected [32]byte) bool {
	return p.RecomputeRoot() == expected
}
