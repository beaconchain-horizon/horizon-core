package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"horizon-core/internal/chain"
	"horizon-core/internal/crypto"
)

// registerChainHandlers wires the chain endpoints on the given router.
// Backend is the source of truth for all chain queries.
func registerChainHandlers(r *gin.Engine, db *gorm.DB) {
	api := r.Group("/api/v1")

	// GET /api/v1/chain/info
	api.GET("/chain/info", func(c *gin.Context) {
		info, err := chain.GetChainInfo(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, info)
	})

	// GET /api/v1/blocks?from=&to=&limit=
	api.GET("/blocks", func(c *gin.Context) {
		from, _ := strconv.ParseInt(c.Query("from"), 10, 64)
		to, _ := strconv.ParseInt(c.Query("to"), 10, 64)
		limit, _ := strconv.Atoi(c.Query("limit"))
		blocks, err := chain.ListBlocks(db, from, to, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"blocks": blocks, "count": len(blocks)})
	})

	// GET /api/v1/blocks/:number
	api.GET("/blocks/:number", func(c *gin.Context) {
		n, err := strconv.ParseInt(c.Param("number"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block number"})
			return
		}
		block, err := chain.GetBlock(db, n)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		txs, _ := chain.GetBlockTxs(db, n)
		c.JSON(http.StatusOK, gin.H{"block": block, "transactions": txs})
	})

	// GET /api/v1/transactions/:tx_id
	api.GET("/transactions/:tx_id", func(c *gin.Context) {
		tx, err := chain.GetTx(db, c.Param("tx_id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}
		c.JSON(http.StatusOK, tx)
	})

	// GET /api/v1/chain/pending
	api.GET("/chain/pending", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pending": chain.PendingCount()})
	})

	// POST /api/v1/license/register
	// Receives a license that was signed OUTSIDE the Backend (in the browser
	// from a USB key). Backend verifies the signature, persists the license,
	// and enqueues a transaction for the next block.
	api.POST("/license/register", func(c *gin.Context) {
		var req struct {
			LicenseID  string `json:"license_id" binding:"required"`
			UserID     string `json:"user_id" binding:"required"`
			ProductID  string `json:"product_id" binding:"required"`
			Volume     int    `json:"volume" binding:"required"`
			Signature  string `json:"signature" binding:"required"`
			SignedData string `json:"signed_data" binding:"required"`
			PublicKey  string `json:"public_key"`
			HardwareID string `json:"hardware_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify signature with the Authority public key (Backend does NOT
		// hold the private key).
		pubHex := req.PublicKey
		if pubHex == "" {
			pubHex = os.Getenv("HORIZON_LICENSE_PUBLIC_KEY")
		}
		pub, err := crypto.Secp256k1PublicKeyFromHex(pubHex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public key: " + err.Error()})
			return
		}
		ok, err := crypto.VerifyData(pub, []byte(req.SignedData), req.Signature)
		if err != nil || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "signature verification failed"})
			return
		}

		// Persist the license.
		lic := LicenseRecord{
			ID:        req.LicenseID,
			ProductID: req.ProductID,
			UserID:    req.UserID,
			Volume:    req.Volume,
			Signature: req.Signature,
			Status:    "active",
			CreatedAt: time.Now().UTC(),
		}
		if err := pgDB.Create(&lic).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "persist: " + err.Error()})
			return
		}

		// Enqueue for the next block.
		if trialTxLimitReached() {
			c.JSON(http.StatusForbidden, gin.H{"error": "TRIAL: transaction limit reached (1,000,000)"})
			return
		}

		if err := chain.Enqueue(chain.PendingTx{
			TxID:      chain.NewTxID(req.SignedData, time.Now()),
			TxType:    "license_issue",
			Payload:   req.SignedData,
			Signature: req.Signature,
			PublicKey: pubHex,
		}); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "enqueue: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"license_id": req.LicenseID,
			"status":     "queued",
			"pending":    chain.PendingCount(),
		})
	})
}

// small wrapper to avoid extra import in handlers file
func osGetenv(k string) string {
	return os.Getenv(k)
}
