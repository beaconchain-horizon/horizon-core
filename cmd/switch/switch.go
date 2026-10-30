package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"horizon-core/internal/merkle"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/net/netutil"
	"gorm.io/gorm"
	"horizon-core/internal/crypto"
)

// ============================================================
// MODELS - SQLite Tables
// ============================================================

type Block struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BlockNum   int       `gorm:"uniqueIndex;column:block_num" json:"block_num"`
	Timestamp  int64     `json:"timestamp"`
	PrevHash   string    `json:"prev_hash"`
	Hash       string    `gorm:"index" json:"hash"`
	MerkleRoot string    `json:"merkle_root"`
	MerkleAlgo string    `json:"merkle_algo"`
	TxCount    int       `json:"tx_count"`
	CreatedAt  time.Time `json:"created_at"`
	Signature  string    `json:"signature,omitempty"`
	SignerAddr string    `json:"signer_addr,omitempty"`
}

type Transaction struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TxID       string    `gorm:"uniqueIndex" json:"tx_id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Amount     float64   `json:"amount"`
	Type       string    `json:"type"`
	Timestamp  int64     `json:"timestamp"`
	BlockIndex int       `gorm:"index" json:"block_index"`
	Synced     bool      `gorm:"index" json:"synced"`
	CreatedAt  time.Time `json:"created_at"`
}

type License struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LicenseID  string    `gorm:"uniqueIndex" json:"license_id"`
	UserID     string    `gorm:"index" json:"user_id"`
	ProductID  string    `json:"product_id"`
	Volume     int       `json:"volume"`
	Used       int       `json:"used"`
	Duration   int       `json:"duration"`
	Signature  string    `json:"signature"`
	MerkleRoot string    `json:"merkle_root"`
	IssuedAt   int64     `json:"issued_at"`
	ExpiresAt  int64     `json:"expires_at"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	HardwareID string    `json:"hardware_id"`
}

type KeyVault struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EncryptedKey string    `json:"encrypted_key"`
	IV           string    `json:"iv"`
	PublicAddr   string    `json:"public_address"`
	Label        string    `json:"label"`
	CreatedAt    time.Time `json:"created_at"`
}

