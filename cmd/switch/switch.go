package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"service": "horizon-switch",
	})
}

func main() {
	http.HandleFunc("/health", healthHandler)
	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Horizon Switch running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
