package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type stats struct {
	ok      int64
	fail    int64
	totalNs int64
}

func main() {
	nodes := flag.Int("nodes", 1000, "iterations for 7-node test")
	batch := flag.Int("batch", 50000, "batch size for switch test")
	conc := flag.Int("c", 50, "concurrency")
	flag.Parse()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  HORIZON CORE — TPS BENCHMARK (Go, no-fork)")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// TEST 1: 7 nodes parallel
	fmt.Printf("[1] 7-Node Parallel Verify (n=%d, c=%d)\n", *nodes, *conc)
	runNodes(*nodes, *conc)

	// TEST 2: Switch tx/batch
	fmt.Printf("\n[2] Switch tx/batch (n=%d)\n", *batch)
	runSwitch(*batch)

	fmt.Println("\n═══════════════════════════════════════════════════════════")
}

func runNodes(n, conc int) {
	payload := `{"left":"A","right":"B","header":"H","message":"M","signature":"S","merkle_type":"binary","used":1,"volume":1000000,"issued_at":1,"expires_at":9999999999,"user_id":"u","product_id":"p","expected_user":"u","expected_product":"p"}`
	ports := []int{9001, 9002, 9003, 9004, 9005, 9006, 9007}

	var st stats
	client := &http.Client{Timeout: 5 * time.Second}
	jobs := make(chan int, conc)
	var wg sync.WaitGroup

	start := time.Now()
	for w := 0; w < conc; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				for _, p := range ports {
					url := fmt.Sprintf("http://localhost:%d/verify", p)
					t0 := time.Now()
					resp, err := client.Post(url, "application/json", bytes.NewReader([]byte(payload)))
					ns := time.Since(t0).Nanoseconds()
					atomic.AddInt64(&st.totalNs, ns)
					if err != nil {
						atomic.AddInt64(&st.fail, 1)
						continue
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					atomic.AddInt64(&st.ok, 1)
				}
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	total := n * len(ports)
	tps := float64(total) / elapsed.Seconds()
	avgMs := float64(st.totalNs) / float64(st.ok+st.fail) / 1e6

	fmt.Printf("    total req:    %d\n", total)
	fmt.Printf("    ok:           %d\n", st.ok)
	fmt.Printf("    fail:         %d\n", st.fail)
	fmt.Printf("    elapsed:      %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("    avg latency:  %.2f ms\n", avgMs)
	fmt.Printf("    ✅ CONSENSUS TPS: %.0f\n", tps)
}

func runSwitch(n int) {
	// setup
	adminHdr := map[string]string{"X-Admin-Token": adminToken, "Content-Type": "application/json"}
	post := func(url, body string) {
		req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte(body)))
		for k, v := range adminHdr {
			req.Header.Set(k, v)
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	post("http://localhost:8091/api/v1/key/setup", `{"private_key":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","password":"bench"}`)
	post("http://localhost:8091/api/v1/key/unlock", `{"password":"bench"}`)
	post("http://localhost:8091/api/v1/admin/license/issue", `{"license_id":"TPS-GO","user_id":"bench","volume":999999999,"duration":365}`)

	// build batch
	type tx struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int    `json:"amount"`
	}
	txs := make([]tx, n)
	for i := range txs {
		txs[i] = tx{From: "bench", To: fmt.Sprintf("r%d", i), Amount: 1}
	}
	body, _ := json.Marshal(map[string]interface{}{"session_token": "go-bench", "txs": txs})

	req, _ := http.NewRequest("POST", "http://localhost:8091/api/v1/tx/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Token", adminToken)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("    error: %v\n", err)
		return
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	tps := float64(n) / elapsed.Seconds()
	fmt.Printf("    batch size:   %d\n", n)
	fmt.Printf("    elapsed:      %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("    response:     %s\n", string(respBody)[:min(120, len(respBody))])
	fmt.Printf("    ✅ SWITCH TPS: %.0f\n", tps)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// adminToken is read from the environment.
// Fallback is provided for local testing only.
var adminToken = os.Getenv("ADMIN_TOKEN")

func init() {
	if adminToken == "" {
		adminToken = "12345678901234567890123456789012"
	}
}
