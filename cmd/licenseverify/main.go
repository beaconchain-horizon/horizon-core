// licenseverify validates a Horizon v1.4 signed license offline. It does not
// call the Backend or trust a public key supplied by the license JSON.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	hcrypto "horizon-core/internal/crypto"
	licenseprotocol "horizon-core/internal/license"
)

func main() {
	licenseFile := flag.String("license", "", "path to signed license JSON")
	publicKeyFile := flag.String("public-key", "", "path to trusted Authority public-key hex file")
	hardwareID := flag.String("hardware-id", "", "expected target hardware ID (recommended)")
	flag.Parse()
	if *licenseFile == "" || *publicKeyFile == "" {
		fmt.Fprintln(os.Stderr, "usage: licenseverify -license license.json -public-key issuer-public.hex [-hardware-id HWID]")
		os.Exit(2)
	}
	data, err := os.ReadFile(*licenseFile)
	if err != nil {
		fail("read license", err)
	}
	var signed licenseprotocol.SignedLicense
	if err := json.Unmarshal(data, &signed); err != nil {
		fail("decode license JSON", err)
	}
	pubData, err := os.ReadFile(*publicKeyFile)
	if err != nil {
		fail("read trusted public key", err)
	}
	pub, err := hcrypto.Secp256k1PublicKeyFromHex(strings.TrimSpace(string(pubData)))
	if err != nil {
		fail("parse trusted secp256k1 public key", err)
	}
	if err := licenseprotocol.Validate(signed, pub, time.Now(), strings.TrimSpace(*hardwareID)); err != nil {
		fmt.Fprintln(os.Stderr, "INVALID:", err)
		os.Exit(1)
	}
	hwidResult := "not locally compared (pass -hardware-id to check the target device)"
	if strings.TrimSpace(*hardwareID) != "" {
		hwidResult = "matched"
	}
	fmt.Printf("VALID v1.4 license: id=%s product=%s user=%s expires_at=%d hwid=%s merkle_root=%s\n",
		signed.Statement.LicenseID, signed.Statement.ProductID, signed.Statement.UserID,
		signed.Statement.ExpiresAt, hwidResult, signed.Statement.MerkleRoot)
}

func fail(prefix string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", prefix, err)
	os.Exit(2)
}
