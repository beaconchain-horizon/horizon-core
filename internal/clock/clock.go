package clock

import (
	"sync"
	"time"
)

var (
	lastTime int64
	mu       sync.Mutex
)

// Now() returns a strictly monotonic Unix timestamp.
// It is immune to Windows clock drift and Timejacking.
func Now() int64 {
	mu.Lock()
	defer mu.Unlock()

	current := time.Now().Unix()
	if current <= lastTime {
		lastTime++
	} else {
		lastTime = current
	}
	return lastTime
}

// Initialize sets the starting point based on the Genesis block.
func Initialize(genesisTime int64) {
	mu.Lock()
	defer mu.Unlock()
	lastTime = genesisTime
}
