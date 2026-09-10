package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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
	ID        string     `json:"id" gorm:"primaryKey"`
	Volume    int        `json:"volume"`
	Used      int        `json:"used"`
	ProductID string     `json:"product_id"`
	UserID    string     `json:"user_id"`
	Root      string     `json:"root"`
	Seed      string     `json:"seed"`
	PrevHash  string     `json:"prev_hash"`
	Hash      string     `json:"hash"`
	Signature string     `json:"signature"`
	PublicKey string     `json:"public_key"`
	CreatedAt time.Time  `json:"created_at"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at"`
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
	expected := os.Getenv("ADMIN_TOKEN")
	if expected == "" {
		return true
	}
	return c.GetHeader("X-Admin-Key") == expected
}

func apiKeyValid(c *gin.Context) bool {
	expected := os.Getenv("API_KEY")
	if expected == "" {
		return true
	}
	return c.GetHeader("X-API-Key") == expected
}

func requireAdmin(c *gin.Context) bool {
	if !adminTokenValid(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin key"})
		return false
	}
	return true
}

func connectPostgres() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	var err error
	pgDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return pgDB.AutoMigrate(
		&Customer{}, &Payment{}, &LicenseRecord{}, &Transaction{}, &LicensePackage{},
		&IndustrialSensor{}, &SensorReading{}, &IndustryAlert{}, &StartToken{},
		&Gateway{}, &Wallet{},
	)
}

func canonicalLicenseData(l LicenseRecord) []byte {
	return []byte(fmt.Sprintf("%s|%s|%s|%d|%s|%s|%s", l.ID, l.ProductID, l.UserID, l.Volume, l.Root, l.Status, timeString(l.ExpiresAt)))
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
	if err := db.InitDB(); err != nil {
		panic(err)
	}
	if err := connectPostgres(); err != nil {
		panic(err)
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
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Volume    int    `json:"volume"`
			Duration  int    `json:"duration"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Volume <= 0 || req.ProductID == "" || req.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id, user_id and positive volume are required"})
			return
		}
		pkg := license.NewPrepaidPackage(req.Volume)
		if err := pkg.Generate(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		pub, _ := db.PublicKeyPEM()
		now := time.Now().UTC()
		var exp *time.Time
		if req.Duration > 0 {
			t := now.Add(time.Duration(req.Duration) * 24 * time.Hour)
			exp = &t
		}
		lic := LicenseRecord{ID: fmt.Sprintf("lic_%d_%d", now.UnixNano(), time.Now().Nanosecond()), Volume: req.Volume, ProductID: req.ProductID, UserID: req.UserID, Root: pkg.RootHex(), Seed: pkg.SeedHex(), CreatedAt: now, Status: "active", PublicKey: pub, ExpiresAt: exp}
		sig, err := crypto.SignData(canonicalLicenseData(lic), db.PrivateKey)
		if err != nil {
			c.JSON(500, gin.H{"error": "signing failed"})
			return
		}
		lic.Signature = sig
		if err := pgDB.Create(&lic).Error; err != nil {
			c.JSON(500, gin.H{"error": "license persistence failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": lic.ID, "volume": lic.Volume, "root": lic.Root, "seed": lic.Seed, "signature": lic.Signature, "public_key": lic.PublicKey, "expires_at": lic.ExpiresAt})
	})

	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		var req struct {
			LicenseID string `json:"license_id"`
			Signature string `json:"signature"`
			Data      string `json:"data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}
		var lic LicenseRecord
		if err := pgDB.First(&lic, "id = ?", req.LicenseID).Error; err != nil {
			c.JSON(404, gin.H{"verified": false, "error": "license not found"})
			return
		}
		pub, err := crypto.PEMToPublicKey(lic.PublicKey)
		if err != nil {
			c.JSON(500, gin.H{"verified": false, "error": "invalid public key"})
			return
		}
		data := []byte(req.Data)
		if len(data) == 0 {
			data = canonicalLicenseData(lic)
		}
		ok := crypto.VerifySignature(pub, data, req.Signature)
		c.JSON(http.StatusOK, gin.H{"verified": ok, "license_id": lic.ID, "status": lic.Status})
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
		if lic.Status != "active" || (lic.ExpiresAt != nil && lic.ExpiresAt.Before(time.Now().UTC())) {
			c.JSON(403, gin.H{"error": "license is not active"})
			return
		}
		if switchEnabled() {
			result, status, err := callSwitch("/api/v1/license/check", map[string]any{"license_id": lic.ID, "user_id": lic.UserID, "signature": lic.Signature, "data": string(canonicalLicenseData(lic))})
			if err != nil {
				c.JSON(502, gin.H{"error": "switch unavailable"})
				return
			}
			valid, _ := result["valid"].(bool)
			if !valid {
				verified, _ := result["verified"].(bool)
				if status >= 400 || !verified {
					c.JSON(403, gin.H{"error": "switch rejected license"})
					return
				}
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
	r.GET("/api/v1/validators", func(c *gin.Context) {
		var v []map[string]any
		if err := pgDB.Raw("SELECT 1 AS index, 'active' AS status").Scan(&v).Error; err != nil {
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
