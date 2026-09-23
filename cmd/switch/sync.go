package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================
// SYNC PUSH — بانک‌ها تراکنش‌های آفلاین می‌فرستن، با validation
// ============================================================

func syncPushHandler(c *gin.Context) {
	var req struct {
		BankID       string `json:"bank_id" binding:"required"`
		Transactions []struct {
			TxID      string  `json:"tx_id" binding:"required"`
			From      string  `json:"from" binding:"required"`
			To        string  `json:"to" binding:"required"`
			Amount    float64 `json:"amount" binding:"required"`
			Type      string  `json:"type"`
			Timestamp int64   `json:"timestamp"`
		} `json:"transactions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accepted := 0
	duplicates := 0
	rejected := 0
	rejectReasons := []gin.H{}

	for _, t := range req.Transactions {
		// ===== ۱. Dedup =====
		var existing Transaction
		if err := db.Where("tx_id = ?", t.TxID).First(&existing).Error; err == nil {
			duplicates++
			continue
		}

		// ===== ۲. Validation =====
		if t.Amount <= 0 {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID, "reason": "amount must be positive",
			})
			continue
		}
		if t.From == t.To {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID, "reason": "cannot transfer to self",
			})
			continue
		}

		// ===== ۳. وجود حساب‌ها =====
		var sender, receiver Account
		if err := db.Where("bank_id = ?", t.From).First(&sender).Error; err != nil {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID, "reason": "sender not found: " + t.From,
			})
			continue
		}
		if err := db.Where("bank_id = ?", t.To).First(&receiver).Error; err != nil {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID, "reason": "receiver not found: " + t.To,
			})
			continue
		}

		// ===== ۴. موجودی کافی =====
		if sender.Balance < t.Amount {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID,
				"reason": fmt.Sprintf("insufficient balance: %.2f < %.2f",
					sender.Balance, t.Amount),
			})
			continue
		}

		// ===== ۵. Atomic: ساخت tx + انتقال =====
		ts := t.Timestamp
		if ts == 0 {
			ts = time.Now().Unix()
		}
		if t.Type == "" {
			t.Type = "transfer"
		}

		tx := &Transaction{
			TxID:       t.TxID,
			From:       t.From,
			To:         t.To,
			Amount:     t.Amount,
			Type:       t.Type,
			Timestamp:  ts,
			BlockIndex: -1,
			Synced:     true,
		}

		err := db.Transaction(func(d *gorm.DB) error {
			if err := d.Create(tx).Error; err != nil {
				return err
			}
			if err := d.Model(&Account{}).Where("bank_id = ?", t.From).
				Update("balance", gorm.Expr("balance - ?", t.Amount)).Error; err != nil {
				return err
			}
			if err := d.Model(&Account{}).Where("bank_id = ?", t.To).
				Update("balance", gorm.Expr("balance + ?", t.Amount)).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			rejected++
			rejectReasons = append(rejectReasons, gin.H{
				"tx_id": t.TxID, "reason": "db error: " + err.Error(),
			})
			continue
		}
		accepted++
	}

	addAudit("sync_push", req.BankID,
		fmt.Sprintf("acc=%d dup=%d rej=%d", accepted, duplicates, rejected),
		c.ClientIP())

	result := gin.H{
		"status":     "ok",
		"bank_id":    req.BankID,
		"accepted":   accepted,
		"duplicates": duplicates,
		"rejected":   rejected,
	}
	if len(rejectReasons) > 0 {
		result["reasons"] = rejectReasons
	}
	c.JSON(http.StatusOK, result)
}
