package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
)

// getHardwareID returns a stable identifier for this machine.
//
// The ID is a SHA-256 hash (truncated to 16 bytes = 32 hex chars)
// of several machine-specific identifiers:
//
//   - Platform-specific extras (see hardware_id_linux.go and
//     hardware_id_windows.go)
//   - All non-loopback MAC addresses (sorted)
//   - Hostname
//   - GOOS
//
// The result is deterministic and stable across reboots.
func getHardwareID() (string, error) {
	parts := []string{}

	// Platform-specific extras
	extras, err := platformHardwareExtras()
	if err == nil {
		parts = append(parts, extras...)
	}

	// All non-loopback MAC addresses, sorted
	macs := []string{}
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			if iface.Flags&net.FlagUp == 0 {
				continue
			}
			if len(iface.HardwareAddr) == 0 {
				continue
			}
			macs = append(macs, iface.HardwareAddr.String())
		}
	}
	sort.Strings(macs)
	parts = append(parts, macs...)

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		parts = append(parts, "host:"+hostname)
	}

	// OS
	parts = append(parts, "os:"+runtime.GOOS)

	if len(parts) == 0 {
		return "", fmt.Errorf("no hardware identifiers available")
	}

	combined := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(combined))

	// 16 bytes = 32 hex chars
	return hex.EncodeToString(hash[:16]), nil
}
