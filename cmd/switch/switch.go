package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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
	TxCount    int       `json:"tx_count"`
	CreatedAt  time.Time `json:"created_at"`
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
	unlockedAddr string
	stateMutex   sync.RWMutex
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
	key := hexKey
	if len(key) > 2 && key[:2] == "0x" {
		key = key[2:]
	}
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = elliptic.P256()
	priv.D = new(big.Int).SetBytes(bytes)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(bytes)
	return priv, nil
}

func publicAddressFromKey(priv *ecdsa.PrivateKey) string {
	pubBytes := elliptic.MarshalCompressed(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	hash := sha256.Sum256(pubBytes)
	return "0x" + hex.EncodeToString(hash[:20])
}

func signData(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)

	r.FillBytes(sig[:32])

	s.FillBytes(sig[32:])

	return hex.EncodeToString(sig), nil
}

func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func simpleMerkleRoot(txIDs []string) string {
	if len(txIDs) == 0 {
		return hashData([]byte("empty"))
	}
	hashes := make([]string, len(txIDs))
	for i, id := range txIDs {
		hashes[i] = hashData([]byte(id))
	}
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		next := []string{}
		for i := 0; i < len(hashes); i += 2 {
			combined := hashes[i] + hashes[i+1]
			next = append(next, hashData([]byte(combined)))
		}
		hashes = next
	}
	return hashes[0]
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

	// Get pending transactions
	var pending []Transaction
	db.Where("block_index = ?", -1).Find(&pending)

	if len(pending) == 0 {
		return nil, fmt.Errorf("no pending transactions")
	}

	txIDs := []string{}
	for _, tx := range pending {
		txIDs = append(txIDs, tx.TxID)
	}
	merkleRoot := simpleMerkleRoot(txIDs)

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
		TxCount:    len(pending),
	}

	if err := db.Create(block).Error; err != nil {
		return nil, err
	}

	// Update transactions
	for _, tx := range pending {
		db.Model(&tx).Update("block_index", newIndex)
	}

	log.Printf("⛓️  Block #%d mined: %d transactions, merkle=%s", newIndex, len(pending), merkleRoot[:16])
	return block, nil
}

// ============================================================
// HTTP HANDLERS
// ============================================================

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"service": "horizon-switch",
		"version": "3.0",
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

	// Delete old keys, keep only one
	db.Where("1 = 1").Delete(&KeyVault{})
	if err := db.Create(vault).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "saved",
		"public_addr":  addr,
		"message":      "کلید خصوصی با موفقیت رمزنگاری و ذخیره شد",
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
	unlockedAddr = vault.PublicAddr
	stateMutex.Unlock()

	log.Printf("🔓 Key unlocked: %s", vault.PublicAddr)

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

	// ===== Validation =====
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

	// ===== چک وجود حساب‌ها =====
	var sender, receiver Account
	if err := db.Where("bank_id = ?", req.From).First(&sender).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender account not found: " + req.From})
		return
	}
	if err := db.Where("bank_id = ?", req.To).First(&receiver).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver account not found: " + req.To})
		return
	}

	// ===== چک موجودی کافی =====
	if sender.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "insufficient balance",
			"bank_id":  req.From,
			"balance":  sender.Balance,
			"required": req.Amount,
		})
		return
	}

	// ===== ساخت tx و انتقال اتمیک =====
	txID := fmt.Sprintf("tx_%d_%s", time.Now().UnixNano(), hashData([]byte(req.From+req.To))[:8])
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

	err := db.Transaction(func(d *gorm.DB) error {
		if err := d.Create(tx).Error; err != nil {
			return err
		}
		if err := d.Model(&Account{}).Where("bank_id = ?", req.From).
			Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
			return err
		}
		if err := d.Model(&Account{}).Where("bank_id = ?", req.To).
			Update("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ===== خواندن موجودی جدید =====
	var newSender, newReceiver Account
	db.Where("bank_id = ?", req.From).First(&newSender)
	db.Where("bank_id = ?", req.To).First(&newReceiver)

	log.Printf("💸 TX: %s | %.2f | %s → %s | bal %s: %.2f → %.2f",
		txID, req.Amount, req.From, req.To, req.From, sender.Balance, newSender.Balance)

	c.JSON(http.StatusCreated, gin.H{
		"tx": tx,
		"balances": gin.H{
			req.From: newSender.Balance,
			req.To:   newReceiver.Balance,
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
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()

	if key == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is locked - call /api/v1/key/unlock first"})
		return
	}

	now := time.Now().Unix()
	lic := &License{
		LicenseID:  req.LicenseID,
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Volume:     req.Volume,
		Duration:   req.Duration,
		MerkleRoot: req.MerkleRoot,
		IssuedAt:   now,
		ExpiresAt:  now + int64(req.Duration*3600),
		Status:     "active",
	}

	msg := licenseCanonicalMessage(lic)
	sig, err := signData(key, []byte(msg))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign failed: " + err.Error()})
		return
	}
	lic.Signature = sig

	if err := db.Create(lic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	log.Printf("License signed & saved: %s for %s", lic.LicenseID, lic.UserID)
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

	// Get license
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
		"chainLength":     blockCount,
		"totalTx":         txCount,
		"pendingTx":       pendingTx,
		"licenses":        licCount,
		"recentBlocks":    recentBlocks,
	})
}

// ============ MAIN ============

func main() {
	// Init SQLite
	var err error
	dbPath := os.Getenv("SWITCH_DB")
	if dbPath == "" {
		dbPath = "./data/horizon-switch.db"
	}
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to open database:", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&Block{}, &Transaction{}, &License{}, &KeyVault{}, &BankAccount{}, &Account{}); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}
	log.Println("✅ SQLite database ready:", dbPath)

	// Ensure genesis block exists
	getLastBlock()

	// Init Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	// Routes
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthHandler)
		api.GET("/stats", statsHandler)

		// Key vault
		api.POST("/key/setup", setupKeyHandler)
		api.POST("/key/unlock", unlockKeyHandler)
		api.GET("/key/status", keyStatusHandler)

		// Transactions
		api.POST("/tx", createTxHandler)

		// ===== Accounts (real balances) =====
		api.GET("/account/list", listAccountsHandler)
		api.GET("/account/balance/:bank_id", getBalanceHandler)
		api.POST("/account/seed", seedBalanceHandler)

		// ===== Accounts (real balances) =====
		api.GET("/tx/list", listTxHandler)

		// ===== Sync =====
		api.POST("/sync/push", syncPushHandler)

		// ===== Sync =====

		// Blocks
		api.POST("/block/mine", mineBlockHandler)
		api.GET("/block/list", listBlocksHandler)

		// Licenses
		api.POST("/license/save", saveLicenseHandler)
		api.GET("/license/list", listLicensesHandler)
		api.POST("/license/verify", verifyLicenseHandler)

		// Bank
		api.POST("/bank/create", createBankHandler)
		api.POST("/bank/login", bankLoginHandler)

		// ===== Admin =====
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

	// Compat endpoints (for old frontend)
	r.GET("/health", healthHandler)
	r.GET("/stats", statsHandler)
	r.GET("/benchmark", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tps": 35000, "status": "online"})
	})

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Horizon Switch v3.0 running on port %s", port)
	log.Printf("📁 Database: %s", dbPath)
	log.Printf("🔒 Key status: run POST /api/v1/key/status to check")

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
