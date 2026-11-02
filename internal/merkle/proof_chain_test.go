package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestBuildProof_MatchesChainFormula proves that a proof built by
// BuildProof, when walked back, produces the same root as the chain
// formula (BinaryRootNoPrefix). This closes the gap where Proof used
// RFC 6962 prefixes while chain does not.
func TestBuildProof_MatchesChainFormula(t *testing.T) {
	cases := [][][]byte{
		{[]byte("a")},
		{[]byte("a"), []byte("b")},
		{[]byte("a"), []byte("b"), []byte("c")},
		{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
		make([][]byte, 5),
		make([][]byte, 100),
		make([][]byte, 184),
	}
	for i := range cases {
		for j := range cases[i] {
			if len(cases[i][j]) == 0 {
				cases[i][j] = []byte{byte(j), byte(j >> 8), 0xAA}
			}
		}
	}

	for ci, leaves := range cases {
		expectedRoot := BinaryRootNoPrefix(leaves)
		expectedBytes, _ := hex.DecodeString(expectedRoot)
		var want [32]byte
		copy(want[:], expectedBytes)

		for index := 0; index < len(leaves); index++ {
			p, err := BuildProof(leaves, index)
			if err != nil {
				t.Fatalf("case %d index %d: BuildProof: %v", ci, index, err)
			}
			if !p.Verify(want) {
				got := p.RecomputeRoot()
				t.Fatalf("case %d index %d (len=%d): proof does not match chain root\n got:  %x\n want: %x",
					ci, index, len(leaves), got, want)
			}
		}
	}
}

// TestBuildProof_RejectsTamperedLeaf proves a tampered leaf fails.
func TestBuildProof_RejectsTamperedLeaf(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	p, err := BuildProof(leaves, 1)
	if err != nil {
		t.Fatal(err)
	}
	expectedRoot := BinaryRootNoPrefix(leaves)
	var want [32]byte
	b, _ := hex.DecodeString(expectedRoot)
	copy(want[:], b)

	// sanity: original passes
	if !p.Verify(want) {
		t.Fatal("original proof did not verify")
	}

	// tamper
	p.Leaf = []byte("X")
	if p.Verify(want) {
		t.Fatal("tampered leaf verified")
	}
}

// TestBuildProof_Determinism proves BuildProof produces the same proof
// twice.
func TestBuildProof_Determinism(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	p1, _ := BuildProof(leaves, 2)
	p2, _ := BuildProof(leaves, 2)
	if len(p1.Siblings) != len(p2.Siblings) {
		t.Fatalf("sibling count differs: %d vs %d", len(p1.Siblings), len(p2.Siblings))
	}
	for i := range p1.Siblings {
		if p1.Siblings[i] != p2.Siblings[i] {
			t.Fatalf("sibling %d differs", i)
		}
	}
}

var _ = sha256.Sum256
