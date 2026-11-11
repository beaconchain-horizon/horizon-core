package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	totalTxs  int
	workers   int
	adminTok  = "bench0000000000000000000000000000000000000000000000000000000bench"
	licID     string
	targetURL = "http://localhost:8091/api/v1/tx"
	
	// آمار
	successCount int64
	failCount    int64
	minLatency   int64 = 999999999999
	maxLatency   int64 = 0
	totalLatency int64 = 0
)

func main() {
	flag.IntVar(&totalTxs, "n", 10000, "Total transactions")
	flag.IntVar(&workers, "w", 100, "Number of parallel workers")
	flag.StringVar(&licID, "lic", "", "License ID")
	flag.Parse()

	if licID == "" {
		fmt.Println("❌ Error: License ID is required (-lic <ID>)")
		return
	}

	fmt.Printf("🔥 Starting %d Txs with %d Workers...\n", totalTxs, workers)

	jobs := make(chan int, totalTxs)
	var wg sync.WaitGroup

	startTime := time.Now()

	// راه‌اندازی ورکرها
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}
			for range jobs {
				executeTx(client)
			}
		}()
	}

	// پر کردن صف کار
	for i := 0; i < totalTxs; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	duration := time.Since(startTime)

	// محاسبات نهایی
	avgLatency := float64(totalLatency) / float64(successCount+failCount) / 1000.0 // to milliseconds
	
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Printf("⏱  Total Time:     %v\n", duration)
	fmt.Printf("✅ Success:        %d\n", successCount)
	fmt.Printf("❌ Failed:         %d\n", failCount)
	fmt.Printf("⚡ TPS:            %.0f\n", float64(totalTxs)/duration.Seconds())
	fmt.Println("─── Latency (ms) ───")
	fmt.Printf("📉 Min:            %.3f ms\n", float64(minLatency)/1000.0)
	fmt.Printf("📊 Avg:            %.3f ms\n", avgLatency)
	fmt.Printf("📈 Max:            %.3f ms\n", float64(maxLatency)/1000.0)
	fmt.Println("═══════════════════════════════════════════")
}

func executeTx(client *http.Client) {
	start := time.Now()

	// 1. دریافت Session Token
	reqSess, _ := http.NewRequest("POST", "http://localhost:8091/api/v1/signing/session", nil)
	reqSess.Header.Set("X-Admin-Token", adminTok)
	respSess, err := client.Do(reqSess)
	if err != nil {
		recordFail(start)
		return
	}
	var sessData map[string]interface{}
	json.NewDecoder(respSess.Body).Decode(&sessData)
	respSess.Body.Close()
	
	sessTok, ok := sessData["session_token"].(string)
	if !ok || sessTok == "" {
		recordFail(start)
		return
	}

	// 2. ارسال تراکنش
	payload := map[string]interface{}{
		"from": "sender_1", "to": "rcv", "amount": 1, "type": "loadtest",
	}
	body, _ := json.Marshal(payload)
	reqTx, _ := http.NewRequest("POST", targetURL, bytes.NewBuffer(body))
	reqTx.Header.Set("Content-Type", "application/json")
	reqTx.Header.Set("X-Admin-Token", adminTok)
	reqTx.Header.Set("X-License-ID", licID)
	reqTx.Header.Set("X-Session-Token", sessTok)

	respTx, err := client.Do(reqTx)
	if err != nil {
		recordFail(start)
		return
	}
	defer respTx.Body.Close()
	io.Copy(io.Discard, respTx.Body)

	if respTx.StatusCode == http.StatusCreated || respTx.StatusCode == http.StatusOK {
		atomic.AddInt64(&successCount, 1)
	} else {
		atomic.AddInt64(&failCount, 1)
	}
	recordLatency(start)
}

func recordFail(start time.Time) {
	atomic.AddInt64(&failCount, 1)
	recordLatency(start)
}

func recordLatency(start time.Time) {
	lat := time.Since(start).Microseconds()
	atomic.AddInt64(&totalLatency, lat)
	for {
		currentMin := atomic.LoadInt64(&minLatency)
		if lat >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, lat) {
			break
		}
	}
	for {
		currentMax := atomic.LoadInt64(&maxLatency)
		if lat <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, lat) {
			break
		}
	}
}
