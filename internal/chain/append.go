package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PendingTx is a transaction waiting to be included in a block.
type PendingTx struct {
	TxID      string
	TxType    string
	Payload   string
	Signature string
	PublicKey string
}

// AppendBlock creates a new block containing the given transactions,
// links it to the previous block, and persists everything in one
// database transaction. Returns the new block.
func AppendBlock(db *gorm.DB, txs []PendingTx) (*Block, error) {
	if len(txs) == 0 {
		return nil, fmt.Errorf("append block: no transactions")
	}

	prev, err := LatestBlock(db)
	if err != nil {
		return nil, fmt.Errorf("append block: get previous: %w", err)
	}

	ts := time.Now().UTC()

	// Build merkle leaves (binary tree).
	leaves := make([][]byte, len(txs))
	for i, t := range txs {
		h := sha256.Sum256([]byte(t.Payload))
		leaves[i] = h[:]
	}

	// Two-level merkle: binary over leaf hashes.
	root := binaryRoot(leaves)

	// Block hash.
	bh := sha256.New()
	bh.Write([]byte(prev.BlockHash))
	bh.Write([]byte(root))
	bh.Write([]byte(ts.Format(time.RFC3339Nano)))
	fmt.Fprintf(bh, "|%d", len(txs))
	blockHash := hex.EncodeToString(bh.Sum(nil))

	block := Block{
		Number:       prev.Number + 1,
		PreviousHash: prev.BlockHash,
		MerkleRoot:   hex.EncodeToString(root),
		MerkleType:   "binary",
		Timestamp:    ts,
		TxCount:      len(txs),
		BlockHash:    blockHash,
		CreatedAt:    ts,
	}

	// Persist block + transactions atomically.
	return &block, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&block).Error; err != nil {
			return fmt.Errorf("create block: %w", err)
		}

		for i, t := range txs {
			leafHash := sha256.Sum256([]byte(t.Payload))
			proof := buildProof(leaves, i)
			proofJSON, _ := json.Marshal(proof)

			txRow := Transaction{
				TxID:        t.TxID,
				BlockNumber: block.Number,
				TxIndex:     i,
				TxType:      t.TxType,
				Payload:     t.Payload,
				Signature:   t.Signature,
				PublicKey:   t.PublicKey,
				MerkleLeaf:  hex.EncodeToString(leafHash[:]),
				MerkleProof: string(proofJSON),
				Timestamp:   ts,
				CreatedAt:   ts,
			}
			if err := tx.Create(&txRow).Error; err != nil {
				return fmt.Errorf("create tx %s: %w", t.TxID, err)
			}
		}

		mr := MerkleRoot{
			RootHash:    block.MerkleRoot,
			BlockNumber: block.Number,
			TreeType:    "binary",
			Timestamp:   ts,
		}
		if err := tx.Create(&mr).Error; err != nil {
			return fmt.Errorf("create merkle root: %w", err)
		}

		return nil
	})
}

// binaryRoot computes a merkle root over the given leaves.
// Odd levels are self-paired.
func binaryRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return make([]byte, 32)
	}
	level := make([][]byte, len(leaves))
	copy(level, leaves)

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
	return level[0]
}

// buildProof returns sibling hashes from leaf to root.
func buildProof(leaves [][]byte, index int) []string {
	if len(leaves) == 1 {
		return []string{}
	}
	var proof []string
	level := make([][]byte, len(leaves))
	copy(level, leaves)
	idx := index

	for len(level) > 1 {
		var sibling []byte
		if idx%2 == 0 {
			if idx+1 < len(level) {
				sibling = level[idx+1]
			} else {
				sibling = level[idx]
			}
		} else {
			sibling = level[idx-1]
		}
		proof = append(proof, hex.EncodeToString(sibling))

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
		idx = idx / 2
	}
	return proof
}
