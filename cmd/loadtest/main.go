package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080/api/v1/tx", "target URL")
	total := flag.Int("n", 10000, "total requests")
	conc := flag.Int("c", 100, "concurrency")
	from := flag.String("from", "bank_markazi", "from bank")
	to := flag.String("to", "bank_melli", "to bank")
	amount := flag.Float64("amount", 1, "amount")
	flag.Parse()

	payload := []byte(fmt.Sprintf(`{"from":"%s","to":"%s","amount":%f,"type":"transfer"}`,
		*from, *to, *amount))

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 2000,
			DisableKeepAlives:   false,
			MaxConnsPerHost:     2000,
		},
		Timeout: 30 * time.Second,
	}

	var wg sync.WaitGroup
	var success, fail atomic.Int64

	perWorker := *total / *conc
	remainder := *total % *conc

	fmt.Printf("Target: %s\n", *url)
	fmt.Printf("Total: %d | Concurrency: %d\n\n", *total, *conc)

	start := time.Now()
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		n := perWorker
		if i == 0 {
			n += remainder
		}
		go func(iterations int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				req, _ := http.NewRequest("POST", *url, bytes.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					fail.Add(1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode == 201 {
					success.Add(1)
				} else {
					fail.Add(1)
				}
			}
		}(n)
	}
	wg.Wait()
	elapsed := time.Since(start)

	tps := float64(*total) / elapsed.Seconds()
	fmt.Printf("=== RESULTS ===\n")
	fmt.Printf("Total requests: %d\n", *total)
	fmt.Printf("Success:        %d\n", success.Load())
	fmt.Printf("Failed:         %d\n", fail.Load())
	fmt.Printf("Time:           %.2fs\n", elapsed.Seconds())
	fmt.Printf("TPS:            %.0f tx/s\n", tps)
	fmt.Printf("Latency avg:    %.2f ms\n", elapsed.Seconds()*1000/float64(*total))
}