type BankAccount struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	BankID       string    `gorm:"uniqueIndex" json:"bank_id"`
	BankName     string    `json:"bank_name"`
	PasswordHash string    `json:"password_hash"`
	LicenseID    string    `gorm:"index" json:"license_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// ============================================================
// GLOBAL STATE
// ============================================================

var (
	db           *gorm.DB
	unlockedKey  *ecdsa.PrivateKey
	unlockedAt   time.Time
	unlockedAddr string
	stateMutex   sync.RWMutex
	ledger       *Ledger
	ledgerStop   chan struct{}
)

// ============================================================
// CRYPTO FUNCTIONS
// ============================================================

func deriveKeyFromPassword(password string) []byte {
	h := sha256.Sum256([]byte(password))
	return h[:]
}

func encryptPrivateKey(privateKeyHex, password string) (string, string, error) {
	key := deriveKeyFromPassword(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(privateKeyHex), nil)
	return hex.EncodeToString(ciphertext), hex.EncodeToString(nonce), nil
}

func decryptPrivateKey(encryptedHex, ivHex, password string) (string, error) {
	key := deriveKeyFromPassword(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}
	nonce, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func hexToECDSA(hexKey string) (*ecdsa.PrivateKey, error) {
	return crypto.Secp256k1PrivateKeyFromHex(hexKey)
}

func publicAddressFromKey(priv *ecdsa.PrivateKey) string {
	return crypto.PublicAddressFromKey(priv)
}

func signData(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)

	sig, err := crypto.SignHashRS(priv, hash[:])
	if err != nil {
		return "", err
	}

	addAudit("SIGN", hex.EncodeToString(hash[:8]), "ecdsa_secp256k1", "")

	return hex.EncodeToString(sig), nil
}

func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func simpleMerkleRoot(txIDs []string, algo string) string {
	if len(txIDs) == 0 {
		return hashData([]byte("empty"))
	}
	if algo == "v2" {
		data := make([][]byte, len(txIDs))
		for i, id := range txIDs {
			data[i] = []byte(id)
		}
		return merkle.RootFromData(data)
	}
	hashes := make([]string, len(txIDs))
	for i, id := range txIDs {
		hashes[i] = hashData([]byte(id))
	}
	return merkle.LegacyRootFromHashes(hashes)
}

// ============================================================
// BLOCKCHAIN FUNCTIONS
// ============================================================

func getLastBlock() (*Block, error) {
	var b Block
	err := db.Order("block_num desc").First(&b).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			genesis := &Block{
				BlockNum:   0,
				Timestamp:  time.Now().Unix(),
				PrevHash:   "genesis",
				Hash:       hashData([]byte("genesis-block")),
				MerkleRoot: hashData([]byte("genesis")),
				MerkleAlgo: "v1",
				TxCount:    0,
			}
			db.Create(genesis)
			return genesis, nil
		}
		return nil, err
	}
	return &b, nil
}

func mineBlock() (*Block, error) {
	last, err := getLastBlock()
	if err != nil {
		return nil, err
	}

	var pending []Transaction
	db.Where("block_index = ?", -1).Find(&pending)

	if len(pending) == 0 {
		return nil, fmt.Errorf("no pending transactions")
	}

	txIDs := []string{}
	for _, tx := range pending {
		txIDs = append(txIDs, tx.TxID)
	}
	merkleAlgo := "v1"
	if last.BlockNum+1 >= 4 {
		merkleAlgo = "v2"
	}
	merkleRoot := simpleMerkleRoot(txIDs, merkleAlgo)

	newIndex := last.BlockNum + 1
	timestamp := time.Now().Unix()
	dataStr := fmt.Sprintf("%d|%d|%s|%s", newIndex, timestamp, last.Hash, merkleRoot)
	blockHash := hashData([]byte(dataStr))

	block := &Block{
		BlockNum:   newIndex,
		Timestamp:  timestamp,
		PrevHash:   last.Hash,
		Hash:       blockHash,
		MerkleRoot: merkleRoot,
		MerkleAlgo: merkleAlgo,
		TxCount:    len(pending),
	}

	// P1.6: sign block hash with unlocked ECDSA key (if available)
	stateMutex.RLock()
	signingKey := unlockedKey
	signerAddr := unlockedAddr
	stateMutex.RUnlock()
	if signingKey != nil {
		if sig, err := signData(signingKey, []byte(block.Hash)); err == nil {
			block.Signature = sig
			block.SignerAddr = signerAddr
		}
	}

	if err := db.Create(block).Error; err != nil {
		return nil, err
	}

	for _, tx := range pending {
		db.Model(&tx).Update("block_index", newIndex)
	}

	log.Printf("Block #%d mined: %d transactions", newIndex, len(pending))
	return block, nil
}

// ============================================================
// HTTP HANDLERS
// ============================================================

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"service": "horizon-switch",
		"version": "1.4",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// ============ KEY VAULT ============

func setupKeyHandler(c *gin.Context) {
	var req struct {
		PrivateKey string `json:"private_key" binding:"required"`
		Password   string `json:"password" binding:"required"`
		Label      string `json:"label"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "private_key and password required"})
		return
	}

	priv, err := hexToECDSA(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid private key: " + err.Error()})
		return
	}

	encrypted, iv, err := encryptPrivateKey(req.PrivateKey, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "encryption failed"})
		return
	}

	addr := publicAddressFromKey(priv)

	vault := &KeyVault{
		EncryptedKey: encrypted,
		IV:           iv,
		PublicAddr:   addr,
		Label:        req.Label,
	}

	db.Where("1 = 1").Delete(&KeyVault{})
	if err := db.Create(vault).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "saved",
		"public_addr": addr,
		"message":     "key encrypted and saved",
	})
}

