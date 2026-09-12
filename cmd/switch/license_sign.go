package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func licenseCanonicalMessage(lic *License) string {
	return fmt.Sprintf(
		"license_id=%s|user_id=%s|product_id=%s|volume=%d|duration=%d|merkle_root=%s|issued_at=%d|expires_at=%d",
		lic.LicenseID, lic.UserID, lic.ProductID,
		lic.Volume, lic.Duration, lic.MerkleRoot,
		lic.IssuedAt, lic.ExpiresAt,
	)
}

func verifySignature(pub *ecdsa.PublicKey, data []byte, sigHex string) (bool, error) {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("bad signature hex: %w", err)
	}
	if len(sig) != 64 {
		return false, fmt.Errorf("bad signature length: got %d, want 64", len(sig))
	}
	hash := sha256.Sum256(data)
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return ecdsa.Verify(pub, hash[:], r, s), nil
}

func verifyLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID string `json:"license_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lic License
	if err := db.Where("license_id = ?", req.LicenseID).First(&lic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "license not found"})
		return
	}

	stateMutex.RLock()
	key := unlockedKey
	addr := unlockedAddr
	stateMutex.RUnlock()

	if key == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is locked - call /api/v1/key/unlock first"})
		return
	}

	msg := licenseCanonicalMessage(&lic)
	sigOK, err := verifySignature(&key.PublicKey, []byte(msg), lic.Signature)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"license_id":      lic.LicenseID,
			"valid":           false,
			"signature_valid": false,
			"error":           err.Error(),
		})
		return
	}

	now := time.Now().Unix()
	expired := now > lic.ExpiresAt
	status := lic.Status
	if expired {
		status = "expired"
	}

	c.JSON(http.StatusOK, gin.H{
		"license_id":      lic.LicenseID,
		"user_id":         lic.UserID,
		"product_id":      lic.ProductID,
		"valid":           sigOK && !expired,
		"signature_valid": sigOK,
		"expired":         expired,
		"status":          status,
		"address":         addr,
		"issued_at":       lic.IssuedAt,
		"expires_at":      lic.ExpiresAt,
		"merkle_root":     lic.MerkleRoot,
	})
}
