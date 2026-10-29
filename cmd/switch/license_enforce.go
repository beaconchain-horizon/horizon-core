package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"gorm.io/gorm"
)

const (
	licenseCheckInterval = 24 * time.Hour

	licenseStatusActive  = "active"
	licenseStatusGrace   = "grace"
	licenseStatusExpired = "expired"
	licenseStatusRevoked = "revoked"
	licenseStatusInvalid = "invalid"
)

var (
	licenseStateMutex sync.RWMutex

	licenseValid bool

	activeLicenseState *License

	lastLicenseError error
)

// isLicenseValid returns the current global license state.
func isLicenseValid() bool {
	licenseStateMutex.RLock()
	defer licenseStateMutex.RUnlock()

	return licenseValid
}

// setLicenseState atomically updates the global license state.
func setLicenseState(
	valid bool,
	lic *License,
	err error,
) {
	licenseStateMutex.Lock()
	defer licenseStateMutex.Unlock()

	licenseValid = valid
	activeLicenseState = lic
	lastLicenseError = err
}

// getLicenseState returns a snapshot of the current state.
func getLicenseState() (bool, *License, error) {
	licenseStateMutex.RLock()
	defer licenseStateMutex.RUnlock()

	return licenseValid, activeLicenseState, lastLicenseError
}

// getActiveLicense loads the newest active license from the
// local database.
func getActiveLicense() (*License, error) {
	if db == nil {
		return nil, errors.New("license database is not initialized")
	}

	var license License

	err := db.
		Where("status = ?", licenseStatusActive).
		Order("issued_at DESC").
		First(&license).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active license found")
		}

		return nil, fmt.Errorf("load active license: %w", err)
	}

	return &license, nil
}

// loadLicensePublicKey loads an ECDSA secp256k1 public key from
// HORIZON_LICENSE_PUBLIC_KEY.
func loadLicensePublicKey() (*ecdsa.PublicKey, error) {

	value := strings.TrimSpace(os.Getenv("HORIZON_LICENSE_PUBLIC_KEY"))

	if value == "" {
		// Fallback: use the currently unlocked key pair
		stateMutex.RLock()
		unlocked := unlockedKey
		stateMutex.RUnlock()

		if unlocked != nil {
			return &unlocked.PublicKey, nil
		}
		return nil, errors.New(
			"HORIZON_LICENSE_PUBLIC_KEY is not configured",
		)
	}

	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")

	raw, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode license public key: %w", err)
	}

	switch len(raw) {
	case 64:
		return newSecp256k1PublicKey(raw[:32], raw[32:])

	case 65:
		if raw[0] != 0x04 {
			return nil, errors.New(
				"invalid uncompressed secp256k1 public key prefix",
			)
		}

		return newSecp256k1PublicKey(raw[1:33], raw[33:65])

	default:
		return nil, fmt.Errorf(
			"unsupported secp256k1 public key length: %d bytes",
			len(raw),
		)
	}
}

