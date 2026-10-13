package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	hcrypto "horizon-core/internal/crypto"
)

type SignableReading struct {
	SensorID  uint    `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
}

type Reading struct {
	SensorID  uint    `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

func main() {
	url := flag.String("url", "", "target URL")
	kp := flag.String("key", "", "path to private key PEM")
	sen := flag.Int("sensor", 1, "sensor id")
	n := flag.Int("n", 1000, "total readings")
	c := flag.Int("c", 50, "concurrency")
	bs := flag.Int("bs", 100, "batch size")
	tok := flag.String("token", "", "X-Admin-Key token")
	dbg := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	if *kp == "" {
		fmt.Fprintln(os.Stderr, "missing --key")
		os.Exit(2)
	}
	pb, readErr := os.ReadFile(*kp)
	if readErr != nil {
		fmt.Fprintln(os.Stderr, "read key:", readErr)
		os.Exit(1)
	}
	pk, keyErr := hcrypto.PEMToPrivateKey(string(pb))
	if keyErr != nil {
		fmt.Fprintln(os.Stderr, "parse key:", keyErr)
		os.Exit(1)
	}

	batches := (*n + *bs - 1) / *bs
	payloads := make([][]byte, batches)

	fmt.Printf("Pre-signing %d batches (%d readings)...\n", batches, *n)

	for i := 0; i < batches; i++ {
		arr := make([]Reading, 0, *bs)
		for j := 0; j < *bs; j++ {
			sr := SignableReading{
				SensorID:  uint(*sen),
				Value:     float64(30 + j%40),
				Nonce:     hex.EncodeToString(randB(16)),
				Timestamp: time.Now().Unix(),
			}
			canonical, _ := json.Marshal(sr)
			signatureHex, signErr := hcrypto.SignData(canonical, pk)
			if signErr != nil {
				fmt.Fprintln(os.Stderr, "sign reading:", signErr)
				os.Exit(1)
			}
			arr = append(arr, Reading{
				SensorID:  sr.SensorID,
				Value:     sr.Value,
				Nonce:     sr.Nonce,
				Timestamp: sr.Timestamp,
				Signature: signatureHex,
			})
		}
		payloads[i], _ = json.Marshal(arr)
		if *dbg && i == 0 {
			fmt.Printf("sample payload: %s\n\n", string(payloads[i]))
		}
	}

	if *dbg {
		fmt.Println("[DEBUG mode — exiting]")
		os.Exit(0)
	}

	cl := &http.Client{Transport: &http.Transport{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 500,
		MaxConnsPerHost:     500,
	}, Timeout: 60 * time.Second}

	var ok, errCnt int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, *c)
	t0 := time.Now()

	for i := 0; i < batches; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(x int) {
			defer wg.Done()
			defer func() { <-sem }()
			q, _ := http.NewRequest("POST", *url, bytes.NewReader(payloads[x]))
			q.Header.Set("Content-Type", "application/json")
			if *tok != "" {
				q.Header.Set("X-Admin-Key", *tok)
			}
			rs, e := cl.Do(q)
			if e != nil {
				atomic.AddInt64(&errCnt, 1)
				return
			}
			body, _ := io.ReadAll(rs.Body)
			rs.Body.Close()
			if rs.StatusCode >= 200 && rs.StatusCode < 300 {
				atomic.AddInt64(&ok, 1)
			} else {
				atomic.AddInt64(&errCnt, 1)
				if *dbg {
					fmt.Printf("status=%d body=%s\n", rs.StatusCode, string(body))
				}
			}
		}(i)
	}
	wg.Wait()
	el := time.Since(t0)
	tps := float64(*n) / el.Seconds()
	fmt.Printf("\n>>> %d readings | c=%d bs=%d | %.2fs | OK=%d ERR=%d | ★TPS=%.0f\n",
		*n, *c, *bs, el.Seconds(), ok, errCnt, tps)
}

func randB(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
