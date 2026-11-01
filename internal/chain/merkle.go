package chain

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeMerkleRoot computes root using binary or ternary hashing.
func ComputeMerkleRoot(leaves [][]byte, treeType string) string {
	if len(leaves) == 0 { return hex.EncodeToString(make([]byte, 32)) }
	level := leaves
	for len(level) > 1 {
		var next [][]byte
		if treeType == "ternary" {
			for i := 0; i < len(level); i += 3 {
				var buf [96]byte
				copy(buf[0:32], level[i])
				if i+1 < len(level) { copy(buf[32:64], level[i+1]) } else { copy(buf[32:64], level[i]) }
				if i+2 < len(level) { copy(buf[64:96], level[i+2]) } else { copy(buf[64:96], level[i]) }
				h := sha256.Sum256(buf[:])
				next = append(next, h[:])
			}
		} else {
			for i := 0; i < len(level); i += 2 {
				var buf [64]byte
				copy(buf[0:32], level[i])
				if i+1 < len(level) { copy(buf[32:64], level[i+1]) } else { copy(buf[32:64], level[i]) }
				h := sha256.Sum256(buf[:])
				next = append(next, h[:])
			}
		}
		level = next
	}
	return hex.EncodeToString(level[0])
}
