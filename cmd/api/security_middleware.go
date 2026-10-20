package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func isMalicious(input string) bool {
	input = strings.ToLower(input)
	patterns := []string{
		"<script", "javascript:", "onerror=", "onload=", "<img", "alert(",
		"union select", "or 1=1", "or 1 = 1", "drop table", "insert into", "delete from", "--", "/*", "xp_cmdshell",
	}
	for _, p := range patterns {
		if strings.Contains(input, p) {
			return true
		}
	}
	return false
}

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, values := range c.Request.URL.Query() {
			for _, v := range values {
				if isMalicious(v) {
					c.JSON(http.StatusForbidden, gin.H{"error": "Security Policy: Malicious payload in query"})
					c.Abort()
					return
				}
			}
		}

		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if isMalicious(string(body)) {
					c.JSON(http.StatusForbidden, gin.H{"error": "Security Policy: Malicious payload in body"})
					c.Abort()
					return
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
		c.Next()
	}
}
