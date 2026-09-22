package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type sigASN struct{ R, S *big.Int }

type SignableReading struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
}

type Reading struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

func main() {
	url := flag.String("url", "", "")
	kp := flag.String("key", "", "")
	sen := flag.String("sensor", "", "")
	n := flag.Int("n", 1000, "")
	c := flag.Int("c", 50, "")
	bs := flag.Int("bs", 100, "")
	dbg := flag.Bool("debug", false, "")
	flag.Parse()

	pb, _ := os.ReadFile(*kp)
	b, _ := pem.Decode(pb)
	pk, _ := x509.ParseECPrivateKey(b.Bytes)

	batches := (*n + *bs - 1) / *bs
	payloads := make([][]byte, batches)

	fmt.Printf("Pre-signing %d batches (%d readings)...\n", batches, *n)

	for i := 0; i < batches; i++ {
		arr := make([]Reading, 0, *bs)
		for j := 0; j < *bs; j++ {
			sr := SignableReading{
				SensorID:  *sen,
				Value:     float64(30 + j%40),
				Nonce:     hex.EncodeToString(randB(16)),
				Timestamp: time.Now().Unix(),
			}
			canonical, _ := json.Marshal(sr)
			h := sha256.Sum256(canonical)
			rr, ss, _ := ecdsa.Sign(rand.Reader, pk, h[:])
			d, _ := asn1.Marshal(sigASN{rr, ss})

			if *dbg && i == 0 && j == 0 {
				fmt.Printf("\n[DEBUG — first payload]\n")
				fmt.Printf("canonical JSON: %s\n", string(canonical))
				fmt.Printf("canonical hex:  %s\n", hex.EncodeToString(canonical))
				fmt.Printf("sha256 hex:     %s\n", hex.EncodeToString(h[:]))
				fmt.Printf("signature hex:  %s\n", hex.EncodeToString(d))
			}

			arr = append(arr, Reading{
				SensorID:  sr.SensorID,
				Value:     sr.Value,
				Nonce:     sr.Nonce,
				Timestamp: sr.Timestamp,
				Signature: hex.EncodeToString(d),
			})
		}
		payloads[i], _ = json.Marshal(arr)
		if *dbg && i == 0 {
			fmt.Printf("sent JSON:      %s\n\n", string(payloads[i]))
		}
	}

	if *dbg {
		fmt.Println("[DEBUG mode — only first batch generated, exiting]")
		os.Exit(0)
	}

	cl := &http.Client{Transport: &http.Transport{
		MaxIdleConns: 500, MaxIdleConnsPerHost: 500, MaxConnsPerHost: 500,
	}, Timeout: 60 * time.Second}

	var ok, err int64
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
			rs, e := cl.Do(q)
			if e != nil {
				atomic.AddInt64(&err, 1)
				return
			}
			io.Copy(io.Discard, rs.Body)
			rs.Body.Close()
			if rs.StatusCode == 200 {
				atomic.AddInt64(&ok, 1)
			} else {
				atomic.AddInt64(&err, 1)
			}
		}(i)
	}
	wg.Wait()
	el := time.Since(t0)
	tps := float64(*n) / el.Seconds()
	fmt.Printf("\n>>> %d readings | c=%d bs=%d | %.2fs | OK=%d ERR=%d | ★TPS=%.0f\n",
		*n, *c, *bs, el.Seconds(), ok, err, tps)
}

func randB(n int) []byte { b := make([]byte, n); rand.Read(b); return b }
