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
		if c.Request.Method != "POST" {
			c.Next()
			return
		}
		b, _ := io.ReadAll(io.LimitReader(c.Request.Body, 65537))
		if len(b) > 65536 {
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
