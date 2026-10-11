// licenseverify validates a Horizon v1.4 signed license offline.
// It accepts the same JSON that offline-license.html emits and that
// the Backend returns from /api/v1/license/generate.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	hcrypto "horizon-core/internal/crypto"
)

type license struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	Signature  string `json:"signature"`
	Status     string `json:"status"`
	HardwareID string `json:"hardware_id"`
}

func buildCanonical(l *license) string {
	exp := time.Unix(l.ExpiresAt, 0).UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s|%s|%s|%d|%s|%s|%s",
		l.LicenseID,
		l.ProductID,
		l.UserID,
		l.Volume,
		l.MerkleRoot,
		l.Status,
		exp,
	)
}

func main() {
	licenseFile := flag.String("license", "", "path to signed license JSON")
	pubKeyHex := flag.String("public-key", "", "trusted Authority public-key hex (128 chars, or file path starting with @)")
	hardwareID := flag.String("hardware-id", "", "expected target hardware ID (optional)")
	canonical := flag.Bool("print-canonical", false, "print canonical payload and exit")
	flag.Parse()

	if *licenseFile == "" || *pubKeyHex == "" {
		fmt.Fprintln(os.Stderr, "usage: licenseverify -license license.json -public-key <hex|@file> [-hardware-id HWID]")
		os.Exit(2)
	}

	data, err := os.ReadFile(*licenseFile)
	if err != nil {
		fail("read license", err)
	}

	var lic license
	if err := json.Unmarshal(data, &lic); err != nil {
		fail("decode license JSON", err)
	}

	canonicalMsg := buildCanonical(&lic)
	if *canonical {
		fmt.Println(canonicalMsg)
		return
	}

	pkHex := *pubKeyHex
	if strings.HasPrefix(pkHex, "@") {
		b, err := os.ReadFile(pkHex[1:])
		if err != nil {
			fail("read public key file", err)
		}
		pkHex = strings.TrimSpace(string(b))
	}

	pub, err := hcrypto.Secp256k1PublicKeyFromHex(pkHex)
	if err != nil {
		fail("parse trusted secp256k1 public key", err)
	}

	// 1) signature
	ok, err := hcrypto.VerifyData(pub, []byte(canonicalMsg), lic.Signature)
	if err != nil {
		fmt.Fprintln(os.Stderr, "INVALID: signature error:", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "INVALID: signature verification failed")
		os.Exit(1)
	}

	// 2) expiry
	now := time.Now().UTC()
	if now.Unix() > lic.ExpiresAt {
		fmt.Fprintf(os.Stderr, "INVALID: expired at %s\n", time.Unix(lic.ExpiresAt, 0).UTC().Format(time.RFC3339))
		os.Exit(1)
	}
	if now.Unix() < lic.IssuedAt {
		fmt.Fprintln(os.Stderr, "INVALID: not yet issued")
		os.Exit(1)
	}

	// 3) HWID (optional)
	hwidResult := "not compared (pass -hardware-id to check)"
	if *hardwareID != "" {
		if lic.HardwareID != "" && lic.HardwareID != *hardwareID {
			fmt.Fprintf(os.Stderr, "INVALID: hardware id mismatch (license=%s current=%s)\n", lic.HardwareID, *hardwareID)
			os.Exit(1)
		}
		hwidResult = "matched"
	}

	// 4) status
	if strings.ToLower(lic.Status) != "active" {
		fmt.Fprintf(os.Stderr, "INVALID: status=%s\n", lic.Status)
		os.Exit(1)
	}

	fmt.Printf("VALID v1.4 license\n")
	fmt.Printf("  license_id:  %s\n", lic.LicenseID)
	fmt.Printf("  user_id:     %s\n", lic.UserID)
	fmt.Printf("  product_id:  %s\n", lic.ProductID)
	fmt.Printf("  volume:      %d\n", lic.Volume)
	fmt.Printf("  expires_at:  %s\n", time.Unix(lic.ExpiresAt, 0).UTC().Format(time.RFC3339))
	fmt.Printf("  hwid:        %s\n", hwidResult)
	fmt.Printf("  signature:   %s...\n", lic.Signature[:16])
	fmt.Printf("  canonical:   %s\n", canonicalMsg)
	_ = hex.EncodeToString
}

func fail(prefix string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", prefix, err)
	os.Exit(2)
}
