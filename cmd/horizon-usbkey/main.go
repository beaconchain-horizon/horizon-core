// horizon-usbkey finds a Horizon secp256k1 private key on any mounted
// USB drive (or the current directory) and prints the corresponding
// 128-hex public key. Used to seed HORIZON_LICENSE_PUBLIC_KEY for Switch.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"horizon-core/internal/crypto"
)

// isPEMContent checks if the file looks like a Horizon secp256k1 PEM.
func isPEMContent(data []byte) bool {
	s := string(data)
	return strings.Contains(s, "PRIVATE KEY") &&
		(strings.Contains(s, "HORIZON SECP256K1") || strings.Contains(s, "BEGIN"))
}

// listFiles returns all regular files in a directory (one level deep).
func listFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			sub, err := os.ReadDir(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, s := range sub {
				if !s.IsDir() {
					out = append(out, filepath.Join(dir, e.Name(), s.Name()))
				}
			}
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out
}

func main() {
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	var roots []string

	// 1) explicit path via env var (highest priority)
	if p := os.Getenv("HORIZON_ISSUER_KEY_PATH"); p != "" {
		roots = append(roots, p)
	}

	// 2) USB drives (D through Z on Windows, /Volumes on macOS)
	switch runtime.GOOS {
	case "windows":
		for _, c := range "DEFGHIJKLMNOPQRSTUVWXYZ" {
			roots = append(roots, fmt.Sprintf("%c:\\", c))
		}
	case "darwin":
		roots = append(roots, "/Volumes")
	case "linux":
		roots = append(roots, "/media", "/mnt", "/run/media")
	}

	// 3) local fallbacks (last)
	roots = append(roots, ".", "./keys", "./test-keys")

	var allFiles []string
	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			allFiles = append(allFiles, root)
			continue
		}
		allFiles = append(allFiles, listFiles(root)...)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "scanned %d files in %d roots\n", len(allFiles), len(roots))
	}

	// deduplicate and sort
	seen := map[string]bool{}
	var unique []string
	for _, f := range allFiles {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	// try each file
	for _, path := range unique {
		base := strings.ToLower(filepath.Base(path))
		// skip obvious non-keys
		if strings.HasSuffix(base, ".go") ||
			strings.HasSuffix(base, ".md") ||
			strings.HasSuffix(base, ".json") ||
			strings.HasSuffix(base, ".yml") ||
			strings.HasSuffix(base, ".yaml") ||
			strings.HasSuffix(base, ".exe") ||
			strings.HasSuffix(base, ".db") {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// try to parse PEM directly
		priv, err := crypto.PEMToPrivateKey(string(data))
		if err != nil {
			// try hex-32 (raw 32-byte hex)
			trimmed := strings.TrimSpace(string(data))
			trimmed = strings.TrimPrefix(trimmed, "0x")
			if len(trimmed) == 64 {
				if p2, err2 := crypto.Secp256k1PrivateKeyFromHex(trimmed); err2 == nil {
					priv = p2
				}
			}
			if priv == nil {
				continue
			}
		}

		var xb, yb [32]byte
		priv.PublicKey.X.FillBytes(xb[:])
		priv.PublicKey.Y.FillBytes(yb[:])
		hexKey := hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:])

		fmt.Fprintf(os.Stderr, "found: %s\n", path)
		fmt.Println(hexKey)
		return
	}

	log.Fatal("no valid Horizon secp256k1 key found in any candidate")
}
