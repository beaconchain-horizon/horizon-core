package auth
import (
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
)
func LicenseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// در حالت تست، اجازه عبور بده تا دیتابیس و CPU تست شوند
		if os.Getenv("BENCHMARK_MODE") == "true" {
			c.Next()
			return
		}
		// حالت عادی (قانون اول)
		if c.GetHeader("X-License-ID") == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Law #1: License required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
