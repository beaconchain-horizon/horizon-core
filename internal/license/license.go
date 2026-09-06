package license

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"

	"horizon-core/internal/crypto"
)

type LicenseInfo struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
	Signature string    `json:"signature"`
	RootHash  string    `json:"root_hash"`
}

type LicenseGenerator struct {
	privateKey *ecdsa.PrivateKey
}

func NewLicenseGenerator(privateKey *ecdsa.PrivateKey) *LicenseGenerator {
	return &LicenseGenerator{privateKey: privateKey}
}

func (g *LicenseGenerator) GenerateLicense(productID, userID string, duration time.Duration) (*LicenseInfo, error) {
	data := map[string]interface{}{
		"product_id": productID,
		"user_id":    userID,
		"created_at": time.Now().UTC(),
		"expires_at": time.Now().UTC().Add(duration),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	license := &LicenseInfo{
		ID:        "lic_" + time.Now().Format("20060102150405"),
		ProductID: productID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(duration),
		Active:    true,
	}

	signature, err := crypto.SignData(jsonData, g.privateKey)
	if err != nil {
		return nil, err
	}
	license.Signature = signature

	return license, nil
}
