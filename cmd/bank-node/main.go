package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ============================================================
// MODELS
// ============================================================

type LocalTx struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TxID      string    `gorm:"uniqueIndex" json:"tx_id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Status    string    `json:"status"` // pending | synced | rejected
	SyncedAt  int64     `json:"synced_at"`
	CreatedAt time.Time `json:"created_at"`
}

// LocalBalance: کش محلی از موجودی بانک ما
type LocalBalance struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	BankID        string  `gorm:"uniqueIndex" json:"bank_id"`
	ConfirmedBal  float64 `json:"confirmed_balance"`  // آخرین sync از مرکزی
	ReservedBal   float64 `json:"reserved_balance"`   // مجموع تراکنش‌های pending
	UpdatedAt     int64   `json:"updated_at"`
}

func (b *LocalBalance) Available() float64 {
	return b.ConfirmedBal - b.ReservedBal
}

// ============================================================
// GLOBALS
// ============================================================

var (
	db         *gorm.DB
	bankID     string
	centralURL string
	nodePort   string
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ============================================================
// MAIN
// ============================================================

func main() {
	bankID = envOr("BANK_ID", "bank_melli")
	centralURL = envOr("CENTRAL_URL", "http://localhost:8080")
	nodePort = envOr("NODE_PORT", "9090")

	dbPath := envOr("BANK_DB", "./data/bank-node.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("db:", err)
	}
	db.AutoMigrate(&LocalTx{}, &LocalBalance{})

	// ساخت رکورد موجودی اگر نیست
	var bal LocalBalance
	if err := db.Where("bank_id = ?", bankID).First(&bal).Error; err == gorm.ErrRecordNotFound {
		db.Create(&LocalBalance{BankID: bankID})
	}

	log.Printf("🏦 Bank node: %s on :%s", bankID, nodePort)

	go syncLoop()
	go balanceRefreshLoop()

	r := gin.Default()
	r.Use(corsMW())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "online", "bank_id": bankID})
	})
	r.POST("/tx", createLocalTx)
	r.GET("/tx/list", listLocalTx)
	r.GET("/status", statusHandler)
	r.GET("/balance", balanceHandler)
	r.POST("/sync", manualSync)

	log.Printf("🚀 Bank node running on :%s", nodePort)
	r.Run(":" + nodePort)
}

func corsMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// ============================================================
// HANDLERS
// ============================================================

func genTxID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("tx_%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func createLocalTx(c *gin.Context) {
	var req struct {
		To     string  `json:"to" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
		Type   string  `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ===== Validation محلی =====
	if req.Amount <= 0 {
		c.JSON(400, gin.H{"error": "amount must be positive"})
		return
	}
	if req.To == bankID {
		c.JSON(400, gin.H{"error": "cannot transfer to self"})
		return
	}
	if req.Type == "" {
		req.Type = "transfer"
	}

	// ===== چک موجودی محلی =====
	var bal LocalBalance
	if err := db.Where("bank_id = ?", bankID).First(&bal).Error; err != nil {
		c.JSON(500, gin.H{"error": "balance not initialized"})
		return
	}
	if bal.Available() < req.Amount {
		c.JSON(409, gin.H{
			"error":     "insufficient available balance",
			"confirmed": bal.ConfirmedBal,
			"reserved":  bal.ReservedBal,
			"available": bal.Available(),
			"required":  req.Amount,
		})
		return
	}

	// ===== ساخت tx + reservation اتمیک =====
	tx := LocalTx{
		TxID:      genTxID(),
		From:      bankID,
		To:        req.To,
		Amount:    req.Amount,
		Type:      req.Type,
		Timestamp: time.Now().Unix(),
		Status:    "pending",
	}

	err := db.Transaction(func(d *gorm.DB) error {
		if err := d.Create(&tx).Error; err != nil {
			return err
		}
		if err := d.Model(&LocalBalance{}).Where("bank_id = ?", bankID).
			Update("reserved_bal", gorm.Expr("reserved_bal + ?", req.Amount)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// خواندن balance جدید
	db.Where("bank_id = ?", bankID).First(&bal)

	log.Printf("📝 Local tx: %s → %s (%.2f) | avail: %.2f",
		bankID, tx.To, tx.Amount, bal.Available())

	c.JSON(201, gin.H{
		"tx": tx,
		"balance": gin.H{
			"confirmed": bal.ConfirmedBal,
			"reserved":  bal.ReservedBal,
			"available": bal.Available(),
		},
	})
}

func listLocalTx(c *gin.Context) {
	var txs []LocalTx
	db.Order("created_at desc").Find(&txs)
	c.JSON(200, gin.H{
		"total":        len(txs),
		"transactions": txs,
	})
}

func statusHandler(c *gin.Context) {
	var total, pending, synced, rejected int64
	db.Model(&LocalTx{}).Count(&total)
	db.Model(&LocalTx{}).Where("status = ?", "pending").Count(&pending)
	db.Model(&LocalTx{}).Where("status = ?", "synced").Count(&synced)
	db.Model(&LocalTx{}).Where("status = ?", "rejected").Count(&rejected)

	var bal LocalBalance
	db.Where("bank_id = ?", bankID).First(&bal)

	c.JSON(200, gin.H{
		"bank_id":      bankID,
		"central_url":  centralURL,
		"total_local":  total,
		"pending_sync": pending,
		"synced":       synced,
		"rejected":     rejected,
		"balance": gin.H{
			"confirmed": bal.ConfirmedBal,
			"reserved":  bal.ReservedBal,
			"available": bal.Available(),
		},
	})
}

func balanceHandler(c *gin.Context) {
	var bal LocalBalance
	if err := db.Where("bank_id = ?", bankID).First(&bal).Error; err != nil {
		c.JSON(404, gin.H{"error": "balance not found"})
		return
	}
	c.JSON(200, gin.H{
		"bank_id":   bankID,
		"confirmed": bal.ConfirmedBal,
		"reserved":  bal.ReservedBal,
		"available": bal.Available(),
		"updated_at": bal.UpdatedAt,
	})
}

func manualSync(c *gin.Context) {
	accepted, dup, rej, err := pushToCentral()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	refreshBalance()
	c.JSON(200, gin.H{
		"accepted":   accepted,
		"duplicates": dup,
		"rejected":   rej,
	})
}

// ============================================================
// SYNC
// ============================================================

func syncLoop() {
	time.Sleep(5 * time.Second)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		accepted, dup, rej, err := pushToCentral()
		if err == nil && accepted+dup+rej > 0 {
			log.Printf("✅ sync: acc=%d dup=%d rej=%d", accepted, dup, rej)
		}
		refreshBalance()
	}
}

func balanceRefreshLoop() {
	time.Sleep(3 * time.Second)
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		refreshBalance()
	}
}

func pushToCentral() (int, int, int, error) {
	var pending []LocalTx
	if err := db.Where("status = ?", "pending").
		Order("id asc").Limit(100).Find(&pending).Error; err != nil {
		return 0, 0, 0, err
	}
	if len(pending) == 0 {
		return 0, 0, 0, nil
	}

	payload := map[string]interface{}{
		"bank_id":      bankID,
		"transactions": pending,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(
		centralURL+"/api/v1/sync/push",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, 0, fmt.Errorf("central returned %d", resp.StatusCode)
	}

	var result struct {
		Accepted   int `json:"accepted"`
		Duplicates int `json:"duplicates"`
		Rejected   int `json:"rejected"`
		Reasons    []struct {
			TxID   string `json:"tx_id"`
			Reason string `json:"reason"`
		} `json:"reasons"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, 0, err
	}

	// بر اساس پاسخ، وضعیت هر tx رو تعیین کن
	rejectedSet := make(map[string]string)
	for _, r := range result.Reasons {
		rejectedSet[r.TxID] = r.Reason
	}

	ids := make([]uint, 0, len(pending))
	for _, tx := range pending {
		ids = append(ids, tx.ID)

		if reason, ok := rejectedSet[tx.TxID]; ok {
			// رد شده
			db.Model(&LocalTx{}).Where("id = ?", tx.ID).
				Updates(map[string]interface{}{
					"status": "rejected",
				})
			log.Printf("❌ TX rejected: %s | %s", tx.TxID, reason)
		} else {
			// پذیرفته یا duplicate → synced
			db.Model(&LocalTx{}).Where("id = ?", tx.ID).
				Updates(map[string]interface{}{
					"status":    "synced",
					"synced_at": time.Now().Unix(),
				})
		}
	}

	return result.Accepted, result.Duplicates, result.Rejected, nil
}

// refreshBalance: موجودی confirmed رو از سوئیچ مرکزی می‌گیره
// و reserved رو بر اساس pendingهای باقی‌مونده بازحساب می‌کنه
func refreshBalance() {
	// ۱. از سوئیچ بگیر
	resp, err := http.Get(centralURL + "/api/v1/account/balance/" + bankID)
	if err == nil && resp.StatusCode == 200 {
		var result struct {
			Balance float64 `json:"balance"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			db.Model(&LocalBalance{}).Where("bank_id = ?", bankID).
				Update("confirmed_bal", result.Balance)
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	// ۲. reserved رو بازحساب کن
	var sumReserved struct{ Total float64 }
	db.Model(&LocalTx{}).
		Where("status = ?", "pending").
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&sumReserved)

	db.Model(&LocalBalance{}).Where("bank_id = ?", bankID).
		Updates(map[string]interface{}{
			"reserved_bal": sumReserved.Total,
			"updated_at":   time.Now().Unix(),
		})
}
