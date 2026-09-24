//go:build linux
// +build linux

package main

import (
	"os"
	"strings"
)

// computeHardwareID reads the system UUID on Linux.
//
// This function is called ONLY ONCE per process lifetime
// (see getHardwareIDOrEmpty in hardware_id.go).
func computeHardwareID() string {
	data, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
