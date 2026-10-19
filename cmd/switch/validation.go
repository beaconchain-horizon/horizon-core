package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

var bad = []string{
	"<", ">", ";", "|", "&", "$", "`",
	"union", "select", "drop", "script", "onerror", "onload",
}

func SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/v1/license/check" {
			c.Next()
			return
		}
		if c.Request.Method != "POST" {
			c.Next()
			return
		}
		limit := int64(65537)
		if c.Request.URL.Path == "/api/v1/tx/batch" || c.Request.URL.Path == "/api/v1/industrial/reading/batch" {
			limit = 100 << 20
		}
		b, _ := io.ReadAll(io.LimitReader(c.Request.Body, limit))
		if int64(len(b)) == limit {
			c.AbortWithStatus(413)
			return
		}
		s := strings.ToLower(string(b))
		for _, x := range bad {
			if strings.Contains(s, x) {
				c.AbortWithStatus(400)
				return
			}
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(b))
		c.Next()
	}
}
