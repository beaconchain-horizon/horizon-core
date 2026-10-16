package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AngelPayload struct {
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

type AngelResp struct {
	Role     string `json:"role"`
	Vote     bool   `json:"vote"`
	Equation string `json:"equation"`
}

var angelCritical = []int{9001, 9002, 9003, 9004}
var angelImportant = []int{9005, 9006, 9007}

func callAngelNode(port int, payload []byte) AngelResp {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/verify", port)
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return AngelResp{Role: "unknown", Vote: false, Equation: "connection error"}
	}
	defer resp.Body.Close()
	var r AngelResp
	json.NewDecoder(resp.Body).Decode(&r)
	return r
}

// RunAngelConsensus sends the payload to all 7 nodes and returns
// (critYes, critTotal, impYes, impTotal, passed).
func RunAngelConsensus(payload AngelPayload, log *gin.Context) (int, int, int, int, bool) {
	body, _ := json.Marshal(payload)

	critYes, critTotal := 0, 0
	impYes, impTotal := 0, 0

	for _, port := range angelCritical {
		critTotal++
		r := callAngelNode(port, body)
		if r.Vote {
			critYes++
		}
	}
	for _, port := range angelImportant {
		impTotal++
		r := callAngelNode(port, body)
		if r.Vote {
			impYes++
		}
	}

	passed := critYes == critTotal && impYes >= 2
	return critYes, critTotal, impYes, impTotal, passed
}
