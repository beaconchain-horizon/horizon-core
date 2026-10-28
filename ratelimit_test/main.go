package main

import (
	"fmt"
	"net/http"
	"sync"
)

const (
	totalReqs = 300
	concurrency = 50
	targetURL = "http://localhost:8080/api/v1/health"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	statusCodes := make(map[int]int)

	jobs := make(chan int, totalReqs)
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: concurrency}}

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				resp, err := client.Get(targetURL)
				if err != nil {
					continue
				}
				mu.Lock()
				statusCodes[resp.StatusCode]++
				mu.Unlock()
				resp.Body.Close()
			}
		}()
	}

	for i := 0; i < totalReqs; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	fmt.Println("=== RATE LIMIT TEST RESULTS ===")
	fmt.Printf("Total Requests: %d\n", totalReqs)
	for code, count := range statusCodes {
		fmt.Printf("HTTP %d: %d\n", code, count)
	}
	
	if statusCodes[429] > 0 {
		fmt.Println("✅ RESULT: Rate Limiting is ACTIVE (Blocked with 429)")
	} else {
		fmt.Println("⚠️ RESULT: No 429 received. (Note: /health might be exempt from rate limiting for monitoring).")
	}
}
