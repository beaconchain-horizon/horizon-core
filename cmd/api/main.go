package main

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"horizon-core/internal/crypto"
	"horizon-core/internal/license"
)

var (
	onlineDB     *gorm.DB
	offlineDB    *gorm.DB
	privateKey   *ecdsa.PrivateKey
)

type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type License struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	Signature string    `json:"signature"`
	RootHash  string    `json:"root_hash"`
	Active    bool      `json:"active"`
	Expiry    time.Time `json:"expiry"`
	CreatedAt time.Time `json:"created_at"`
}

func connectOnlineDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://root:CHANGE_ME@horizon:5432/postgres?sslmode=disable"
	}
	var err error
	if strings.HasPrefix(dsn, "sqlite://") {
		dsn = strings.TrimPrefix(dsn, "sqlite://")
		onlineDB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	} else {
		onlineDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		log.Printf("Error connecting to primary DB: %v", err)
		return
	}
	onlineDB.AutoMigrate(&Transaction{}, &License{})
	log.Println("Connected to primary DB")
}

func connectOfflineDB() {
	var err error
	offlineDB, err = gorm.Open(sqlite.Open("horizon_offline.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to SQLite:", err)
	}
	offlineDB.AutoMigrate(&Transaction{})
	log.Println("Connected to SQLite")
}

func main() {
	// Load .env if exists
	// (We'll just rely on environment variables)

	connectOnlineDB()
	connectOfflineDB()

	// Generate or load private key
	privPEM, _ := crypto.GenerateKeyPair()
	privateKey, _ = crypto.PEMToPrivateKey(privPEM)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key", "X-Admin-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "online"})
	})

	// Transactions (hybrid)
	r.POST("/api/v1/transactions", func(c *gin.Context) {
		var tx Transaction
		c.ShouldBindJSON(&tx)
		if onlineDB == nil {
			tx.Status = "offline"
			offlineDB.Create(&tx)
			c.JSON(201, gin.H{"message": "Saved offline", "transaction": tx})
			return
		}
		tx.Status = "confirmed"
		onlineDB.Create(&tx)
		c.JSON(201, gin.H{"message": "Saved online", "transaction": tx})
	})

	r.GET("/api/v1/transactions", func(c *gin.Context) {
		var offline, online []Transaction
		offlineDB.Find(&offline)
		if onlineDB != nil {
			onlineDB.Find(&online)
		}
		c.JSON(200, gin.H{"offline": offline, "online": online})
	})

	// License Generation with Merkle + ECDSA
	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Duration  int    `json:"duration"` // in hours
		}
		c.ShouldBindJSON(&req)

		// Create a PrepaidPackage (Merkle tree)
		pkg := license.NewPrepaidPackage(100) // Example volume
		if err := pkg.Generate(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate Merkle root"})
			return
		}

		// Sign the root with ECDSA
		sig, err := crypto.SignData(pkg.Root, privateKey)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to sign license"})
			return
		}

		lic := License{
			ID:        "lic_" + time.Now().Format("20060102150405"),
			ProductID: req.ProductID,
			UserID:    req.UserID,
			Signature: sig,
			RootHash:  pkg.RootHex(),
			Active:    true,
			Expiry:    time.Now().Add(time.Duration(req.Duration) * time.Hour),
			CreatedAt: time.Now(),
		}
		onlineDB.Create(&lic)
		c.JSON(201, lic)
	})

	r.GET("/api/v1/licenses", func(c *gin.Context) {
		var licenses []License
		onlineDB.Find(&licenses)
		c.JSON(200, licenses)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
