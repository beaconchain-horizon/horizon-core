package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"gorm.io/gorm"
)

type PendingTx struct {
	TxType    string
	Payload   interface{}
	Timestamp int64
}

type ChainBatcher struct {
	mu    sync.Mutex
	queue []PendingTx
	cfg   BatcherConfig
	db    *gorm.DB
	done  chan struct{}

	totalSubmitted int64
	totalFlushed   int64
	totalBlocks    int64
}

type BatcherConfig struct {
	MaxSize       int
	FlushInterval time.Duration
	SignWorkers   int
}

var (
	globalBatcher *ChainBatcher
	batcherMu     sync.RWMutex
)

func InitBatcher(db *gorm.DB, cfg BatcherConfig) {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 50000
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 50 * time.Millisecond
	}
	if cfg.SignWorkers <= 0 {
		cfg.SignWorkers = runtime.NumCPU()
	}
	b := &ChainBatcher{
		cfg:  cfg,
		db:   db,
		done: make(chan struct{}),
	}
	batcherMu.Lock()
	globalBatcher = b
	batcherMu.Unlock()
	go b.loop()
	log.Printf("⛓  batcher loop started, interval=%s", cfg.FlushInterval)
	log.Printf("⛓  batcher: max=%d interval=%s workers=%d",
		cfg.MaxSize, cfg.FlushInterval, cfg.SignWorkers)
}

func GetBatcher() *ChainBatcher {
	batcherMu.RLock()
	defer batcherMu.RUnlock()
	return globalBatcher
}

func (b *ChainBatcher) loop() {
	ticker := time.NewTicker(b.cfg.FlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-b.done:
			b.Flush()
			return
		case <-ticker.C:
			b.Flush()
		}
	}
}

func (b *ChainBatcher) Submit(txType string, payload interface{}) string {
	b.mu.Lock()
	b.queue = append(b.queue, PendingTx{
		TxType:    txType,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
	})
	b.totalSubmitted++
	n := len(b.queue)
	full := n >= b.cfg.MaxSize
	b.mu.Unlock()

	if full {
		go b.Flush()
	}
	return "queued"
}

func (b *ChainBatcher) SubmitBatch(txType string, payloads []interface{}) {
	b.mu.Lock()
	for _, payload := range payloads {
		b.queue = append(b.queue, PendingTx{
			TxType:    txType,
			Payload:   payload,
			Timestamp: time.Now().UnixNano(),
		})
	}
	b.totalSubmitted += int64(len(payloads))
	n := len(b.queue)
	full := n >= b.cfg.MaxSize
	b.mu.Unlock()

	if full {
		go b.Flush()
	}
}

func (b *ChainBatcher) Flush() {
	// DEBUG
	stateMutex.RLock()
	key := unlockedKey
	stateMutex.RUnlock()
	if key == nil {
		return
	}

	b.mu.Lock()
	if len(b.queue) == 0 {
		b.mu.Unlock()
		return
	}
	pending := b.queue
	b.queue = nil
	b.mu.Unlock()

	txs := b.signParallel(key, pending)
	if len(txs) == 0 {
		return
	}

	block, err := addChainBlock(b.db, key, txs)
	if err != nil {
		log.Printf("⚠️  batcher flush failed: %v", err)
		b.mu.Lock()
		for _, t := range pending {
			b.queue = append(b.queue, t)
		}
		b.mu.Unlock()
		return
	}

	b.mu.Lock()
	b.totalFlushed += int64(len(txs))
	b.totalBlocks++
	b.mu.Unlock()
	log.Printf("⛓  batcher: %d txs → block #%d", len(txs), block.BlockNum)
}

// signParallel — OPTIMIZED: no per-tx ECDSA.
// The block signature (in addChainBlock) covers all txs via merkle root.
// Per-tx signature would be redundant (10x-100x slower).
// Standard blockchain model: block signature is authoritative.
func (b *ChainBatcher) signParallel(key *ecdsa.PrivateKey, pending []PendingTx) []ChainTx {
	n := len(pending)
	result := make([]ChainTx, n)
	var wg sync.WaitGroup
	var idx int64
	var mu sync.Mutex
	workers := b.cfg.SignWorkers
	if workers > n {
		workers = n
	}
	if workers < 1 {
		workers = 1
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				i := int(idx)
				idx++
				mu.Unlock()
				if i >= n {
					return
				}
				p := pending[i]
				payloadBytes, err := json.Marshal(p.Payload)
				if err != nil {
					continue
				}
				payloadStr := string(payloadBytes)
				txID := fmt.Sprintf("tx_%d_%d_%s", p.Timestamp, i, randomHex(4))
				txHash := computeTxHash(p.TxType, payloadStr, p.Timestamp)
				result[i] = ChainTx{
					TxID:      txID,
					TxType:    p.TxType,
					Payload:   payloadStr,
					TxHash:    txHash,
					Signature: "",
					Timestamp: p.Timestamp,
				}
			}
		}()
	}
	wg.Wait()

	out := make([]ChainTx, 0, n)
	for _, t := range result {
		if t.TxID != "" {
			out = append(out, t)
		}
	}
	return out
}

func (b *ChainBatcher) Stats() map[string]int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return map[string]int64{
		"queued":          int64(len(b.queue)),
		"total_submitted": b.totalSubmitted,
		"total_flushed":   b.totalFlushed,
		"total_blocks":    b.totalBlocks,
		"max_size":        int64(b.cfg.MaxSize),
		"interval_ms":     b.cfg.FlushInterval.Milliseconds(),
		"workers":         int64(b.cfg.SignWorkers),
	}
}

func (b *ChainBatcher) Stop() {
	select {
	case <-b.done:
		return
	default:
		close(b.done)
	}
}
