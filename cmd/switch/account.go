package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Account struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	BankID    string  `gorm:"uniqueIndex;not null" json:"bank_id"`
	Balance   float64 `gorm:"not null;default:0" json:"balance"`
	UpdatedAt int64   `json:"updated_at"`
}

func ensureAccount(bankID string) error {
	var acc Account
	err := db.Where("bank_id = ?", bankID).First(&acc).Error
	if err == gorm.ErrRecordNotFound {
		return db.Create(&Account{BankID: bankID, Balance: 0}).Error
	}
	return err
}

func transferFunds(from, to string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if from == to {
		return fmt.Errorf("cannot transfer to self")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var sender, receiver Account

		if err := tx.Where("bank_id = ?", from).First(&sender).Error; err != nil {
			return fmt.Errorf("sender not found: %s", from)
		}
		if err := tx.Where("bank_id = ?", to).First(&receiver).Error; err != nil {
			return fmt.Errorf("receiver not found: %s", to)
		}
		if sender.Balance < amount {
			return fmt.Errorf("insufficient balance: %.2f < %.2f", sender.Balance, amount)
		}

		if err := tx.Model(&Account{}).Where("bank_id = ?", from).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&Account{}).Where("bank_id = ?", to).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
		return nil
	})
}

func getBalanceHandler(c *gin.Context) {
	bankID := c.Param("bank_id")
	var acc Account
	if err := db.Where("bank_id = ?", bankID).First(&acc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"bank_id": acc.BankID,
		"balance": acc.Balance,
	})
}

func listAccountsHandler(c *gin.Context) {
	var accounts []Account
	db.Order("balance desc").Find(&accounts)
	c.JSON(http.StatusOK, gin.H{
		"total":    len(accounts),
		"accounts": accounts,
	})
}

func seedBalanceHandler(c *gin.Context) {
	var req struct {
		BankID string  `json:"bank_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be >= 0"})
		return
	}

	if err := ensureAccount(req.BankID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&Account{}).Where("bank_id = ?", req.BankID).
		Update("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var acc Account
	db.Where("bank_id = ?", req.BankID).First(&acc)

	addAudit("seed_balance", req.BankID,
		fmt.Sprintf("+%.2f", req.Amount), c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"bank_id": acc.BankID,
		"balance": acc.Balance,
	})
}
