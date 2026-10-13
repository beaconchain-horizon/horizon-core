// Command chainbench measures end-to-end blockchain TPS.
// It signs licenses and posts them to /api/v1/license/register,
// then measures how fast the chain persists them into blocks.
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

func main() {
	url := flag.String("url", "http://localhost:8080", "Backend URL")
	keyPath := flag.String("key", ".horizon-dev-key.pem", "signing key PEM")
	n := flag.Int("n", 5000, "total licenses")
	c := flag.Int("c", 50, "workers")
	flag.Parse()

	// Load key
	pb, err := os.ReadFile(*keyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read key:", err)
		os.Exit(1)
	}
	priv, err := hcrypto.PEMToPrivateKey(string(pb))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse key:", err)
		os.Exit(1)
	}
	var xb, yb [32]byte
	priv.PublicKey.X.FillBytes(xb[:])
	priv.PublicKey.Y.FillBytes(yb[:])
	pubHex := hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:])

	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println(" Horizon Chain — TPS Benchmark")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf(" backend:    %s\n", *url)
	fmt.Printf(" total:      %d licenses\n", *n)
	fmt.Printf(" workers:    %d\n", *c)
	fmt.Println()

	// ═══ Phase 1: Pre-sign ═══
	t1 := time.Now()
	payloads := make([][]byte, *n)
	w := 8
	chunk := (*n + w - 1) / w
	expireUnix := time.Now().Add(365 * 24 * time.Hour).Unix()
	expireISO := time.Unix(expireUnix, 0).UTC().Format(time.RFC3339)

	var wg1 sync.WaitGroup
	for k := 0; k < w; k++ {
		s, e := k*chunk, (k+1)*chunk
		if e > *n {
			e = *n
		}
		if s >= *n {
			break
		}
		wg1.Add(1)
		go func(start, end int) {
			defer wg1.Done()
			for i := start; i < end; i++ {
				licID := fmt.Sprintf("HORIZON-BENCH-%08d-%04x", i, randUint16())
				data := licID + "|horizon-core|bench|1000||active|" + expireISO
				sig, err := hcrypto.SignDataRS([]byte(data), priv)
				if err != nil {
					continue
				}
				body, _ := json.Marshal(map[string]any{
					"license_id":  licID,
					"user_id":     "bench",
					"product_id":  "horizon-core",
					"volume":      1000,
					"signature":   hex.EncodeToString(sig),
					"signed_data": data,
					"public_key":  pubHex,
				})
				payloads[i] = body
			}
		}(s, e)
	}
	wg1.Wait()
	fmt.Printf("─── Phase 1: signed %d in %s\n", *n, time.Since(t1).Round(time.Millisecond))

	// ═══ Phase 2: Submit ═══
	fmt.Println("─── Phase 2: submitting")
	transport := &http.Transport{
		MaxIdleConns:        *c * 2,
		MaxIdleConnsPerHost: *c * 2,
		MaxConnsPerHost:     *c * 2,
	}
	client := &http.Client{Transport: transport, Timeout: 60 * time.Second}

	jobs := make(chan int, *c*2)
	var ok, errCnt int64
	var wg2 sync.WaitGroup
	t2 := time.Now()

	for k := 0; k < *c; k++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for idx := range jobs {
				req, _ := http.NewRequest("POST", *url+"/api/v1/license/register", bytes.NewReader(payloads[idx]))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&errCnt, 1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&ok, 1)
				} else {
					atomic.AddInt64(&errCnt, 1)
					body, _ := io.ReadAll(resp.Body)
					atomic.AddInt64(&errCnt, 1)
					if atomic.LoadInt64(&errCnt) <= 5 {
						fmt.Printf("    ERR status=%d body=%s\n", resp.StatusCode, string(body))
					}
				}
			}
		}()
	}
	for i := 0; i < *n; i++ {
		jobs <- i
	}
	close(jobs)
	wg2.Wait()
	elapsed := time.Since(t2)

	httpTPS := float64(ok) / elapsed.Seconds()
	fmt.Printf("    http:    OK=%d ERR=%d | %.2fs | ★HTTP-TPS=%.0f\n",
		ok, errCnt, elapsed.Seconds(), httpTPS)

	// ═══ Phase 3: Wait for chain to drain ═══
	fmt.Println("─── Phase 3: waiting for chain to persist into blocks")
	var finalHeight, finalTx float64
	startWait := time.Now()
	for i := 0; i < 120; i++ {
		time.Sleep(1 * time.Second)
		info, err := getJSON(*url + "/api/v1/chain/info")
		if err != nil {
			continue
		}
		pending, _ := getJSON(*url + "/api/v1/chain/pending")
		p, _ := pending["pending"].(float64)
		h, _ := info["height"].(float64)
		tx, _ := info["tx_total"].(float64)
		finalHeight, finalTx = h, tx
		if i%3 == 0 {
			fmt.Printf("    t=%2ds  height=%.0f  tx=%.0f  pending=%.0f\n",
				i, h, tx, p)
		}
		if p == 0 && tx >= float64(*n) {
			break
		}
	}
	waitTime := time.Since(startWait)
	chainTPS := finalTx / (elapsed + waitTime).Seconds()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println(" RESULTS")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf(" HTTP TPS (accept rate):    %.0f tx/s\n", httpTPS)
	fmt.Printf(" Chain TPS (persist rate):  %.0f tx/s\n", chainTPS)
	fmt.Printf(" total elapsed:             %s\n", (elapsed + waitTime).Round(time.Second))
	fmt.Printf(" final height:              %.0f blocks\n", finalHeight)
	fmt.Printf(" total tx on chain:         %.0f\n", finalTx)
	fmt.Printf(" avg block size:            %.1f tx/block\n", finalTx/finalHeight)
	fmt.Println("═══════════════════════════════════════════════")
}

func randUint16() uint16 {
	b := make([]byte, 2)
	rand.Read(b)
	return uint16(b[0])<<8 | uint16(b[1])
}

func getJSON(url string) (map[string]any, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
