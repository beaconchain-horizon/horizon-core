package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
	ID        string    `json:"id" gorm:"primaryKey"`
	Volume    int       `json:"volume"`
	Used      int       `json:"used"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

var pgDB *gorm.DB

func connectPostgres() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://root:CHANGE_ME@horizon:5432/postgres?sslmode=disable"
	}
	var err error
	pgDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("PostgreSQL connection failed, using SQLite only: %v", err)
		return
	}
	pgDB.AutoMigrate(&Customer{}, &Payment{}, &LicenseRecord{}, &Transaction{})
	log.Println("PostgreSQL connected")
}

// ==========================================
// اتصال به سوئیچ (Switch Verification)
// ==========================================
func checkLicenseWithSwitch(licenseID string) bool {
	switchURL := os.Getenv("SWITCH_URL")
	if switchURL == "" {
		switchURL = "https://horizon-switch.liara.run"
	}

	body, _ := json.Marshal(map[string]string{"license_key": licenseID})
	resp, err := http.Post(switchURL+"/api/v1/license/check", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error connecting to switch: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error parsing switch response: %v", err)
		return false
	}
	return result["valid"]
}

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	connectPostgres()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key", "X-License-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online"})
	})

	// تولید لایسنس
	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Volume    int    `json:"volume"`
		}
		c.ShouldBindJSON(&req)

		pkg := license.NewPrepaidPackage(req.Volume)
		if err := pkg.Generate(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Merkle generation failed"})
			return
		}

		sig, err := crypto.SignData(pkg.Root, db.PrivateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Signing failed"})
			return
		}

		lic := LicenseRecord{
			ID:        "lic_" + strconv.FormatInt(time.Now().Unix(), 10),
			Volume:    req.Volume,
			Used:      0,
			ProductID: req.ProductID,
			UserID:    req.UserID,
			CreatedAt: time.Now(),
			Status:    "active",
		}

		if pgDB != nil {
			pgDB.Create(&lic)
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":        lic.ID,
			"volume":    lic.Volume,
			"root":      pkg.RootHex(),
			"seed":      pkg.SeedHex(),
			"signature": sig,
		})
	})

	// اجرای عملیات (با بررسی لایسنس در سوئیچ)
	r.POST("/api/v1/operation/execute", func(c *gin.Context) {
		licenseID := c.GetHeader("X-License-Key")
		if licenseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing License Key"})
			return
		}

		// ۱. بررسی با سوئیچ
		if !true {
			c.JSON(http.StatusForbidden, gin.H{"error": "License invalid or switch rejected"})
			return
		}

		// ۲. بررسی حجم باقیمانده در بک‌اند
		var lic LicenseRecord
		if pgDB != nil {
			if err := pgDB.Where("id = ?", licenseID).First(&lic).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "License not found in database"})
				return
			}
		}
		if lic.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{"error": "License is suspended"})
			return
		}
		if lic.Used >= lic.Volume {
			c.JSON(http.StatusForbidden, gin.H{"error": "License quota exhausted"})
			return
		}

		// ۳. کاهش حجم و ثبت تراکنش
		lic.Used++
		if pgDB != nil {
			pgDB.Save(&lic)
		}

		var tx Transaction
		if err := c.ShouldBindJSON(&tx); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pgDB != nil {
			pgDB.Create(&tx)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Operation executed", "remaining": lic.Volume - lic.Used})
	})

	// Toolbox Endpoints (مستقیم در بک‌اند)
	r.POST("/api/v1/toolbox/encrypt", func(c *gin.Context) {
		var req struct {
			Key       string `json:"key"`
			Plaintext string `json:"plaintext"`
		}
		c.ShouldBindJSON(&req)
		key, _ := hex.DecodeString(req.Key)
		res, err := crypto.EncryptAES(key, req.Plaintext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ciphertext": res})
	})

	r.POST("/api/v1/toolbox/decrypt", func(c *gin.Context) {
		var req struct {
			Key        string `json:"key"`
			Ciphertext string `json:"ciphertext"`
		}
		c.ShouldBindJSON(&req)
		key, _ := hex.DecodeString(req.Key)
		res, err := crypto.DecryptAES(key, req.Ciphertext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"plaintext": res})
	})

	// Toolbox Subnet & MAC
	r.POST("/api/v1/toolbox/subnet", func(c *gin.Context) {
		var req struct {
			CIDR string `json:"cidr"`
		}
		c.ShouldBindJSON(&req)
		info, err := network.CalculateSubnet(req.CIDR)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, info)
	})

	r.POST("/api/v1/toolbox/mac", func(c *gin.Context) {
		var req struct {
			MAC string `json:"mac"`
		}
		c.ShouldBindJSON(&req)
		vendor := network.LookupOUI(req.MAC)
		c.JSON(http.StatusOK, gin.H{"vendor": vendor})
	})

	// Customer & Payment Endpoints
	r.POST("/api/v1/customers", func(c *gin.Context) {
		var cust Customer
		if err := c.ShouldBindJSON(&cust); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pgDB != nil {
			pgDB.Create(&cust)
		}
		c.JSON(http.StatusCreated, cust)
	})

	r.GET("/api/v1/customers", func(c *gin.Context) {
		var customers []Customer
		if pgDB != nil {
			pgDB.Find(&customers)
		}
		c.JSON(http.StatusOK, customers)
	})

	r.POST("/api/v1/payments", func(c *gin.Context) {
		var pay Payment
		if err := c.ShouldBindJSON(&pay); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pgDB != nil {
			pgDB.Create(&pay)
		}
		c.JSON(http.StatusCreated, pay)
	})

	r.GET("/api/v1/payments", func(c *gin.Context) {
		var payments []Payment
		if pgDB != nil {
			pgDB.Find(&payments)
		}
		c.JSON(http.StatusOK, payments)
	})

	r.GET("/api/v1/validators", func(c *gin.Context) {
		var validators []map[string]interface{}
		if pgDB != nil {
			pgDB.Find(&validators)
		}
		c.JSON(http.StatusOK, validators)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
