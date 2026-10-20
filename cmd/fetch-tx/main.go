package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fetch-tx <block_number>")
		os.Exit(1)
	}
	blockNum := os.Args[1]
	url := fmt.Sprintf("http://localhost:8080/api/v1/blocks/%s", blockNum)
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching block: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}
	
	pretty, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(pretty))
}
