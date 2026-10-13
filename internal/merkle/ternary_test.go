package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func h3(a, b, c []byte) [32]byte {
	buf := append(append(append([]byte{}, a...), b...), c...)
	return sha256.Sum256(buf)
}

func TestTernary_Empty(t *testing.T) {
	got := TernaryRootNoPrefix(nil)
	want := hex.EncodeToString(make([]byte, 32))
	if got != want {
		t.Fatalf("empty got %s", got)
	}
}

func TestTernary_SingleLeaf(t *testing.T) {
	d := []byte("a")
	l := sha256.Sum256(d)
	got := TernaryRootNoPrefix([][]byte{d})
	want := hex.EncodeToString(l[:])
	if got != want {
		t.Fatalf("single got %s want %s", got, want)
	}
}

func TestTernary_TwoLeaves(t *testing.T) {
	a, b := []byte("a"), []byte("b")
	ha, hb := sha256.Sum256(a), sha256.Sum256(b)
	want := hex.EncodeToString(func() []byte { r := h3(ha[:], hb[:], hb[:]); return r[:] }())
	got := TernaryRootNoPrefix([][]byte{a, b})
	if got != want {
		t.Fatalf("two got %s want %s", got, want)
	}
}

func TestTernary_ThreeLeaves(t *testing.T) {
	a, b, c := []byte("a"), []byte("b"), []byte("c")
	ha, hb, hc := sha256.Sum256(a), sha256.Sum256(b), sha256.Sum256(c)
	want := hex.EncodeToString(func() []byte { r := h3(ha[:], hb[:], hc[:]); return r[:] }())
	got := TernaryRootNoPrefix([][]byte{a, b, c})
	if got != want {
		t.Fatalf("three got %s want %s", got, want)
	}
}

func TestTernary_FourLeaves(t *testing.T) {
	a, b, c, d := []byte("a"), []byte("b"), []byte("c"), []byte("d")
	ha, hb, hc, hd := sha256.Sum256(a), sha256.Sum256(b), sha256.Sum256(c), sha256.Sum256(d)
	g0 := h3(ha[:], hb[:], hc[:])
	g1 := h3(hd[:], hd[:], hd[:])
	want := hex.EncodeToString(func() []byte { r := h3(g0[:], g1[:], g1[:]); return r[:] }())
	got := TernaryRootNoPrefix([][]byte{a, b, c, d})
	if got != want {
		t.Fatalf("four got %s want %s", got, want)
	}
}

func TestTernary_Determinism(t *testing.T) {
	data := [][]byte{[]byte("x"), []byte("y"), []byte("z")}
	if TernaryRootNoPrefix(data) != TernaryRootNoPrefix(data) {
		t.Fatal("non-deterministic")
	}
}

func TestTernary_DifferentDataDifferentRoot(t *testing.T) {
	r1 := TernaryRootNoPrefix([][]byte{[]byte("a"), []byte("b"), []byte("c")})
	r2 := TernaryRootNoPrefix([][]byte{[]byte("a"), []byte("b"), []byte("d")})
	if r1 == r2 {
		t.Fatal("collision")
	}
}

func TestTernary_DiffersFromBinary(t *testing.T) {
	data := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if TernaryRootNoPrefix(data) == BinaryRootNoPrefix(data) {
		t.Fatal("ternary must differ from binary")
	}
}
