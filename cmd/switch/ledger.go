package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const numShards = 16

type Ledger struct {
	shards [numShards]map[string]float64
	mu     [numShards]sync.RWMutex
	ops    atomic.Int64
}

func NewLedger() *Ledger {
	l := &Ledger{}
	for i := 0; i < numShards; i++ {
		l.shards[i] = make(map[string]float64)
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
	return l.shards[i][bankID]
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
	l.shards[i][bankID] = amount
	l.mu[i].Unlock()
}

func (l *Ledger) AddBalance(bankID string, amount float64) {
	i := shardIdx(bankID)
	l.mu[i].Lock()
	l.shards[i][bankID] += amount
	l.mu[i].Unlock()
}

func (l *Ledger) Transfer(from, to string, amount float64) error {
	if amount <= 0 {
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
			return fmt.Errorf("receiver account not found: %s", to)
		}
		if bal < amount {
			return fmt.Errorf("insufficient balance: %.2f < %.2f", bal, amount)
		}
		l.shards[iFrom][from] = bal - amount
		l.shards[iFrom][to] += amount
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
			return fmt.Errorf("receiver account not found: %s", to)
		}
		if bal < amount {
			return fmt.Errorf("insufficient balance: %.2f < %.2f", bal, amount)
		}
		l.shards[iFrom][from] = bal - amount
		l.shards[iTo][to] += amount
	}
	l.ops.Add(1)
	return nil
}

func (l *Ledger) Snapshot() map[string]float64 {
	out := make(map[string]float64)
	for i := 0; i < numShards; i++ {
		l.mu[i].RLock()
		for k, v := range l.shards[i] {
			out[k] = v
		}
		l.mu[i].RUnlock()
	}
	return out
}

func (l *Ledger) LoadSnapshot(data map[string]float64) {
	for i := 0; i < numShards; i++ {
		l.mu[i].Lock()
		l.shards[i] = make(map[string]float64)
		l.mu[i].Unlock()
	}
	for k, v := range data {
		i := shardIdx(k)
		l.mu[i].Lock()
		l.shards[i][k] = v
		l.mu[i].Unlock()
	}
}

func (l *Ledger) Ops() int64 {
	return l.ops.Load()
}
