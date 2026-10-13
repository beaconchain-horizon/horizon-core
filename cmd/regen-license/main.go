package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	hcrypto "horizon-core/internal/crypto"
	"horizon-core/internal/merkle"
)

type payload struct {
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

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "usage: regen <private.pem> <db> <hwid> <user_id>")
		os.Exit(1)
	}
	keyPath, dbPath, hwid, userID := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	pemData, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read key:", err)
		os.Exit(1)
	}

	priv, err := hcrypto.PEMToPrivateKey(string(pemData))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse key failed:", err)
		os.Exit(1)
	}

	now := time.Now()
	issued := now.Unix()
	expires := now.Add(365 * 24 * time.Hour).Unix()
	licID := fmt.Sprintf("TPS-%d-01", now.UnixMilli())

	leaves := [][]byte{
		[]byte("license_id=" + licID),
		[]byte("user_id=" + userID),
		[]byte("product_id=horizon-core"),
		[]byte("volume=1000000"),
		[]byte("duration=365"),
		[]byte(fmt.Sprintf("issued_at=%d", issued)),
		[]byte(fmt.Sprintf("expires_at=%d", expires)),
		[]byte("hardware_id=" + hwid),
	}
	root := merkle.BinaryRootNoPrefix(leaves)

	p := payload{
		LicenseID: licID, UserID: userID, ProductID: "horizon-core",
		Volume: 1000000, Duration: 365,
		MerkleRoot: root, IssuedAt: issued, ExpiresAt: expires, HardwareID: hwid,
	}
	msg, _ := json.Marshal(p)
	sigHex, err := hcrypto.SignData(msg, priv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sign:", err)
		os.Exit(1)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		os.Exit(1)
	}

	// delete old licenses for this user
	if err := db.Exec("DELETE FROM licenses WHERE user_id = ?", userID).Error; err != nil {
		// table may have different name; try without where
		db.Exec("DELETE FROM licenses")
	}

	res := db.Exec(`INSERT INTO licenses
		(license_id, user_id, product_id, volume, used, duration, signature, merkle_root, issued_at, expires_at, status, hardware_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		licID, userID, "horizon-core", 1000000, 0, 365,
		sigHex, root, issued, expires, "active", hwid, time.Now())

	if res.Error != nil {
		fmt.Fprintln(os.Stderr, "insert:", res.Error)
		os.Exit(1)
	}
	os.Stdout.WriteString(fmt.Sprintf("%s|%s|%s|%d|%d\n", licID, sigHex, root, issued, expires))
}
