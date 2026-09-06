package main

import (
	"bytes"
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

func main() {
	// مسیر سلامت (Health Check)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "online"})
	})

	// مسیر تأیید لایسنس
	http.HandleFunc("/api/v1/license/check", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			LicenseKey string `json:"license_key"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(map[string]bool{"valid": true})
	})

	// مسیر رمزنگاری AES
	http.HandleFunc("/api/v1/toolbox/encrypt", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Key       string `json:"key"`
			Plaintext string `json:"plaintext"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		key, _ := hex.DecodeString(req.Key)
		res, err := AESEncrypt(key, []byte(req.Plaintext))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"ciphertext": res})
	})

	http.HandleFunc("/api/v1/toolbox/decrypt", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Key        string `json:"key"`
			Ciphertext string `json:"ciphertext"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		key, _ := hex.DecodeString(req.Key)
		res, err := AESDecrypt(key, req.Ciphertext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"plaintext": string(res)})
	})

	port := os.Getenv("SWITCH_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Horizon Switch running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
