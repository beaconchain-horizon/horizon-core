package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	pendingMu     sync.Mutex
	pendingQueue  []PendingTx
	pendingNotify chan struct{}
	ConsensusGate func(txs []PendingTx) (bool, string)
)

// Enqueue adds a transaction to the pending queue for the next block.
// Returns an error if the queue is full.
func Enqueue(tx PendingTx) error {
	pendingMu.Lock()
	defer pendingMu.Unlock()

	if len(pendingQueue) >= 50000 {
		return fmt.Errorf("pending queue full")
	}
	if err := persistPending(tx); err != nil {
		return fmt.Errorf("persist pending: %w", err)
	}
	pendingQueue = append(pendingQueue, tx)

	if pendingNotify != nil {
		select {
		case pendingNotify <- struct{}{}:
		default:
		}
	}
	return nil
}

// PendingCount returns the number of transactions waiting for the next block.
func PendingCount() int {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	return len(pendingQueue)
}

func drainPending() []PendingTx {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	out := pendingQueue
	pendingQueue = nil
	return out
}

// StartBlockLoop launches a background goroutine that:
//   - flushes a new block every `interval` seconds, OR
//   - flushes immediately when pending count reaches `maxSize`
func StartBlockLoop(db *gorm.DB, interval time.Duration, maxSize int) {
	if err := InitPendingStore(db); err != nil {
		log.Printf("[chain] pending store init failed: %v", err)
	}
	if pendingNotify == nil {
		pendingNotify = make(chan struct{}, 1)
	}

	// Ticker-based flush
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			flush(db, maxSize)
		}
	}()

	// Event-based flush (when threshold reached)
	go func() {
		for range pendingNotify {
			pendingMu.Lock()
			n := len(pendingQueue)
			pendingMu.Unlock()
			if n >= maxSize {
				flush(db, maxSize)
			}
		}
	}()

	log.Printf("[chain] block loop started: interval=%s maxSize=%d", interval, maxSize)
}

func flush(db *gorm.DB, maxSize int) {
	txs := drainPending()
	if len(txs) == 0 {
		return
	}

	// Split into chunks of maxSize.
	for len(txs) > 0 {
		n := maxSize
		if n > len(txs) {
			n = len(txs)
		}
		chunk := txs[:n]
		txs = txs[n:]

		// ─── Consensus Gate ───
		if ConsensusGate != nil {
			passed, info := ConsensusGate(chunk)
			if !passed {
				log.Printf("[chain] consensus gate rejected: %s", info)
				pendingMu.Lock()
				pendingQueue = append(chunk, pendingQueue...)
				pendingMu.Unlock()
				return
			}
		}

		block, err := AppendBlock(db, chunk)
		if err != nil {
			log.Printf("[chain] append block failed: %v", err)
			// Put the txs back at the head of the queue.
			pendingMu.Lock()
			pendingQueue = append(chunk, pendingQueue...)
			pendingMu.Unlock()
			return
		}
		log.Printf("[chain] block %d created: %d tx, root=%s…",
			block.Number, block.TxCount, block.MerkleRoot[:12])
		ids := make([]string, len(chunk))
		for i, tx := range chunk {
			ids[i] = tx.TxID
		}
		deletePendingByIDs(ids)
	}
}

// NewTxID generates a deterministic tx id from payload + timestamp.
// Caller is responsible for uniqueness (Backend provides user_id + nonce).
func NewTxID(payload string, ts time.Time) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d", payload, ts.UnixNano())))
	return "tx_" + hex.EncodeToString(h[:16])
}
