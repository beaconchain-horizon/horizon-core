package main

import (
	"crypto/ecdsa"
	"fmt"
	hcrypto "horizon-core/internal/crypto"
	"log"
	"time"
)

const keyRotationDays = 90

// ShouldRotateKey checks if the current key is older than 90 days
func ShouldRotateKey(createdAt time.Time) bool {
	return time.Since(createdAt).Hours() > float64(keyRotationDays*24)
}

// GenerateNewSigningKey creates a fresh secp256k1 ECDSA key
func GenerateNewSigningKey() (*ecdsa.PrivateKey, string, error) {
	priv, err := hcrypto.GenerateSecp256k1ECDSAKey()
	if err != nil {
		return nil, "", fmt.Errorf("secp256k1 keygen failed: %w", err)
	}

	addr := hcrypto.PublicAddressFromKey(priv)

	return priv, addr, nil
}

// ExportPrivateKeyPEM converts a private key to PEM format
func ExportPrivateKeyPEM(priv *ecdsa.PrivateKey) ([]byte, error) {
	value, err := hcrypto.PrivateKeyPEM(priv)
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// CheckKeyRotation logs a warning if the key needs rotation
func CheckKeyRotation() {
	var vault KeyVault
	if err := db.Order("created_at desc").First(&vault).Error; err != nil {
		return
	}
	if ShouldRotateKey(vault.CreatedAt) {
		log.Printf("[KEY-ROTATION] ⚠️  Key older than %d days — consider rotating", keyRotationDays)
	} else {
		daysLeft := keyRotationDays - int(time.Since(vault.CreatedAt).Hours()/24)
		log.Printf("[KEY-ROTATION] ✅ Key valid for %d more days", daysLeft)
	}
}
