package main

import (
	"strings"
	"sync"
)

var (
	cachedHardwareID string
	hardwareIDOnce   sync.Once
)

// getHardwareIDOrEmpty returns the hardware ID, cached after first call.
//
// On the first call, it computes the hardware ID by calling
// the platform-specific computeHardwareID() function.
// Subsequent calls return the cached value.
//
// This is critical for performance: computing the hardware ID
// spawns a system process (wmic, dmidecode) which takes ~200-300ms.
// Without caching, every license save request would be slow.
func getHardwareIDOrEmpty() string {
	hardwareIDOnce.Do(func() {
		cachedHardwareID = strings.TrimSpace(computeHardwareID())
	})
	return cachedHardwareID
}
