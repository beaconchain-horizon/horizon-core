package auth

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

// WAFMiddleware blocks malicious payloads before they reach the core
func WAFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// بررسی هدرها و مسیر برای کاراکترهای خطرناک
		target := c.Request.URL.Path + c.Request.URL.RawQuery
		
		// الگوهای ساده برای بلاک کردن SQLi و XSS و Traversal
		badPatterns := []string{
			"<script>", "SELECT ", "DROP ", "UNION ", "../", "..\\",
			"cmd.exe", "/bin/bash", "169.254.169.254", "<!ENTITY",
		}
		
		for _, pattern := range badPatterns {
			if strings.Contains(strings.ToLower(target), strings.ToLower(pattern)) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "WAF: Malicious payload blocked"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
