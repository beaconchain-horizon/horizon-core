package main

import (
    "net/http"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
)

func adminAuth(c *gin.Context) {
    p := c.Request.URL.Path
    if p == "/api/v1/health" || strings.HasPrefix(p, "/api/v1/customer/login") {
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
