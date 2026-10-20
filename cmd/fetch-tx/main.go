package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type BlockResponse struct {
	Number       int64 `json:"number"`
	Transactions []struct {
		TxID      string `json:"tx_id"`
		Signature string `json:"signature"`
		Payload   string `json:"payload"`
	} `json:"transactions"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fetch-tx <block_number> <tx_index>")
		os.Exit(1)
	}
	blockNum := os.Args[1]
	txIdx, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid tx_index")
		os.Exit(1)
	}

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
	filename := fmt.Sprintf("block-%s-tx-%d.json", blockNum, txIdx)

	pretty, _ := json.MarshalIndent(tx, "", "  ")
	if err := os.WriteFile(filename, pretty, 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved transaction %s to %s\n", tx.TxID, filename)
}