func unlockKeyHandler(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}

	var vault KeyVault
	if err := db.First(&vault).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no key found, please setup first"})
		return
	}

	privateKeyHex, err := decryptPrivateKey(vault.EncryptedKey, vault.IV, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	priv, err := hexToECDSA(privateKeyHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "key parse failed"})
		return
	}

	stateMutex.Lock()
	unlockedKey = priv
	unlockedAt = time.Now()
	unlockedAddr = vault.PublicAddr
	stateMutex.Unlock()

	// init blockchain chain on first unlock
	if err := ensureChainGenesis(db, priv); err != nil {
		log.Printf("⚠️  chain genesis: %v", err)
	}

	log.Printf("Key unlocked: %s", vault.PublicAddr)

	checkLicenseNow()

	c.JSON(http.StatusOK, gin.H{
		"status":      "unlocked",
		"public_addr": vault.PublicAddr,
	})
}

func keyStatusHandler(c *gin.Context) {
	stateMutex.RLock()
	locked := unlockedKey == nil
	addr := unlockedAddr
	stateMutex.RUnlock()

	var count int64
	db.Model(&KeyVault{}).Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"has_key":     count > 0,
		"unlocked":    !locked,
		"public_addr": addr,
	})
}

// ============ TRANSACTIONS ============

func createTxHandler(c *gin.Context) {
	var req struct {
		From   string  `json:"from" binding:"required"`
		To     string  `json:"to" binding:"required"`
		Amount float64 `json:"amount"`
		Type   string  `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from, to required"})
		return
	}

	// === License Enforcement ===
	if !isLicenseValid() {
		resp := buildLicenseBlockedResponse()
		c.JSON(http.StatusPaymentRequired, resp)
		return
	}

	// === Grace Period Check ===
	if allowed, reason := isTransactionAllowedInGrace(); !allowed {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":  "license grace period restriction",
			"reason": reason,
		})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}
	if req.From == req.To {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot transfer to self"})
		return
	}
	if req.Type == "" {
		req.Type = "transfer"
	}

	// ═══════════════════════════════════════════════════════════════════
	// P1.3 FIX: Enforce one-time signing session token.
	// The client MUST first call POST /signing/session to obtain a token,
	// then send it in the X-Session-Token header of this request.
	// Without a valid, unconsumed, non-expired token, the transaction is rejected.
	// ═══════════════════════════════════════════════════════════════════
	sessionTok := c.GetHeader("X-Session-Token")
	if sessionTok == "" {
		addAudit("TX_DENIED", "no-session", "reason=missing_session_token", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing X-Session-Token header",
			"hint":  "call POST /api/v1/signing/session first",
		})
		return
	}
	if !ConsumeSessionToken(sessionTok) {
		addAudit("TX_DENIED", sessionTok[:min(16, len(sessionTok))], "reason=invalid_or_expired_session", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid or expired signing session token",
			"hint":  "session tokens are single-use and expire after 5 minutes",
		})
		return
	}

	// ═══════════════════════════════════════════════════════════════════
	// P1.2 FIX: Enforce Dual Control for high-value transactions.
	// If amount >= dualThreshold, the client must have already obtained
	// approval via POST /dual/request and POST /dual/approve by a second
	// admin. The approval ID must be sent in the X-Dual-Approval header.
	// ═══════════════════════════════════════════════════════════════════
	if RequiresDualControl(req.Amount) {
		approvalID := c.GetHeader("X-Dual-Approval")
		if approvalID == "" {
			addAudit("TX_DENIED", "no-dual", fmt.Sprintf("amount=%.2f requires dual control", req.Amount), c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "dual control required for high-value transaction",
				"threshold": dualThreshold,
				"amount":    req.Amount,
				"hint":      "call POST /api/v1/dual/request then POST /api/v1/dual/approve, then retry with X-Dual-Approval header",
			})
			return
		}

		dualMutex.Lock()
		approval, ok := dualPending[approvalID]
		dualMutex.Unlock()

		if !ok {
			addAudit("TX_DENIED", approvalID, "reason=dual_approval_not_found", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "dual approval ID not found or expired",
			})
			return
		}
		if approval.Status != "approved" {
			addAudit("TX_DENIED", approvalID, "reason=dual_not_yet_approved", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "dual approval not yet granted",
				"status": approval.Status,
			})
			return
		}
		if approval.ApprovedBy == approval.RequestedBy {
			addAudit("TX_DENIED", approvalID, "reason=dual_same_approver", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{
				"error": "dual approval must be granted by a different admin",
			})
			return
		}

		// Consume the approval — one-time use
		dualMutex.Lock()
		delete(dualPending, approvalID)
		dualMutex.Unlock()
		addAudit("DUAL_CONSUMED", approvalID, fmt.Sprintf("amount=%.2f", req.Amount), c.ClientIP())
	}

	// ═══ Actual transaction ═══
	// Reserve capacity before execution
	if err := ReserveLicenseUsage(1); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "license volume exceeded", "details": err.Error()})
		return
	}

	// ATOMIC RESERVATION BEFORE EXECUTION
	if err := ReserveLicenseUsage(1); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "license volume exceeded", "details": err.Error()})
		return
	}

	if err := ledger.Transfer(req.From, req.To, req.Amount); err != nil {
		ReleaseLicenseUsage(1) // Release on failure
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Release capacity on failure
	if err := ReleaseLicenseUsage(1); err != nil {
		fmt.Printf("WARN: failed to release license usage: %v\n", err)
	}

	txID := fmt.Sprintf("tx_%d_%s", time.Now().UnixNano(), hashData([]byte(req.From + req.To))[:8])
	tx := &Transaction{
		TxID:       txID,
		From:       req.From,
		To:         req.To,
		Amount:     req.Amount,
		Type:       req.Type,
		Timestamp:  time.Now().Unix(),
		BlockIndex: -1,
		Synced:     false,
	}
	queueTransaction(tx)

	// append transaction batch to blockchain chain
	stateMutex.RLock()
	chainTxKey := unlockedKey
	stateMutex.RUnlock()
	if chainTxKey != nil {
		batchPayload := map[string]interface{}{
			"tx_id":  txID,
			"from":   req.From,
			"to":     req.To,
			"amount": req.Amount,
			"type":   req.Type,
		}
		if chTx, _, err := appendChainTx(db, chainTxKey, TxBatch, batchPayload); err != nil {
			log.Printf("⚠️  chain: tx append failed: %v", err)
		} else {
			log.Printf("⛓  tx %s appended in block #%d", txID, chTx.BlockNum)
		}
	}

	addAudit("TX_CREATED", txID, fmt.Sprintf("from=%s to=%s amount=%.2f", req.From, req.To, req.Amount), c.ClientIP())

	newFrom := ledger.GetBalance(req.From)
	newTo := ledger.GetBalance(req.To)

	c.JSON(http.StatusCreated, gin.H{
		"tx": tx,
		"balances": gin.H{
			req.From: newFrom,
			req.To:   newTo,
		},
	})
}

func listTxHandler(c *gin.Context) {
	var txs []Transaction
	db.Order("timestamp desc").Limit(50).Find(&txs)
	c.JSON(http.StatusOK, gin.H{
		"total":        len(txs),
		"transactions": txs,
	})
}

// ============ BLOCKS ============

func mineBlockHandler(c *gin.Context) {
	stateMutex.RLock()
	hasKey := unlockedKey != nil
	stateMutex.RUnlock()

	if !hasKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "key not unlocked, please unlock first"})
		return
	}

	block, err := mineBlock()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, block)
}

func listBlocksHandler(c *gin.Context) {
	var blocks []Block
	db.Order("block_num desc").Limit(50).Find(&blocks)
	c.JSON(http.StatusOK, gin.H{
		"chainLength": len(blocks),
		"blocks":      blocks,
	})
}

// ============ LICENSES ============

func saveLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID  string `json:"license_id" binding:"required"`
		UserID     string `json:"user_id" binding:"required"`
		ProductID  string `json:"product_id"`
		Volume     int    `json:"volume"`
		Duration   int    `json:"duration"`
		MerkleRoot string `json:"merkle_root"`
		Signature  string `json:"signature" binding:"required"`
		SignedData string `json:"signed_data" binding:"required"`
		HardwareID string `json:"hardware_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sd map[string]interface{}
	json.Unmarshal([]byte(req.SignedData), &sd)

	var issuedAt, expiresAt int64
	var duration int
	var merkleRoot, hardwareID string

	if val, ok := sd["issued_at"].(float64); ok {
		issuedAt = int64(val)
	}
	if val, ok := sd["expires_at"].(float64); ok {
		expiresAt = int64(val)
	}
	if val, ok := sd["duration"].(float64); ok {
		duration = int(val)
	}
	if val, ok := sd["merkle_root"].(string); ok {
		merkleRoot = val
	}
	if val, ok := sd["hardware_id"].(string); ok {
		hardwareID = val
	}

	if merkleRoot == "" {
		merkleRoot = req.MerkleRoot
	}
	if hardwareID == "" {
		hardwareID = req.HardwareID
	}

	lic := &License{
		LicenseID:  req.LicenseID,
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Volume:     req.Volume,
		Duration:   duration,
		MerkleRoot: merkleRoot,
		IssuedAt:   issuedAt,
		ExpiresAt:  expiresAt,
		Status:     "active",
		HardwareID: hardwareID,
		Signature:  req.Signature,
	}

	if err := db.Create(lic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed: " + err.Error()})
		return
	}

	log.Printf("License saved: %s", lic.LicenseID)
	go checkLicenseNow()
	c.JSON(http.StatusCreated, lic)
}

func listLicensesHandler(c *gin.Context) {
	var licenses []License
	db.Order("created_at desc").Find(&licenses)
	c.JSON(http.StatusOK, gin.H{
		"total":    len(licenses),
		"licenses": licenses,
	})
}

// ============ BANK ACCOUNT ============

func bankLoginHandler(c *gin.Context) {
	var req struct {
		BankID   string `json:"bank_id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var bank BankAccount
	if err := db.Where("bank_id = ?", req.BankID).First(&bank).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bank not found"})
		return
	}

	passHash := hex.EncodeToString(deriveKeyFromPassword(req.Password))
	if bank.PasswordHash != passHash {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	var license License
	db.Where("user_id = ?", req.BankID).Order("created_at desc").First(&license)

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"bank":    bank,
		"license": license,
	})
}

func createBankHandler(c *gin.Context) {
	var req struct {
		BankID   string `json:"bank_id" binding:"required"`
		BankName string `json:"bank_name" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passHash := hex.EncodeToString(deriveKeyFromPassword(req.Password))

	bank := &BankAccount{
		BankID:       req.BankID,
		BankName:     req.BankName,
		PasswordHash: passHash,
	}

	if err := db.Create(bank).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":    "created",
		"bank_id":   req.BankID,
		"bank_name": req.BankName,
	})
}

