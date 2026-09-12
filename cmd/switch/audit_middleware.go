package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	auditLogAllRequests bool
)

func initAuditMiddleware() {
	auditLogAllRequests = os.Getenv("AUDIT_ALL_REQUESTS") == "true"
}

func auditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Only log non-GET by default (to avoid noise)
		m := c.Request.Method
		if !auditLogAllRequests && (m == "GET" || m == "OPTIONS" || m == "HEAD") {
			return
		}

		status := c.Writer.Status()
		latency := time.Since(start)

		addAudit(
			m+" "+c.FullPath(),
			c.Request.URL.Path,
			"status="+itoa(status)+" latency="+latency.String(),
			c.ClientIP(),
		)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
