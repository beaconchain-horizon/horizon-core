package main

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ═══════════════════════════════════════════════════════════════════
// Horizon Unified Chain — License Registry on Blockchain
//
// Every action (user, license, branch, tx batch) is a transaction
// in a single chain, signed by the owner private key.
// Switch verifies all against this chain.
// ═══════════════════════════════════════════════════════════════════

// ─── Tx Types ───
const (
	TxUserRegister   = "user.register"
	TxLicenseIssue   = "license.issue"
	TxLicenseRenew   = "license.renew"
	TxLicenseRevoke  = "license.revoke"
	TxBranchRegister = "branch.register"
	TxBatch          = "tx.batch"
	TxKeyRotate      = "key.rotate"
	TxAudit          = "audit.event"
)

// ─── Chain Block ───
type ChainBlock struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BlockNum   int64     `gorm:"uniqueIndex;not null" json:"block_num"`
	PrevHash   string    `gorm:"not null" json:"prev_hash"`
	MerkleRoot string    `json:"merkle_root"`
	TxCount    int       `json:"tx_count"`
	Timestamp  int64     `json:"timestamp"`
	Hash       string    `gorm:"not null;index" json:"hash"`
	Signature  string    `json:"signature"`
	SignerAddr string    `json:"signer_addr"`
	CreatedAt  time.Time `json:"created_at"`
}

