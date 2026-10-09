// licensegen — Offline License Generator
//
// Purpose:
//
//	Sign Horizon Core licenses using an ECDSA secp256k1 private key
//	WITHOUT any server connection. Designed for air-gapped USB use.
//
// Usage:
//
//	licensegen \
//	  -key <hex-private-key> \
//	  -license-id BANK-MELLI-SHOBE-01 \
//	  -user-id bank-melli \
//	  -product-id horizon-core \
//	  -volume 10000000 \
//	  -days 365 \
//	  -hardware-id <optional-hex> \
//	  -out license.json
//
// Output:
//
//	A JSON file containing the fully signed license, ready for
//	POST /api/v1/license/save or /api/v1/admin/license/issue.
package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"horizon-core/internal/crypto"
	"os"
	"time"
)

type LicenseSigningPayload struct {
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

type SignedLicense struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id,omitempty"`
	Signature  string `json:"signature"`
	SignerAddr string `json:"signer_addr"`
}

func hexToECDSA(hexKey string) (*ecdsa.PrivateKey, error) {
	return crypto.Secp256k1PrivateKeyFromHex(hexKey)
}

func trim0x(s string) string {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return s[2:]
	}
	return s
}

func signData(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	sig, err := crypto.SignDataRS(data, priv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func publicAddressFromKey(priv *ecdsa.PrivateKey) string {
	return crypto.PublicAddressFromKey(priv)
}

func main() {
	var (
		keyHex     = flag.String("key", "", "ECDSA secp256k1 private key (64 hex chars, with or without 0x)")
		keyFile    = flag.String("key-file", "", "path to file containing hex private key (alternative to -key)")
		licenseID  = flag.String("license-id", "", "license ID (required)")
		userID     = flag.String("user-id", "", "user/org ID (required)")
		productID  = flag.String("product-id", "horizon-core", "product ID")
		volume     = flag.Int("volume", 0, "max readings/transactions (required, >0)")
		days       = flag.Int("days", 365, "validity in days")
		hwid       = flag.String("hardware-id", "", "bind license to a specific hardware ID (optional)")
		merkleRoot = flag.String("merkle-root", "", "merkle root (optional, computed if empty)")
		out        = flag.String("out", "license.json", "output JSON file")
	)
	flag.Parse()

	// ── read key ──
	var privHex string
	if *keyHex != "" {
		privHex = *keyHex
	} else if *keyFile != "" {
		data, err := os.ReadFile(*keyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read key-file: %v\n", err)
			os.Exit(1)
		}
		privHex = string(data)
	} else {
		fmt.Fprintln(os.Stderr, "error: -key or -key-file required")
		flag.Usage()
		os.Exit(1)
	}

	priv, err := hexToECDSA(privHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid private key: %v\n", err)
		os.Exit(1)
	}

	// ── validate ──
	if *licenseID == "" || *userID == "" || *volume <= 0 {
		fmt.Fprintln(os.Stderr, "error: -license-id, -user-id, -volume (>0) required")
		os.Exit(1)
	}

	now := time.Now().Unix()
	lic := LicenseSigningPayload{
		LicenseID:  *licenseID,
		UserID:     *userID,
		ProductID:  *productID,
		Volume:     *volume,
		Duration:   *days,
		MerkleRoot: *merkleRoot,
		IssuedAt:   now,
		ExpiresAt:  now + int64(*days)*86400,
		HardwareID: *hwid,
	}

	// compute merkle root if empty (must match licenseMerkleRoot in server)
	if lic.MerkleRoot == "" {
		lic.MerkleRoot = computeMerkleRoot(lic)
	}

	// canonical JSON (must match licenseCanonicalMessage on server)
	canon, err := json.Marshal(lic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}

	sig, err := signData(priv, canon)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign: %v\n", err)
		os.Exit(1)
	}

	signed := SignedLicense{
		LicenseID:  lic.LicenseID,
		UserID:     lic.UserID,
		ProductID:  lic.ProductID,
		Volume:     lic.Volume,
		Duration:   lic.Duration,
		MerkleRoot: lic.MerkleRoot,
		IssuedAt:   lic.IssuedAt,
		ExpiresAt:  lic.ExpiresAt,
		HardwareID: lic.HardwareID,
		Signature:  sig,
		SignerAddr: publicAddressFromKey(priv),
	}

	outJSON, _ := json.MarshalIndent(signed, "", "  ")
	if err := os.WriteFile(*out, outJSON, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ License signed successfully\n")
	fmt.Printf("   File:        %s\n", *out)
	fmt.Printf("   License:     %s\n", signed.LicenseID)
	fmt.Printf("   User:        %s\n", signed.UserID)
	fmt.Printf("   Volume:      %d\n", signed.Volume)
	fmt.Printf("   Valid until: %s\n", time.Unix(signed.ExpiresAt, 0).Format(time.RFC3339))
	fmt.Printf("   Signer:      %s\n", signed.SignerAddr)
	fmt.Printf("   Signature:   %s...\n", sig[:24])
}

// computeMerkleRoot — placeholder, mirrors server licenseMerkleRoot
// If the server computes a different root, use -merkle-root to provide it.
func computeMerkleRoot(lic LicenseSigningPayload) string {
	// simple SHA-256 over concatenation — SERVER MUST MATCH THIS
	// (see cmd/switch/license_sign.go: licenseMerkleRoot)
	// For safety, admin can pass -merkle-root explicitly.
	h := sha256.New()
	fmt.Fprintf(h, "license_id=%s|user_id=%s|product_id=%s|volume=%d|duration=%d|issued_at=%d|expires_at=%d|hardware_id=%s",
		lic.LicenseID, lic.UserID, lic.ProductID, lic.Volume, lic.Duration,
		lic.IssuedAt, lic.ExpiresAt, lic.HardwareID)
	return hex.EncodeToString(h.Sum(nil))
}
