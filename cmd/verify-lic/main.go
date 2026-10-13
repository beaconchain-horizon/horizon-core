package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"horizon-core/internal/crypto"
	"horizon-core/internal/license"
)

type flat struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	Signature  string `json:"signature"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: verify-lic <license.json> <pubkey-hex>")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read:", err)
		os.Exit(1)
	}

	var f flat
	if err := json.Unmarshal(raw, &f); err != nil {
		fmt.Println("parse:", err)
		os.Exit(1)
	}

	st := license.Statement{
		LicenseID:  f.LicenseID,
		UserID:     f.UserID,
		ProductID:  f.ProductID,
		Volume:     f.Volume,
		Duration:   f.Duration,
		MerkleRoot: f.MerkleRoot,
		IssuedAt:   f.IssuedAt,
		ExpiresAt:  f.ExpiresAt,
		HardwareID: f.HardwareID,
	}

	want := license.MerkleRoot(st)
	fmt.Println("envelope  :", f.MerkleRoot)
	fmt.Println("recomputed:", want)
	fmt.Println("merkle match:", f.MerkleRoot == want)

	msg, _ := license.CanonicalBytes(st)
	fmt.Println("canonical :", string(msg))

	hash := sha256.Sum256(msg)
	pub, err := crypto.Secp256k1PublicKeyFromHex(os.Args[2])
	if err != nil {
		fmt.Println("pubkey:", err)
		os.Exit(1)
	}
	sig, _ := hex.DecodeString(f.Signature)
	ok := crypto.VerifyHashRS(pub, hash[:], sig)
	fmt.Println("sig valid :", ok)

	if f.MerkleRoot == want && ok {
		fmt.Println("RESULT: VALID")
		os.Exit(0)
	}
	fmt.Println("RESULT: INVALID")
	os.Exit(1)
}
