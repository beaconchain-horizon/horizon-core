package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// INVOICE MODEL
// ============================================================

// Invoice represents a pending or paid renewal invoice.
type Invoice struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	InvoiceID   string    `gorm:"uniqueIndex" json:"invoice_id"`
	LicenseID   string    `gorm:"index" json:"license_id"`
	UserID      string    `gorm:"index" json:"user_id"`
	AmountIRR   int64     `json:"amount_irr"`
	DurationH   int       `json:"duration_hours"`
	Status      string    `json:"status"` // pending | paid | cancelled
	PaymentRef  string    `json:"payment_ref"`
	CreatedAt   time.Time `json:"created_at"`
	PaidAt      int64     `json:"paid_at"`
}

// ============================================================
// PRICING
// ============================================================

// pricePerDayToman is the base price per day of license
// validity, in Toman.
const pricePerDayToman int64 = 1000000

// priceForDuration returns the total price in Toman for
// the given duration, in hours.
func priceForDuration(hours int) int64 {
	if hours <= 0 {
		return 0
	}
	days := int64(hours) / 24
	if days < 1 {
		days = 1
	}
	return days * pricePerDayToman
}

// ============================================================
// ID GENERATORS
// ============================================================

func generateInvoiceID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("INV-%d-%s",
		time.Now().Unix(),
		hex.EncodeToString(b),
	)
}

func nextLicenseID(prevID string) string {
	// If prevID ends with -R<N>, increment N.
	// Otherwise append -R2.
	idx := -1
	for i := len(prevID) - 1; i >= 0; i-- {
		if prevID[i] == '-' {
			idx = i
			break
		}
	}
	if idx == -1 {
		return prevID + "-R2"
	}
	suffix := prevID[idx+1:]
	if len(suffix) < 2 || suffix[0] != 'R' {
		return prevID + "-R2"
	}
	num := 0
	for _, c := range suffix[1:] {
		if c < '0' || c > '9' {
			return prevID + "-R2"
		}
		num = num*10 + int(c-'0')
	}
	return fmt.Sprintf("%s-R%d", prevID[:idx], num+1)
}

// ============================================================
// HANDLERS
// ============================================================

