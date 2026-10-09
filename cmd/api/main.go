package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"

	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
	"horizon-core/internal/license"
	"horizon-core/internal/network"
)

type Customer struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Wallet    string    `json:"wallet"`
	CreatedAt time.Time `json:"created_at"`
}
type Payment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CustomerID uint      `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	TxHash     string    `json:"tx_hash"`
	CreatedAt  time.Time `json:"created_at"`
}
type LicenseRecord struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	Volume     int        `json:"volume"`
	Used       int        `json:"used"`
	ProductID  string     `json:"product_id"`
	UserID     string     `json:"user_id"`
	Root       string     `json:"root"`
	Seed       string     `json:"seed"`
	PrevHash   string     `json:"prev_hash"`
	Hash       string     `json:"hash"`
	Signature  string     `json:"signature"`
	PublicKey  string     `json:"public_key"`
	MerkleRoot string     `json:"merkle_root"`
	IssuedAt   int64      `json:"issued_at"`
	Duration   int        `json:"duration"`
	HardwareID string     `json:"hardware_id"`
	CreatedAt  time.Time  `json:"created_at"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	LicenseID string    `json:"license_id"`
	CreatedAt time.Time `json:"created_at"`
}
type LicensePackage struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	CustomerType string    `json:"customer_type"`
	CustomerID   string    `json:"customer_id"`
	Volume       int       `json:"volume"`
	Duration     int       `json:"duration"`
	TotalPrice   float64   `json:"total_price"`
	PricePerTx   float64   `json:"price_per_tx"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
type IndustrialSensor struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Unit      string    `json:"unit"`
	MinValue  float64   `json:"min_value"`
	MaxValue  float64   `json:"max_value"`
	Location  string    `json:"location"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
