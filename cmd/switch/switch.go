package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Switch OK"))
	})

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Switch running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
