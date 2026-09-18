package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"horizon-core/internal/tenant"
)

func registerTenantRoutes(api *gin.RouterGroup) {
	t := api.Group("/tenant")
	t.POST("/heartbeat", receiveHeartbeat)
	t.GET("/list", listTenants)
	t.POST("/create", createTenant)
	t.GET("/:id", getTenant)
	t.GET("/:id/heartbeats", listHeartbeats)
}

func receiveHeartbeat(c *gin.Context) {
	token := c.GetHeader("X-Agent-Token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing agent token"})
		return
	}

	tn, err := tenant.FindByToken(db, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid agent token"})
		return
	}

	var req struct {
		TenantID string `json:"tenant_id"`
		AgentURL string `json:"agent_url"`
		Version  string `json:"version"`
		Status   string `json:"status"`
		Payload  string `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := tenant.RecordHeartbeat(db, tn.TenantID, req.AgentURL, req.Version, req.Status, req.Payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "received", "tenant_id": tn.TenantID})
}

func listTenants(c *gin.Context) {
	var items []tenant.Tenant
	db.Find(&items)

	// اضافه کردن وضعیت heartbeat
	type row struct {
		tenant.Tenant
		Online bool `json:"online"`
	}
	rows := make([]row, 0, len(items))
	for _, t := range items {
		rows = append(rows, row{Tenant: t, Online: !tenant.IsStale(&t, 5*time.Minute)})
	}
	c.JSON(http.StatusOK, gin.H{"tenants": rows, "total": len(rows)})
}

func createTenant(c *gin.Context) {
	var t tenant.Tenant
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if t.TenantID == "" || t.Name == "" || t.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id, name, type required"})
		return
	}
	if t.AgentToken == "" {
		t.AgentToken = tenant.GenerateToken()
	}
	t.IsActive = true
	if err := db.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func getTenant(c *gin.Context) {
	t, err := tenant.FindByID(db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func listHeartbeats(c *gin.Context) {
	var items []tenant.Heartbeat
	db.Where("tenant_id = ?", c.Param("id")).Order("received_at desc").Limit(50).Find(&items)
	c.JSON(http.StatusOK, gin.H{"heartbeats": items, "total": len(items)})
}
