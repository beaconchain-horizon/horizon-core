//go:build windows

package main

import (
	"os/exec"
	"strings"
)

// platformHardwareExtras returns Windows-specific identifiers.
//
// It uses WMIC to read:
//   - Motherboard UUID (csproduct uuid)
//   - CPU ProcessorId (cpu get ProcessorId)
//
// If WMIC fails, it returns an empty slice (not an error),
// so hardware ID can still be computed from MAC + hostname.
func platformHardwareExtras() ([]string, error) {
	extras := []string{}

	// Motherboard UUID
	if out, err := exec.Command(
		"wmic", "csproduct", "get", "uuid",
	).Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.EqualFold(line, "UUID") {
				continue
			}
			extras = append(extras, "mb:"+line)
			break
		}
	}

	// CPU ProcessorId
	if out, err := exec.Command(
		"wmic", "cpu", "get", "ProcessorId",
	).Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.EqualFold(line, "ProcessorId") {
				continue
			}
			extras = append(extras, "cpu:"+line)
			break
		}
	}

	return extras, nil
}