// ─── Tx in a block ───
type ChainTx struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TxID      string    `gorm:"uniqueIndex;not null" json:"tx_id"`
	BlockNum  int64     `gorm:"index;not null" json:"block_num"`
	TxType    string    `gorm:"index;not null" json:"tx_type"`
	Payload   string    `gorm:"type:text;not null" json:"payload"` // JSON
	TxHash    string    `gorm:"not null;index" json:"tx_hash"`
	Signature string    `json:"signature"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Chain state ───
var (
	chainMutex    sync.Mutex
	chainLoaded   bool
	chainOwnerPub string // hex public key of owner
)

// ─── Canonical tx message for signing ───
func canonicalTxMessage(txType, txID, payload string, ts int64) string {
	return fmt.Sprintf("HZTX1|%s|%s|%s|%d", txType, txID, payload, ts)
}

// ─── Canonical block message ───
func canonicalBlockMessage(blockNum int64, prevHash, merkleRoot string, ts int64) string {
	return fmt.Sprintf("HZBLK1|%d|%s|%s|%d", blockNum, prevHash, merkleRoot, ts)
}

// ─── Migrate chain tables ───
func migrateChain(db *gorm.DB) error {
	if err := db.AutoMigrate(&ChainBlock{}, &ChainTx{}); err != nil {
		return err
	}
	return nil
}

// ─── Compute tx hash ───
func computeTxHash(txType, payload string, ts int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("HZTXH1|%s|%s|%d", txType, payload, ts)))
	return hex.EncodeToString(h[:])
}

// ─── Compute merkle root of txs ───
func computeBlockMerkleRoot(txHashes []string) string {
	if len(txHashes) == 0 {
		return ""
	}
	// simple two-level: hash of concatenated hashes (deterministic)
	h := sha256.New()
	for _, th := range txHashes {
		h.Write([]byte(th))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ─── Compute block hash ───
func computeBlockHash(blockNum int64, prevHash, merkleRoot string, ts int64) string {
	h := sha256.Sum256([]byte(canonicalBlockMessage(blockNum, prevHash, merkleRoot, ts)))
	return hex.EncodeToString(h[:])
}

// ═══════════════════════════════════════════════════════════════════
// GENESIS
// ═══════════════════════════════════════════════════════════════════

func createGenesisBlock(db *gorm.DB, ownerKey *ecdsa.PrivateKey) (*ChainBlock, error) {
	ts := time.Now().Unix()
	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	merkle := ""
	hash := computeBlockHash(0, prevHash, merkle, ts)
	sig, err := signData(ownerKey, []byte(canonicalBlockMessage(0, prevHash, merkle, ts)))
	if err != nil {
		return nil, err
	}
	addr := publicAddressFromKey(ownerKey)
	block := &ChainBlock{
		BlockNum:   0,
		PrevHash:   prevHash,
		MerkleRoot: merkle,
		TxCount:    0,
		Timestamp:  ts,
		Hash:       hash,
		Signature:  sig,
		SignerAddr: addr,
	}
	if err := db.Create(block).Error; err != nil {
		return nil, err
	}
	log.Printf("⛓  Genesis block created (signer=%s)", addr[:10])
	return block, nil
}

// ═══════════════════════════════════════════════════════════════════
// ADD BLOCK (with txs)
// ═══════════════════════════════════════════════════════════════════

func addChainBlock(db *gorm.DB, ownerKey *ecdsa.PrivateKey, txs []ChainTx) (*ChainBlock, error) {
	chainMutex.Lock()
	defer chainMutex.Unlock()

	// get last block
	var last ChainBlock
	if err := db.Order("block_num DESC").First(&last).Error; err != nil {
		return nil, fmt.Errorf("no chain: %w", err)
	}

	// compute tx hashes + sign each tx
	txHashes := make([]string, 0, len(txs))
	for i := range txs {
		txHashes = append(txHashes, txs[i].TxHash)
	}
	merkle := computeBlockMerkleRoot(txHashes)

	blockNum := last.BlockNum + 1
	ts := time.Now().Unix()
	blockHash := computeBlockHash(blockNum, last.Hash, merkle, ts)
	sig, err := signData(ownerKey, []byte(canonicalBlockMessage(blockNum, last.Hash, merkle, ts)))
	if err != nil {
		return nil, err
	}
	addr := publicAddressFromKey(ownerKey)

	block := &ChainBlock{
		BlockNum:   blockNum,
		PrevHash:   last.Hash,
		MerkleRoot: merkle,
		TxCount:    len(txs),
		Timestamp:  ts,
		Hash:       blockHash,
		Signature:  sig,
		SignerAddr: addr,
	}
	if err := db.Create(block).Error; err != nil {
		return nil, err
	}

	// attach block_num and insert txs
	for i := range txs {
		txs[i].BlockNum = blockNum
		if err := db.Create(&txs[i]).Error; err != nil {
			return nil, err
		}
	}

	log.Printf("⛓  Block #%d added (%d txs, merkle=%s)", blockNum, len(txs), merkle[:12])
	return block, nil
}

// ═══════════════════════════════════════════════════════════════════
// APPEND A SINGLE TX (creates a block with 1 tx)
// ═══════════════════════════════════════════════════════════════════

func appendChainTx(db *gorm.DB, ownerKey *ecdsa.PrivateKey, txType string, payloadObj interface{}) (*ChainTx, *ChainBlock, error) {
	payloadBytes, err := json.Marshal(payloadObj)
	if err != nil {
		return nil, nil, err
	}
	payload := string(payloadBytes)
	ts := time.Now().Unix()
	txID := fmt.Sprintf("tx_%d_%s", ts, randomHex(6))
	txHash := computeTxHash(txType, payload, ts)
	sig, err := signData(ownerKey, []byte(canonicalTxMessage(txType, txID, payload, ts)))
	if err != nil {
		return nil, nil, err
	}

	tx := ChainTx{
		TxID:      txID,
		TxType:    txType,
		Payload:   payload,
		TxHash:    txHash,
		Signature: sig,
		Timestamp: ts,
	}

	block, err := addChainBlock(db, ownerKey, []ChainTx{tx})
	if err != nil {
		return nil, nil, err
	}
	tx.BlockNum = block.BlockNum
	return &tx, block, nil
}

// ═══════════════════════════════════════════════════════════════════
// VERIFY CHAIN
// ═══════════════════════════════════════════════════════════════════

func verifyChain(db *gorm.DB, ownerPub *ecdsa.PublicKey) (bool, error) {
	var blocks []ChainBlock
	if err := db.Order("block_num ASC").Find(&blocks).Error; err != nil {
		return false, err
	}
	if len(blocks) == 0 {
		return false, fmt.Errorf("empty chain")
	}

	var prev ChainBlock
	for i, b := range blocks {
		// check block num
		if b.BlockNum != int64(i) {
			return false, fmt.Errorf("block %d: wrong num %d", i, b.BlockNum)
		}
		// check prev hash
		if i > 0 && b.PrevHash != prev.Hash {
			return false, fmt.Errorf("block %d: prev hash mismatch", i)
		}
		// verify signature
		msg := canonicalBlockMessage(b.BlockNum, b.PrevHash, b.MerkleRoot, b.Timestamp)
		ok, err := verifySignature(ownerPub, []byte(msg), b.Signature)
		if err != nil || !ok {
			return false, fmt.Errorf("block %d: signature invalid", i)
		}
		prev = b
	}
	return true, nil
}

// ═══════════════════════════════════════════════════════════════════
// FIND LICENSE STATE FROM CHAIN
// ═══════════════════════════════════════════════════════════════════

type LicenseState struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
	Status     string `json:"status"` // active | revoked | expired
	IssuedIn   int64  `json:"issued_in_block"`
	UpdatedIn  int64  `json:"updated_in_block"`
}

func findLicenseState(db *gorm.DB, licenseID string) (*LicenseState, error) {
	var txs []ChainTx
	if err := db.
		Where("tx_type IN ?", []string{TxLicenseIssue, TxLicenseRenew, TxLicenseRevoke}).
		Order("block_num ASC").
		Find(&txs).Error; err != nil {
		return nil, err
	}
	var st *LicenseState
	for _, tx := range txs {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(tx.Payload), &payload); err != nil {
			continue
		}
		lid, _ := payload["license_id"].(string)
		if lid != licenseID {
			continue
		}
		switch tx.TxType {
		case TxLicenseIssue, TxLicenseRenew:
			st = &LicenseState{
				LicenseID:  lid,
				UserID:     strVal(payload, "user_id"),
				ProductID:  strVal(payload, "product_id"),
				Volume:     intVal(payload, "volume"),
				Duration:   intVal(payload, "duration"),
				IssuedAt:   int64Val(payload, "issued_at"),
				ExpiresAt:  int64Val(payload, "expires_at"),
				HardwareID: strVal(payload, "hardware_id"),
				Status:     "active",
				IssuedIn:   tx.BlockNum,
				UpdatedIn:  tx.BlockNum,
			}
		case TxLicenseRevoke:
			if st != nil {
				st.Status = "revoked"
				st.UpdatedIn = tx.BlockNum
			}
		}
	}
	if st == nil {
		return nil, fmt.Errorf("license not found in chain")
	}
	// check expiry
	if st.Status == "active" && time.Now().Unix() > st.ExpiresAt {
		st.Status = "expired"
	}
	return st, nil
}

// ─── helpers ───
func strVal(m map[string]interface{}, k string) string {
	v, _ := m[k].(string)
	return v
}
func intVal(m map[string]interface{}, k string) int {
	v, _ := m[k].(float64)
	return int(v)
}
func int64Val(m map[string]interface{}, k string) int64 {
	v, _ := m[k].(float64)
	return int64(v)
}

// ═══════════════════════════════════════════════════════════════════
// INIT
// ═══════════════════════════════════════════════════════════════════

func initLicenseChain(db *gorm.DB) error {
	if err := migrateChain(db); err != nil {
		return err
	}
	var count int64
	db.Model(&ChainBlock{}).Count(&count)
	if count == 0 {
		// create genesis from currently unlocked key (or skip if locked)
		stateMutex.RLock()
		k := unlockedKey
		stateMutex.RUnlock()
		if k == nil {
			log.Println("⛓  chain: no unlocked key, genesis deferred until unlock")
			return nil
		}
		_, err := createGenesisBlock(db, k)
		return err
	}
	chainLoaded = true
	return nil
}

// EnsureGenesis — called from key unlock
func ensureChainGenesis(db *gorm.DB, ownerKey *ecdsa.PrivateKey) error {
	var count int64
	db.Model(&ChainBlock{}).Count(&count)
	if count == 0 {
		_, gerr := createGenesisBlock(db, ownerKey)
		return gerr
	}
	return nil
}

// ─── helper: random hex string ───
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		return fmt.Sprintf("%016x", time.Now().UnixNano())[:n*2]
	}
	return hex.EncodeToString(b)
}
