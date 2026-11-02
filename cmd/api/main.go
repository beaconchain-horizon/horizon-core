package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"horizon-core/internal/auth"
	"horizon-core/internal/db"
	"github.com/gin-gonic/gin"
)

var CanonicalLawHash = ""

func init() {
	hash := sha256.Sum256([]byte("LAW-001: Zero-Storage; LAW-002: Hardware-License; LAW-003: All Comms Require License"))
	CanonicalLawHash = hex.EncodeToString(hash[:])
}

func main() {
	db.Init()
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" { nodeID = "node-main" }

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MB Limit for Overflow tests
	
	r.Use(auth.WAFMiddleware()) // اعمال فایروال
	r.Use(auth.LicenseAuthMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "online", "node": nodeID, "law_hash": CanonicalLawHash})
	})

	r.POST("/api/v1/license/verify", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "verified"})
	})

	log.Printf("✅ %s listening on :%s (WAF Active)", nodeID, port)
	r.Run(":" + port)
}
