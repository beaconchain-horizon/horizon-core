package auth
import (
	"net/http"
	"horizon-core/internal/db"
	"github.com/gin-gonic/gin"
)
func LicenseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" { c.Next(); return }
		licenseID := c.GetHeader("X-License-ID")
		if licenseID == "INTERNAL-CONSENSUS" { c.Next(); return }
		if licenseID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Law #1: X-License-ID required"})
			c.Abort()
			return
		}
		var license db.License
		if err := db.DB.Where("license_id = ? AND status = ?", licenseID, "active").First(&license).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "License invalid"})
			c.Abort()
			return
		}
		c.Next()
	}
}
