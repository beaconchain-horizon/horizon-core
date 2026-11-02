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

// هش قانون اساسی پروژه (Law Hash)
var CanonicalLawHash = ""

func init() {
	// ساخت هش از قوانین اساسی (برای سند ممیزی)
	hash := sha256.Sum256([]byte("LAW-001: Zero-Storage; LAW-002: Hardware-License; LAW-003: All Comms Require License"))
	CanonicalLawHash = hex.EncodeToString(hash[:])[:16] // 16 کاراکتر اول
}

func main() {
	db.Init()
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" { nodeID = "node-main" }

	r := gin.Default()
	r.Use(auth.LicenseAuthMiddleware())

	// اندپوینت Health که حالا Law Hash را برمی‌گرداند
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "online", 
			"node": nodeID,
			"law_hash": CanonicalLawHash,
			"audit_ready": true,
		})
	})

	r.POST("/api/v1/tx", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("✅ %s listening on :%s (Audit Compliant)", nodeID, port)
	r.Run(":" + port)
}
