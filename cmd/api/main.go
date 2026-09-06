package main

import (
	"encoding/hex"
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

type Validator struct {
	Index   int     `json:"index" gorm:"primaryKey"`
	Status  string  `json:"status"`
	Balance float64 `json:"balance"`
}

type License struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Volume    int       `json:"volume"`
	Root      string    `json:"root"`
	Seed      string    `json:"seed"`
	CreatedAt time.Time `json:"created_at"`
	Signature string    `json:"signature"`
	Status    string    `json:"status"`
	Expiry    time.Time `json:"expiry"`
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
	pgDB.AutoMigrate(&Customer{}, &Payment{}, &Validator{}, &License{}, &Transaction{})
	log.Println("PostgreSQL connected")
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
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online"})
	})

	// License Generation
	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Volume    int    `json:"volume"`
			Duration  int    `json:"duration"`
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

		lic := License{
			ID:        "lic_" + strconv.FormatInt(time.Now().Unix(), 10),
			Volume:    pkg.Volume,
			Root:      pkg.RootHex(),
			Seed:      pkg.SeedHex(),
			CreatedAt: time.Now(),
			Signature: sig,
			Status:    "active",
			Expiry:    time.Now().Add(time.Duration(req.Duration) * time.Hour),
		}
		if pgDB != nil {
			pgDB.Create(&lic)
		}
		c.JSON(http.StatusCreated, lic)
	})

	// List Licenses
	r.GET("/api/v1/licenses", func(c *gin.Context) {
		var licenses []License
		if pgDB != nil {
			pgDB.Find(&licenses)
		}
		c.JSON(http.StatusOK, licenses)
	})

	// Verify License
	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		var req struct {
			License string `json:"license"`
		}
		c.ShouldBindJSON(&req)
		var lic License
		if pgDB != nil {
			if err := pgDB.Where("id = ?", req.License).First(&lic).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "License not found"})
				return
			}
		}
		valid := lic.Status == "active" && time.Now().Before(lic.Expiry)
		c.JSON(http.StatusOK, gin.H{"valid": valid, "expiry": lic.Expiry})
	})

	// Suspend All
	r.POST("/api/v1/license/suspend-all", func(c *gin.Context) {
		if pgDB != nil {
			pgDB.Model(&License{}).Where("status = ?", "active").Update("status", "suspended")
		}
		c.JSON(http.StatusOK, gin.H{"message": "All licenses suspended"})
	})

	// Renew All
	r.POST("/api/v1/license/renew-all", func(c *gin.Context) {
		if pgDB != nil {
			pgDB.Model(&License{}).Update("status", "active")
		}
		c.JSON(http.StatusOK, gin.H{"message": "All licenses renewed"})
	})

	// Toolbox Endpoints
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

	// Customer Endpoints
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

	// Payment Endpoints
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

	// Transaction Endpoints
	r.POST("/api/v1/transactions", func(c *gin.Context) {
		var tx Transaction
		if err := c.ShouldBindJSON(&tx); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pgDB != nil {
			pgDB.Create(&tx)
		}
		c.JSON(http.StatusCreated, tx)
	})

	r.GET("/api/v1/transactions", func(c *gin.Context) {
		var transactions []Transaction
		if pgDB != nil {
			pgDB.Find(&transactions)
		}
		c.JSON(http.StatusOK, transactions)
	})

	// Validator Endpoints
	r.GET("/api/v1/validators", func(c *gin.Context) {
		var validators []Validator
		if pgDB != nil {
			pgDB.Find(&validators)
		}
		c.JSON(http.StatusOK, validators)
	})

	r.POST("/api/v1/validators", func(c *gin.Context) {
		var v Validator
		if err := c.ShouldBindJSON(&v); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pgDB != nil {
			pgDB.Create(&v)
		}
		c.JSON(http.StatusCreated, v)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
