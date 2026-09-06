package main

import (
	"crypto/ecdsa"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"horizon-core/internal/crypto"
	"horizon-core/internal/license"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB
var privateKey *ecdsa.PrivateKey

type Transaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func connectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "sqlite://horizon_offline.db"
	}
	var err error
	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect DB:", err)
	}
	db.AutoMigrate(&Transaction{})
}

func main() {
	godotenv.Load()
	connectDB()
	privateKey, _ = crypto.GenerateKeyPair()

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/api/v1/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "online"}) })

	r.POST("/api/v1/transactions", func(c *gin.Context) {
		var tx Transaction
		c.ShouldBindJSON(&tx)
		tx.Status = "confirmed"
		db.Create(&tx)
		c.JSON(201, gin.H{"message": "Saved online", "transaction": tx})
	})

	r.GET("/api/v1/transactions", func(c *gin.Context) {
		var txs []Transaction
		db.Find(&txs)
		c.JSON(200, txs)
	})

	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			ProductID string `json:"product_id"`
			UserID    string `json:"user_id"`
			Volume    int    `json:"volume"`
		}
		c.ShouldBindJSON(&req)
		if req.Volume == 0 { req.Volume = 100 }

		pkg := license.NewPrepaidPackage(req.Volume)
		pkg.Generate()
		sig, _ := crypto.SignData([]byte(pkg.RootHex()+":"+strconv.Itoa(req.Volume)), privateKey)

		c.JSON(201, gin.H{
			"id": "lic_" + strconv.FormatInt(time.Now().Unix(), 10),
			"product_id": req.ProductID,
			"user_id": req.UserID,
			"root": pkg.RootHex(),
			"seed": pkg.SeedHex(),
			"signature": sig,
		})
	})

	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		var req struct { Root string `json:"root"` }
		c.ShouldBindJSON(&req)
		c.JSON(200, gin.H{"valid": len(req.Root) == 64})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" { port = "8080" }
	r.Run(":" + port)
}
