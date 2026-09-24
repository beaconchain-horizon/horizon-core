package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"horizon-core/internal/merkle"

	"github.com/gin-gonic/gin"
)

// licenseSigningPayload contains ONLY immutable license fields.
//
// IMPORTANT:
// Used and Status are intentionally excluded.
//
// Used changes as the license is consumed.
// Status can change between active/grace/expired/revoked.
//
// Neither field belongs to the cryptographic signature.
type licenseSigningPayload struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id,omitempty"`
}

// licenseMerkleRoot — RFC 6962 root for license fields
func licenseMerkleRoot(lic *License) string {
	leaves := [][]byte{
		[]byte("license_id=" + lic.LicenseID),
		[]byte("user_id=" + lic.UserID),
		[]byte("product_id=" + lic.ProductID),
		[]byte("volume=" + strconv.Itoa(lic.Volume)),
		[]byte("duration=" + strconv.Itoa(lic.Duration)),
		[]byte("issued_at=" + strconv.FormatInt(lic.IssuedAt, 10)),
		[]byte("expires_at=" + strconv.FormatInt(lic.ExpiresAt, 10)),
		[]byte("hardware_id=" + lic.HardwareID),
	}
	return merkle.RootFromData(leaves)
}

// licenseCanonicalMessage returns the canonical JSON payload
// used for BOTH signing and verification.
//
// This function is the single source of truth for the license
// signature format.
//
// Do not add Used or Status here.
func licenseCanonicalMessage(lic *License) string {
	payload := licenseSigningPayload{
		LicenseID:  lic.LicenseID,
		UserID:     lic.UserID,
		ProductID:  lic.ProductID,
		Volume:     lic.Volume,
		Duration:   lic.Duration,
		MerkleRoot: lic.MerkleRoot,
		IssuedAt:   lic.IssuedAt,
		ExpiresAt:  lic.ExpiresAt,
		HardwareID: lic.HardwareID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("marshal license signing payload: %v", err))
	}

	return string(data)
}

// verifySignature verifies an ECDSA P-256 signature.
//
// The input data is SHA-256 hashed before verification.
//
// Signature format:
//
//	R || S
//
// where R and S are each exactly 32 bytes.
func verifySignature(
	pub *ecdsa.PublicKey,
	data []byte,
	sigHex string,
) (bool, error) {
	sigHex = normalizeHex(sigHex)

	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("bad signature hex: %w", err)
	}

	if len(sig) != 64 {
		return false, fmt.Errorf(
			"bad signature length: got %d, want 64",
			len(sig),
		)
	}

	if pub == nil {
		return false, fmt.Errorf("public key is nil")
	}

	hash := sha256.Sum256(data)

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])

	return ecdsa.Verify(
		pub,
		hash[:],
		r,
		s,
	), nil
}

// normalizeHex removes an optional 0x/0X prefix.
func normalizeHex(value string) string {
	if len(value) >= 2 {
		if value[:2] == "0x" || value[:2] == "0X" {
			return value[2:]
		}
	}

	return value
}

// verifyLicenseHandler verifies a license stored in the local
// database using the same canonical JSON message used when the
// license was signed.
func verifyLicenseHandler(c *gin.Context) {
	var req struct {
		LicenseID string `json:"license_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var lic License

	if err := db.
		Where("license_id = ?", req.LicenseID).
		First(&lic).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "license not found",
		})
		return
	}

	stateMutex.RLock()
	key := unlockedKey
	addr := unlockedAddr
	stateMutex.RUnlock()

	if key == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key is locked - call /api/v1/key/unlock first",
		})
		return
	}

	// IMPORTANT:
	// Signing and verification use the exact same canonical
	// JSON function.
	msg := licenseCanonicalMessage(&lic)

	sigOK, err := verifySignature(
		&key.PublicKey,
		[]byte(msg),
		lic.Signature,
	)

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
