package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func adminAuth(c *gin.Context) {
	p := c.Request.URL.Path

	// OPTIONS requests (CORS preflight) — always allow
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	// Public paths — no auth required
	if p == "/api/v1/health" ||
		p == "/api/v1/industrial/panel" ||
		p == "/api/v1/industrial/reading" ||
		p == "/api/v1/industrial/reading/batch" ||
		p == "/api/v1/admin/login" ||
		p == "/api/v1/admin/logout" ||
		strings.HasPrefix(p, "/api/v1/customer/login") {
		c.Next()
		return
	}

	token := c.GetHeader("X-Admin-Token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Check 1: env token (ADMIN_TOKEN)
	if token == os.Getenv("ADMIN_TOKEN") {
		c.Next()
		return
	}

	// Check 2: session token (from adminLoginHandler)
	if validateAdminSession(token) {
		c.Next()
		return
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
