package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"horizon-core/internal/crypto"
)

type Req struct {
	TxID       string `json:"tx_id"`
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int64  `json:"volume"`
	Duration   int    `json:"duration"`
	Signature  string `json:"signature"`
	PublicKey  string `json:"public_key"`
	MerkleRoot string `json:"merkle_root"`
	MerkleType string `json:"merkle_type"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
}

func main() {
	port := os.Getenv("NODE_PORT")
	role := os.Getenv("NODE_ROLE")
	if port == "" || role == "" {
		log.Fatal("FAIL-CLOSED: env missing")
	}

	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		var req Req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(map[string]any{"vote": false, "reason": "bad json"})
			return
		}

		var vote bool

		switch role {
		case "merkle-type":
			vote = (req.MerkleType == "binary" || req.MerkleType == "ternary")
		case "merkle-root":
			vote = len(req.MerkleRoot) == 64
		case "ecdsa":
			pub, err := crypto.Secp256k1PublicKeyFromHex(req.PublicKey)
			if err == nil {
				// EXACT format matching the Offline Dashboard
				expTime := time.Unix(req.ExpiresAt, 0).UTC().Format(time.RFC3339)
				payload := fmt.Sprintf("%s|%s|%s|%d|%s|active|%s", req.LicenseID, req.ProductID, req.UserID, req.Volume, req.MerkleRoot, expTime)
				vote, _ = crypto.VerifyData(pub, []byte(payload), req.Signature)
			}
		case "txid":
			vote = strings.HasPrefix(req.TxID, "tx_") && len(req.TxID) > 10
		case "volume":
			vote = req.Volume > 0
		case "time":
			now := time.Now().Unix()
			vote = now >= req.IssuedAt && now <= req.ExpiresAt
		case "user-product":
			vote = req.UserID != "" && req.ProductID != ""
		default:
			vote = false
		}

		json.NewEncoder(w).Encode(map[string]any{"vote": vote, "role": role})
	})

	log.Printf("✅ Node [%s] on port %s", role, port)
	http.ListenAndServe(":"+port, nil)
}
