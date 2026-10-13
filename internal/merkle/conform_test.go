package merkle_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"horizon-core/internal/merkle"
)

// combineLikeChain replicates the pairing loop from internal/chain.binaryRoot
// (which expects PRE-HASHED leaves and does not hash its inputs).
func combineLikeChain(hashes [][]byte) string {
	if len(hashes) == 0 {
		return hex.EncodeToString(make([]byte, 32))
	}
	level := make([][]byte, len(hashes))
	copy(level, hashes)
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			var pair []byte
			if i+1 < len(level) {
				pair = append(level[i], level[i+1]...)
			} else {
				pair = append(level[i], level[i]...)
			}
			h := sha256.Sum256(pair)
			next = append(next, h[:])
		}
		level = next
	}
	return hex.EncodeToString(level[0])
}

// chainRootForPayloads reproduces the FULL chain formula for a list of
// payloads: hash each payload (like append.go does), then combine.
func chainRootForPayloads(payloads [][]byte) string {
	if len(payloads) == 0 {
		return combineLikeChain(nil)
	}
	hashes := make([][]byte, len(payloads))
	for i, p := range payloads {
		h := sha256.Sum256(p)
		hashes[i] = h[:]
	}
	return combineLikeChain(hashes)
}

func TestBinaryRootNoPrefix_MatchesChainFormula(t *testing.T) {
	cases := [][][]byte{
		{},
		{[]byte("a")},
		{[]byte("a"), []byte("b")},
		{[]byte("a"), []byte("b"), []byte("c")},
		{[]byte("a"), []byte("b"), []byte("c"), []byte("d")},
		make([][]byte, 5),
		make([][]byte, 7),
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

	for i, payloads := range cases {
		got := merkle.BinaryRootNoPrefix(payloads)
		want := chainRootForPayloads(payloads)
		if got != want {
			t.Fatalf("case %d (len=%d): mismatch\n got:  %s\n want: %s",
				i, len(payloads), got, want)
		}
	}
}
