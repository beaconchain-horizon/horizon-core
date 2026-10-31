package db

import "time"

// Block represents a block in the chain
type Block struct {
	Number       uint64    `gorm:"primaryKey" json:"number"`
	PreviousHash string    `json:"previous_hash"`
	MerkleRoot   string    `json:"merkle_root"`
	MerkleType   string    `json:"merkle_type"`
	TxCount      uint64    `json:"tx_count"`
	Timestamp    time.Time `json:"timestamp"`
	BlockHash    string    `json:"block_hash"`
	AuthoritySig string    `json:"authority_signature,omitempty"`
}

// Transaction represents a license or a transaction in the chain
type Transaction struct {
	TxID       string    `gorm:"primaryKey" json:"tx_id"`
	BlockNum   uint64    `json:"block_number"`
	LicenseID  string    `json:"license_id"`
	UserID     string    `json:"user_id"`
	ProductID  string    `json:"product_id"`
	HardwareID string    `json:"hardware_id"`
	Volume     int64     `json:"volume"`
	Duration   int       `json:"duration"`
	Signature  string    `json:"signature"`
	PublicKey  string    `json:"public_key"`
	CreatedAt  time.Time `json:"created_at"`
}

// Agent represents an agent in the network
type Agent struct {
	AgentID       string    `gorm:"primaryKey" json:"agent_id"`
	NetworkUser   string    `json:"network_user"`
	APIKey        string    `json:"api_key"`
	LicenseID     string    `json:"license_id"`
	CreditBalance int64     `json:"credit_balance"`
	Status        string    `json:"status"`
	LastSeen      time.Time `json:"last_seen"`
}

// SystemLaw represents immutable laws in the database
type SystemLaw struct {
	LawID       string    `gorm:"primaryKey" json:"law_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
