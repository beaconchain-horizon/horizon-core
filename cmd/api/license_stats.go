package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerLicenseStats(r *gin.Engine, db *gorm.DB) {
	r.GET("/api/v1/license/stats", func(c *gin.Context) {
		type row struct {
			UserID string
			Cnt    int64
		}
		var rows []row
		db.Model(&LicenseRecord{}).
			Select("user_id, COUNT(*) as cnt").
			Group("user_id").
			Scan(&rows)

		byUser := gin.H{}
		var total int64
		for _, rw := range rows {
			nextNum := rw.Cnt + 1
			prefix := strings.ToUpper(strings.ReplaceAll(rw.UserID, "-", "-"))
			nextID := fmt.Sprintf("HORIZON-%s-%03d", prefix, nextNum)
			byUser[rw.UserID] = gin.H{
				"count":   rw.Cnt,
				"next_id": nextID,
			}
			total += rw.Cnt
		}
		c.JSON(http.StatusOK, gin.H{
			"total":   total,
			"by_user": byUser,
		})
	})
}
