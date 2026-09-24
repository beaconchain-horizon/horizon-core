//go:build windows
// +build windows

package main

import (
	"os/exec"
	"strings"
)

// computeHardwareID reads the system UUID on Windows via wmic.
//
// This function is called ONLY ONCE per process lifetime
// (see getHardwareIDOrEmpty in hardware_id.go).
func computeHardwareID() string {
	cmd := exec.Command("wmic", "csproduct", "get", "uuid")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return ""
	}
	return strings.TrimSpace(lines[1])
}
