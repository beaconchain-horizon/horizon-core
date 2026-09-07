package main

import (
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "horizon-switch",
		})
	})

	r.POST("/verify", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"verified": true,
			"message":  "License verified successfully",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Switch starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"service": "Horizon Switch"})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "online"})
	})

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Horizon Switch running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
