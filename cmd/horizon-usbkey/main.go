// horizon-usbkey finds a Horizon secp256k1 private key on any mounted
// USB drive (or the current directory) and prints the corresponding
// 128-hex public key. Prefers .pem/.key and horizon/authority-named files.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"horizon-core/internal/crypto"
)

type candidate struct {
	path     string
	priority int // lower = tried first
}

func scoreFile(name string) int {
	l := strings.ToLower(name)
	switch {
	case strings.HasSuffix(l, ".pem"):
		return 0
	case strings.HasSuffix(l, ".key"):
		return 1
	case strings.Contains(l, "horizon") || strings.Contains(l, "authority") || strings.Contains(l, "issuer"):
		return 2
	case strings.HasSuffix(l, ".txt"):
		return 5
	default:
		return 10
	}
}

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

func parseKeyFile(path string) (string, bool) {
	base := strings.ToLower(filepath.Base(path))
	if strings.HasSuffix(base, ".go") ||
		strings.HasSuffix(base, ".md") ||
		strings.HasSuffix(base, ".json") ||
		strings.HasSuffix(base, ".yml") ||
		strings.HasSuffix(base, ".yaml") ||
		strings.HasSuffix(base, ".exe") ||
		strings.HasSuffix(base, ".db") {
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	// try PEM first
	priv, err := crypto.PEMToPrivateKey(string(data))
	if err != nil {
		// try hex-32
		trimmed := strings.TrimSpace(string(data))
		trimmed = strings.TrimPrefix(trimmed, "0x")
		if len(trimmed) == 64 {
			if p2, err2 := crypto.Secp256k1PrivateKeyFromHex(trimmed); err2 == nil {
				priv = p2
			}
		}
		if priv == nil {
			return "", false
		}
	}

	var xb, yb [32]byte
	priv.PublicKey.X.FillBytes(xb[:])
	priv.PublicKey.Y.FillBytes(yb[:])
	return hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:]), true
}

func main() {
	verbose := flag.Bool("v", false, "verbose output")
	all := flag.Bool("all", false, "print all candidates found")
	flag.Parse()

	// If HORIZON_ISSUER_KEY_PATH is set, use ONLY that file.
	if p := os.Getenv("HORIZON_ISSUER_KEY_PATH"); p != "" {
		if h, ok := parseKeyFile(p); ok {
			fmt.Fprintf(os.Stderr, "found (env): %s\n", p)
			fmt.Println(h)
			return
		}
		log.Fatalf("HORIZON_ISSUER_KEY_PATH=%s did not contain a valid key", p)
	}

	var roots []string
	if p := os.Getenv("HORIZON_ISSUER_KEY_PATH"); p != "" {
		roots = append(roots, p)
	}
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

	// dedup
	seen := map[string]bool{}
	var files []candidate
	for _, f := range allFiles {
		if seen[f] {
			continue
		}
		seen[f] = true
		files = append(files, candidate{path: f, priority: scoreFile(filepath.Base(f))})
	}

	// sort by priority, then path
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].priority != files[j].priority {
			return files[i].priority < files[j].priority
		}
		return files[i].path < files[j].path
	})

	if *verbose {
		fmt.Fprintf(os.Stderr, "scanned %d files, sorted by priority\n", len(files))
	}

	var found []struct{ path, hex string }
	for _, f := range files {
		if h, ok := parseKeyFile(f.path); ok {
			if *all {
				fmt.Fprintf(os.Stderr, "found: %s (priority=%d)\n", f.path, f.priority)
				fmt.Println(h)
				found = append(found, struct{ path, hex string }{f.path, h})
				continue
			}
			fmt.Fprintf(os.Stderr, "found: %s (priority=%d)\n", f.path, f.priority)
			fmt.Println(h)
			return
		}
	}

	if *all && len(found) > 0 {
		return
	}

	log.Fatal("no valid Horizon secp256k1 key found")
}