// ============ STATS ============

func statsHandler(c *gin.Context) {
	var blockCount, txCount, licCount, pendingTx int64
	db.Model(&Block{}).Count(&blockCount)
	db.Model(&Transaction{}).Count(&txCount)
	db.Model(&License{}).Count(&licCount)
	db.Model(&Transaction{}).Where("block_index = -1").Count(&pendingTx)

	var recentBlocks []Block
	db.Order("block_num desc").Limit(10).Find(&recentBlocks)

	c.JSON(http.StatusOK, gin.H{
		"chainLength":  blockCount,
		"totalTx":      txCount,
		"pendingTx":    pendingTx,
		"licenses":     licCount,
		"recentBlocks": recentBlocks,
	})
}

// ============ MAIN ============

func main() {
	// ─── P0.4 FAIL-CLOSED: refuse to start without required secrets ───
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		adminToken = os.Getenv("API_KEY")
	}
	if adminToken == "" {
		log.Fatal("P0.4 FAIL-CLOSED: ADMIN_TOKEN (or API_KEY) must be set. Refusing to start with open auth.")
	}
	if len(adminToken) < 32 {
		log.Fatal("P0.4 FAIL-CLOSED: ADMIN_TOKEN must be at least 32 characters.")
	}
	log.Printf("P0.4: auth enabled (token length=%d)", len(adminToken))
	// ──────────────────────────────────────────────────────────────

	// Init SQLite
	var err error
	dbPath := os.Getenv("SWITCH_DB")
	if dbPath == "" {
		dbPath = "./data/horizon-switch.db"
	}

	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	if err := configureSQLite(db); err != nil {
		log.Printf("configureSQLite failed: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&Block{}, &Transaction{}, &License{}, &KeyVault{}, &BankAccount{}, &Account{}, &Invoice{}, &Customer{}); err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("SQLite database ready:", dbPath)

	// Auto-setup key from env if present
	autoSetupKeyFromEnv()

	// === License Enforcement ===
	licenseStopCh := make(chan struct{})
	go startLicenseChecker(licenseStopCh)
	log.Println("License enforcement started")
	CheckKeyRotation()
	startAutoLock(30)
	initDualControl()
	initSigningSessions()

	// === In-Memory Ledger ===
	ledger = NewLedger()
	initAirGap()

	// init license chain (blockchain registry)
	if err := initLicenseChain(db); err != nil {

		log.Printf("⚠️  license chain init: %v", err)
	}
	initChainConfig()
	initCustomers()

	// init adaptive chain batcher (async batching)
	InitBatcher(db, BatcherConfig{
		MaxSize:       50000,
		FlushInterval: 200 * time.Millisecond,
	})

	// start industrial anchor worker
	startIndustrialAnchorWorker()
	initAuditMiddleware()
	if err := loadLedgerFromDB(ledger); err != nil {
		log.Printf("ledger load warning: %v", err)
	}
	initTxQueue()
	ledgerStop = make(chan struct{})
	startLedgerLoops(ledger, ledgerStop)
	log.Println("Ledger initialized (in-memory + async persist)")

	// Ensure genesis block exists
	getLastBlock()

	// Init Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if origin == "" || origin == "null" {
				return true
			}
			return strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				origin == "https://horizon-backend.liara.run" ||
				origin == "https://horizon-switch.liara.run"
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Customer-Token", "X-Admin-Token"},
		AllowCredentials: false,
	}))
	r.Use(adminAuth)
	r.Use(airgapMiddleware())
	r.Use(chainIDMiddleware())
	r.Use(auditMiddleware())
	rl := NewRateLimiter(500000, 500000)
	r.Use(SecurityHeadersMiddleware())
	r.Use(RateLimitMiddleware(rl))
	r.Use(SanitizeInputMiddleware())

	// Routes
	api := r.Group("/api/v1")
	api.POST("/license/check", licenseCheckHandler)
	{
		registerIndustrialRoutes(api)

		registerTenantRoutes(api)

		startIndustrialBackgroundJobs()
		api.GET("/health", healthHandler)
		api.GET("/chain/info", chainInfoHandler)
		api.GET("/chain/stats", func(c *gin.Context) {
			if b := GetBatcher(); b != nil {
				c.JSON(200, b.Stats())
				return
			}
			c.JSON(200, gin.H{"error": "batcher not initialized"})
		})

		api.POST("/chain/flush", func(c *gin.Context) {
			if b := GetBatcher(); b != nil {
				b.Flush()
				c.JSON(http.StatusOK, gin.H{"status": "flushed"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "batcher not initialized"})
		})
		api.GET("/stats", statsHandler)

		// License enforcement status
		api.GET("/license/enforce/status", func(c *gin.Context) {
			valid, lic, err := getLicenseState()
			var errStr interface{}
			if err != nil {
				errStr = err.Error()
			}
			c.JSON(http.StatusOK, gin.H{
				"valid":   valid,
				"license": lic,
				"error":   errStr,
			})
		})

		// License grace status
		api.GET("/license/grace-status", func(c *gin.Context) {
			lic := getActiveLicenseOrNil()
			status := getGraceStatus(lic)
			c.JSON(http.StatusOK, status)
		})

		// Key vault
		api.POST("/key/setup", setupKeyHandler)
		api.POST("/key/unlock", unlockKeyHandler)
		api.GET("/key/status", keyStatusHandler)
		api.POST("/key/lock", lockKeyHandler)
		api.POST("/dual/request", requestApprovalHandler)
		api.POST("/dual/approve", approveHandler)
		api.GET("/dual/pending", listPendingHandler)
		api.POST("/signing/session", requestSigningSessionHandler)

		// Transactions
		api.POST("/tx", createTxHandler)
		api.POST("/tx/batch", createTxBatchHandler)
		api.GET("/tx/list", listTxHandler)

		// Accounts
		api.GET("/account/list", listAccountsHandler)
		api.GET("/account/balance/:bank_id", getBalanceHandler)
		api.POST("/account/seed", seedBalanceHandler)

		// Sync
		api.POST("/sync/push", syncPushHandler)

		// Blocks
		api.POST("/block/mine", mineBlockHandler)
		api.GET("/block/list", listBlocksHandler)

		// Licenses
		api.POST("/license/save", saveLicenseHandler)
		api.GET("/license/list", listLicensesHandler)
		api.POST("/license/verify", verifyLicenseHandler)

		// License renewal
		api.POST("/license/renew/request", requestRenewalHandler)
		api.POST("/license/renew/confirm", confirmRenewalHandler)
		api.GET("/license/renew/history", renewalHistoryHandler)
		api.GET("/license/status", licenseStatusHandler)

		// Payment
		api.POST("/payment/initiate", initiatePaymentHandler)
		api.POST("/payment/mock/confirm", mockConfirmHandler)
		api.GET("/payment/success", paymentSuccessPageHandler)

		// Bank
		api.POST("/bank/create", createBankHandler)
		api.POST("/bank/login", bankLoginHandler)

		// Admin
		api.POST("/admin/login", adminLoginHandler)
		api.POST("/admin/logout", adminLogoutHandler)

		admin := api.Group("/admin")
		admin.Use(adminAuthRequired())
		{
			admin.GET("/banks", listBanksHandler)
			admin.GET("/audit", adminAuditHandler)
			admin.POST("/license/issue", adminIssueLicenseHandler)
			admin.POST("/license/revoke", revokeLicenseHandler)
		}
	}

	// Customer panel
	apiCustomer := r.Group("/api/v1/customer")
	{
		apiCustomer.POST("/login", customerLoginHandler)
		apiCustomer.POST("/logout", customerLogoutHandler)

		authCustomer := apiCustomer.Group("")
		authCustomer.Use(customerAuthRequired())
		{
			authCustomer.GET("/me", customerMeHandler)
			authCustomer.GET("/licenses", customerLicensesHandler)
			authCustomer.GET("/invoices", customerInvoicesHandler)
			authCustomer.GET("/license/download/:license_id", customerDownloadLicenseHandler)
			authCustomer.POST("/license/renew/:license_id", customerRenewHandler)
		}
	}

	// Compat endpoints
	r.GET("/health", healthHandler)
	r.GET("/stats", statsHandler)
	r.GET("/benchmark", benchmarkHandler)
	r.GET("/payment/mock", mockPaymentPageHandler)
	r.POST("/api/v1/bnpl/anchor", adminAuthRequired(), bnplAnchorHandler)

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Horizon Switch v1.4 running on port %s", port)
	log.Printf("Database: %s", dbPath)

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Starting server with connection limit (500)...")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Listen failed:", err)
	}
	limitedListener := netutil.LimitListener(listener, 500)
	// graceful shutdown with chain flush
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("⏹  shutdown signal received, flushing batcher...")
		if b := GetBatcher(); b != nil {
			b.Flush()
			b.Stop()
		}
		srv.Shutdown(context.Background())
	}()

	if err := srv.Serve(limitedListener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// startAutoLock runs a goroutine that locks the key after inactivity
func startAutoLock(timeoutMinutes int) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			stateMutex.Lock()
			if unlockedKey != nil && !unlockedAt.IsZero() {
				if time.Since(unlockedAt).Minutes() > float64(timeoutMinutes) {
					unlockedKey = nil
					unlockedAddr = ""
					unlockedAt = time.Time{}
					log.Printf("[AUTO-LOCK] Key locked due to %d min inactivity", timeoutMinutes)
				}
			}
			stateMutex.Unlock()
		}
	}()
}

