package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// RFC 6962 Merkle Tree
// leaf   = SHA256(0x00 || data)
// parent = SHA256(0x01 || left || right)

const (
	leafPrefix = byte(0x00)
	nodePrefix = byte(0x01)
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

func LeafHash(data []byte) string {
	h := sha256.New()
	h.Write([]byte{leafPrefix})
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func NodeHash(leftHex, rightHex string) (string, error) {
	left, err := hex.DecodeString(leftHex)
	if err != nil {
		return "", fmt.Errorf("invalid left hash: %w", err)
	}
	right, err := hex.DecodeString(rightHex)
	if err != nil {
		return "", fmt.Errorf("invalid right hash: %w", err)
	}
	if len(left) != sha256.Size || len(right) != sha256.Size {
		return "", fmt.Errorf("hashes must be 32 bytes")
	}
	h := sha256.New()
	h.Write([]byte{nodePrefix})
	h.Write(left)
	h.Write(right)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func NewTree(data [][]byte) *Tree {
	tree := &Tree{}
	if len(data) == 0 {
		return tree
	}
	tree.Leaves = make([]*Node, len(data))
	for i, d := range data {
		tree.Leaves[i] = &Node{Hash: LeafHash(d)}
	}
	tree.Root = buildRFC6962(tree.Leaves)
	return tree
}

func buildRFC6962(nodes []*Node) *Node {
	if len(nodes) == 0 {
		return nil
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	k := largestPow2LessThan(len(nodes))
	left := buildRFC6962(nodes[:k])
	right := buildRFC6962(nodes[k:])
	hash, err := NodeHash(left.Hash, right.Hash)
	if err != nil {
		panic(err)
	}
	return &Node{Hash: hash, Left: left, Right: right}
}

func largestPow2LessThan(n int) int {
	if n <= 1 {
		return 0
	}
	k := 1
	for (k << 1) < n {
		k <<= 1
	}
	return k
}

func (t *Tree) GetRootHash() string {
	if t == nil || t.Root == nil {
		sum := sha256.Sum256(nil)
		return hex.EncodeToString(sum[:])
	}
	return t.Root.Hash
}

func (t *Tree) GenerateProof(index int) []ProofStep {
	if t == nil || index < 0 || index >= len(t.Leaves) {
		return nil
	}
	var proof []ProofStep
	var walk func(nodes []*Node, idx int)
	walk = func(nodes []*Node, idx int) {
		if len(nodes) <= 1 {
			return
		}
		k := largestPow2LessThan(len(nodes))
		if idx < k {
			right := buildRFC6962(nodes[k:])
			proof = append(proof, ProofStep{Hash: right.Hash, IsLeft: false})
			walk(nodes[:k], idx)
			return
		}
		left := buildRFC6962(nodes[:k])
		proof = append(proof, ProofStep{Hash: left.Hash, IsLeft: true})
		walk(nodes[k:], idx-k)
	}
	walk(t.Leaves, index)
	return proof
}

func VerifyProof(leafHash string, proof []ProofStep, rootHash string) bool {
	current := leafHash
	for _, step := range proof {
		var next string
		var err error
		if step.IsLeft {
			next, err = NodeHash(step.Hash, current)
		} else {
			next, err = NodeHash(current, step.Hash)
		}
		if err != nil {
			return false
		}
		current = next
	}
	return current == rootHash
}

func VerifyProofData(data []byte, proof []ProofStep, rootHash string) bool {
	return VerifyProof(LeafHash(data), proof, rootHash)
}

func RootFromData(data [][]byte) string {
	return NewTree(data).GetRootHash()
}

func RootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		return NewTree(nil).GetRootHash()
	}
	data := make([][]byte, len(hashes))
	for i, v := range hashes {
		decoded, err := hex.DecodeString(v)
		if err != nil {
			return ""
		}
		data[i] = decoded
	}
	return RootFromData(data)
}

// LegacyRootFromHashes — Horizon v1 Merkle (blocks #0-#3)
func LegacyRootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		h := sha256.Sum256([]byte("empty"))
		return hex.EncodeToString(h[:])
	}
	nodes := make([]string, len(hashes))
	copy(nodes, hashes)
	for len(nodes) > 1 {
		next := make([]string, 0, (len(nodes)+1)/2)
		for i := 0; i < len(nodes); i += 2 {
			if i+1 >= len(nodes) {
				next = append(next, nodes[i])
				continue
			}
			combined := append([]byte(nodes[i]), []byte(nodes[i+1])...)
			hash := sha256.Sum256(combined)
			next = append(next, hex.EncodeToString(hash[:]))
		}
		nodes = next
	}
	return nodes[0]
}
