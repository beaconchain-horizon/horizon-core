package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"horizon-core/internal/industrial"
)

func registerIndustrialRoutes(api *gin.RouterGroup) {
	_ = db.AutoMigrate(
		&industrial.Site{},
		&industrial.Sensor{},
		&industrial.Reading{},
		&industrial.Alert{},
		&industrial.TamperEvent{},
		&industrial.ControlCommand{},
	)

	ind := api.Group("/industrial")
	ind.GET("/sites", listIndustrialSites)
	ind.POST("/sites", createIndustrialSite)
	ind.GET("/sensors", listIndustrialSensors)
	ind.POST("/sensors", createIndustrialSensor)
	ind.POST("/reading", ingestReading)
	ind.POST("/reading/batch", ingestReadingBatch)
	ind.GET("/reading/:sensor_id", listReadings)
	ind.GET("/alerts", listIndustrialAlerts)
	ind.POST("/alerts/ack", ackIndustrialAlert)
	ind.POST("/control/command", executeControlCommand)
	ind.GET("/control/log", listControlLog)
	ind.GET("/tamper", listTamperEvents)
}

func listIndustrialSites(c *gin.Context) {
	var items []industrial.Site
	db.Find(&items)
	c.JSON(http.StatusOK, gin.H{"sites": items, "total": len(items)})
}

func createIndustrialSite(c *gin.Context) {
	var s industrial.Site
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.SiteID == "" || s.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site_id and name required"})
		return
	}
	s.IsActive = true
	if err := db.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func listIndustrialSensors(c *gin.Context) {
	var items []industrial.Sensor
	q := db
	if siteID := c.Query("site_id"); siteID != "" {
		q = q.Where("site_id = ?", siteID)
	}
	q.Find(&items)
	c.JSON(http.StatusOK, gin.H{"sensors": items, "total": len(items)})
}

func createIndustrialSensor(c *gin.Context) {
	var s industrial.Sensor
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s.SensorID == "" || s.PublicKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sensor_id and public_key required"})
		return
	}
	s.IsActive = true
	if err := db.Create(&s).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func ingestReading(c *gin.Context) {
	var req struct {
		SensorID  string  `json:"sensor_id"`
		Value     float64 `json:"value"`
		Nonce     string  `json:"nonce"`
		Timestamp int64   `json:"timestamp"`
		Signature string  `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var sensor industrial.Sensor
	if err := db.Where("sensor_id = ?", req.SensorID).First(&sensor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}
	signable := industrial.SignableReading{
		SensorID: req.SensorID, Value: req.Value,
		Nonce: req.Nonce, Timestamp: req.Timestamp,
	}
	if !industrial.VerifyReading(sensor.PublicKey, signable, req.Signature) {
		db.Create(&industrial.TamperEvent{SensorID: req.SensorID,
			Reason: "invalid_signature", Detail: "verify failed"})
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid signature"})
		return
	}
	if !industrial.NonceFresh(req.Nonce, req.Timestamp) {
		db.Create(&industrial.TamperEvent{SensorID: req.SensorID,
			Reason: "replay", Detail: req.Nonce})
		c.JSON(http.StatusForbidden, gin.H{"error": "replay detected"})
		return
	}
	reading := industrial.Reading{
		SensorID: req.SensorID, Value: req.Value,
		Nonce: req.Nonce, Timestamp: req.Timestamp,
		Signature: req.Signature, Verified: true,
		RecordedAt: time.Now().UTC(),
	}
	db.Create(&reading)
	alert := industrial.EvaluateReading(sensor, req.Value)
	if alert != nil {
		db.Create(alert)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "accepted", "verified": true,
		"reading_id": reading.ID,
		"hash": industrial.HashReading(signable),
		"alert": alert,
	})
}

func ingestReadingBatch(c *gin.Context) {
	var reqs []struct {
		SensorID  string  `json:"sensor_id"`
		Value     float64 `json:"value"`
		Nonce     string  `json:"nonce"`
		Timestamp int64   `json:"timestamp"`
		Signature string  `json:"signature"`
	}
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accepted, rejected := 0, 0
	for _, r := range reqs {
		var sensor industrial.Sensor
		if err := db.Where("sensor_id = ?", r.SensorID).First(&sensor).Error; err != nil {
			rejected++
			continue
		}
		sig := industrial.SignableReading{SensorID: r.SensorID, Value: r.Value,
			Nonce: r.Nonce, Timestamp: r.Timestamp}
		if !industrial.VerifyReading(sensor.PublicKey, sig, r.Signature) {
			db.Create(&industrial.TamperEvent{SensorID: r.SensorID, Reason: "invalid_signature"})
			rejected++
			continue
		}
		if !industrial.NonceFresh(r.Nonce, r.Timestamp) {
			db.Create(&industrial.TamperEvent{SensorID: r.SensorID, Reason: "replay"})
			rejected++
			continue
		}
		db.Create(&industrial.Reading{
			SensorID: r.SensorID, Value: r.Value, Nonce: r.Nonce,
			Timestamp: r.Timestamp, Signature: r.Signature,
			Verified: true, RecordedAt: time.Now().UTC(),
		})
		if a := industrial.EvaluateReading(sensor, r.Value); a != nil {
			db.Create(a)
		}
		accepted++
	}
	c.JSON(http.StatusOK, gin.H{"accepted": accepted, "rejected": rejected})
}

func listReadings(c *gin.Context) {
	var items []industrial.Reading
	db.Where("sensor_id = ?", c.Param("sensor_id")).
		Order("id desc").Limit(500).Find(&items)
	c.JSON(http.StatusOK, gin.H{"readings": items, "total": len(items)})
}

func listIndustrialAlerts(c *gin.Context) {
	var items []industrial.Alert
	db.Order("id desc").Limit(200).Find(&items)
	c.JSON(http.StatusOK, gin.H{"alerts": items, "total": len(items)})
}

func ackIndustrialAlert(c *gin.Context) {
	var req struct{ ID uint `json:"id"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := db.Model(&industrial.Alert{}).Where("id = ?", req.ID).Update("ack", true)
	c.JSON(http.StatusOK, gin.H{"updated": res.RowsAffected})
}

func executeControlCommand(c *gin.Context) {
	var cmd industrial.ControlCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	licOK := false
	if cmd.LicenseID != "" {
		var lic License
		if err := db.Where("license_id = ? AND status = ?", cmd.LicenseID, "active").
			First(&lic).Error; err == nil {
			licOK = true
		}
	}
	if err := industrial.ValidateCommand(cmd, licOK); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	cmd.Executed = true
	cmd.CreatedAt = time.Now().UTC()
	if err := db.Create(&cmd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "executed", "command_id": cmd.ID})
}

func listControlLog(c *gin.Context) {
	var items []industrial.ControlCommand
	db.Order("id desc").Limit(200).Find(&items)
	c.JSON(http.StatusOK, gin.H{"commands": items, "total": len(items)})
}

func listTamperEvents(c *gin.Context) {
	var items []industrial.TamperEvent
	db.Order("id desc").Limit(200).Find(&items)
	c.JSON(http.StatusOK, gin.H{"events": items, "total": len(items)})
}
