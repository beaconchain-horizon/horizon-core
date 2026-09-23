package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

func adminAuth(c *gin.Context) {
	p := c.Request.URL.Path
	if p == "/api/v1/health" || p == "/api/v1/industrial/panel" || p == "/api/v1/industrial/reading" || p == "/api/v1/industrial/reading/batch" || strings.HasPrefix(p, "/api/v1/customer/login") {
		c.Next()
		return
	}
	token := c.GetHeader("X-Admin-Token")
	if token == "" || token != os.Getenv("ADMIN_TOKEN") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Next()
}
