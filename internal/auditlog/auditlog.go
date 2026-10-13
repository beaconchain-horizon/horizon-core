package auditlog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	Seq       int64  `json:"seq"`
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Payload   string `json:"payload"`
	PrevHash  string `json:"prev_hash"`
	EntryHash string `json:"entry_hash"`
}

type Log struct {
	mu      sync.Mutex
	seq     int64
	prev    string
	entries []Entry
}

func New() *Log { return &Log{prev: ""} }

func (l *Log) Append(action, actor, payload string) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	e := Entry{
		Seq:       l.seq + 1,
		Timestamp: time.Now().Unix(),
		Action:    action,
		Actor:     actor,
		Payload:   payload,
		PrevHash:  l.prev,
	}
	raw, _ := json.Marshal([]any{e.Seq, e.Timestamp, e.Action, e.Actor, e.Payload, e.PrevHash})
	h := sha256.Sum256(raw)
	e.EntryHash = hex.EncodeToString(h[:])
	l.entries = append(l.entries, e)
	l.seq = e.Seq
	l.prev = e.EntryHash
	return e
}

func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

func (l *Log) Verify() int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	prev := ""
	for _, e := range l.entries {
		if e.PrevHash != prev {
			return e.Seq
		}
		raw, _ := json.Marshal([]any{e.Seq, e.Timestamp, e.Action, e.Actor, e.Payload, e.PrevHash})
		h := sha256.Sum256(raw)
		if hex.EncodeToString(h[:]) != e.EntryHash {
			return e.Seq
		}
		prev = e.EntryHash
	}
	return 0
}

func (l *Log) String() string {
	if len(l.prev) < 12 {
		return fmt.Sprintf("audit entries=%d", l.seq)
	}
	return fmt.Sprintf("audit entries=%d tip=%s", l.seq, l.prev[:12])
}
