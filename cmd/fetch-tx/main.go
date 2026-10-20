package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type BlockResponse struct {
	Number       int64  `json:"number"`
	MerkleRoot   string `json:"merkle_root"`
	Transactions []struct {
		TxID      string `json:"tx_id"`
		Signature string `json:"signature"`
		Payload   string `json:"payload"`
	} `json:"transactions"`
}

type VerifierInput struct {
	TxID         string   `json:"tx_id"`
	Signature    string   `json:"signature"`
	Payload      string   `json:"payload"`
	PublicKey    string   `json:"public_key"`
	MerkleLeaf   string   `json:"merkle_leaf"`
	MerkleProof  []string `json:"merkle_proof"`
	ExpectedRoot string   `json:"expected_root"`
	Index        int      `json:"index"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fetch-tx <block_number> <tx_index>")
		os.Exit(1)
	}
	blockNum := os.Args[1]
	txIdx, _ := strconv.Atoi(os.Args[2])

	url := fmt.Sprintf("http://localhost:8080/api/v1/blocks/%s", blockNum)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching block: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var block BlockResponse
	if err := json.Unmarshal(body, &block); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	if txIdx < 0 || txIdx >= len(block.Transactions) {
		fmt.Printf("Transaction index out of bounds: %d\n", txIdx)
		os.Exit(1)
	}

	tx := block.Transactions[txIdx]

	// محاسبه مرکل لیف برای این تراکنش
	h := sha256.Sum256([]byte(tx.TxID + tx.Payload + tx.Signature))
	merkleLeaf := hex.EncodeToString(h[:])

	pubKey := os.Getenv("HORIZON_LICENSE_PUBLIC_KEY")
	if pubKey == "" {
		pubKey = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
	}

	out := VerifierInput{
		TxID:         tx.TxID,
		Signature:    tx.Signature,
		Payload:      tx.Payload,
		PublicKey:    pubKey,
		MerkleLeaf:   merkleLeaf,
		MerkleProof:  []string{}, // در این پیاده‌سازی API فعلی، پروف از بک‌اند دریافت نمی‌شود
		ExpectedRoot: block.MerkleRoot,
		Index:        txIdx,
	}

	filename := fmt.Sprintf("block-%s-tx-%d-proof.json", blockNum, txIdx)
	pretty, _ := json.MarshalIndent(out, "", "  ")
	os.WriteFile(filename, pretty, 0644)
	fmt.Printf("Saved verifier input to %s\n", filename)
}
