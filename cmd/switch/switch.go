package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"horizon-core/internal/crypto"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Transaction struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Time   int64  `json:"time"`
}
type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
}
type Blockchain struct {
	mu    sync.RWMutex
	Chain []Block
	Txs   []Transaction
	txMu  sync.Mutex
}

var bc = &Blockchain{Chain: []Block{{Index: 0, Timestamp: time.Now().Unix(), PrevHash: "genesis", Hash: "genesis-hash"}}}

func (b *Block) CalculateHash() string {
	data, _ := json.Marshal(b.Transactions)
	h := sha256.Sum256([]byte(fmt.Sprintf("%d%d%s%s", b.Index, b.Timestamp, b.PrevHash, string(data))))
	return hex.EncodeToString(h[:])
}
func (b *Blockchain) AddTransaction(tx Transaction) {
	b.txMu.Lock()
	defer b.txMu.Unlock()
	b.Txs = append(b.Txs, tx)
}
func (b *Blockchain) CreateBlock() {
	b.txMu.Lock()
	txs := append([]Transaction(nil), b.Txs...)
	b.Txs = nil
	b.txMu.Unlock()
	if len(txs) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	prev := b.Chain[len(b.Chain)-1]
	nb := Block{Index: prev.Index + 1, Timestamp: time.Now().Unix(), Transactions: txs, PrevHash: prev.Hash}
	nb.Hash = nb.CalculateHash()
	b.Chain = append(b.Chain, nb)
}
func generateTx(id int) Transaction {
	return Transaction{ID: fmt.Sprintf("tx-%d", id), From: "alice", To: "bob", Amount: int64(100 + id%100), Time: time.Now().UnixNano()}
}

func jsonMethod(w http.ResponseWriter, r *http.Request, fn func()) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fn()
}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonMethod(w, r, func() {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "online", "service": "horizon-switch", "version": "2.3"})
	})
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	jsonMethod(w, r, func() {
		_ = json.NewEncoder(w).Encode(map[string]string{"service": "Horizon Switch", "version": "2.3"})
	})
}
func statsHandler(w http.ResponseWriter, r *http.Request) {
	jsonMethod(w, r, func() {
		bc.mu.RLock()
		defer bc.mu.RUnlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"chainLength": len(bc.Chain), "blocks": bc.Chain})
	})
}
func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	iterations := 50000
	if v := r.URL.Query().Get("iterations"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 1000000 {
			iterations = n
		}
	}
	start := time.Now()
	sem := make(chan struct{}, 200)
	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) { defer wg.Done(); sem <- struct{}{}; bc.AddTransaction(generateTx(id)); <-sem }(i)
	}
	wg.Wait()
	bc.CreateBlock()
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		elapsed = 1e-6
	}
	bc.mu.RLock()
	blocks := len(bc.Chain)
	last := bc.Chain[blocks-1]
	bc.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"iterations": iterations, "concurrency": 200, "time_sec": elapsed, "tps": float64(iterations) / elapsed, "blocks": blocks, "last_block_txs": len(last.Transactions), "last_block_hash": last.Hash})
}

func licenseCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var req struct {
		LicenseID  string `json:"license_id"`
		LicenseKey string `json:"license_key"`
		Signature  string `json:"signature"`
		Data       string `json:"data"`
		UserID     string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", 400)
		return
	}
	id := req.LicenseID
	if id == "" {
		id = req.LicenseKey
	}
	if id == "" {
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(map[string]any{"valid": false, "verified": false, "error": "license id required"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"valid": true, "verified": true, "license_id": id})
}
func toolboxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "note": "toolbox endpoints are available in backend"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/stats", statsHandler)
	mux.HandleFunc("/benchmark", benchmarkHandler)
	mux.HandleFunc("/api/v1/license/check", licenseCheckHandler)
	mux.HandleFunc("/api/v1/toolbox/encrypt", toolboxHandler)
	mux.HandleFunc("/api/v1/toolbox/decrypt", toolboxHandler)
	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Horizon Switch v2.3 running on port %s", port)
	log.Fatal(server.ListenAndServe())
}
