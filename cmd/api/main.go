package main

import (
	"encoding/hex"
	"log"
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

// ------------------- Models -------------------
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
	Index      int     `json:"index" gorm:"primaryKey"`
	Status     string  `json:"status"`
	Balance    float64 `json:"balance"`
}

type License struct {
	ID        string    `json:"id"`
	Volume    int       `json:"volume"`
	Root      string    `json:"root"`
	Seed      string    `json:"seed"`
	CreatedAt time.Time `json:"created_at"`
	Signature string    `json:"signature"`
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
	pgDB.AutoMigrate(&Customer{}, &Payment{}, &Validator{})
	log.Println("PostgreSQL connected")
}

func main() {
	// Connect to SQLite (offline/simple storage)
	if err := db.InitDB(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	// Connect to PostgreSQL (for customers, payments, validators)
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
		c.JSON(200, gin.H{"status": "online"})
	})

	// License Generation
	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			Volume int `json:"volume"`
		}
		c.ShouldBindJSON(&req)
		pkg := license.NewPrepaidPackage(req.Volume)
		if err := pkg.Generate(); err != nil {
			c.JSON(500, gin.H{"error": "Merkle generation failed"})
			return
		}
		sig, err := crypto.SignData(pkg.Root, db.PrivateKey)
		if err != nil {
			c.JSON(500, gin.H{"error": "Signing failed"})
			return
		}
		lic := License{
			ID:        "lic_" + strconv.FormatInt(time.Now().Unix(), 10),
			Volume:    pkg.Volume,
			Root:      pkg.RootHex(),
			Seed:      pkg.SeedHex(),
			CreatedAt: time.Now(),
			Signature: sig,
		}
		// Store in SQLite if Postgres fails, else in Postgres
		if pgDB != nil {
			pgDB.Create(&lic)
		} else {
			// Fallback to SQLite (not fully implemented in db.go yet)
			log.Println("License saved in memory/SQLite only")
		}
		c.JSON(201, lic)
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
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ciphertext": res})
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
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"plaintext": res})
	})

	r.POST("/api/v1/toolbox/subnet", func(c *gin.Context) {
		var req struct {
			CIDR string `json:"cidr"`
		}
		c.ShouldBindJSON(&req)
		info, err := network.CalculateSubnet(req.CIDR)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	})

	r.POST("/api/v1/toolbox/mac", func(c *gin.Context) {
		var req struct {
			MAC string `json:"mac"`
		}
		c.ShouldBindJSON(&req)
		vendor := network.LookupOUI(req.MAC)
		c.JSON(200, gin.H{"vendor": vendor})
	})

	// ------------------- New Endpoints -------------------
	if pgDB != nil {
		// Customers
		r.POST("/api/v1/customers", func(c *gin.Context) {
			var cust Customer
			if err := c.ShouldBindJSON(&cust); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			pgDB.Create(&cust)
			c.JSON(201, cust)
		})

		r.GET("/api/v1/customers", func(c *gin.Context) {
			var customers []Customer
			pgDB.Find(&customers)
			c.JSON(200, customers)
		})

		// Payments
		r.POST("/api/v1/payments", func(c *gin.Context) {
			var pay Payment
			if err := c.ShouldBindJSON(&pay); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			pgDB.Create(&pay)
			c.JSON(201, pay)
		})

		r.GET("/api/v1/payments", func(c *gin.Context) {
			var payments []Payment
			pgDB.Find(&payments)
			c.JSON(200, payments)
		})

		// Validators
		r.GET("/api/v1/validators", func(c *gin.Context) {
			var validators []Validator
			pgDB.Find(&validators)
			c.JSON(200, validators)
		})

		r.POST("/api/v1/validators", func(c *gin.Context) {
			var v Validator
			if err := c.ShouldBindJSON(&v); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			pgDB.Create(&v)
			c.JSON(201, v)
		})
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
