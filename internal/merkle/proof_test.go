package merkle

import (
	"bytes"
	"testing"
)

func TestProof_RoundTrip(t *testing.T) {
	leaves := [][]byte{
		[]byte("alpha"),
		[]byte("bravo"),
		[]byte("charlie"),
		[]byte("delta"),
		[]byte("echo"),
	}

	for i := range leaves {
		p, err := BuildProof(leaves, i)
		if err != nil {
			t.Fatalf("index %d: %v", i, err)
		}
		root := p.RecomputeRoot()
		if !p.Verify(root) {
			t.Fatalf("index %d: self-verify failed", i)
		}
	}
}

func TestProof_RejectsTamperedLeaf(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	p, _ := BuildProof(leaves, 2)
	root := p.RecomputeRoot()

	p.Leaf = []byte("tampered")
	if p.Verify(root) {
		t.Fatal("tampered leaf accepted")
	}
}

func TestProof_RejectsWrongIndex(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	p, _ := BuildProof(leaves, 0)
	root := p.RecomputeRoot()

	p.Index = 3
	if p.Verify(root) {
		t.Fatal("wrong index accepted")
	}
}

func TestProof_RejectsDifferentRoot(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d")}
	p, _ := BuildProof(leaves, 1)

	otherLeaves := [][]byte{[]byte("x"), []byte("y"), []byte("z"), []byte("w")}
	otherProof, _ := BuildProof(otherLeaves, 1)
	otherRoot := otherProof.RecomputeRoot()

	if p.Verify(otherRoot) {
		t.Fatal("proof verified against wrong root")
	}
	if bytes.Equal(p.RecomputeRoot()[:], otherRoot[:]) {
		t.Fatal("roots collide")
	}
}

func TestProof_SingleLeaf(t *testing.T) {
	leaves := [][]byte{[]byte("only")}
	p, err := BuildProof(leaves, 0)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	root := p.RecomputeRoot()
	if !p.Verify(root) {
		t.Fatal("single-leaf self-verify failed")
	}
	if len(p.Siblings) != 0 {
		t.Fatalf("single-leaf proof should have 0 siblings, got %d", len(p.Siblings))
	}
}

func TestProof_RejectsOutOfRangeIndex(t *testing.T) {
	leaves := [][]byte{[]byte("a"), []byte("b")}
	if _, err := BuildProof(leaves, 2); err == nil {
		t.Fatal("out-of-range index accepted")
	}
	if _, err := BuildProof(leaves, -1); err == nil {
		t.Fatal("negative index accepted")
	}
	if _, err := BuildProof(nil, 0); err == nil {
		t.Fatal("empty leaves accepted")
	}
}
