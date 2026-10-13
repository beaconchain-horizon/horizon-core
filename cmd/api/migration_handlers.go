package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"horizon-core/internal/migration"
)

var migrationManager = migration.NewManager()

// registerMigrationHandlers wires the migration endpoints into the
// Backend API. These are the "intersection" endpoints that move data
// between the two SQLite sources and the PostgreSQL target.
func registerMigrationHandlers(r *gin.Engine) {
	g := r.Group("/api/v1/migrate")

	// POST /api/v1/migrate/start
	// Body:
	//   {
	//     "source_path": "/tmp/horizon-chain.db",
	//     "target_dsn":  "postgresql://user:pass@host:5432/horizon?sslmode=disable",
	//     "tables":      ["blocks","transactions","merkle_roots","geneses"],
	//     "batch_size":  1000,
	//     "workers":     8
	//   }
	g.POST("/start", func(c *gin.Context) {
		var req struct {
			SourcePath string   `json:"source_path" binding:"required"`
			TargetDSN  string   `json:"target_dsn"  binding:"required"`
			Tables     []string `json:"tables"      binding:"required"`
			BatchSize  int      `json:"batch_size"`
			Workers    int      `json:"workers"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.BatchSize <= 0 {
			req.BatchSize = 1000
		}
		if req.Workers <= 0 {
			req.Workers = 8
		}

		job := &migration.Job{
			ID:         "mig_" + randHex(8),
			SourcePath: req.SourcePath,
			TargetDSN:  req.TargetDSN,
			Tables:     req.Tables,
			BatchSize:  req.BatchSize,
			Workers:    req.Workers,
			Status:     "pending",
		}
		if err := migrationManager.Start(job); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"job_id":  job.ID,
			"status":  "started",
			"tables":  job.Tables,
			"workers": job.Workers,
		})
	})

	// GET /api/v1/migrate/status/:id
	g.GET("/status/:id", func(c *gin.Context) {
		j, ok := migrationManager.Get(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		c.JSON(http.StatusOK, j)
	})

	// GET /api/v1/migrate/list
	g.GET("/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"jobs": migrationManager.List()})
	})

	// POST /api/v1/migrate/cancel/:id
	g.POST("/cancel/:id", func(c *gin.Context) {
		j, ok := migrationManager.Get(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		j.Status = "cancelled"
		c.JSON(http.StatusOK, gin.H{"job_id": j.ID, "status": "cancelled"})
	})
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
