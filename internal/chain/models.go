// Package chain defines the block/transaction data model for
// Horizon Core's append-only ledger.
package chain

import "time"

// Block is a single unit of the append-only chain.
type Block struct {
	Number       int64     `json:"number" gorm:"primaryKey;autoIncrement:false"`
	PreviousHash string    `json:"previous_hash" gorm:"index"`
	MerkleRoot   string    `json:"merkle_root"`
	MerkleType   string    `json:"merkle_type"` // "binary" or "ternary"
	Timestamp    time.Time `json:"timestamp" gorm:"index"`
	TxCount      int       `json:"tx_count"`
	BlockHash    string    `json:"block_hash" gorm:"index"`
	AuthoritySig string    `json:"authority_signature"` // optional, empty for genesis-only
	CreatedAt    time.Time `json:"created_at"`
}

// Transaction is a single signed event on the chain.
type Transaction struct {
	TxID        string    `json:"tx_id" gorm:"primaryKey"`
	BlockNumber int64     `json:"block_number" gorm:"index"`
	TxIndex     int       `json:"tx_index"`
	TxType      string    `json:"tx_type" gorm:"index"` // license_issue | license_revoke | operation | sensor_reading
	Payload     string    `json:"payload"`              // canonical bytes that were signed
	Signature   string    `json:"signature"`            // 128 hex
	PublicKey   string    `json:"public_key"`           // 128 hex
	MerkleLeaf  string    `json:"merkle_leaf"`          // 64 hex
	MerkleProof string    `json:"merkle_proof"`         // JSON array of siblings
	Timestamp   time.Time `json:"timestamp" gorm:"index"`
	CreatedAt   time.Time `json:"created_at"`
}

// MerkleRoot records every root ever produced (for audit).
type MerkleRoot struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	RootHash    string    `json:"root_hash" gorm:"index"`
	BlockNumber int64     `json:"block_number" gorm:"index"`
	TreeType    string    `json:"tree_type"`
	Signature   string    `json:"signature"` // Authority signature over root
	Timestamp   time.Time `json:"timestamp"`
}

// Genesis is the very first block metadata (single row).
type Genesis struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	ChainID         string    `json:"chain_id"`
	AuthorityPubKey string    `json:"authority_pubkey"` // 128 hex
	Timestamp       time.Time `json:"timestamp"`
	Signature       string    `json:"signature"`
}
