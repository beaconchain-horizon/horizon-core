package db

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"sync"

	"horizon-core/internal/crypto"
)

var (
	PrivateKey *ecdsa.PrivateKey
	mu         sync.RWMutex
)

func InitDB() error {
	mu.Lock()
	defer mu.Unlock()

	if pemValue := os.Getenv("LICENSE_PRIVATE_KEY_PEM"); pemValue != "" {
		key, err := crypto.PEMToPrivateKey(pemValue)
		if err != nil {
			return fmt.Errorf("invalid LICENSE_PRIVATE_KEY_PEM: %w", err)
		}
		PrivateKey = key
		return nil
	}
	pemValue, err := crypto.GenerateKeyPair()
	if err != nil {
		return err
	}
	key, err := crypto.PEMToPrivateKey(pemValue)
	if err != nil {
		return err
	}
	PrivateKey = key
	return nil
}

func PublicKeyPEM() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if PrivateKey == nil {
		return "", fmt.Errorf("private key unavailable")
	}
	return crypto.PublicKeyPEM(PrivateKey)
}

func Close() {}
