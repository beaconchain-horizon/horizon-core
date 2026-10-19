package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Block struct {
	Number       int64       `json:"number"`
	MerkleRoot   string      `json:"merkle_root"`
	Transactions []Tx        `json:"transactions"`
}

type Tx struct {
	TxIndex     int      `json:"tx_index"`
	Payload     string   `json:"payload"`
	PublicKey   string   `json:"public_key"`
	MerkleLeaf  string   `json:"merkle_leaf"`
	MerkleProof []string `json:"merkle_proof"`
}

type Input struct {
	Payload      string   `json:"payload"`
	Signature    string   `json:"signature"`
	MerkleLeaf   string   `json:"merkle_leaf"`
	MerkleProof  []string `json:"merkle_proof"`
	ExpectedRoot string   `json:"expected_root"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: fetch-tx <base-url> <block> [block...]")
	}
	base := os.Args[1]
	dir := ".verify-tmp"
	os.MkdirAll(dir, 0755)

	for _, b := range os.Args[2:] {
		url := fmt.Sprintf("%s/api/v1/blocks/%s", base, b)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("block %s: %v\n", b, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var blk Block
		if err := json.Unmarshal(body, &blk); err != nil {
			continue
		}
		if len(blk.Transactions) == 0 {
			fmt.Printf("block %s: no txs\n", b)
		}
		tx := blk.Transactions[0]
		in := Input{
			Payload:      tx.Payload,
			Signature:    tx.Signature,
			PublicKey:    tx.PublicKey,
			MerkleLeaf:   tx.MerkleLeaf,
			MerkleProof:  tx.MerkleProof,
			Index:        tx.TxIndex,
		}
		out, _ := json.MarshalIndent(in, "", "  ")
		os.WriteFile(p, out, 0644)
		fmt.Printf("block %s: index=%d count=%d -> %s\n", b, tx.TxIndex, len(blk.Transactions), p)
}
