package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// CUSTOMER MODEL
// ============================================================

type Customer struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CustomerID   string    `gorm:"uniqueIndex" json:"customer_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type CustomerSeed struct {
	CustomerID string `json:"customer_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Password   string `json:"password"`
}

// ============================================================
// CUSTOMER SESSIONS
// ============================================================

type customerSession struct {
	CustomerID string
	ExpiresAt  time.Time
}

var (
	customerSessions   = make(map[string]customerSession)
	customerSessionsMu sync.RWMutex
)

func createCustomerSession(customerID string) string {
	b := make([]byte, 16)
	rand.Read(b)
	tok := hex.EncodeToString(b)

	customerSessionsMu.Lock()
	customerSessions[tok] = customerSession{
		CustomerID: customerID,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	customerSessionsMu.Unlock()

	return tok
}

func validateCustomerSession(tok string) (string, bool) {
	customerSessionsMu.RLock()
	defer customerSessionsMu.RUnlock()

	s, ok := customerSessions[tok]
	if !ok {
		return "", false
	}
	if time.Now().After(s.ExpiresAt) {
		return "", false
	}
	return s.CustomerID, true
}

// ============================================================
// INIT
// ============================================================

func initCustomers() {
	var count int64
	db.Model(&Customer{}).Count(&count)
	if count > 0 {
		log.Printf("customers: %d already exist", count)
		return
	}

	path := os.Getenv("CUSTOMERS_CONFIG")
	if path == "" {
		path = "./config/customers.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("customers: no config file (%s), skipping seed", path)
		return
	}

	var seeds []CustomerSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		log.Printf("customers: parse error: %v", err)
		return
	}

	for _, s := range seeds {
		if s.CustomerID == "" || s.Password == "" {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(s.Password), 10)
		if err != nil {
			continue
		}
		ctype := s.Type
		if ctype == "" {
			ctype = "bank"
		}
		if err := db.Create(&Customer{
			CustomerID:   s.CustomerID,
			Name:         s.Name,
			Type:         ctype,
			PasswordHash: string(hash),
		}).Error; err != nil {
			log.Printf("customers: failed to seed %s: %v", s.CustomerID, err)
			continue
		}
		log.Printf("customers: seeded %s", s.CustomerID)
	}

	log.Printf("customers: seeded %d customers", len(seeds))
}

// ============================================================
// MIDDLEWARE
// ============================================================

func customerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := c.GetHeader("X-Customer-Token")
		if tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "no token",
			})
			return
		}
		cid, ok := validateCustomerSession(tok)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}
		c.Set("customer_id", cid)
		c.Next()
	}
}

// ============================================================
// HANDLERS
// ============================================================

// POST /api/v1/customer/login
func customerLoginHandler(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cust Customer
	if err := db.Where("customer_id = ?", req.CustomerID).First(&cust).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "customer not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(cust.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}

	tok := createCustomerSession(cust.CustomerID)

	addAudit("customer_login", cust.CustomerID, "login ok", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"token":       tok,
		"customer_id": cust.CustomerID,
		"name":        cust.Name,
		"type":        cust.Type,
	})
}

// POST /api/v1/customer/logout
func customerLogoutHandler(c *gin.Context) {
	tok := c.GetHeader("X-Customer-Token")
	customerSessionsMu.Lock()
	delete(customerSessions, tok)
	customerSessionsMu.Unlock()
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// GET /api/v1/customer/me
func customerMeHandler(c *gin.Context) {
	cid := c.GetString("customer_id")

	var cust Customer
	if err := db.Where("customer_id = ?", cid).First(&cust).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id": cust.CustomerID,
		"name":        cust.Name,
		"type":        cust.Type,
	})
}

// GET /api/v1/customer/licenses
func customerLicensesHandler(c *gin.Context) {
	cid := c.GetString("customer_id")

	var licenses []License
	db.Where("user_id = ?", cid).Order("issued_at desc").Find(&licenses)

	type licenseView struct {
		License
		DaysRemaining int         `json:"days_remaining"`
		IsValid       bool        `json:"is_valid"`
		Grace         GraceStatus `json:"grace"`
	}

	now := time.Now().Unix()
	out := make([]licenseView, 0, len(licenses))

	for _, lic := range licenses {
		days := int((lic.ExpiresAt - now) / 86400)
		g := getGraceStatus(&lic)
		valid := lic.Status == "active" && !g.IsReadOnly

		out = append(out, licenseView{
			License:       lic,
			DaysRemaining: days,
			IsValid:       valid,
			Grace:         g,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"customer_id": cid,
		"total":       len(out),
		"licenses":    out,
	})
}

// GET /api/v1/customer/invoices
func customerInvoicesHandler(c *gin.Context) {
	cid := c.GetString("customer_id")

	var invoices []Invoice
	db.Where("user_id = ?", cid).Order("created_at desc").Find(&invoices)

	c.JSON(http.StatusOK, gin.H{
		"customer_id": cid,
		"total":       len(invoices),
		"invoices":    invoices,
	})
}

// GET /api/v1/customer/license/download/:license_id
func customerDownloadLicenseHandler(c *gin.Context) {
	cid := c.GetString("customer_id")
	licenseID := c.Param("license_id")

	var lic License
	if err := db.Where(
		"license_id = ? AND user_id = ?",
		licenseID, cid,
	).First(&lic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}

	chain := GetChainConfig()

	download := gin.H{
		"version":     "1.0",
		"chain_id":    chain.ChainID,
		"license_id":  lic.LicenseID,
		"user_id":     lic.UserID,
		"product_id":  lic.ProductID,
		"volume":      lic.Volume,
		"duration":    lic.Duration,
		"merkle_root": lic.MerkleRoot,
		"issued_at":   lic.IssuedAt,
		"expires_at":  lic.ExpiresAt,
		"status":      lic.Status,
		"hardware_id": lic.HardwareID,
		"issued_by":   chain.AdminPublicKey,
		"signature":   lic.Signature,
	}

	c.Header("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s.lic\"", lic.LicenseID))
	c.JSON(http.StatusOK, download)
}

// POST /api/v1/customer/license/renew/:license_id
func customerRenewHandler(c *gin.Context) {
	cid := c.GetString("customer_id")
	licenseID := c.Param("license_id")

	var req struct {
		DurationHours int `json:"duration_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.DurationHours = 8760
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 8760
	}

	// Verify ownership
	var lic License
	if err := db.Where(
		"license_id = ? AND user_id = ?",
		licenseID, cid,
	).First(&lic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}

	// Create invoice
	inv := &Invoice{
		InvoiceID: generateInvoiceID(),
		LicenseID: licenseID,
		UserID:    cid,
		AmountIRR: priceForDuration(req.DurationHours),
		DurationH: req.DurationHours,
		Status:    "pending",
	}
	if err := db.Create(inv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-initiate mock payment
	authority := generateAuthority()
	db.Model(&Invoice{}).
		Where("invoice_id = ?", inv.InvoiceID).
		Update("payment_ref", authority)

	redirectURL := fmt.Sprintf(
		"/payment/mock?authority=%s&invoice_id=%s",
		authority, inv.InvoiceID,
	)

	addAudit("customer_renew", licenseID,
		fmt.Sprintf("invoice=%s", inv.InvoiceID),
		c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"status":       "ok",
		"invoice_id":   inv.InvoiceID,
		"amount_toman": inv.AmountIRR,
		"authority":    authority,
		"redirect_url": redirectURL,
	})
}
