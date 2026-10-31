package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"horizon-core/internal/crypto"
)

type AgentVerifyReq struct {
	AgentID    string `json:"agent_id"`
	LicenseID  string `json:"license_id"`
	PublicKey  string `json:"public_key"`
	Signature  string `json:"signature"`
	HardwareID string `json:"hardware_id"`
}

func main() {
	port := os.Getenv("AGENT_NODE_PORT")
	if port == "" {
		port = "9008"
	}

	http.HandleFunc("/inspect-agent", func(w http.ResponseWriter, r *http.Request) {
		var req AgentVerifyReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}

		pub, err := crypto.Secp256k1PublicKeyFromHex(req.PublicKey)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"status": "rejected", "reason": "invalid public key"})
			return
		}

		payload := fmt.Sprintf("%s|%s|%s", req.AgentID, req.LicenseID, req.HardwareID)
		ok, _ := crypto.VerifyData(pub, []byte(payload), req.Signature)

		if ok {
			json.NewEncoder(w).Encode(map[string]any{"status": "verified", "agent_id": req.AgentID})
		} else {
			json.NewEncoder(w).Encode(map[string]any{"status": "rejected", "reason": "signature mismatch"})
		}
	})

	log.Printf("👼 Agent Inspector Node on port %s", port)
	http.ListenAndServe(":"+port, nil)
}
