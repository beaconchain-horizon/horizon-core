package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ConsensusTx struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	LicenseID string  `json:"license_id"`
	Type      string  `json:"type"`
	Payload   string  `json:"payload"`
}

type ConsensusLicense struct {
	LicenseID     string `json:"license_id"`
	UserID        string `json:"user_id"`
	ProductID     string `json:"product_id"`
	Volume        int64  `json:"volume"`
	Used          int64  `json:"used"`
	Duration      int64  `json:"duration"`
	IssuedAt      int64  `json:"issued_at"`
	ExpiresAt     int64  `json:"expires_at"`
	Status        string `json:"status"`
	Signature     string `json:"signature"`
	MerkleRoot    string `json:"merkle_root"`
	WOTSSignature string `json:"wots_signature"`
	WOTSPubKey    string `json:"wots_pubkey"`
}

type ConsensusPayload struct {
	BlockNumber uint64             `json:"block_number"`
	BlockHash   string             `json:"block_hash"`
	PrevHash    string             `json:"prev_hash"`
	MerkleRoot  string             `json:"merkle_root"`
	MerkleType  string             `json:"merkle_type"`
	Timestamp   int64              `json:"timestamp"`
	TxCount     int                `json:"tx_count"`
	Signature   string             `json:"signature"`
	SignerAddr  string             `json:"signer_addr"`
	BlockLicenseID string `json:"block_license_id"`
	Licenses    []ConsensusLicense `json:"licenses"`
	Txs         []ConsensusTx      `json:"txs"`
}

type ConsensusVote struct {
	Role   string `json:"role"`
	Vote   string `json:"vote"`
	Reason string `json:"reason"`
}

func RequestConsensus(payload ConsensusPayload) (bool, map[string]string, error) {
	ports := []int{9001, 9002, 9003, 9004, 9005, 9006, 9007, 9008}
	votes := make(map[string]string)
	
	client := &http.Client{Timeout: 5 * time.Second}
	body, _ := json.Marshal(payload)

	criticalAccept := 0
	majorityAccept := 0

	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%d/consensus", port)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := client.Do(req)
		if err != nil {
			votes[fmt.Sprintf(":%d", port)] = "ERROR: " + err.Error()
			continue
		}
		
		var v ConsensusVote
		json.NewDecoder(resp.Body).Decode(&v)
		resp.Body.Close()
		
		votes[v.Role] = v.Vote + " (" + v.Reason + ")"
		
		if v.Vote == "ACCEPT" {
			switch v.Role {
			case "Merkle", "Hash", "ECDSA", "Type":
				criticalAccept++
			case "Volume", "Time", "User":
				majorityAccept++
			}
		} else if v.Role == "AgentGuardian" {
		    return false, votes, fmt.Errorf("AgentGuardian vetoed: %s", v.Reason)
		}
	}

	if criticalAccept < 4 {
		return false, votes, fmt.Errorf("critical nodes failed: %d/4", criticalAccept)
	}
	if majorityAccept < 2 {
		return false, votes, fmt.Errorf("majority nodes failed: %d/2", majorityAccept)
	}

	return true, votes, nil
}
