package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// مسیر سلامت
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "horizon-switch",
		})
	})

	// مسیر تأیید لایسنس (برای اتصال بک‌اند)
	r.POST("/verify", func(c *gin.Context) {
		var req struct {
			LicenseID string `json:"license_id"`
			Signature string `json:"signature"`
			Data      string `json:"data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// در اینجا منطق تأیید را اضافه کنید (فعلاً همیشه true)
		c.JSON(http.StatusOK, gin.H{
			"verified": true,
			"message":  "License verified successfully",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Switch starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
