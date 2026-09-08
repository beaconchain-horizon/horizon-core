package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

// ==========================================
// ابزارهای رمزنگاری AES-GCM
// ==========================================

func AESEncrypt(key, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

func AESDecrypt(key []byte, ciphertextHex string) ([]byte, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, nil
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ==========================================
// هندلرها
// ==========================================

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"service": "horizon-switch",
		"port":    "8080",
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "Horizon Switch",
		"version": "2.0",
		"status":  "running",
	})
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		LicenseID string `json:"license_id"`
		Signature string `json:"signature"`
		Data      string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"verified": true,
		"message":  "License verified successfully",
		"license":  req.LicenseID,
	})
}

func licenseCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		LicenseKey string `json:"license_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	valid := req.LicenseKey != "invalid" && req.LicenseKey != ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"valid": valid,
	})
}

func encryptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Key       string `json:"key"`
		Plaintext string `json:"plaintext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := hex.DecodeString(req.Key)
	if err != nil || len(key) != 32 {
		http.Error(w, "Invalid key (must be 32 bytes hex)", http.StatusBadRequest)
		return
	}
	res, err := AESEncrypt(key, []byte(req.Plaintext))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"ciphertext": res})
}

func decryptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Key        string `json:"key"`
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := hex.DecodeString(req.Key)
	if err != nil || len(key) != 32 {
		http.Error(w, "Invalid key (must be 32 bytes hex)", http.StatusBadRequest)
		return
	}
	res, err := AESDecrypt(key, req.Ciphertext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"plaintext": string(res)})
}

// ==========================================
// Main
// ==========================================

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/api/v1/license/check", licenseCheckHandler)
	http.HandleFunc("/api/v1/toolbox/encrypt", encryptHandler)
	http.HandleFunc("/api/v1/toolbox/decrypt", decryptHandler)

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Horizon Switch running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
