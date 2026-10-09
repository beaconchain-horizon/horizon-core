package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"horizon-core/internal/crypto"
	licenseprotocol "horizon-core/internal/license"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func protocolStatement(lic *License) licenseprotocol.Statement {
	return licenseprotocol.Statement{
		LicenseID: lic.LicenseID, UserID: lic.UserID, ProductID: lic.ProductID,
		Volume: lic.Volume, Duration: lic.Duration, MerkleRoot: lic.MerkleRoot,
		IssuedAt: lic.IssuedAt, ExpiresAt: lic.ExpiresAt, HardwareID: lic.HardwareID,
	}
}

// These wrappers preserve existing Switch call sites while delegating protocol
// serialization and Merkle calculation to the shared internal/license package.
func licenseMerkleRoot(lic *License) string {
	return licenseprotocol.MerkleRoot(protocolStatement(lic))
}

func licenseCanonicalMessage(lic *License) string {
	message, err := licenseprotocol.CanonicalBytes(protocolStatement(lic))
	if err != nil {
		return ""
	}
	return string(message)
}

func verifySignature(pub *ecdsa.PublicKey, data []byte, sigHex string) (bool, error) {
	if pub == nil {
		return false, errors.New("trusted public key is nil")
	}
	return crypto.VerifyData(pub, data, sigHex)
}

// verifyLicenseHandler verifies a license already present in Switch's local
// registry. It uses the provisioned Authority public key, never the unlocked
// Switch block-signing key.
func verifyLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID string `json:"license_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "license_id is required"})
		return
	}
	var lic License
	if err := db.Where("license_id = ?", req.LicenseID).First(&lic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"verified": false, "valid": false, "error": "license not found"})
		return
	}
	publicKey, err := loadLicensePublicKey()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"verified": false, "valid": false, "error": err.Error()})
		return
	}
	status := lic.Status
	if status == "revoked" || status == "invalid" || status == "suspended" {
		c.JSON(http.StatusOK, gin.H{"verified": false, "valid": false, "license_id": lic.LicenseID, "status": status})
		return
	}
	hardwareID := strings.TrimSpace(getHardwareIDOrEmpty())
	if hardwareID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"verified": false, "valid": false, "error": "Switch hardware ID unavailable"})
		return
	}
	envelope := licenseprotocol.SignedLicense{Statement: protocolStatement(&lic), Signature: lic.Signature}
	if err := licenseprotocol.Validate(envelope, publicKey, time.Now(), hardwareID); err != nil {
		c.JSON(http.StatusOK, gin.H{"verified": false, "valid": false, "license_id": lic.LicenseID, "status": status, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"verified": true, "valid": true, "license_id": lic.LicenseID,
		"status": status, "merkle_root": lic.MerkleRoot, "hardware_bound": true,
	})
}

// licenseCheckHandler validates Backend's signed statement against the
// independently provisioned Authority public key and this Switch's hardware
// ID. It persists first-seen licenses so local revocation is authoritative.
func licenseCheckHandler(c *gin.Context) {
	var signed licenseprotocol.SignedLicense
	if err := c.ShouldBindJSON(&signed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"verified": false, "valid": false, "error": "expected statement and signature"})
		return
	}
	publicKey, err := loadLicensePublicKey()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"verified": false, "valid": false, "error": err.Error()})
		return
	}
	hardwareID := strings.TrimSpace(getHardwareIDOrEmpty())
	if hardwareID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"verified": false, "valid": false, "error": "Switch hardware ID is unavailable; refusing license check"})
		return
	}
	if err := licenseprotocol.Validate(signed, publicKey, time.Now(), hardwareID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"verified": false, "valid": false, "license_id": signed.Statement.LicenseID, "error": err.Error()})
		return
	}

	s := signed.Statement
	var existing License
	err = db.Where("license_id = ?", s.LicenseID).First(&existing).Error
	if err == nil {
		if existing.Status == "revoked" || existing.Status == "invalid" || existing.Status == "suspended" {
			c.JSON(http.StatusForbidden, gin.H{"verified": true, "valid": false, "license_id": s.LicenseID, "error": "license is locally revoked or suspended"})
			return
		}
		message, _ := licenseprotocol.CanonicalBytes(s)
		if existing.Signature != signed.Signature || licenseCanonicalMessage(&existing) != string(message) {
			c.JSON(http.StatusConflict, gin.H{"verified": true, "valid": false, "license_id": s.LicenseID, "error": "license ID already registered with different signed content"})
			return
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = License{
			LicenseID: s.LicenseID, UserID: s.UserID, ProductID: s.ProductID,
			Volume: s.Volume, Used: 0, Duration: s.Duration, Signature: signed.Signature,
			MerkleRoot: s.MerkleRoot, IssuedAt: s.IssuedAt, ExpiresAt: s.ExpiresAt,
			Status: "active", CreatedAt: time.Unix(s.IssuedAt, 0).UTC(), HardwareID: s.HardwareID,
		}
		if err := db.Create(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"verified": true, "valid": false, "license_id": s.LicenseID, "error": "license registry persistence failed"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"verified": false, "valid": false, "error": "license registry query failed"})
		return
	}

	_ = checkLicenseNow()
	c.JSON(http.StatusOK, gin.H{"verified": true, "valid": true, "license_id": s.LicenseID, "merkle_root": s.MerkleRoot, "hardware_bound": true})
}

func mustCanonical(s licenseprotocol.Statement) []byte {
	b, err := licenseprotocol.CanonicalBytes(s)
	if err != nil {
		panic(fmt.Sprintf("canonicalize validated license: %v", err))
	}
	return b
}
