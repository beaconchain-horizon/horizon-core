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
	SessionToken string `json:"session_token"`
	Txs          []Tx   `json:"txs"`
}

const (
	urlBatch  = "http://localhost:8091/api/v1/tx/batch"
	urlSeed   = "http://localhost:8091/api/v1/account/seed"
	urlSess   = "http://localhost:8091/api/v1/signing/session"
	token     = "bench0000000000000000000000000000000000000000000000000000000bench"
	license   = "HORIZON-MASTER-001"
)

var client = &http.Client{Timeout: 60 * time.Second, Transport: &http.Transport{MaxIdleConns: 500, MaxIdleConnsPerHost: 500}}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: loadtest [raw|full] [count]")
		os.Exit(1)
	}
	mode := os.Args[1]
	count, _ := strconv.Atoi(os.Args[2])

	seedAccounts(100)

	start := time.Now()
	var success int64

	if mode == "raw" {
		success = runRaw(count)
	} else {
		success = runFull(count)
	}

	duration := time.Since(start)
	fmt.Println("\n══════════ نتیجه تست ══════════")
	fmt.Printf("⏱  زمان کل:     %v\n", duration)
	fmt.Printf("✅ موفق:        %d تراکنش\n", success)
	fmt.Printf("⚡ TPS واقعی:   %.2f تراکنش بر ثانیه\n", float64(success)/duration.Seconds())
}

func seedAccounts(n int) {
	fmt.Printf("⏳ ساخت %d حساب فرستنده...\n", n)
	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			payload := fmt.Sprintf(`{"bank_id":"sender_%d","amount":1000000000}`, id)
			req, _ := http.NewRequest("POST", urlSeed, bytes.NewBufferString(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Admin-Token", token)
			client.Do(req)
		}(i)
	}
	wg.Wait()
	fmt.Println("✅ حساب‌ها آماده شد.")
}

func runRaw(total int) int64 {
	fmt.Printf("🚀 حالت RAW: شلیک %d تراکنش (Batch)...\n", total)
	batchSize := 10000
	numWorkers := 20
	var success int64
	var wg sync.WaitGroup
	ch := make(chan int, total/batchSize)

	for i := 0; i < total/batchSize; i++ { ch <- i }
	close(ch)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				// ۱. گرفتن Session Token
				reqSess, _ := http.NewRequest("POST", urlSess, nil)
				reqSess.Header.Set("X-Admin-Token", token)
				respSess, err := client.Do(reqSess)
				if err != nil { continue }
				bodySess, _ := io.ReadAll(respSess.Body)
				respSess.Body.Close()
				
				var sessData map[string]string
				json.Unmarshal(bodySess, &sessData)
				sessToken := sessData["session_token"]

				// ۲. ساخت تراکنش‌ها
				var txs []Tx
				for i := 0; i < batchSize; i++ {
					txs = append(txs, Tx{From: fmt.Sprintf("sender_%d", rand.Intn(100)+1), To: "rcv", Amount: 1.0, Type: "transfer"})
				}
				
				// ۳. ارسال Batch با فرمت درست
				payload, _ := json.Marshal(BatchReq{SessionToken: sessToken, Txs: txs})
				req, _ := http.NewRequest("POST", urlBatch, bytes.NewBuffer(payload))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Admin-Token", token)
				req.Header.Set("X-License-ID", license)

				resp, err := client.Do(req)
				if err == nil {
					io.ReadAll(resp.Body)
					resp.Body.Close()
					if resp.StatusCode == 200 || resp.StatusCode == 201 { atomic.AddInt64(&success, int64(batchSize)) } else { if atomic.LoadInt64(&success) == 0 { b, _ := io.ReadAll(resp.Body); fmt.Printf("BATCH ERROR %d: %s\n", resp.StatusCode, string(b)) } }
				}
			}
		}()
	}
	wg.Wait()
	return success
}

func runFull(total int) int64 {
	return 0 // فعلاً نیاز نیست
}
