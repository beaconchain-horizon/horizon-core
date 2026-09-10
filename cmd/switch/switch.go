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

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
}

type Transaction struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Time   int64  `json:"time"`
}

type Blockchain struct {
	Chain []Block `json:"chain"`
	mu    sync.Mutex
	txs   []Transaction
	txMu  sync.Mutex
}

var bc = &Blockchain{Chain: []Block{}}

func (b *Block) CalculateHash() string {
	record := fmt.Sprintf("%d%d%s%v", b.Index, b.Timestamp, b.PrevHash, b.Transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func NewBlock(prevBlock Block, txs []Transaction) Block {
	block := Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		PrevHash:     prevBlock.Hash,
	}
	block.Hash = block.CalculateHash()
	return block
}

func (bc *Blockchain) AddBlock(txs []Transaction) Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	var prevBlock Block
	if len(bc.Chain) == 0 {
		prevBlock = Block{Index: -1, Hash: "genesis"}
	} else {
		prevBlock = bc.Chain[len(bc.Chain)-1]
	}
	newBlock := NewBlock(prevBlock, txs)
	bc.Chain = append(bc.Chain, newBlock)
	return newBlock
}

func (bc *Blockchain) AddTransaction(tx Transaction) {
	bc.txMu.Lock()
	bc.txs = append(bc.txs, tx)
	bc.txMu.Unlock()
}

func (bc *Blockchain) GetPendingTxs() []Transaction {
	bc.txMu.Lock()
	defer bc.txMu.Unlock()
	txs := bc.txs
	bc.txs = []Transaction{}
	return txs
}

func generateTestTx(id int) Transaction {
	return Transaction{
		ID:     fmt.Sprintf("tx-%d", id),
		From:   "alice",
		To:     "bob",
		Amount: int64(100 + id%100),
		Time:   time.Now().Unix(),
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"service": "horizon-switch",
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

func benchmarkTPS(iterations int, concurrency int) float64 {
	var wg sync.WaitGroup
	start := time.Now()
	sem := make(chan struct{}, concurrency)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			tx := generateTestTx(id)
			bc.AddTransaction(tx)
			<-sem
		}(i)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Wait for all transactions to be added
	elapsed := time.Since(start).Seconds()

	txs := bc.GetPendingTxs()
	if len(txs) > 0 {
		bc.AddBlock(txs)
	}

	return float64(iterations) / elapsed
}

func benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	iterations := 10000
	concurrency := 200
	results := struct {
		Iterations  int     `json:"iterations"`
		Concurrency int     `json:"concurrency"`
		TPS         float64 `json:"tps"`
		TimeSec     float64 `json:"time_sec"`
		Blocks      int     `json:"blocks"`
	}{
		Iterations:  iterations,
		Concurrency: concurrency,
	}

	start := time.Now()
	tps := benchmarkTPS(iterations, concurrency)
	elapsed := time.Since(start).Seconds()

	results.TPS = tps
	results.TimeSec = elapsed
	results.Blocks = len(bc.Chain)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "Horizon Switch with Blockchain & TPS Benchmark",
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
	log.Printf("Horizon Switch running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
