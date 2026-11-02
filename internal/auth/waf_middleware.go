package auth

import (
	"io"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
)

func WAFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ۱. چک کردن حجم
		if c.Request.ContentLength > 1024*1024 {
			// خواندن و دور ریختن بدنه برای جلوگیری از خطای TCP Reset
			if c.Request.Body != nil {
				io.Copy(io.Discard, c.Request.Body)
				c.Request.Body.Close()
			}
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "WAF: Payload too large"})
			c.Abort()
			return
		}

		// ۲. بررسی الگوهای خطرناک در URL
		target := c.Request.URL.Path + c.Request.URL.RawQuery
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
