package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// XSSMiddleware blocks requests containing common XSS payloads
func XSSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// بررسی وجود تگ‌های خطرناک
				bodyStr := string(body)
				if strings.Contains(bodyStr, "<script") || strings.Contains(bodyStr, "onerror=") || strings.Contains(bodyStr, "javascript:") || strings.Contains(bodyStr, "<img") {
					c.JSON(http.StatusForbidden, gin.H{"error": "XSS payload detected and blocked"})
					c.Abort()
					return
				}
				// بازگرداندن بدنه برای ادامه پردازش
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
		c.Next()
	}
}