// requestRenewalHandler creates a new pending invoice for
// renewing the given license.
//
// POST /api/v1/license/renew/request
// Body: {"license_id": "...", "duration_hours": 8760}
func requestRenewalHandler(c *gin.Context) {
	var req struct {
		LicenseID     string `json:"license_id" binding:"required"`
		DurationHours int    `json:"duration_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.DurationHours <= 0 {
		req.DurationHours = 8760 // default: 1 year
	}

	// Verify the license exists
	var lic License
	if err := db.Where("license_id = ?", req.LicenseID).First(&lic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}

	// Create invoice
	inv := &Invoice{
		InvoiceID: generateInvoiceID(),
		LicenseID: req.LicenseID,
		UserID:    lic.UserID,
		AmountIRR: priceForDuration(req.DurationHours),
		DurationH: req.DurationHours,
		Status:    "pending",
	}

	if err := db.Create(inv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	addAudit("renew_request", req.LicenseID,
		fmt.Sprintf("invoice=%s amount=%d", inv.InvoiceID, inv.AmountIRR),
		c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"invoice_id":     inv.InvoiceID,
		"license_id":     inv.LicenseID,
		"user_id":        inv.UserID,
		"amount_irr":     inv.AmountIRR,
		"amount_toman":   inv.AmountIRR,
		"duration_hours": inv.DurationH,
		"status":         inv.Status,
		"payment_url":    "", // populated by Phase F (ZarinPal)
	})
}

// confirmRenewalHandler confirms a paid invoice and issues a
// new license.
//
// POST /api/v1/license/renew/confirm
// Body: {"invoice_id": "...", "payment_ref": "..."}
func confirmRenewalHandler(c *gin.Context) {
	var req struct {
		InvoiceID  string `json:"invoice_id" binding:"required"`
		PaymentRef string `json:"payment_ref"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the invoice
	var inv Invoice
	if err := db.Where("invoice_id = ?", req.InvoiceID).First(&inv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	if inv.Status == "paid" {
		c.JSON(http.StatusConflict, gin.H{"error": "invoice already paid"})
		return
	}

	if inv.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice is not pending"})
		return
	}

	// Find the old license
	var oldLic License
	if err := db.Where("license_id = ?", inv.LicenseID).First(&oldLic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "original license not found"})
		return
	}

	// Check key
	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()

	if key == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is locked"})
		return
	}

	// Create new license
	now := time.Now().Unix()
	newLicID := nextLicenseID(oldLic.LicenseID)

	newLic := &License{
		LicenseID:  newLicID,
		UserID:     oldLic.UserID,
		ProductID:  oldLic.ProductID,
		Volume:     oldLic.Volume,
		Duration:   inv.DurationH,
		MerkleRoot: oldLic.MerkleRoot,
		IssuedAt:   now,
		ExpiresAt:  now + int64(inv.DurationH)*3600,
		Status:     "active",
		HardwareID: oldLic.HardwareID,
	}

	msg := licenseCanonicalMessage(newLic)
	sig, err := signData(key, []byte(msg))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign failed: " + err.Error()})
		return
	}
	newLic.Signature = sig

	// Mark old license as expired
	if err := db.Model(&License{}).
		Where("license_id = ?", oldLic.LicenseID).
		Update("status", "expired").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Save new license
	if err := db.Create(newLic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Mark invoice as paid
	if err := db.Model(&Invoice{}).
		Where("invoice_id = ?", inv.InvoiceID).
		Updates(map[string]interface{}{
			"status":      "paid",
			"payment_ref": req.PaymentRef,
			"paid_at":     now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	addAudit("renew_confirm", newLicID,
		fmt.Sprintf("invoice=%s payment=%s", inv.InvoiceID, req.PaymentRef),
		c.ClientIP())

	// Re-check license state immediately
	go checkLicenseNow()

	c.JSON(http.StatusOK, gin.H{
		"status":         "renewed",
		"old_license_id": oldLic.LicenseID,
		"new_license":    newLic,
		"invoice_id":     inv.InvoiceID,
	})
}

// renewalHistoryHandler returns all invoices for a license.
//
// GET /api/v1/license/renew/history?license_id=...
func renewalHistoryHandler(c *gin.Context) {
	licenseID := c.Query("license_id")

	var invoices []Invoice
	q := db.Order("created_at desc")
	if licenseID != "" {
		q = q.Where("license_id = ?", licenseID)
	}
	q.Find(&invoices)

	c.JSON(http.StatusOK, gin.H{
		"total":    len(invoices),
		"invoices": invoices,
	})
}

// licenseStatusHandler returns the current license with days
// remaining and grace status.
//
// GET /api/v1/license/status
func licenseStatusHandler(c *gin.Context) {
	lic := getActiveLicenseOrNil()
	if lic == nil {
		c.JSON(http.StatusOK, gin.H{
			"has_license": false,
			"message":     "no active license",
		})
		return
	}

	now := time.Now().Unix()
	daysRemaining := int((lic.ExpiresAt - now) / 86400)

	graceStatus := getGraceStatus(lic)

	c.JSON(http.StatusOK, gin.H{
		"has_license":     true,
		"license_id":      lic.LicenseID,
		"user_id":         lic.UserID,
		"product_id":      lic.ProductID,
		"volume":          lic.Volume,
		"used":            lic.Used,
		"issued_at":       lic.IssuedAt,
		"expires_at":      lic.ExpiresAt,
		"status":          lic.Status,
		"hardware_id":     lic.HardwareID,
		"days_remaining":  daysRemaining,
		"grace":           graceStatus,
	})
}

// ============================================================
// SHARED: issue a new license for a paid invoice
// ============================================================

// issueRenewalLicense creates and signs a new license for a
// paid invoice, and marks the old license as expired.
//
// Used by both confirmRenewalHandler and mockConfirmHandler.
func issueRenewalLicense(inv *Invoice) (*License, error) {
	if inv == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	var oldLic License
	if err := db.Where("license_id = ?", inv.LicenseID).First(&oldLic).Error; err != nil {
		return nil, fmt.Errorf("original license not found: %w", err)
	}

	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()

	if key == nil {
		return nil, fmt.Errorf("key is locked")
	}

	now := time.Now().Unix()
	newLicID := nextLicenseID(oldLic.LicenseID)

	newLic := &License{
		LicenseID:  newLicID,
		UserID:     oldLic.UserID,
		ProductID:  oldLic.ProductID,
		Volume:     oldLic.Volume,
		Duration:   inv.DurationH,
		MerkleRoot: oldLic.MerkleRoot,
		IssuedAt:   now,
		ExpiresAt:  now + int64(inv.DurationH)*3600,
		Status:     "active",
		HardwareID: oldLic.HardwareID,
	}

	msg := licenseCanonicalMessage(newLic)
	sig, err := signData(key, []byte(msg))
	if err != nil {
		return nil, fmt.Errorf("sign failed: %w", err)
	}
	newLic.Signature = sig

	// Mark old license as expired
	if err := db.Model(&License{}).
		Where("license_id = ?", oldLic.LicenseID).
		Update("status", "expired").Error; err != nil {
		return nil, fmt.Errorf("update old license: %w", err)
	}

	// Save new license
	if err := db.Create(newLic).Error; err != nil {
		return nil, fmt.Errorf("save new license: %w", err)
	}

	// Re-check license state immediately
	go checkLicenseNow()

	return newLic, nil
}
