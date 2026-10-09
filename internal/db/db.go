package db

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"
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

	pemValue := strings.TrimSpace(os.Getenv("LICENSE_PRIVATE_KEY_PEM"))
	if keyFile := strings.TrimSpace(os.Getenv("LICENSE_PRIVATE_KEY_FILE")); keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("read LICENSE_PRIVATE_KEY_FILE: %w", err)
		}
		pemValue = string(data)
	}
	if strings.TrimSpace(pemValue) == "" {
		return fmt.Errorf("LICENSE_PRIVATE_KEY_FILE or LICENSE_PRIVATE_KEY_PEM is required; refusing to generate an ephemeral Authority key")
	}
	key, err := crypto.PEMToPrivateKey(pemValue)
	if err != nil {
		return fmt.Errorf("invalid Authority secp256k1 private key: %w", err)
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
