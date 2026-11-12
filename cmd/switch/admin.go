package main

import (
	"encoding/json"
	"horizon-core/internal/clock"
	"fmt"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// ADMIN SESSION STORE (in-memory)
// ============================================================

var (
	adminSessions = make(map[string]time.Time)
	adminMutex    sync.RWMutex

	auditLog   []AuditEntry
	auditMutex sync.Mutex
)

type AuditEntry struct {
	ID        int64  `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Details   string `json:"details"`
	IP        string `json:"ip"`
}

func sessionTTL() int {
	s := os.Getenv("ADMIN_SESSION_TTL")
	if s == "" {
		return 3600
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return 3600
	}
	return v
}

func checkAdminPassword(input string) bool {
	envHash := os.Getenv("ADMIN_PASSWORD_HASH")
	if envHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(envHash), []byte(input)) == nil
}

func createAdminSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	tok := hex.EncodeToString(b)
	adminMutex.Lock()
	adminSessions[tok] = time.Now().Add(time.Duration(sessionTTL()) * time.Second)
	adminMutex.Unlock()
	return tok
}

func validateAdminSession(tok string) bool {
	adminMutex.RLock()
	exp, ok := adminSessions[tok]
	adminMutex.RUnlock()
	return ok && time.Now().Before(exp)
}

func revokeAdminSession(tok string) {
	adminMutex.Lock()
	delete(adminSessions, tok)
	adminMutex.Unlock()
}

// ============================================================
// AUDIT
// ============================================================

func addAudit(action, target, details, ip string) {
	auditMutex.Lock()
	defer auditMutex.Unlock()
	auditLog = append(auditLog, AuditEntry{
		ID:        int64(len(auditLog) + 1),
		Timestamp: time.Now().Unix(),
		Action:    action,
		Target:    target,
		Details:   details,
		IP:        ip,
	})
	if len(auditLog) > 500 {
		auditLog = auditLog[len(auditLog)-500:]
	}
}

// ============================================================
// MIDDLEWARE
// ============================================================

func adminAuthRequired() gin.HandlerFunc {
	envToken := os.Getenv("ADMIN_TOKEN")
	return func(c *gin.Context) {
		tok := c.GetHeader("X-Admin-Token")
		if tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token"})
			return
		}

		// Check 1: env ADMIN_TOKEN (master token)
		isValid := false
		if envToken != "" && tok == envToken {
			isValid = true
		}
		// Check 2: session token (from adminLoginHandler)
		if !isValid && validateAdminSession(tok) {
			isValid = true
		}

		if !isValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		stateMutex.RLock()
		key := unlockedKey
		stateMutex.RUnlock()
		if key == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "key is locked — call /api/v1/key/unlock first"})
			return
		}
		c.Next()
	}
}

// ============================================================
// HANDLERS
// ============================================================

func adminLoginHandler(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !checkAdminPassword(req.Password) {
		addAudit("login_fail", "-", "wrong password", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
		return
	}
	tok := createAdminSession()
	addAudit("login", "-", "admin logged in", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"token":      tok,
		"expires_in": sessionTTL(),
	})
}

func adminLogoutHandler(c *gin.Context) {
	tok := c.GetHeader("X-Admin-Token")
	revokeAdminSession(tok)
	addAudit("logout", "-", "admin logged out", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

func listBanksHandler(c *gin.Context) {
	var banks []BankAccount
	db.Order("created_at desc").Find(&banks)

	type bankWithLicense struct {
		BankAccount
		Balance float64  `json:"balance"`
		License *License `json:"license,omitempty"`
	}

	out := make([]bankWithLicense, 0, len(banks))
	for _, b := range banks {
		var lic License
		var ptr *License
		if err := db.Where("user_id = ?", b.BankID).
			Order("created_at desc").
			First(&lic).Error; err == nil {
			ptr = &lic
		}

		var acc Account
		db.Where("bank_id = ?", b.BankID).First(&acc)

		out = append(out, bankWithLicense{
			BankAccount: b,
			Balance:     acc.Balance,
			License:     ptr,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total": len(out),
		"banks": out,
	})
}

func adminAuditHandler(c *gin.Context) {
	auditMutex.Lock()
	defer auditMutex.Unlock()
	out := make([]AuditEntry, len(auditLog))
	for i, e := range auditLog {
		out[len(auditLog)-1-i] = e
	}
	c.JSON(http.StatusOK, gin.H{
		"total":   len(out),
		"entries": out,
	})
}

func adminIssueLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID  string `json:"license_id" binding:"required"`
		UserID     string `json:"user_id" binding:"required"`
		ProductID  string `json:"product_id"`
		Volume     int    `json:"volume"`
		Duration   int    `json:"duration"`
		LicenseType string `json:"license_type"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "key locked"})
		return
	}

	if req.ProductID == "" {
		req.ProductID = "horizon-core"
	}
	if req.Duration == 0 {
		req.Duration = 8760
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
		ExpiresAt:  now + int64(req.Duration*86400),
		Status:     "pending",
			LicenseType: req.LicenseType,
	}

	if lic.MerkleRoot == "" {
		lic.MerkleRoot = licenseMerkleRoot(lic)
	}

	msg := licenseCanonicalMessage(lic)
	sig, err := signData(key, []byte(msg))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign failed: " + err.Error()})
		return
	}
	lic.Signature = sig

	if err := db.Create(lic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed: " + err.Error()})
		return
	}

	// append license to blockchain chain
	if chainTx, _, err := appendChainTx(db, key, TxLicenseIssue, lic); err != nil {
		log.Printf("⚠️  chain: license append failed: %v", err)
	} else {
		log.Printf("⛓  license %s appended in block #%d", lic.LicenseID, chainTx.BlockNum)
	}

	// ثبت لایسنس به عنوان یک تراکنش در صف اجماع
	tx := &Transaction{
		TxID:       fmt.Sprintf("lic_%s_%d", lic.LicenseID, time.Now().UnixNano()),
		From:       "system",
		To:         lic.UserID,
		Amount:     0,
		Type:       TxLicenseIssue,
		Timestamp:  time.Now().Unix(),
		BlockIndex: -1,
		Synced:     false,
	}
	if err := db.Create(tx).Error; err != nil {
		log.Printf("⚠️  failed to queue license tx: %v", err)
	} else {
		log.Printf("⛓  license %s queued as tx (awaiting 8-node consensus)", lic.LicenseID)
	}

		addAudit("issue_license", lic.LicenseID, "for user "+lic.UserID, c.ClientIP())

	// refresh license state in memory (so isLicenseValid becomes true)
	checkLicenseNow()

	payloadBytes, _ := json.Marshal(lic)
	queueTransaction(&Transaction{
		TxID:      "lic_mint_" + lic.LicenseID,
		From:      "system",
		To:        "network",
		Amount:    0,
		Type:      "license_mint",
		LicenseID: lic.LicenseID,
		Payload:   string(payloadBytes),
		Timestamp: clock.Now(),
	})

	

	c.JSON(http.StatusCreated, lic)
}

func revokeLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID string `json:"license_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := db.Model(&License{}).
		Where("license_id = ?", req.LicenseID).
		Update("status", "revoked")
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}
	// append revoke to blockchain chain
	stateMutex.RLock()
	revokeKey := unlockedKey
	stateMutex.RUnlock()
	if revokeKey != nil {
		revokePayload := map[string]interface{}{
			"license_id": req.LicenseID,
			"action":     "revoke",
		}
		if chainTx, _, err := appendChainTx(db, revokeKey, TxLicenseRevoke, revokePayload); err != nil {
			log.Printf("⚠️  chain: revoke append failed: %v", err)
		} else {
			log.Printf("⛓  revoke appended in block #%d", chainTx.BlockNum)
		}
	}

	addAudit("revoke_license", req.LicenseID, "license revoked", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"status":     "revoked",
		"license_id": req.LicenseID,
	})
}
