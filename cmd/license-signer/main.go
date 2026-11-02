package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"horizon-core/internal/crypto"
)

type signRequest struct {
	LicenseID  string `json:"license_id" binding:"required"`
	UserID     string `json:"user_id" binding:"required"`
	ProductID  string `json:"product_id" binding:"required"`
	Volume     int    `json:"volume" binding:"required"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	HardwareID string `json:"hardware_id"`
}

type signResponse struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
	SignedData string `json:"signed_data"`
	Signature  string `json:"signature"`
	PublicKeyX string `json:"public_key_x"`
	PublicKeyY string `json:"public_key_y"`
	SignerAddr string `json:"signer_address"`
}

var (
	privateKey    *ecdsa.PrivateKey
	keySourcePath string
)

// findUSBKey scans common USB mount points and the current directory
// for a PEM file that contains a HORIZON SECP256K1 PRIVATE KEY.
func findUSBKey() (string, error) {
	var roots []string

	if envPath := os.Getenv("HORIZON_ISSUER_KEY_PATH"); envPath != "" {
		roots = append(roots, envPath)
	}

	switch runtime.GOOS {
	case "windows":
		for _, letter := range "DEFGHIJKLMNOPQRSTUVWXYZ" {
			roots = append(roots, fmt.Sprintf("%c:\\", letter))
		}
	case "darwin":
		roots = append(roots, "/Volumes")
	case "linux":
		roots = append(roots, "/media", "/mnt", "/run/media")
	}

	roots = append(roots, ".", "./keys", "./keys/authority")

	var candidates []string

	for _, root := range roots {
		info, err := os.Stat(root)
		if err != nil {
			continue
		}

		if !info.IsDir() {
			// مسیر فایل مستقیم است
			candidates = append(candidates, root)
			continue
		}

		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() {
				// داخل پوشه‌های USB هم نگاه کن
				subEntries, err := os.ReadDir(filepath.Join(root, e.Name()))
				if err != nil {
					continue
				}
				for _, se := range subEntries {
					name := se.Name()
					if isKeyFile(name) {
						candidates = append(candidates, filepath.Join(root, e.Name(), name))
					}
				}
				continue
			}
			if isKeyFile(e.Name()) {
				candidates = append(candidates, filepath.Join(root, e.Name()))
			}
		}
	}

	sort.Strings(candidates)

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if !strings.Contains(string(data), "PRIVATE KEY") {
			continue
		}
		if _, err := crypto.PEMToPrivateKey(string(data)); err != nil {
			continue
		}
		return path, nil
	}

	return "", fmt.Errorf("no valid Horizon secp256k1 private key found on any USB or local path")
}

func isKeyFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".pem") ||
		strings.HasSuffix(lower, ".key") ||
		strings.Contains(lower, "authority") ||
		strings.Contains(lower, "issuer") ||
		strings.Contains(lower, "horizon")
}

func loadPrivateKey(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	key, err := crypto.PEMToPrivateKey(string(data))
	if err != nil {
		return err
	}
	privateKey = key
	keySourcePath = path
	return nil
}

func canonicalSignedData(r signRequest, expiresAt int64) string {
	return fmt.Sprintf("%s|%s|%s|%d|%s|active|%s",
		r.LicenseID,
		r.ProductID,
		r.UserID,
		r.Volume,
		r.MerkleRoot,
		time.Unix(expiresAt, 0).UTC().Format(time.RFC3339),
	)
}

func main() {
	port := os.Getenv("SIGNER_PORT")
	if port == "" {
		port = "9999"
	}

	// تلاش برای پیدا کردن کلید در استارتاپ
	if path, err := findUSBKey(); err == nil {
		if err := loadPrivateKey(path); err != nil {
			log.Printf("WARN: found key at %s but failed to load: %v", path, err)
		} else {
			log.Printf("KEY LOADED from %s", path)
		}
	} else {
		log.Printf("NO KEY at startup: %v", err)
		log.Printf("insert USB and call GET /reload-key")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: false,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "online", "service": "license-signer"})
	})

	r.GET("/status", func(c *gin.Context) {
		if privateKey == nil {
			c.JSON(http.StatusOK, gin.H{
				"key_loaded": false,
				"usb_path":   "",
				"hint":       "insert USB with the Authority PEM and call /reload-key",
			})
			return
		}
		var xb, yb [32]byte
		privateKey.PublicKey.X.FillBytes(xb[:])
		privateKey.PublicKey.Y.FillBytes(yb[:])
		c.JSON(http.StatusOK, gin.H{
			"key_loaded":     true,
			"usb_path":       keySourcePath,
			"public_key_hex": hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:]),
			"signer_address": crypto.PublicAddressFromKey(privateKey),
		})
	})

	r.POST("/reload-key", func(c *gin.Context) {
		path, err := findUSBKey()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"key_loaded": false, "error": err.Error()})
			return
		}
		if err := loadPrivateKey(path); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"key_loaded": false, "error": err.Error()})
			return
		}
		var xb, yb [32]byte
		privateKey.PublicKey.X.FillBytes(xb[:])
		privateKey.PublicKey.Y.FillBytes(yb[:])
		c.JSON(http.StatusOK, gin.H{
			"key_loaded":     true,
			"usb_path":       keySourcePath,
			"public_key_hex": hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:]),
			"signer_address": crypto.PublicAddressFromKey(privateKey),
		})
	})

	r.POST("/sign", func(c *gin.Context) {
		if privateKey == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "no key loaded. Insert USB and call /reload-key",
			})
			return
		}

		var req signRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().UTC()
		issuedAt := now.Unix()

		var expiresAt int64
		if req.Duration > 0 {
			expiresAt = now.Add(time.Duration(req.Duration) * 24 * time.Hour).Unix()
		} else {
			expiresAt = now.Add(365 * 24 * time.Hour).Unix()
		}

		data := canonicalSignedData(req, expiresAt)
		sig, err := crypto.SignDataRS([]byte(data), privateKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sign: " + err.Error()})
			return
		}

		var xb, yb [32]byte
		privateKey.PublicKey.X.FillBytes(xb[:])
		privateKey.PublicKey.Y.FillBytes(yb[:])

		c.JSON(http.StatusOK, signResponse{
			LicenseID:  req.LicenseID,
			UserID:     req.UserID,
			ProductID:  req.ProductID,
			Volume:     req.Volume,
			Duration:   req.Duration,
			MerkleRoot: req.MerkleRoot,
			IssuedAt:   issuedAt,
			ExpiresAt:  expiresAt,
			HardwareID: req.HardwareID,
			SignedData: data,
			Signature:  hex.EncodeToString(sig),
			PublicKeyX: hex.EncodeToString(xb[:]),
			PublicKeyY: hex.EncodeToString(yb[:]),
			SignerAddr: crypto.PublicAddressFromKey(privateKey),
		})
	})

	log.Printf("license-signer listening on 127.0.0.1:%s", port)
	log.Printf("insert USB with Authority PEM; then POST /reload-key")
	if err := r.Run("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
