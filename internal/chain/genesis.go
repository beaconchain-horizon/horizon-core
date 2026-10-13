package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GenesisChainID is the identifier for this chain instance.
// In production it should be unique per deployment.
const GenesisChainID = "horizon-core-v1.4"

// EnsureGenesis creates the genesis block if it does not exist.
// Returns the genesis block.
func EnsureGenesis(db *gorm.DB, authorityPubKeyHex string) (*Block, error) {
	var existing Block
	err := db.Where("number = ?", 0).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("genesis lookup: %w", err)
	}

	ts := time.Now().UTC()

	// Genesis merkle root = hash of chain_id + pubkey + timestamp.
	h := sha256.New()
	h.Write([]byte(GenesisChainID))
	h.Write([]byte("|"))
	h.Write([]byte(authorityPubKeyHex))
	h.Write([]byte("|"))
	h.Write([]byte(ts.Format(time.RFC3339)))
	root := hex.EncodeToString(h.Sum(nil))

	block := Block{
		Number:       0,
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
		MerkleRoot:   root,
		MerkleType:   "genesis",
		Timestamp:    ts,
		TxCount:      0,
		BlockHash:    root, // block 0 hash = its merkle root
		CreatedAt:    ts,
	}

	if err := db.Create(&block).Error; err != nil {
		return nil, fmt.Errorf("create genesis block: %w", err)
	}

	gen := Genesis{
		ID:              1,
		ChainID:         GenesisChainID,
		AuthorityPubKey: authorityPubKeyHex,
		Timestamp:       ts,
	}
	_ = db.Where("id = ?", 1).Delete(&Genesis{}).Error
	if err := db.Create(&gen).Error; err != nil {
		return nil, fmt.Errorf("create genesis record: %w", err)
	}

	return &block, nil
}

// LatestBlock returns the highest-numbered block.
func LatestBlock(db *gorm.DB) (*Block, error) {
	var b Block
	if err := db.Order("number DESC").First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}
