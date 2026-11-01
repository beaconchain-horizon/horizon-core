package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"horizon-core/internal/chain"
	"horizon-core/internal/crypto"
	"horizon-core/internal/db"
)

// ============================================================================
// TYPES
// ============================================================================

// LicenseReq is the exact JSON structure the offline dashboard signs.
// Note: It does NOT include merkle_type, because the dashboard doesn't sign it.
type LicenseReq struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int64  `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
	Signature  string `json:"signature"`
	// We also accept these for block creation (not signed by dashboard)
	MerkleType string `json:"merkle_type"`
}

// ============================================================================
// HELPERS
// ============================================================================

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func createGenesisIfNeeded() {
	var count int64
	db.DB.Model(&db.Block{}).Count(&count)
	if count == 0 {
		ts := time.Now().UTC()
		zeroHash := "0000000000000000000000000000000000000000000000000000000000000000"
		emptyRoot := hex.EncodeToString(make([]byte, 32))
		h := sha256.New()
		h.Write([]byte(zeroHash))
		h.Write([]byte(emptyRoot))
		h.Write([]byte(ts.Format(time.RFC3339Nano)))
		fmt.Fprintf(h, "|%d", 0)
		genesisHash := hex.EncodeToString(h.Sum(nil))
		genesis := db.Block{
			Number:       0,
			PreviousHash: zeroHash,
			MerkleRoot:   emptyRoot,
			MerkleType:   "genesis",
			TxCount:      0,
			Timestamp:    ts,
			BlockHash:    genesisHash,
		}
		db.DB.Create(&genesis)
		log.Printf("🌱 Genesis Block created: #0 hash=%s", genesisHash[:16])
	}
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	db.Init()
	createGenesisIfNeeded()

	// 1. Chain Info
	http.HandleFunc("/api/v1/chain/info", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		var latest db.Block
		db.DB.Order("number desc").First(&latest)
		var txCount int64
		db.DB.Model(&db.Transaction{}).Count(&txCount)
		json.NewEncoder(w).Encode(map[string]any{
			"chain_id": "horizon-core-v1.4",
			"height": latest.Number,
			"block_count": latest.Number + 1,
			"tx_total": txCount,
			"latest_hash": latest.BlockHash,
			"latest_root": latest.MerkleRoot,
		})
	})

	// 2. Blocks
	http.HandleFunc("/api/v1/blocks", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		var blocks []db.Block
		db.DB.Order("number desc").Limit(20).Find(&blocks)
		json.NewEncoder(w).Encode(blocks)
	})

	// 3. Agents
	http.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		var agents []db.Agent
		db.DB.Find(&agents)
		json.NewEncoder(w).Encode(agents)
	})

	// 4. License Registration (The core of this phase)
	http.HandleFunc("/api/v1/license/register", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LicenseReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 1) Verify the signature using the exact JSON structure (without merkle_type)
		verifyPayload := struct {
			LicenseID  string `json:"license_id"`
			UserID     string `json:"user_id"`
			ProductID  string `json:"product_id"`
			Volume     int64  `json:"volume"`
			Duration   int    `json:"duration"`
			MerkleRoot string `json:"merkle_root"`
			IssuedAt   int64  `json:"issued_at"`
			ExpiresAt  int64  `json:"expires_at"`
			HardwareID string `json:"hardware_id"`
		}{
			LicenseID:  req.LicenseID,
			UserID:     req.UserID,
			ProductID:  req.ProductID,
			Volume:     req.Volume,
			Duration:   req.Duration,
			MerkleRoot: req.MerkleRoot,
			IssuedAt:   req.IssuedAt,
			ExpiresAt:  req.ExpiresAt,
			HardwareID: req.HardwareID,
		}

		payloadBytes, _ := json.Marshal(verifyPayload)
		pubKey := os.Getenv("HORIZON_PUBLIC_KEY")
		if pubKey == "" {
			// Using the key from the dashboard for now
			pubKey = "0x04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
		}

		valid, err := crypto.VerifyOfflineLicense(pubKey, payloadBytes, req.Signature)
		if err != nil || !valid {
			log.Printf("❌ Signature verification failed for %s: %v", req.LicenseID, err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid signature"})
			return
		}

		// 2) If valid, create the transaction and block
		merkleType := req.MerkleType
		if merkleType == "" {
			merkleType = "binary" // default
		}

		// For a single transaction, the Merkle Root is just the hash of the transaction payload.
		// The `req.MerkleRoot` provided by the dashboard is actually this hash.
		// We compute the block's Merkle Root based on the request type.
		txPayload := payloadBytes
		leafHash := sha256.Sum256(txPayload)
		
		// For a single-tx block, the root is the leaf hash.
		// But if we want to test the chain logic, we can use the chain package:
		merkleRoot := chain.ComputeMerkleRoot([][]byte{leafHash[:]}, merkleType)
		
		// Check if the merkleRoot matches the one provided by the dashboard (if merkle_type is binary)
		if merkleType == "binary" {
			// The dashboard's merkle_root is just the SHA256 of the payload.
			if req.MerkleRoot != hex.EncodeToString(leafHash[:]) {
				log.Printf("⚠️ Warning: Provided MerkleRoot does not match computed leaf hash.")
			}
		}

		// 3) Save transaction
		tx := db.Transaction{
			TxID:       fmt.Sprintf("tx_%d_%s", time.Now().UnixNano(), req.LicenseID),
			BlockNum:   0, // Will be set after block creation
			LicenseID:  req.LicenseID,
			UserID:     req.UserID,
			ProductID:  req.ProductID,
			Volume:     req.Volume,
			Duration:   req.Duration,
			Signature:  req.Signature,
			PublicKey:  pubKey,
			CreatedAt:  time.Now(),
		}
		db.DB.Create(&tx)

		// 4) Create block
		var latest db.Block
		db.DB.Order("number desc").First(&latest)
		newBlockNum := latest.Number + 1
		
		block := db.Block{
			Number:       newBlockNum,
			PreviousHash: latest.BlockHash,
			MerkleRoot:   merkleRoot,
			MerkleType:   merkleType,
			TxCount:      1,
			Timestamp:    time.Now().UTC(),
			BlockHash:    "", // computed below
		}

		// Compute block hash
		h := sha256.New()
		h.Write([]byte(block.PreviousHash))
		h.Write([]byte(block.MerkleRoot))
		h.Write([]byte(block.Timestamp.Format(time.RFC3339Nano)))
		fmt.Fprintf(h, "|%d", block.TxCount)
		block.BlockHash = hex.EncodeToString(h.Sum(nil))

		db.DB.Create(&block)

		// Update transaction with block number
		db.DB.Model(&tx).Update("block_num", block.Number)

		// 5) Return success
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "success",
			"tx_id":       tx.TxID,
			"block_num":   block.Number,
			"merkle_root": merkleRoot,
			"merkle_type": merkleType,
			"block_hash":  block.BlockHash,
		})
	})

	log.Println("🚀 Backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
