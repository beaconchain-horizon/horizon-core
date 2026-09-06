package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
	"horizon-core/internal/license"
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
	if err := db.InitDB(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	defer db.Close()

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
			Volume:    pkg.Volume, // اینجا باید pkg.Volume باشه، نه req.Volume
			Root:      pkg.RootHex(),
			Seed:      pkg.SeedHex(),
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
