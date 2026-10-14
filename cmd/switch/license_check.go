package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"horizon-core/internal/crypto"
)

type licenseCheckRequest struct {
	LicenseID string `json:"license_id" binding:"required"`
	UserID    string `json:"user_id"`
	Signature string `json:"signature" binding:"required"`
	Data      string `json:"data" binding:"required"`
}

func licenseCheckHandler(c *gin.Context) {
	var req licenseCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false, "verified": false,
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	pub, err := loadLicensePublicKey()
	if err != nil {
		log.Printf("[LICENSE-CHECK] public key unavailable: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"valid": false, "verified": false,
			"error": "public key not configured",
		})
		return
	}

	ok, err := crypto.VerifyData(pub, []byte(req.Data), req.Signature)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid": false, "verified": false,
			"license_id": req.LicenseID,
			"error": "signature error: " + err.Error(),
		})
		return
	}
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"valid": false, "verified": false,
			"license_id": req.LicenseID,
			"error": "signature verification failed",
		})
		return
	}

	// If the license already exists in the local DB, run policy checks
	// (expiry, status, hardware binding). If it does not exist yet, we
	// accept the freshly verified license and inject it below.
	var lic License
	if err := db.Where("license_id = ?", req.LicenseID).First(&lic).Error; err == nil {
		if err := validateLicense(&lic, time.Now()); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"valid": false, "verified": true,
				"license_id": req.LicenseID,
				"error": "policy rejection: " + err.Error(),
			})
			return
		}
	}

	// v2 injection: persist the verified license into the local DB and
	// flip the global license state so /tx and /tx/batch accept traffic.
	// This is the trust handoff between the Backend (source of truth) and
	// the Switch (customer-side enforcement gateway).
	if injErr := InjectV2License(req.Data, req.Signature); injErr != nil {
		log.Printf("[license/check] v2 injection failed: %v", injErr)
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true, "verified": true,
		"license_id": req.LicenseID,
		"user_id": req.UserID,
		"checked_at": time.Now().Unix(),
	})
}
