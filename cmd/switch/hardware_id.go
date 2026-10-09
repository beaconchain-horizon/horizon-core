package main

import (
	"strings"
	"os"
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
		// Container/cloud deployments may expose a stable hardware identity
		// through deployment secrets/configuration. If set, this explicit
		// identity takes precedence over host-specific discovery.
		override := strings.TrimSpace(os.Getenv("HORIZON_HARDWARE_ID"))
		if override != "" {
			cachedHardwareID = override
			return
		}
		cachedHardwareID = strings.TrimSpace(computeHardwareID())
	})
	return cachedHardwareID
}
