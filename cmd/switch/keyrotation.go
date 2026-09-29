package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"time"
)

const keyRotationDays = 90

// ShouldRotateKey checks if the current key is older than 90 days
func ShouldRotateKey(createdAt time.Time) bool {
	return time.Since(createdAt).Hours() > float64(keyRotationDays*24)
}

// GenerateNewSigningKey creates a fresh ECDSA P-256 key
func GenerateNewSigningKey() (*ecdsa.PrivateKey, string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("keygen failed: %w", err)
	}

	pubBytes := elliptic.Marshal(priv.Curve, priv.X, priv.Y)
	addr := hex.EncodeToString(pubBytes[:20])

	return priv, "0x" + addr, nil
}

// ExportPrivateKeyPEM converts a private key to PEM format
func ExportPrivateKeyPEM(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}), nil
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
