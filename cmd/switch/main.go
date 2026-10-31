package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8091"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"online","service":"switch"}`))
	})

	log.Printf("Switch running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
