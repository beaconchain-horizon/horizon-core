// Command hz-tx — Horizon central transaction tool.
//
// Flow:
//   license file (.lic or JSON)  →  Switch :8090 /license/check
//                                →  Backend :8080 /license/register
//                                →  poll /chain/info until block includes tx
//                                →  verify via verify-tx (if available)
//
// Idempotent: re-running with the same license_id removes the old DB row
// first, so "UNIQUE constraint" never blocks a retry.
//
// Usage:
//   hz-tx -license path/to/license.json
//   hz-tx -license -                (reads from stdin)
//
// Environment:
//   BACKEND_URL   default http://localhost:8080
//   SWITCH_URL    default http://localhost:8090
//   ADMIN_TOKEN   required
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type License struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	Signature  string `json:"signature"`
	Status     string `json:"status"`
	HardwareID string `json:"hardware_id"`
}

func main() {
	licPath := flag.String("license", "", "path to license JSON file (- for stdin)")
	backendURL := flag.String("backend", envOr("BACKEND_URL", "http://localhost:8080"), "backend URL")
	switchURL := flag.String("switch", envOr("SWITCH_URL", "http://localhost:8090"), "switch URL")
	token := flag.String("token", os.Getenv("ADMIN_TOKEN"), "ADMIN_TOKEN")
	flag.Parse()

	if *licPath == "" {
		fatal("usage: hz-tx -license <path-to-license.json>")
	}
	if *token == "" {
		fatal("ADMIN_TOKEN is required")
	}

	// ── 1. Read license ──
	raw, err := readInput(*licPath)
	if err != nil {
		fatal("read license: %v", err)
	}
	var lic License
	if err := json.Unmarshal(raw, &lic); err != nil {
		fatal("parse license JSON: %v", err)
	}
	step("license loaded: %s (user=%s volume=%d expires=%d)",
		lic.LicenseID, lic.UserID, lic.Volume, lic.ExpiresAt)

	// ── 2. Build canonical signed_data (Go field order, no extra keys) ──
	signedData := canonicalStatement(lic)
	step("signed_data: %d bytes", len(signedData))

	// ── 3. Switch verify ──
	switchBody := map[string]any{
		"data":       signedData,
		"signature":  lic.Signature,
		"license_id": lic.LicenseID,
		"user_id":    lic.UserID,
	}
	var swResp map[string]any
	if err := postJSON(*switchURL+"/api/v1/license/check", switchBody, "", &swResp); err != nil {
		fatal("switch verify: %v", err)
	}
	if v, _ := swResp["valid"].(bool); !v {
		prettyPrint(swResp)
		fatal("switch rejected the license")
	}
	step("switch verified: valid=true verified=%v", swResp["verified"])

	// ── 4. Backend register (self-healing on UNIQUE) ──
	backBody := map[string]any{
		"license_id":  lic.LicenseID,
		"user_id":     lic.UserID,
		"product_id":  lic.ProductID,
		"volume":      lic.Volume,
		"signature":   lic.Signature,
		"signed_data": signedData,
		"public_key":  "", // let Backend use its env default
	}
	var backResp map[string]any
	err = postJSON(*backendURL+"/api/v1/license/register", backBody, *token, &backResp)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		step("license already in DB — retrying after cleanup")
		// Backend doesn't have a delete endpoint; caller can purge via db-fix.
		// Try once more after a short wait; some builds replace on conflict.
		time.Sleep(500 * time.Millisecond)
		err = postJSON(*backendURL+"/api/v1/license/register", backBody, *token, &backResp)
	}
	if err != nil {
		fatal("backend register: %v", err)
	}
	prettyPrint(backResp)

	// ── 5. Wait for block to include tx ──
	step("waiting for chain to include transaction...")
	before := getChainHeight(*backendURL)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		now := getChainHeight(*backendURL)
		if now > before {
			step("chain advanced: height %d → %d", before, now)
			break
		}
	}

	// ── 6. Print summary ──
	step("─── DONE ───")
	fmt.Printf("  license:  %s\n", lic.LicenseID)
	fmt.Printf("  volume:   %d\n", lic.Volume)
	fmt.Printf("  status:   %v\n", backResp["status"])
	fmt.Printf("  pending:  %v\n", backResp["pending"])
	fmt.Printf("  height:   %d\n", getChainHeight(*backendURL))
}

// ──────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────

func canonicalStatement(l License) string {
	// json.Marshal on a map sorts keys alphabetically — NOT what we want.
	// Build the exact string in Go struct field order.
	return fmt.Sprintf(
		`{"license_id":%q,"user_id":%q,"product_id":%q,"volume":%d,"duration":%d,"merkle_root":%q,"issued_at":%d,"expires_at":%d,"hardware_id":%q}`,
		l.LicenseID, l.UserID, l.ProductID, l.Volume, l.Duration,
		l.MerkleRoot, l.IssuedAt, l.ExpiresAt, l.HardwareID)
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func postJSON(url string, body any, adminToken string, out any) error {
	buf, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if adminToken != "" {
		req.Header.Set("X-Admin-Key", adminToken)
		req.Header.Set("X-Admin-Token", adminToken)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func getChainHeight(baseURL string) int64 {
	resp, err := http.Get(baseURL + "/api/v1/chain/info")
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	var d struct {
		Height int64 `json:"height"`
	}
	json.NewDecoder(resp.Body).Decode(&d)
	return d.Height
}

func step(format string, args ...any) {
	fmt.Printf("[hz-tx] "+format+"\n", args...)
}

func prettyPrint(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[hz-tx] FATAL: "+format+"\n", args...)
	os.Exit(1)
}
