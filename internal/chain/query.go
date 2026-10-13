package chain

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ChainInfo summarises the current state of the chain.
type ChainInfo struct {
	ChainID    string `json:"chain_id"`
	Height     int64  `json:"height"`
	LatestHash string `json:"latest_hash"`
	LatestRoot string `json:"latest_root"`
	TxTotal    int64  `json:"tx_total"`
}

// GetChainInfo returns a summary of the chain.
func GetChainInfo(db *gorm.DB) (*ChainInfo, error) {
	var gen Genesis
	if err := db.First(&gen, 1).Error; err != nil {
		return nil, fmt.Errorf("genesis: %w", err)
	}

	latest, err := LatestBlock(db)
	if err != nil {
		return nil, fmt.Errorf("latest: %w", err)
	}

	var txTotal int64
	if err := db.Model(&Transaction{}).Count(&txTotal).Error; err != nil {
		return nil, err
	}

	return &ChainInfo{
		ChainID:    gen.ChainID,
		Height:     latest.Number,
		LatestHash: latest.BlockHash,
		LatestRoot: latest.MerkleRoot,
		TxTotal:    txTotal,
	}, nil
}

// GetBlock returns a block by number.
func GetBlock(db *gorm.DB, number int64) (*Block, error) {
	var b Block
	if err := db.Where("number = ?", number).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBlockTxs returns all transactions in a block, ordered.
func GetBlockTxs(db *gorm.DB, number int64) ([]Transaction, error) {
	var out []Transaction
	if err := db.Where("block_number = ?", number).Order("tx_index ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetTx returns a transaction by tx_id.
func GetTx(db *gorm.DB, txID string) (*Transaction, error) {
	var t Transaction
	if err := db.Where("tx_id = ?", txID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// ListBlocks returns blocks in a range [from, to], newest first if desc.
func ListBlocks(db *gorm.DB, from, to int64, limit int) ([]Block, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var out []Block
	q := db.Model(&Block{})
	if from > 0 {
		q = q.Where("number >= ?", from)
	}
	if to > 0 {
		q = q.Where("number <= ?", to)
	}
	if err := q.Order("number DESC").Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ErrNotFound is a sentinel for missing records.
var ErrNotFound = errors.New("not found")
