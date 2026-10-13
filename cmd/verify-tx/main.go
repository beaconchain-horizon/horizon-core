package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"horizon-core/internal/crypto"
)

type input struct {
	Payload      string   `json:"payload"`
	Signature    string   `json:"signature"`
	PublicKey    string   `json:"public_key"`
	MerkleLeaf   string   `json:"merkle_leaf"`
	MerkleProof  []string `json:"merkle_proof"`
	ExpectedRoot string   `json:"expected_root"`
}

var exitCode int

func ok(s string)             { fmt.Printf("  [PASS] %s\n", s) }
func bad(s string, a ...any)  { exitCode = 1; fmt.Printf("  [FAIL] %s\n", fmt.Sprintf(s, a...)) }
func fail(s string, a ...any) { fmt.Fprintf(os.Stderr, s+"\n", a...); os.Exit(2) }
func mustHex(s string) []byte { b, _ := hex.DecodeString(strings.TrimSpace(s)); return b }

func walkProof(leafHex string, proof []string) string {
	cur, _ := hex.DecodeString(leafHex)
	for _, s := range proof {
		sib, _ := hex.DecodeString(s)
		pair := append(append([]byte{}, cur...), sib...)
		h := sha256.Sum256(pair)
		cur = h[:]
	}
	return hex.EncodeToString(cur)
}

func main() {
	f := flag.String("f", "", "input JSON file (default: stdin)")
	flag.Parse()

	var r *os.File
	var err error
	if *f == "" {
		r = os.Stdin
	} else {
		r, err = os.Open(*f)
		if err != nil {
			fail("open input: %v", err)
		}
		defer r.Close()
	}

	var in input
	if err := json.NewDecoder(r).Decode(&in); err != nil {
		fail("decode input: %v", err)
	}

	fmt.Println("── verify-tx ──────────────────────────────────")

	h := sha256.Sum256([]byte(in.Payload))
	wantLeaf := hex.EncodeToString(h[:])
	if wantLeaf == in.MerkleLeaf {
		ok("leaf = SHA256(payload)")
	} else {
		bad("leaf mismatch\n    computed: %s\n    expected: %s", wantLeaf, in.MerkleLeaf)
	}

	pub, err := crypto.Secp256k1PublicKeyFromHex(in.PublicKey)
	if err != nil {
		bad("public key: %v", err)
	} else {
		ok("public key parsed (secp256k1)")
		digest := sha256.Sum256([]byte(in.Payload))
		if crypto.VerifyHashRS(pub, digest[:], mustHex(in.Signature)) {
			ok("ECDSA signature verifies against public key")
		} else {
			bad("ECDSA signature verification failed")
		}
	}

	if len(in.MerkleProof) == 0 {
		bad("merkle proof empty")
	} else if root := walkProof(in.MerkleLeaf, in.MerkleProof); root == in.ExpectedRoot {
		ok("merkle proof walks to root")
	} else {
		bad("merkle proof root mismatch\n    computed: %s\n    expected: %s", root, in.ExpectedRoot)
	}

	fmt.Println("───────────────────────────────────────────────")
	if exitCode == 0 {
		fmt.Println("RESULT: VALID")
	} else {
		fmt.Println("RESULT: INVALID")
	}
	os.Exit(exitCode)
}
