package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Tx struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}
type BatchReq struct {
	Transactions []Tx `json:"transactions"`
}

func main() {
	totalTx := 1000000
	if len(os.Args) > 1 {
		totalTx, _ = strconv.Atoi(os.Args[1])
	}
	urlBatch := "http://localhost:8091/api/v1/tx/batch"
	urlSeed := "http://localhost:8091/api/v1/account/seed"
	token := "bench0000000000000000000000000000000000000000000000000000000bench"
	license := "HORIZON-MASTER-001"
	batchSize := 5000
	numWorkers := 20

	client := &http.Client{Timeout: 60 * time.Second, Transport: &http.Transport{MaxIdleConns: 200, MaxIdleConnsPerHost: 200}}

	fmt.Println("⏳ در حال ساخت ۱۰۰ حساب فرستنده...")
	var wgSeed sync.WaitGroup
	for i := 1; i <= 100; i++ {
		wgSeed.Add(1)
		go func(id int) {
			defer wgSeed.Done()
			payload := fmt.Sprintf(`{"bank_id":"sender_%d","amount":1000000000}`, id)
			req, _ := http.NewRequest("POST", urlSeed, bytes.NewBufferString(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Admin-Token", token)
			client.Do(req)
		}(i)
	}
	wgSeed.Wait()
	fmt.Println("✅ حساب‌ها آماده شد.")

	fmt.Printf("🚀 شلیک %d تراکنش | Batch: %d | Workers: %d\n", totalTx, batchSize, numWorkers)
	
	var successTx int64
	start := time.Now()
	var wg sync.WaitGroup
	txChannel := make(chan int, totalTx/batchSize)

	for i := 0; i < totalTx/batchSize; i++ { txChannel <- i }
	close(txChannel)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range txChannel {
				var txs []Tx
				for i := 0; i < batchSize; i++ {
					txs = append(txs, Tx{From: fmt.Sprintf("sender_%d", rand.Intn(100)+1), To: "rcv", Amount: 1.0, Type: "transfer"})
				}
				payload, _ := json.Marshal(BatchReq{Transactions: txs})
				req, _ := http.NewRequest("POST", urlBatch, bytes.NewBuffer(payload))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Admin-Token", token)
				req.Header.Set("X-License-ID", license)
				resp, err := client.Do(req)
				if err == nil {
					io.ReadAll(resp.Body)
					resp.Body.Close()
					if resp.StatusCode == 200 { atomic.AddInt64(&successTx, int64(batchSize)) }
				}
			}
		}()
	}
	wg.Wait()
	
	fmt.Println("\n══════════ نتیجه تست ══════════")
	fmt.Printf("⏱  زمان کل:     %v\n", time.Since(start))
	fmt.Printf("✅ موفق:        %d تراکنش\n", successTx)
	fmt.Printf("⚡ TPS واقعی:   %.2f تراکنش بر ثانیه\n", float64(successTx)/time.Since(start).Seconds())
}
