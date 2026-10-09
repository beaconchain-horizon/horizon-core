// licensegen creates a signed Horizon Core v1.4 license using the shared protocol.
package main

import (
	"encoding/json"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	hcrypto "horizon-core/internal/crypto"
	protocol "horizon-core/internal/license"
)

func main() {
	keyHex := flag.String("key", "", "secp256k1 private key hex")
	keyFile := flag.String("key-file", "", "file containing secp256k1 private key hex")
	licenseID := flag.String("license-id", "", "unique license ID")
	userID := flag.String("user-id", "", "tenant/customer ID")
	productID := flag.String("product-id", "horizon-core", "product ID")
	volume := flag.Int("volume", 0, "max transactions/readings, >0")
	days := flag.Int("days", 365, "validity duration in days, >0")
	hwid := flag.String("hardware-id", "", "required target hardware ID")
	out := flag.String("out", "license.json", "output JSON file (must not already exist)")
	flag.Parse()

	rawKey := strings.TrimSpace(*keyHex)
	if rawKey == "" && *keyFile != "" {
		raw, err := os.ReadFile(*keyFile)
		if err != nil { fail("read key file", err) }
		rawKey = strings.TrimSpace(string(raw))
	}
	if rawKey == "" { fail("input", fmt.Errorf("-key or -key-file is required")) }
	var priv *ecdsa.PrivateKey
	var err error
	if strings.HasPrefix(rawKey, "-----BEGIN ") {
		priv, err = hcrypto.PEMToPrivateKey(rawKey)
	} else {
		priv, err = hcrypto.Secp256k1PrivateKeyFromHex(rawKey)
	}
	if err != nil { fail("parse Authority key", err) }
	if strings.TrimSpace(*licenseID) == "" || strings.TrimSpace(*userID) == "" ||
		strings.TrimSpace(*productID) == "" || strings.TrimSpace(*hwid) == "" ||
		*volume <= 0 || *days <= 0 {
		fail("input", fmt.Errorf("license-id, user-id, product-id, hardware-id, positive volume and positive days are required"))
	}
	if _, err := os.Stat(*out); err == nil {
		fail("output", fmt.Errorf("refusing to overwrite existing file %q", *out))
	} else if !os.IsNotExist(err) {
		fail("output", err)
	}

	now := time.Now().UTC().Unix()
	s := protocol.Statement{
		LicenseID: strings.TrimSpace(*licenseID),
		UserID: strings.TrimSpace(*userID),
		ProductID: strings.TrimSpace(*productID),
		Volume: *volume,
		Duration: *days,
		IssuedAt: now,
		ExpiresAt: now + int64(*days)*86400,
		HardwareID: strings.TrimSpace(*hwid),
	}
	s.MerkleRoot = protocol.MerkleRoot(s)
	message, err := protocol.CanonicalBytes(s)
	if err != nil { fail("canonicalize license", err) }
	signature, err := hcrypto.SignData(message, priv)
	if err != nil { fail("sign license", err) }

	envelope := protocol.SignedLicense{Statement:s, Signature:signature}
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil { fail("serialize license", err) }
	data = append(data, '\n')
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { fail("create license output", err) }
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(*out)
		fail("write license output", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		fail("sync license output", err)
	}
	if err := f.Close(); err != nil { fail("close license output", err) }

	fmt.Printf("Horizon Core v1.4 license created: %s\n", *out)
	fmt.Printf("License ID: %s\n", s.LicenseID)
	fmt.Printf("Product: %s | User: %s | Volume: %d | Duration: %d days\n", s.ProductID, s.UserID, s.Volume, s.Duration)
	fmt.Printf("Expires: %s UTC\n", time.Unix(s.ExpiresAt, 0).UTC().Format(time.RFC3339))
	fmt.Printf("Merkle root: %s\n", s.MerkleRoot)
	fmt.Println("Envelope format: shared v1.4 secp256k1 Statement + Signature")
}

func fail(label string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", label, err)
	os.Exit(1)
}
