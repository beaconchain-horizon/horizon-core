package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
)

type License struct {
	ID        string    `json:"id"`
	Volume    int       `json:"volume"`
	Root      string    `json:"root"`
	Seed      string    `json:"seed"`
	CreatedAt time.Time `json:"created_at"`
	Signature string    `json:"signature"`
}

func main() {
	// Connect to DB (SQLite)
	if err := db.InitDB(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

	// Generate private key if not exists
	if db.PrivateKey == nil {
		privPEM, err := crypto.GenerateKeyPair()
		if err != nil {
			log.Fatal("Key generation failed:", err)
		}
		db.SavePrivateKey(privPEM)
	}

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

	r.POST("/api/v1/license/generate", func(c *gin.Context) {
		var req struct {
			Volume int `json:"volume"`
		}
		c.ShouldBindJSON(&req)

		// Generate Merkle root using internal/crypto
		seed := make([]byte, 32)
		// (Simplified for this example: use rand.Read)
		// You can implement your PrepaidPackage logic here

		// Sign the seed with ECDSA
		sig, err := crypto.SignData(seed, db.PrivateKey)
		if err != nil {
			c.JSON(500, gin.H{"error": "Signing failed"})
			return
		}

		lic := License{
			ID:        "lic_" + strconv.FormatInt(time.Now().Unix(), 10),
			Volume:    req.Volume,
			Root:      "ROOT_PLACEHOLDER", // Replace with actual Merkle Root
			Seed:      "SEED_PLACEHOLDER",
			CreatedAt: time.Now(),
			Signature: sig,
		}
		c.JSON(201, lic)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
