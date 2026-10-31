package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
)

type GuardianReq struct {
	UserID      string `json:"user_id"`
	LicenseID   string `json:"license_id"`
	PublicKey   string `json:"public_key"`
	Signature   string `json:"signature"`
	SignedData  string `json:"signed_data"`
	TargetAgent string `json:"target_agent"`
	Command     string `json:"command"`
}

func main() {
	db.Init()

	port := os.Getenv("AGENT_NODE_PORT")
	if port == "" {
		port = "9008"
	}

	http.HandleFunc("/api/v1/guardian/dispatch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req GuardianReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"bad json"}`, 400)
			return
		}

		// 1) Verify guardian signature against signed_data
		pub, err := crypto.Secp256k1PublicKeyFromHex(req.PublicKey)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"status": "rejected", "reason": "invalid pubkey"})
			return
		}
		ok, _ := crypto.VerifyData(pub, []byte(req.SignedData), req.Signature)
		if !ok {
			json.NewEncoder(w).Encode(map[string]any{"status": "rejected", "reason": "invalid signature"})
			return
		}

		// 2) Find specialized agent
		var agent db.Agent
		if err := db.DB.Where("role = ? AND status = ?", req.TargetAgent, "Active").First(&agent).Error; err != nil {
			json.NewEncoder(w).Encode(map[string]any{"status": "no_agent", "reason": "agent not found"})
			return
		}

		log.Printf("Guardian -> [%s] (%s): %s", agent.AgentID, agent.Role, req.Command)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "dispatched",
			"agent":  agent.AgentID,
			"role":   agent.Role,
		})
	})

	log.Printf("Guardian Node on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
