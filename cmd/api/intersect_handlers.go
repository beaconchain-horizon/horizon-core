package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"horizon-core/internal/intersect"
)

var intersectInst *intersect.Intersect

// initIntersect reads INTERSECT_DSN env var and boots the intersection layer.
// If INTERSECT_DSN is empty, the layer is disabled and endpoints return 503.
func initIntersect() error {
	dsn := os.Getenv("INTERSECT_DSN")
	if dsn == "" {
		return nil
	}
	x, err := intersect.New(intersect.Config{
		TargetDSN:     dsn,
		FlushInterval: 200 * time.Millisecond,
		FlushBatch:    2000,
		FlushWorkers:  8,
		BufferSize:    200000,
	})
	if err != nil {
		return err
	}
	intersectInst = x
	return nil
}

// registerIntersectHandlers wires the intersection endpoints.
func registerIntersectHandlers(r *gin.Engine) {
	api := r.Group("/api/v1/intersect")

	// POST /api/v1/intersect/:table  — single row
	api.POST("/:table", func(c *gin.Context) {
		if intersectInst == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "intersection layer disabled"})
			return
		}
		var row map[string]any
		if err := c.ShouldBindJSON(&row); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := intersectInst.Enqueue(c.Param("table"), row); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
	})

	// POST /api/v1/intersect/:table/batch  — many rows
	api.POST("/:table/batch", func(c *gin.Context) {
		if intersectInst == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "intersection layer disabled"})
			return
		}
		var rows []map[string]any
		if err := c.ShouldBindJSON(&rows); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := intersectInst.EnqueueBatch(c.Param("table"), rows); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"status": "queued",
			"count":  len(rows),
		})
	})

	// GET /api/v1/intersect/stats
	api.GET("/stats", func(c *gin.Context) {
		if intersectInst == nil {
			c.JSON(http.StatusOK, gin.H{"enabled": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"enabled": true,
			"stats":   intersectInst.Stats(),
		})
	})

	// GET /api/v1/intersect/read/:table?limit=100
	api.GET("/read/:table", func(c *gin.Context) {
		if intersectInst == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "intersection layer disabled"})
			return
		}
		limit := 100
		if l := c.Query("limit"); l != "" {
			if n, err := parseIntSafe(l); err == nil && n > 0 && n <= 1000 {
				limit = n
			}
		}
		var out []map[string]any
		err := intersectInst.DB().
			Table(c.Param("table")).
			Limit(limit).
			Order("1 DESC").
			Find(&out).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"table": c.Param("table"),
			"count": len(out),
			"rows":  out,
		})
	})
}

func parseIntSafe(s string) (int, error) {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, http.ErrNotSupported
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
