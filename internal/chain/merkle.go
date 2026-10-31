package chain

import (
	"crypto/sha256"
	"encoding/hex"
)

func BinaryRoot(leaves [][]byte) string {
	if len(leaves) == 0 {
		return ""
	}
	level := leaves
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