// lockKeyHandler locks the key manually
func lockKeyHandler(c *gin.Context) {
	stateMutex.Lock()
	wasUnlocked := unlockedKey != nil
	unlockedKey = nil
	unlockedAddr = ""
	unlockedAt = time.Time{}
	stateMutex.Unlock()

	if wasUnlocked {
		log.Printf("[MANUAL-LOCK] Key locked by admin")
	}
	c.JSON(http.StatusOK, gin.H{"status": "locked"})
}

// ═══════════════════════════════════════════════════════════════════
// createTxBatchHandler — batch transactions
// Accepts up to 1000 txs in one HTTP request.
// Creates ONE block with all txs (vs N blocks).
// ~100x faster than individual /tx for high volume.
// ═══════════════════════════════════════════════════════════════════

func createTxBatchHandler(c *gin.Context) {
	var req struct {
		SessionToken string `json:"session_token" binding:"required"`
		Txs          []struct {
			From   string  `json:"from"`
			To     string  `json:"to"`
			Amount float64 `json:"amount"`
			Type   string  `json:"type"`
		} `json:"txs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// License check
	if !isLicenseValid() {
		c.JSON(http.StatusPaymentRequired, buildLicenseBlockedResponse())
		return
	}

	// Session token (single-use)
	if !ConsumeSessionToken(req.SessionToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session token"})
		return
	}

	if len(req.Txs) == 0 || len(req.Txs) > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "txs must be 1..100000"})
		return
	}

	// ATOMIC RESERVATION BEFORE EXECUTION
	if err := ReserveLicenseUsage(len(req.Txs)); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "license volume exceeded", "details": err.Error()})
		return
	}

	// Process all txs in ledger
	processed := make([]*Transaction, 0, len(req.Txs))
	failed := 0
	for i, t := range req.Txs {
		if t.Amount <= 0 || t.From == "" || t.To == "" || t.From == t.To {
			failed++
			continue
		}
		if err := ledger.Transfer(t.From, t.To, t.Amount); err != nil {
			failed++
			continue
		}
		txID := fmt.Sprintf("tx_%d_%d_%s", time.Now().UnixNano(), i, hashData([]byte(t.From + t.To))[:6])
		if t.Type == "" {
			t.Type = "transfer"
		}
		tx := &Transaction{
			TxID:       txID,
			From:       t.From,
			To:         t.To,
			Amount:     t.Amount,
			Type:       t.Type,
			Timestamp:  time.Now().Unix(),
			BlockIndex: -1,
		}
		queueTransaction(tx)
		processed = append(processed, tx)
	}

	// Release capacity for failed transactions
	if failed > 0 {
		if err := ReleaseLicenseUsage(failed); err != nil {
			fmt.Printf("WARN: failed to release license usage: %v\n", err)
		}
	}

	// Chain: batch payload
	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()

	chainInfo := "offline"
	if key != nil && len(processed) > 0 {
		if b := GetBatcher(); b != nil {
			for _, tx := range processed {
				b.Submit(TxBatch, map[string]interface{}{
					"tx_id":  tx.TxID,
					"from":   tx.From,
					"to":     tx.To,
					"amount": tx.Amount,
				})
			}
			chainInfo = fmt.Sprintf("batched:%d", len(processed))
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"processed": len(processed),
		"failed":    failed,
		"chain":     chainInfo,
	})
}
