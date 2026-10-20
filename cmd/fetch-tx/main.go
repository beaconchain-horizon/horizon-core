package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: fetch-tx <block_number> <tx_index>")
		os.Exit(1)
	}
	blockNum := os.Args[1]
	txIndex := os.Args[2]
	url := fmt.Sprintf("http://localhost:8080/api/v1/blocks/%s", blockNum)
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching block: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	// ذخیره خروجی در یک فایل برای verifier مستقل
	filename := fmt.Sprintf("block-%s-tx-%s.json", blockNum, txIndex)
	pretty, _ := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile(filename, pretty, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved to %s\n", filepath.Join(".", filename))
}
