//go:build linux

package main

import (
	"os"
	"strings"
)

// platformHardwareExtras returns Linux-specific identifiers.
//
// It reads:
//   - /sys/class/dmi/id/product_uuid (motherboard UUID)
//   - /proc/cpuinfo (serial, model name)
//
// If any file is not readable, it is skipped without error.
func platformHardwareExtras() ([]string, error) {
	extras := []string{}

	// Motherboard UUID
	if data, err := os.ReadFile(
		"/sys/class/dmi/id/product_uuid",
	); err == nil {
		v := strings.TrimSpace(string(data))
		if v != "" {
			extras = append(extras, "mb:"+v)
		}
	}

	// CPU info
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "serial") ||
				strings.HasPrefix(line, "model name") {
				extras = append(
					extras,
					"cpu:"+strings.TrimSpace(line),
				)
			}
		}
	}

	return extras, nil
}
