package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	mu    sync.Mutex
	Chain []Block
	Txs   []Transaction
	txMu  sync.Mutex
}

var bc = &Blockchain{
	Chain: []Block{{
		Index:     0,
		Timestamp: time.Now().Unix(),
		PrevHash:  "genesis",
		Hash:      "genesis-hash",
	}},
}

func (b *Block) CalculateHash() string {
	data, _ := json.Marshal(b.Transactions)
	record := fmt.Sprintf("%d%d%s%s", b.Index, b.Timestamp, b.PrevHash, string(data))
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.txMu.Lock()
	bc.Txs = append(bc.Txs, tx)
	bc.txMu.Unlock()
}

func (bc *Blockchain) CreateBlock() {
	bc.txMu.Lock()
	txs := bc.Txs
	bc.Txs = []Transaction{}
	bc.txMu.Unlock()

	if len(txs) == 0 {
		return
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	prev := bc.Chain[len(bc.Chain)-1]
	newBlock := Block{
		Index:        prev.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		PrevHash:     prev.Hash,
	}
	newBlock.Hash = newBlock.CalculateHash()
	bc.Chain = append(bc.Chain, newBlock)
}

func generateTx(id int) Transaction {
	return Transaction{
		ID:     fmt.Sprintf("tx-%d", id),
		From:   "alice",
		To:     "bob",
		Amount: int64(100 + id%100),
		Time:   time.Now().UnixNano(),
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"service": "horizon-switch",
		"version": "2.2",
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	bc.mu.Lock()
	defer bc.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chainLength": len(bc.Chain),
		"blocks":      bc.Chain,
	})
}

func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	iterations := 50000
	concurrency := 200

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	start := time.Now()

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			tx := generateTx(id)
			bc.AddTransaction(tx)
			<-sem
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start).Seconds()

	bc.CreateBlock()

	tps := float64(iterations) / elapsed

	bc.mu.Lock()
	blocks := len(bc.Chain)
	lastBlock := bc.Chain[blocks-1]
	bc.mu.Unlock()

	result := map[string]interface{}{
		"iterations":      iterations,
		"concurrency":     concurrency,
		"time_sec":        elapsed,
		"tps":             tps,
		"blocks":          blocks,
		"last_block_txs":  len(lastBlock.Transactions),
		"last_block_hash": lastBlock.Hash,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "Horizon Switch",
		"version": "2.2",
	})
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/benchmark", benchmarkHandler)

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Horizon Switch v2.2 running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
