package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"horizon-core/internal/db"
)

type LicenseReq struct {
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
	TxID       string `json:"tx_id"`
}

func main() {
	db.Init()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "online", "version": "v2"})
	})

	http.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		var agents []db.Agent
		db.DB.Find(&agents)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	})

	http.HandleFunc("/api/v1/transactions", func(w http.ResponseWriter, r *http.Request) {
		var txs []db.Transaction
		db.DB.Find(&txs)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(txs)
	})

	http.HandleFunc("/api/v1/laws", func(w http.ResponseWriter, r *http.Request) {
		var laws []db.SystemLaw
		db.DB.Find(&laws)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(laws)
	})

	http.HandleFunc("/api/v1/license/register", func(w http.ResponseWriter, r *http.Request) {
		var req LicenseReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", 400)
			return
		}

		// Generate Transaction ID
		h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d", req.Signature, time.Now().UnixNano())))
		req.TxID = "tx_" + hex.EncodeToString(h[:])

		// Send to 7 nodes (Strict Consensus)
		votes := 0
		body, _ := json.Marshal(req)
		for p := 9001; p <= 9007; p++ {
			resp, err := http.Post(fmt.Sprintf("http://localhost:%d/verify", p), "application/json", bytes.NewReader(body))
			if err == nil {
				var res map[string]any
				json.NewDecoder(resp.Body).Decode(&res)
				if v, ok := res["vote"].(bool); ok && v {
					votes++
				}
				resp.Body.Close()
			}
		}

		if votes >= 4 {
			tx := db.Transaction{
				TxID: req.TxID, LicenseID: req.LicenseID, UserID: req.UserID,
				ProductID: req.ProductID, HardwareID: req.HardwareID,
				Volume: req.Volume, Duration: req.Duration,
				Signature: req.Signature, PublicKey: req.PublicKey,
				CreatedAt: time.Now(),
			}
			db.DB.Create(&tx)

			w.WriteHeader(201)
			json.NewEncoder(w).Encode(map[string]any{"status": "success", "tx_id": req.TxID, "votes": votes})
		} else {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(map[string]any{"status": "failed", "votes": votes})
		}
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Backend API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
