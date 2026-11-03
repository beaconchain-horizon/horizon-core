package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const numShards = 16

// ═══════════════════════════════════════════════════════════════════
// Money Precision Layer
//
// Internal storage: int64 minor units
// External API: float64 (backward compatible)
//
// MoneyScale = 1  → integer Rial (bank standard)
// For cents: change to 100
// ═══════════════════════════════════════════════════════════════════

const MoneyScale = 1

func toMinor(amount float64) int64 {
	if amount >= 0 {
		return int64(amount*float64(MoneyScale) + 0.5)
	}
	return int64(amount*float64(MoneyScale) - 0.5)
}

func fromMinor(minor int64) float64 {
	return float64(minor) / float64(MoneyScale)
}

type Ledger struct {
	shards [numShards]map[string]int64
	mu     [numShards]sync.RWMutex
	ops    atomic.Int64
}

func NewLedger() *Ledger {
	l := &Ledger{}
	for i := 0; i < numShards; i++ {
		l.shards[i] = make(map[string]int64)
	}
	return l
}

func shardIdx(bankID string) int {
	h := uint32(2166136261)
	for i := 0; i < len(bankID); i++ {
		h ^= uint32(bankID[i])
		h *= 16777619
	}
	return int(h % numShards)
}

func (l *Ledger) GetBalance(bankID string) float64 {
	i := shardIdx(bankID)
	l.mu[i].RLock()
	defer l.mu[i].RUnlock()
	return fromMinor(l.shards[i][bankID])
}

func (l *Ledger) HasAccount(bankID string) bool {
	i := shardIdx(bankID)
	l.mu[i].RLock()
	defer l.mu[i].RUnlock()
	_, ok := l.shards[i][bankID]
	return ok
}

func (l *Ledger) SetBalance(bankID string, amount float64) {
	i := shardIdx(bankID)
	l.mu[i].Lock()
	l.shards[i][bankID] = toMinor(amount)
	l.mu[i].Unlock()
}

func (l *Ledger) AddBalance(bankID string, amount float64) {
	i := shardIdx(bankID)
	l.mu[i].Lock()
	l.shards[i][bankID] += toMinor(amount)
	l.mu[i].Unlock()
}

func (l *Ledger) Transfer(from, to string, amount float64) error {
	minorAmount := toMinor(amount)
	if minorAmount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if from == to {
		return fmt.Errorf("cannot transfer to self")
	}
	iFrom := shardIdx(from)
	iTo := shardIdx(to)

	if iFrom == iTo {
		l.mu[iFrom].Lock()
		defer l.mu[iFrom].Unlock()
		bal, ok := l.shards[iFrom][from]
		if !ok {
			return fmt.Errorf("sender account not found: %s", from)
		}
		if _, ok := l.shards[iFrom][to]; !ok {
			l.shards[iFrom][to] = 0
		}
		if bal < minorAmount {
			return fmt.Errorf("insufficient balance: %d < %d", bal, minorAmount)
		}
		l.shards[iFrom][from] = bal - minorAmount
		l.shards[iFrom][to] += minorAmount
	} else {
		first, second := iFrom, iTo
		if first > second {
			first, second = second, first
		}
		l.mu[first].Lock()
		l.mu[second].Lock()
		defer l.mu[first].Unlock()
		defer l.mu[second].Unlock()

		bal, ok := l.shards[iFrom][from]
		if !ok {
			return fmt.Errorf("sender account not found: %s", from)
		}
		if _, ok := l.shards[iTo][to]; !ok {
			l.shards[iTo][to] = 0
		}
		if bal < minorAmount {
			return fmt.Errorf("insufficient balance: %d < %d", bal, minorAmount)
		}
		l.shards[iFrom][from] = bal - minorAmount
		l.shards[iTo][to] += minorAmount
	}
	l.ops.Add(1)
	return nil
}

func (l *Ledger) Snapshot() map[string]float64 {
	out := make(map[string]float64)
	for i := 0; i < numShards; i++ {
		l.mu[i].RLock()
		for k, v := range l.shards[i] {
			out[k] = fromMinor(v)
		}
		l.mu[i].RUnlock()
	}
	return out
}

func (l *Ledger) LoadSnapshot(data map[string]float64) {
	for i := 0; i < numShards; i++ {
		l.mu[i].Lock()
		l.shards[i] = make(map[string]int64)
		l.mu[i].Unlock()
	}
	for k, v := range data {
		i := shardIdx(k)
		l.mu[i].Lock()
		l.shards[i][k] = toMinor(v)
		l.mu[i].Unlock()
	}
}

func (l *Ledger) Ops() int64 {
	return l.ops.Load()
}
