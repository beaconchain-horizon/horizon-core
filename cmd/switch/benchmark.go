package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Global sink to prevent compiler optimization (Dead Code Elimination)
var benchmarkSink string
var benchmarkSinkMu sync.Mutex

// benchWorker: real cryptographic work
func benchWorker(priv *ecdsa.PrivateKey, iterations int) (string, int64) {
	var localSink string
	hasher := sha256.New()
	start := time.Now()

	for i := 0; i < iterations; i++ {
		// 1. Build a message
		msg := []byte(fmt.Sprintf("tx-%d-%d", i, start.UnixNano()))

		// 2. Hash the message
		hasher.Reset()
		hasher.Write(msg)
		digest := hasher.Sum(nil)

		// 3. Sign with ECDSA P-256 (real signature)
		r, s, err := ecdsa.Sign(rand.Reader, priv, digest)
		if err != nil {
			continue
		}

		// 4. Build Merkle-style hash: sha256(digest || r || s)
		rBytes := r.Bytes()
		sBytes := s.Bytes()
		merkleInput := append(digest, rBytes...)
		merkleInput = append(merkleInput, sBytes...)
		merkleHash := sha256.Sum256(merkleInput)

		// 5. Accumulate into localSink so compiler can't remove
		localSink = hex.EncodeToString(merkleHash[:8])
	}

	elapsed := time.Since(start).Nanoseconds()
	return localSink, elapsed
}

// benchmarkHandler: HTTP endpoint
// GET /benchmark?iterations=5000&concurrency=4
func benchmarkHandler(c *gin.Context) {
	iterStr := c.DefaultQuery("iterations", "5000")
	concStr := c.DefaultQuery("concurrency", "1")

	iterations, err := strconv.Atoi(iterStr)
	if err != nil || iterations < 1 || iterations > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "iterations must be 1..1000000"})
		return
	}
	concurrency, err := strconv.Atoi(concStr)
	if err != nil || concurrency < 1 || concurrency > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "concurrency must be 1..64"})
		return
	}

	// Generate one ECDSA key for all workers
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "keygen failed"})
		return
	}

	// Warm-up (2 seconds or 10% of iterations, whichever is smaller)
	warmupIters := iterations / 10
	if warmupIters < 100 {
		warmupIters = 100
	}
	if warmupIters > 10000 {
		warmupIters = 10000
	}
	benchWorker(priv, warmupIters)

	// Real test
	perWorker := iterations / concurrency
	if perWorker < 1 {
		perWorker = 1
	}

	var wg sync.WaitGroup
	var totalNs atomic.Int64
	var totalOps atomic.Int64

	startWall := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sink, ns := benchWorker(priv, perWorker)
			benchmarkSinkMu.Lock()
			benchmarkSink = sink
			benchmarkSinkMu.Unlock()
			totalNs.Add(ns)
			totalOps.Add(int64(perWorker))
		}()
	}
	wg.Wait()

	wallNs := time.Since(startWall).Nanoseconds()
	wallSec := float64(wallNs) / 1e9

	tps := float64(totalOps.Load()) / wallSec

	c.JSON(http.StatusOK, gin.H{
		"iterations":   totalOps.Load(),
		"concurrency":  concurrency,
		"wall_sec":     fmt.Sprintf("%.4f", wallSec),
		"tps":          fmt.Sprintf("%.2f", tps),
		"cpu_cores":    runtime.NumCPU(),
		"go_version":   runtime.Version(),
		"sink_prefix":  benchmarkSink[:min(16, len(benchmarkSink))],
		"note":         "real ECDSA P-256 + SHA-256 + Merkle hash",
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
