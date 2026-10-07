// Horizon BNPL - Merkle Tree (binary + ternary)
// Anchors pending txs into blocks with Merkle root

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type MerkleBlock struct {
	Height     int64    `json:"height"`
	Timestamp  int64    `json:"timestamp"`
	PrevHash   string   `json:"prev_hash"`
	MerkleRoot string   `json:"merkle_root"`
	Fanout     int      `json:"fanout"`
	TxIDs      []string `json:"tx_ids"`
	Hash       string   `json:"hash"`
	Signature  string   `json:"signature"`
}

func leafHash(txID string) string {
	h := sha256.Sum256([]byte("leaf:" + txID))
	return hex.EncodeToString(h[:])
}

func parentHash(children []string) string {
	joined := ""
	for i, c := range children {
		if i > 0 {
			joined += "|"
		}
		joined += c
	}
	h := sha256.Sum256([]byte("node:" + joined))
	return hex.EncodeToString(h[:])
}

// merkleRoot - builds root with given fanout (2 = binary, 3 = ternary)
func merkleRoot(txIDs []string, fanout int) string {
	if len(txIDs) == 0 {
		return parentHash([]string{"empty"})
	}
	if fanout < 2 {
		fanout = 2
	}
	level := make([]string, len(txIDs))
	for i, id := range txIDs {
		level[i] = leafHash(id)
	}

	for len(level) > 1 {
		var next []string
		for i := 0; i < len(level); i += fanout {
			end := i + fanout
			if end > len(level) {
				end = len(level)
			}
			group := level[i:end]
			// RFC 6962 style: do not duplicate last leaf
			next = append(next, parentHash(group))
		}
		level = next
	}
	return level[0]
}

func blockHash(b *MerkleBlock) string {
	raw := fmt.Sprintf("%d|%d|%s|%s|%d",
		b.Height, b.Timestamp, b.PrevHash, b.MerkleRoot, b.Fanout)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// anchorTxs - takes list of tx ids, builds block, signs it
func anchorTxs(txIDs []string, fanout int) (*MerkleBlock, error) {
	if len(txIDs) == 0 {
		return nil, errString("no txs to anchor")
	}

	mu.RLock()
	prevHash := "0"
	height := int64(0)
	if len(blocks) > 0 {
		last := blocks[len(blocks)-1]
		prevHash = last.Hash
		height = last.Height + 1
	}
	mu.RUnlock()

	b := &MerkleBlock{
		Height:     height,
		Timestamp:  time.Now().Unix(),
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot(txIDs, fanout),
		Fanout:     fanout,
		TxIDs:      txIDs,
	}
	b.Hash = blockHash(b)

	sig, err := signData([]byte(b.Hash))
	if err != nil {
		return nil, err
	}
	b.Signature = sig

	mu.Lock()
	blocks = append(blocks, b)
	for _, id := range txIDs {
		if tx, ok := txs[id]; ok {
			tx.BlockHash = b.Hash
		}
	}
	mu.Unlock()

	return b, nil
}

// merkleProof - returns sibling hashes proving a leaf is in the tree
func merkleProof(txIDs []string, target string, fanout int) ([]string, error) {
	if len(txIDs) == 0 {
		return nil, errString("empty tree")
	}
	if fanout < 2 {
		fanout = 2
	}

	idx := -1
	for i, id := range txIDs {
		if id == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, errString("leaf not found")
	}

	level := make([]string, len(txIDs))
	for i, id := range txIDs {
		level[i] = leafHash(id)
	}

	var proof []string
	for len(level) > 1 {
		var next []string
		groupStart := (idx / fanout) * fanout
		groupEnd := groupStart + fanout
		if groupEnd > len(level) {
			groupEnd = len(level)
		}
		for i := groupStart; i < groupEnd; i++ {
			if i != idx {
				proof = append(proof, level[i])
			}
		}
		for i := 0; i < len(level); i += fanout {
			end := i + fanout
			if end > len(level) {
				end = len(level)
			}
			next = append(next, parentHash(level[i:end]))
		}
		idx = groupStart / fanout
		level = next
	}
	return proof, nil
}
