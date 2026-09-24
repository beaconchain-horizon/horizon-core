package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

type Node struct {
	Hash  string
	Left  *Node
	Right *Node
}

type Tree struct {
	Root   *Node
	Leaves []*Node
}

type ProofStep struct {
	Hash   string
	IsLeft bool
}

func NewTree(data [][]byte) *Tree {
	if len(data) == 0 {
		return &Tree{}
	}
	leaves := make([]*Node, len(data))
	for i, d := range data {
		hash := sha256.Sum256(d)
		leaves[i] = &Node{Hash: hex.EncodeToString(hash[:])}
	}
	tree := &Tree{Leaves: leaves}
	tree.Root = tree.buildTree(leaves)
	return tree
}

func (t *Tree) buildTree(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	var nextLevel []*Node
	for i := 0; i < len(nodes); i += 2 {
		if i+1 < len(nodes) {
			combined := append([]byte(nodes[i].Hash), []byte(nodes[i+1].Hash)...)
			hash := sha256.Sum256(combined)
			parent := &Node{
				Hash:  hex.EncodeToString(hash[:]),
				Left:  nodes[i],
				Right: nodes[i+1],
			}
			nextLevel = append(nextLevel, parent)
		} else {
			nextLevel = append(nextLevel, nodes[i])
		}
	}
	return t.buildTree(nextLevel)
}

func (t *Tree) GetRootHash() string {
	if t.Root == nil {
		return ""
	}
	return t.Root.Hash
}

func (t *Tree) GenerateProof(index int) []ProofStep {
	if index < 0 || index >= len(t.Leaves) {
		return nil
	}
	var proof []ProofStep
	currentLevel := t.Leaves
	currentIndex := index
	for len(currentLevel) > 1 {
		var nextLevel []*Node
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				combined := append([]byte(currentLevel[i].Hash), []byte(currentLevel[i+1].Hash)...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, &Node{
					Hash:  hex.EncodeToString(hash[:]),
					Left:  currentLevel[i],
					Right: currentLevel[i+1],
				})
			} else {
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}
		if currentIndex%2 == 0 {
			if currentIndex+1 < len(currentLevel) {
				proof = append(proof, ProofStep{
					Hash:   currentLevel[currentIndex+1].Hash,
					IsLeft: false,
				})
			}
		} else {
			proof = append(proof, ProofStep{
				Hash:   currentLevel[currentIndex-1].Hash,
				IsLeft: true,
			})
		}
		currentIndex = currentIndex / 2
		currentLevel = nextLevel
	}
	return proof
}

func VerifyProof(leafHash string, proof []ProofStep, rootHash string) bool {
	currentHash := leafHash
	for _, step := range proof {
		var combined []byte
		if step.IsLeft {
			combined = append([]byte(step.Hash), []byte(currentHash)...)
		} else {
			combined = append([]byte(currentHash), []byte(step.Hash)...)
		}
		hash := sha256.Sum256(combined)
		currentHash = hex.EncodeToString(hash[:])
	}
	return currentHash == rootHash
}

// RootFromHashes — helper that takes a list of hex hashes
// and returns the merkle root. Used by switch.
func RootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	data := make([][]byte, len(hashes))
	for i, h := range hashes {
		data[i] = []byte(h)
	}
	t := NewTree(data)
	return t.GetRootHash()
}
