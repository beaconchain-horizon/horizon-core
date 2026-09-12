package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// corsMiddleware handles CORS including preflight for custom headers
// like X-Admin-Token. Allows file:// (Origin: null) too.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, "+
				"X-CSRF-Token, Authorization, accept, origin, Cache-Control, "+
				"X-Requested-With, X-Admin-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Admin-Token")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