type SensorReading struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	SensorID   uint       `json:"sensor_id"`
	Value      float64    `json:"value"`
	RecordedAt time.Time  `json:"recorded_at"`
	IsOffline  bool       `json:"is_offline"`
	SyncedAt   *time.Time `json:"synced_at"`
}
type IndustryAlert struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SensorID     uint      `json:"sensor_id"`
	Value        float64   `json:"value"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	Acknowledged bool      `json:"acknowledged"`
	CreatedAt    time.Time `json:"created_at"`
}
type StartToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Token     string    `json:"token"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}
type Gateway struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
type Wallet struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Network   string    `json:"network"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

var pgDB *gorm.DB
var requestMu sync.Mutex

func switchEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_SWITCH")))
	return v != "false" && v != "0" && v != "no"
}

func adminTokenValid(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("ADMIN_TOKEN"))
	return len(expected) >= 32 && c.GetHeader("X-Admin-Key") == expected
}

func apiKeyValid(c *gin.Context) bool {
	expected := strings.TrimSpace(os.Getenv("API_KEY"))
	return len(expected) >= 32 && c.GetHeader("X-API-Key") == expected
}

func requireAdmin(c *gin.Context) bool {
	if !adminTokenValid(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin key"})
		return false
	}
	return true
}

func connectPostgres() error {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Println("DATABASE_URL not set - falling back to SQLite at ./horizon-dev.db")
		pgDB, err = gorm.Open(sqlite.Open("horizon-dev.db"), &gorm.Config{})
	} else {
		pgDB, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	}
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	return nil
}

func statementFromRecord(l LicenseRecord) license.Statement {
	var expiresAt int64
	if l.ExpiresAt != nil {
		expiresAt = l.ExpiresAt.Unix()
	}
	return license.Statement{
		LicenseID: l.ID, UserID: l.UserID, ProductID: l.ProductID,
		Volume: l.Volume, Duration: l.Duration, MerkleRoot: l.MerkleRoot,
		IssuedAt: l.IssuedAt, ExpiresAt: expiresAt, HardwareID: l.HardwareID,
	}
}

func canonicalLicenseData(l LicenseRecord) []byte {
	data, _ := license.CanonicalBytes(statementFromRecord(l))
	return data
}

func errorMessage(err error) string {
	if err == nil { return "" }
	return err.Error()
}
func timeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func callSwitch(path string, payload any) (map[string]any, int, error) {
	base := strings.TrimRight(os.Getenv("SWITCH_URL"), "/")
	if base == "" {
		return nil, 0, errors.New("SWITCH_URL is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(http.MethodPost, base+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(os.Getenv("ADMIN_TOKEN")); token != "" {
		req.Header.Set("X-Admin-Token", token)
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, err
	}
	return result, resp.StatusCode, nil
}

func mustReadFile(path string) []byte { b, _ := os.ReadFile(path); return b }

func main() {
	// ─── P0.4 FAIL-CLOSED: refuse to start without required secrets ───
	adminToken := strings.TrimSpace(os.Getenv("ADMIN_TOKEN"))
	apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
	if len(adminToken) < 32 { log.Fatal("P0.4 FAIL-CLOSED: ADMIN_TOKEN must be set to at least 32 characters.") }
	if len(apiKey) < 32 { log.Fatal("P0.4 FAIL-CLOSED: API_KEY must be set to at least 32 characters.") }
	log.Printf("P0.4: independent admin/API authentication enabled")
	// ──────────────────────────────────────────────────────────────

	if err := db.InitDB(); err != nil {
		panic(err)
	}
	if err := connectPostgres(); err != nil {
		panic(err)
	}

	if err := pgDB.AutoMigrate(&LicenseRecord{}, &Customer{}, &Payment{}, &Transaction{}, &LicensePackage{},
		&IndustrialSensor{}, &SensorReading{}, &IndustryAlert{}, &StartToken{}, &Gateway{}, &Wallet{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{AllowOrigins: []string{"*"}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key", "X-License-Key"}, AllowCredentials: false, MaxAge: 12 * time.Hour}))

	r.GET("/api/v1/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "online", "service": "horizon-backend"}) })

	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		if !apiKeyValid(c) && !adminTokenValid(c) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		var req struct {
			ProductID  string `json:"product_id"`
			UserID     string `json:"user_id"`
			Volume     int    `json:"volume"`
			Duration   int    `json:"duration"`
			HardwareID string `json:"hardware_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Volume <= 0 || req.ProductID == "" ||
			req.UserID == "" || req.Duration <= 0 || strings.TrimSpace(req.HardwareID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id, user_id, positive volume/duration, and hardware_id are required"})
			return
		}
		if req.Duration > 36500 || req.Volume > 1000000000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "volume or duration exceeds policy limits"})
			return
		}
		pkg := license.NewPrepaidPackage(req.Volume)
		if err := pkg.Generate(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "license package generation failed"})
			return
		}
		pub, err := db.PublicKeyPEM()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authority public key unavailable"})
			return
		}
		now := time.Now().UTC().Truncate(time.Second)
		expires := time.Unix(now.Unix()+int64(req.Duration)*86400, 0).UTC()
		lic := LicenseRecord{
			ID: fmt.Sprintf("HZ14-%d", now.UnixNano()), Volume: req.Volume, ProductID: strings.TrimSpace(req.ProductID),
			UserID: strings.TrimSpace(req.UserID), Root: pkg.RootHex(), Seed: pkg.SeedHex(), CreatedAt: now,
			IssuedAt: now.Unix(), Duration: req.Duration, HardwareID: strings.TrimSpace(req.HardwareID),
			Status: "active", PublicKey: pub, ExpiresAt: &expires,
		}
		statement := statementFromRecord(lic)
		statement.MerkleRoot = license.MerkleRoot(statement)
		lic.MerkleRoot = statement.MerkleRoot
		message, err := license.CanonicalBytes(statement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "canonical license encoding failed"})
			return
		}
		signature, err := crypto.SignData(message, db.PrivateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authority signing failed"})
			return
		}
		lic.Signature = signature
		if err := pgDB.Create(&lic).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "license persistence failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"license": license.SignedLicense{Statement: statement, Signature: signature},
			"package_root": lic.Root, "seed": lic.Seed, "issuer_public_key": lic.PublicKey,
			"status": lic.Status,
		})
	})

	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		var req struct { LicenseID string `json:"license_id" binding:"required"` }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"verified": false, "valid": false, "error": "license_id is required"})
			return
		}
		var lic LicenseRecord
		if err := pgDB.First(&lic, "id = ?", req.LicenseID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"verified": false, "valid": false, "error": "license not found"})
			return
		}
		if lic.Status != "active" || lic.ExpiresAt == nil {
			c.JSON(http.StatusOK, gin.H{"verified": false, "valid": false, "license_id": lic.ID, "status": lic.Status})
			return
		}
		if db.PrivateKey == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"verified": false, "valid": false, "error": "Authority key unavailable"})
			return
		}
		err := license.Validate(license.SignedLicense{Statement: statementFromRecord(lic), Signature: lic.Signature},
			&db.PrivateKey.PublicKey, time.Now(), "")
		c.JSON(http.StatusOK, gin.H{"verified": err == nil, "valid": err == nil, "license_id": lic.ID,
			"status": lic.Status, "error": errorMessage(err), "merkle_root": lic.MerkleRoot})
	})

	r.GET("/api/v1/licenses", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var items []LicenseRecord
		if err := pgDB.Order("created_at desc").Find(&items).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, items)
	})
	r.POST("/api/v1/license/suspend-all", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		res := pgDB.Model(&LicenseRecord{}).Where("status <> ?", "suspended").Update("status", "suspended")
		if res.Error != nil {
			c.JSON(500, gin.H{"error": res.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"updated": res.RowsAffected, "status": "suspended"})
	})
	r.POST("/api/v1/license/renew-all", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		res := pgDB.Model(&LicenseRecord{}).Where("status = ?", "suspended").Update("status", "active")
		if res.Error != nil {
			c.JSON(500, gin.H{"error": res.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"updated": res.RowsAffected, "status": "active"})
	})

	r.POST("/api/v1/operation/execute", func(c *gin.Context) {
		licenseID := c.GetHeader("X-License-Key")
		if licenseID == "" {
			c.JSON(401, gin.H{"error": "missing license key"})
			return
		}
		var tx Transaction
		if err := c.ShouldBindJSON(&tx); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		var lic LicenseRecord
		if err := pgDB.First(&lic, "id = ?", licenseID).Error; err != nil {
			c.JSON(404, gin.H{"error": "license not found"})
			return
		}
		if lic.Status != "active" || lic.ExpiresAt == nil || !lic.ExpiresAt.After(time.Now().UTC()) {
			c.JSON(403, gin.H{"error": "license is not active"})
			return
		}
		if switchEnabled() {
			result, status, err := callSwitch("/api/v1/license/check", license.SignedLicense{Statement: statementFromRecord(lic), Signature: lic.Signature})
			if err != nil {
				c.JSON(502, gin.H{"error": "switch unavailable"})
				return
			}
			valid, _ := result["valid"].(bool)
			if status < 200 || status >= 300 || !valid {
				c.JSON(http.StatusForbidden, gin.H{"error": "Switch rejected license", "reason": result["error"]})
				return
			}
		}
		requestMu.Lock()
		defer requestMu.Unlock()
		if lic.Used >= lic.Volume {
			c.JSON(403, gin.H{"error": "license quota exhausted"})
			return
		}
		lic.Used++
		if err := pgDB.Save(&lic).Error; err != nil {
			c.JSON(500, gin.H{"error": "quota update failed"})
			return
		}
		tx.LicenseID = licenseID
		if err := pgDB.Create(&tx).Error; err != nil {
			c.JSON(500, gin.H{"error": "transaction persistence failed"})
			return
		}
		c.JSON(200, gin.H{"message": "operation executed", "remaining": lic.Volume - lic.Used, "transaction_id": tx.ID})
	})

	r.POST("/api/v1/toolbox/encrypt", func(c *gin.Context) {
		if !apiKeyValid(c) {
			c.JSON(401, gin.H{"error": "invalid API key"})
			return
		}
		var req struct{ Key, Plaintext string }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		key, err := hex.DecodeString(req.Key)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid key"})
			return
		}
		out, err := crypto.EncryptAES(key, req.Plaintext)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ciphertext": out})
	})
	r.POST("/api/v1/toolbox/decrypt", func(c *gin.Context) {
		if !apiKeyValid(c) {
			c.JSON(401, gin.H{"error": "invalid API key"})
			return
		}
		var req struct{ Key, Ciphertext string }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		key, err := hex.DecodeString(req.Key)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid key"})
			return
		}
		out, err := crypto.DecryptAES(key, req.Ciphertext)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"plaintext": out})
	})
	r.POST("/api/v1/toolbox/subnet", func(c *gin.Context) {
		var req struct {
			CIDR string `json:"cidr"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		out, err := network.CalculateSubnet(req.CIDR)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, out)
	})
	r.POST("/api/v1/toolbox/mac", func(c *gin.Context) {
		var req struct {
			MAC string `json:"mac"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		c.JSON(200, gin.H{"vendor": network.LookupOUI(req.MAC)})
	})

	r.POST("/api/v1/customers", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v Customer
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := pgDB.Create(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, v)
	})
	r.GET("/api/v1/customers", func(c *gin.Context) {
		var v []Customer
		if err := pgDB.Find(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, v)
	})
	r.POST("/api/v1/payments", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v Payment
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := pgDB.Create(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, v)
	})
	r.GET("/api/v1/payments", func(c *gin.Context) {
		var v []Payment
		if err := pgDB.Find(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, v)
	})

	r.POST("/api/v1/packages", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v LicensePackage
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := pgDB.Create(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, v)
	})
	r.GET("/api/v1/packages", func(c *gin.Context) {
		var v []LicensePackage
		if err := pgDB.Find(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, v)
	})
	r.PUT("/api/v1/packages/:id", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v LicensePackage
		if err := pgDB.First(&v, c.Param("id")).Error; err != nil {
			c.JSON(404, gin.H{"error": "package not found"})
			return
		}
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := pgDB.Save(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, v)
	})
	r.DELETE("/api/v1/packages/:id", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		if err := pgDB.Delete(&LicensePackage{}, c.Param("id")).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Status(204)
	})

	r.GET("/api/v1/gateways", func(c *gin.Context) { var v []Gateway; pgDB.Find(&v); c.JSON(200, v) })
	r.POST("/api/v1/gateways", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v Gateway
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		pgDB.Create(&v)
		c.JSON(201, v)
	})
	r.GET("/api/v1/wallets", func(c *gin.Context) { var v []Wallet; pgDB.Find(&v); c.JSON(200, v) })
	r.POST("/api/v1/wallets", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v Wallet
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		pgDB.Create(&v)
		c.JSON(201, v)
	})
	r.POST("/api/v1/purchase", func(c *gin.Context) {
		if !apiKeyValid(c) {
			c.JSON(401, gin.H{"error": "invalid API key"})
			return
		}
		c.JSON(501, gin.H{"error": "purchase gateway integration requires configured provider adapters"})
	})
	r.POST("/api/v1/ai/ask", func(c *gin.Context) {
		if !apiKeyValid(c) {
			c.JSON(401, gin.H{"error": "invalid API key"})
			return
		}
		base := strings.TrimRight(os.Getenv("AI_BASE_URL"), "/")
		key := os.Getenv("AI_API_KEY")
		model := os.Getenv("AI_MODEL_ID")
		if base == "" || key == "" || model == "" {
			c.JSON(503, gin.H{"error": "AI service is not configured"})
			return
		}
		var reqBody map[string]any
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		reqBody["model"] = model
		b, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", base+"/chat/completions", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+key)
		resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
		if err != nil {
			c.JSON(502, gin.H{"error": "AI upstream unavailable"})
			return
		}
		defer resp.Body.Close()
		c.Status(resp.StatusCode)
		var out any
		json.NewDecoder(resp.Body).Decode(&out)
		c.JSON(resp.StatusCode, out)
	})
	r.POST("/api/v1/start", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v StartToken
		v.Token = fmt.Sprintf("start_%d", time.Now().UnixNano())
		if err := pgDB.Create(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, v)
	})
	r.GET("/api/v1/industry/sensors", func(c *gin.Context) { var v []IndustrialSensor; pgDB.Find(&v); c.JSON(200, v) })
	r.POST("/api/v1/industry/sensors", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v IndustrialSensor
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		pgDB.Create(&v)
		c.JSON(201, v)
	})
	r.GET("/api/v1/industry/readings", func(c *gin.Context) {
		var v []SensorReading
		q := pgDB
		if id := c.Query("sensor_id"); id != "" {
			q = q.Where("sensor_id = ?", id)
		}
		q.Find(&v)
		c.JSON(200, v)
	})
	r.POST("/api/v1/industry/readings/batch", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var v []SensorReading
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if err := pgDB.Create(&v).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(201, v)
	})
	r.GET("/api/v1/industry/alerts", func(c *gin.Context) { var v []IndustryAlert; pgDB.Order("created_at desc").Find(&v); c.JSON(200, v) })
	r.POST("/api/v1/industry/alerts/ack", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var req struct {
			ID uint `json:"id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		res := pgDB.Model(&IndustryAlert{}).Where("id = ?", req.ID).Update("acknowledged", true)
		c.JSON(200, gin.H{"updated": res.RowsAffected})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

var _ *ecdsa.PublicKey
