package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// isMalicious checks for common SQLi and XSS patterns
func isMalicious(input string) bool {
	input = strings.ToLower(input)
	patterns := []string{
		// XSS Patterns
		"<script", "javascript:", "onerror=", "onload=", "<img", "alert(",
		// SQL Injection Patterns
		"union select", "or 1=1", "or 1 = 1", "drop table", "insert into", "delete from", "--", "/*", "xp_cmdshell",
	}
	for _, p := range patterns {
		if strings.Contains(input, p) {
			return true
		}
	}
	return false
}

// SecurityMiddleware blocks malicious payloads in Query, Body, and Headers
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check URL Query Parameters (e.g., ?q=1' OR 1=1--)
		for _, values := range c.Request.URL.Query() {
			for _, v := range values {
				if isMalicious(v) {
					c.JSON(http.StatusForbidden, gin.H{"error": "Security Policy: Malicious payload detected in query"})
					c.Abort()
					return
				}
			}
		}

		// 2. Check Request Body (for POST/PUT)
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if isMalicious(string(body)) {
					c.JSON(http.StatusForbidden, gin.H{"error": "Security Policy: Malicious payload detected in body"})
					c.Abort()
					return
				}
				// Restore body for the actual handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		c.Next()
	}
}
