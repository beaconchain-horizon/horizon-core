package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Number       uint64
	PreviousHash string
	MerkleRoot   string
	MerkleType   string
	TxCount      uint64
	Timestamp    time.Time
	BlockHash    string
}

func NewBlock(prevHash, merkleRoot string, txCount uint64) *Block {
	b := &Block{
		Number:       0, // to be set by caller
		PreviousHash: prevHash,
		MerkleRoot:   merkleRoot,
		MerkleType:   "binary",
		TxCount:      txCount,
		Timestamp:    time.Now().UTC(),
	}
	b.BlockHash = b.computeHash()
	return b
}

func (b *Block) computeHash() string {
	h := sha256.New()
	h.Write([]byte(b.PreviousHash))
	h.Write([]byte(b.MerkleRoot))
	h.Write([]byte(b.Timestamp.Format(time.RFC3339Nano)))
	fmt.Fprintf(h, "|%d", b.TxCount)
	return hex.EncodeToString(h.Sum(nil))
}
