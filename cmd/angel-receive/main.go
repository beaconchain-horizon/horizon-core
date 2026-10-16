package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type License struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int64  `json:"volume"`
	Duration   int64  `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
	Signature  string `json:"signature"`
	Used       int64  `json:"used"`
}

type NodeReq struct {
	Left            string `json:"left"`
	Right           string `json:"right"`
	Header          string `json:"header"`
	Message         string `json:"message"`
	Signature       string `json:"signature"`
	MerkleType      string `json:"merkle_type"`
	Used            int64  `json:"used"`
	Volume          int64  `json:"volume"`
	IssuedAt        int64  `json:"issued_at"`
	ExpiresAt       int64  `json:"expires_at"`
	UserID          string `json:"user_id"`
	ProductID       string `json:"product_id"`
	ExpectedUser    string `json:"expected_user"`
	ExpectedProduct string `json:"expected_product"`
}

type NodeResp struct {
	Role     string `json:"role"`
	Vote     bool   `json:"vote"`
	Equation string `json:"equation"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: angel-receive <license.json>")
		os.Exit(1)
	}

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	var lic License
	if err := json.Unmarshal(raw, &lic); err != nil {
		fmt.Println("json error:", err)
		os.Exit(1)
	}

	msgObj := struct {
		LicenseID  string `json:"license_id"`
		UserID     string `json:"user_id"`
		ProductID  string `json:"product_id"`
		Volume     int64  `json:"volume"`
		Duration   int64  `json:"duration"`
		MerkleRoot string `json:"merkle_root"`
		IssuedAt   int64  `json:"issued_at"`
		ExpiresAt  int64  `json:"expires_at"`
		HardwareID string `json:"hardware_id"`
	}{
		LicenseID:  lic.LicenseID,
		UserID:     lic.UserID,
		ProductID:  lic.ProductID,
		Volume:     lic.Volume,
		Duration:   lic.Duration,
		MerkleRoot: lic.MerkleRoot,
		IssuedAt:   lic.IssuedAt,
		ExpiresAt:  lic.ExpiresAt,
		HardwareID: lic.HardwareID,
	}
	msgBytes, _ := json.Marshal(msgObj)
	msg := string(msgBytes)

	req := NodeReq{
		Left:            lic.LicenseID,
		Right:           lic.UserID,
		Header:          lic.LicenseID,
		Message:         msg,
		Signature:       lic.Signature,
		MerkleType:      "binary",
		Used:            lic.Used,
		Volume:          lic.Volume,
		IssuedAt:        lic.IssuedAt,
		ExpiresAt:       lic.ExpiresAt,
		UserID:          lic.UserID,
		ProductID:       lic.ProductID,
		ExpectedUser:    lic.UserID,
		ExpectedProduct: lic.ProductID,
	}
	reqBytes, _ := json.Marshal(req)

	fmt.Println("License:", lic.LicenseID)
	fmt.Println("User:", lic.UserID)
	fmt.Println()
	fmt.Println("Message:", msg)
	fmt.Println()

	critical := []int{9001, 9002, 9003, 9004}
	important := []int{9005, 9006, 9007}

	critYes, critTotal := 0, 0
	impYes, impTotal := 0, 0

	fmt.Println("=== CRITICAL (need 100%) ===")
	for _, port := range critical {
		critTotal++
		r := callNode(port, reqBytes)
		if r.Vote {
			critYes++
			fmt.Printf("  [%d] %s YES\n", port, r.Role)
		} else {
			fmt.Printf("  [%d] %s NO (%s)\n", port, r.Role, r.Equation)
		}
	}

	fmt.Println("=== IMPORTANT (need 2/3) ===")
	for _, port := range important {
		impTotal++
		r := callNode(port, reqBytes)
		if r.Vote {
			impYes++
			fmt.Printf("  [%d] %s YES\n", port, r.Role)
		} else {
			fmt.Printf("  [%d] %s NO (%s)\n", port, r.Role, r.Equation)
		}
	}

	fmt.Println()
	fmt.Printf("Critical: %d/%d | Important: %d/%d\n", critYes, critTotal, impYes, impTotal)

	if critYes == critTotal && impYes >= 2 {
		fmt.Println("CONSENSUS REACHED - ACCEPT")
		os.Exit(0)
	}
	fmt.Println("NO CONSENSUS - REJECT (fail-closed)")
	os.Exit(1)
}

func callNode(port int, body []byte) NodeResp {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/verify", port)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return NodeResp{Vote: false, Equation: "connection error"}
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var r NodeResp
	json.Unmarshal(b, &r)
	return r
}
