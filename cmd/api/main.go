package main

import (
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
	"horizon-core/internal/license"
	"horizon-core/internal/network"
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
			Volume:    pkg.Volume,
			Root:      pkg.RootHex(),
			Seed:      pkg.SeedHex(),
			CreatedAt: time.Now(),
			Signature: sig,
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

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
