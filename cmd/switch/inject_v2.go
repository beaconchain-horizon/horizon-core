package main

import (
	"encoding/json"
	"log"
	"time"
)

type InjectedLicense struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
}

func InjectV2License(data string, signature string) error {
	var lic InjectedLicense
	if err := json.Unmarshal([]byte(data), &lic); err != nil {
		return err
	}
	var existingUsed int
	var old License
	if err := db.Where("license_id = ?", lic.LicenseID).First(&old).Error; err == nil {
		existingUsed = old.Used
	}
	row := &License{
		LicenseID:  lic.LicenseID,
		UserID:     lic.UserID,
		ProductID:  lic.ProductID,
		Volume:     lic.Volume,
		Used:       existingUsed,
		Duration:   lic.Duration,
		Signature:  signature,
		MerkleRoot: lic.MerkleRoot,
		IssuedAt:   lic.IssuedAt,
		ExpiresAt:  lic.ExpiresAt,
		Status:     "active",
		HardwareID: lic.HardwareID,
		CreatedAt:  time.Now(),
	}
	db.Where("license_id = ?", lic.LicenseID).Delete(&License{})
	if err := db.Create(row).Error; err != nil {
		return err
	}
	setLicenseState(true, row, nil)
	log.Printf("[inject_v2] accepted license %s (user=%s volume=%d)",
		lic.LicenseID, lic.UserID, lic.Volume)
	return nil
}
