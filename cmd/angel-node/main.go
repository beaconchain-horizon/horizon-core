package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
)

// محاسبه پویای هش قانون
func lawHash() string {
	rules := "v=1.0\n" +
		"merkle|9001|SHA256(left||right)|CRITICAL\n" +
		"hash|9002|SHA256(header)|CRITICAL\n" +
		"ecdsa|9003|secp256k1.Verify|CRITICAL\n" +
		"type|9004|type in binary or ternary|CRITICAL\n" +
		"volume|9005|used < volume|IMPORTANT\n" +
		"time|9006|issued <= now < expires|IMPORTANT\n" +
		"user|9007|user_id match|IMPORTANT\n"
	h := sha256.Sum256([]byte(rules))
	return hex.EncodeToString(h[:])
}

func main() {
	role := os.Getenv("NODE_ROLE")
	port := os.Getenv("NODE_PORT")
	if role == "" || port == "" {
		log.Fatal("FAIL-CLOSED: NODE_ROLE and NODE_PORT required")
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"role":"%s", "port":"%s", "status":"online", "law_hash":"%s"}`, role, port, lawHash()[:16])
	})

	log.Printf("👼 Node [%s] running on :%s", role, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
