package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"horizon-core/internal/db"
)

func main() {
	// Initialize Database
	db.Init()

	// Default Windows port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "online", "version": "v2"})
	})

	http.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		var agents []db.Agent
		if err := db.DB.Find(&agents).Error; err != nil {
			http.Error(w, "Failed to fetch agents", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	})

	http.HandleFunc("/api/v1/laws", func(w http.ResponseWriter, r *http.Request) {
		var laws []db.SystemLaw
		if err := db.DB.Find(&laws).Error; err != nil {
			http.Error(w, "Failed to fetch laws", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(laws)
	})

	log.Printf("🚀 Backend API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
