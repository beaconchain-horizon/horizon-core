package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type VR struct {
	Role   string `json:"role"`
	Vote   string `json:"vote"`
	Reason string `json:"reason"`
}

type P struct {
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	TxCount     int    `json:"tx_count"`
}

var (
	blocks int64
	txs    int64
	mu     sync.Mutex
)

func main() {
	port := os.Getenv("AGENT_NODE_PORT")
	if port == "" { port = "9008" }
	http.HandleFunc("/health", health)
	http.HandleFunc("/consensus", consensus)
	http.HandleFunc("/stats", statsHandler)
	log.Printf("Agent Guardian on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"role": "AgentGuardian", "status": "online"})
}

func consensus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var p P
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &p); err != nil {
		json.NewEncoder(w).Encode(VR{Role: "AgentGuardian", Vote: "REJECT", Reason: "bad json"})
		return
	}
	mu.Lock()
	blocks++
	txs += int64(p.TxCount)
	b, t := blocks, txs
	mu.Unlock()
	if p.TxCount > 1000000 {
		json.NewEncoder(w).Encode(VR{Role: "AgentGuardian", Vote: "REJECT", Reason: "abnormal"})
		return
	}
	json.NewEncoder(w).Encode(VR{Role: "AgentGuardian", Vote: "ACCEPT", Reason: fmt.Sprintf("ok blocks=%d txs=%d", b, t)})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{"role": "AgentGuardian", "blocks": blocks, "txs": txs})
}
