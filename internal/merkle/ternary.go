package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

// Ternary Merkle tree (fanout = 3).
//
// Formula matches v2 binary philosophy — no domain separation:
//   leaf(data)     = SHA256(data)
//   node(a, b, c)  = SHA256(a || b || c)
//
// Incomplete right-edge groups are self-padded, consistent with
// the binary odd-level rule SHA256(last || last):
//   1 leftover -> node(x, x, x)
//   2 leftover -> node(a, b, b)
//
// Tested against hand-computed references in ternary_test.go.

func ternaryNode(a, b, c [32]byte) [32]byte {
	var buf [96]byte
	copy(buf[0:32], a[:])
	copy(buf[32:64], b[:])
	copy(buf[64:96], c[:])
	return sha256.Sum256(buf[:])
}

func TernaryRootNoPrefix(data [][]byte) string {
	if len(data) == 0 {
		return hex.EncodeToString(make([]byte, 32))
	}
	level := make([][32]byte, len(data))
	for i, d := range data {
		level[i] = sha256.Sum256(d)
	}
	for len(level) > 1 {
		next := make([][32]byte, 0, (len(level)+2)/3)
		for i := 0; i < len(level); i += 3 {
			a := level[i]
			var b, c [32]byte
			if i+1 < len(level) {
				b = level[i+1]
			} else {
				b = a
			}
			if i+2 < len(level) {
				c = level[i+2]
			} else {
				c = b
			}
			next = append(next, ternaryNode(a, b, c))
		}
		level = next
	}
	return hex.EncodeToString(level[0][:])
}
