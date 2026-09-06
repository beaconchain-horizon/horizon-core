package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "online"})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"service": "Horizon Switch"})
	})

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Horizon Switch running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