// newSecp256k1PublicKey constructs an ECDSA secp256k1 public key from
// X and Y coordinates.
func newSecp256k1PublicKey(
	xBytes []byte,
	yBytes []byte,
) (*ecdsa.PublicKey, error) {
	if len(xBytes) != 32 || len(yBytes) != 32 {
		return nil, errors.New(
			"secp256k1 public key coordinates must be 32 bytes",
		)
	}

	curve := secp256k1.S256()

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	if !curve.IsOnCurve(x, y) {
		return nil, errors.New(
			"license public key is not on secp256k1 curve",
		)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// verifyLicenseEnforcementSignature verifies the license using
// the SAME canonical JSON payload used by license_sign.go.
func verifyLicenseEnforcementSignature(license *License) error {
	if license == nil {
		return errors.New("license is nil")
	}

	if strings.TrimSpace(license.Signature) == "" {
		return errors.New("license signature is empty")
	}

	publicKey, err := loadLicensePublicKey()
	if err != nil {
		return err
	}

	msg := licenseCanonicalMessage(license)

	valid, err := verifySignature(
		publicKey,
		[]byte(msg),
		license.Signature,
	)

	if err != nil {
		return err
	}

	if !valid {
		return errors.New("license signature verification failed")
	}

	return nil
}

// validateLicense performs all Phase-A license checks.
func validateLicense(license *License, now time.Time) error {
	if license == nil {
		return errors.New("license is nil")
	}

	if strings.TrimSpace(license.LicenseID) == "" {
		return errors.New("license_id is empty")
	}

	if strings.TrimSpace(license.ProductID) == "" {
		return errors.New("product_id is empty")
	}

	if license.Volume <= 0 {
		return errors.New("license volume must be greater than zero")
	}

	if license.Used < 0 {
		return errors.New("license used value cannot be negative")
	}

	if license.Used > license.Volume {
		return fmt.Errorf(
			"license volume exceeded: used=%d volume=%d",
			license.Used,
			license.Volume,
		)
	}

	if license.IssuedAt <= 0 {
		return errors.New("license issued_at is invalid")
	}

	if license.ExpiresAt <= 0 {
		return errors.New("license expires_at is invalid")
	}

	if license.ExpiresAt <= license.IssuedAt {
		return errors.New("license expires_at must be after issued_at")
	}

	issuedAt := time.Unix(license.IssuedAt, 0)
	expiresAt := time.Unix(license.ExpiresAt, 0)

	if now.Before(issuedAt) {
		return errors.New("license is not active yet")
	}

	status := strings.ToLower(strings.TrimSpace(license.Status))

	if status == licenseStatusRevoked {
		return errors.New("license is revoked")
	}

	if status == licenseStatusInvalid {
		return errors.New("license is invalid")
	}

	if err := verifyLicenseEnforcementSignature(license); err != nil {
		return err
	}

	// P0.3 FIX: Hardware ID enforcement is independent of signature check.
	// Before: HWID was only checked inside the signature-error branch,
	// meaning any valid signature bypassed hardware binding entirely.
	if license.HardwareID != "" {
		currentHWID := getHardwareIDOrEmpty()
		if currentHWID == "" {
			return fmt.Errorf("hardware id: cannot read current hardware")
		}
		if currentHWID != license.HardwareID {
			return fmt.Errorf(
				"hardware id mismatch: license bound to %s, current is %s",
				license.HardwareID,
				currentHWID,
			)
		}
	}

	// Grace period: allow up to gracePeriodDays after expiry
	if now.After(expiresAt) {
		daysOverdue := int(now.Sub(expiresAt).Hours() / 24)
		if daysOverdue > gracePeriodDays {
			return fmt.Errorf(
				"license expired %d days ago (grace period exceeded)",
				daysOverdue,
			)
		}
		log.Printf(
			"[LICENSE] grace period: %d days overdue, %d days left",
			daysOverdue,
			gracePeriodDays-daysOverdue,
		)
	}

	return nil
}

// checkLicenseNow loads and validates the active license and
// updates the global enforcement state.
func checkLicenseNow() error {
	license, err := getActiveLicense()

	if err != nil {
		setLicenseState(false, nil, err)

		log.Printf("[LICENSE] validation failed: %v", err)

		return err
	}

	if err := validateLicense(license, time.Now()); err != nil {
		setLicenseState(false, license, err)

		log.Printf(
			"[LICENSE] validation failed for %s: %v",
			license.LicenseID,
			err,
		)

		return err
	}

	setLicenseState(true, license, nil)

	log.Printf(
		"[LICENSE] valid: id=%s product=%s expires_at=%s volume=%d used=%d",
		license.LicenseID,
		license.ProductID,
		time.Unix(license.ExpiresAt, 0).UTC().Format(time.RFC3339),
		license.Volume,
		license.Used,
	)

	return nil
}

// startLicenseChecker starts the periodic license validation
// loop. The first check is performed immediately.
func startLicenseChecker(stopCh chan struct{}) {
	if stopCh == nil {
		log.Printf("[LICENSE] checker cannot start: stop channel is nil")
		return
	}

	if err := checkLicenseNow(); err != nil {
		log.Printf("[LICENSE] initial validation failed: %v", err)
	}

	ticker := time.NewTicker(licenseCheckInterval)
	defer ticker.Stop()

	log.Printf(
		"[LICENSE] periodic checker started; interval=%s",
		licenseCheckInterval,
	)

	for {
		select {
		case <-ticker.C:
			if err := checkLicenseNow(); err != nil {
				log.Printf(
					"[LICENSE] periodic validation failed: %v",
					err,
				)
			}

		case <-stopCh:
			log.Printf("[LICENSE] periodic checker stopped")
			return
		}
	}
}

// licenseExpiredResponse is the response structure used when
// transaction execution is blocked by license enforcement.
type licenseExpiredResponse struct {
	Error     string `json:"error"`
	ExpiresAt int64  `json:"expires_at"`
	DaysLeft  int    `json:"days_left"`
}

// buildLicenseBlockedResponse creates the standard expired
// license response.
func buildLicenseBlockedResponse() licenseExpiredResponse {
	valid, license, lastErr := getLicenseState()

	errMsg := "License invalid or expired"
	if lastErr != nil {
		errMsg = lastErr.Error() // حالا خطای واقعی را نشان می‌دهد
	}

	if valid || license == nil {
		return licenseExpiredResponse{
			Error:     errMsg,
			ExpiresAt: 0,
			DaysLeft:  0,
		}
	}

	now := time.Now()
	expiresAt := time.Unix(license.ExpiresAt, 0)
	daysLeft := int(expiresAt.Sub(now).Hours() / 24)

	return licenseExpiredResponse{
		Error:     errMsg,
		ExpiresAt: license.ExpiresAt,
		DaysLeft:  daysLeft,
	}
}

// IncrementLicenseUsage increases the 'Used' field in the DB for the active license.
func IncrementLicenseUsage(count int) error {
	if count <= 0 {
		return nil
	}
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	res := db.Model(&License{}).Where("status = ?", "active").UpdateColumn("used", gorm.Expr("used + ?", count))
	if res.Error != nil {
		fmt.Printf("WARN: failed to increment license usage: %v\n", res.Error)
		return res.Error
	}
	return nil
}
