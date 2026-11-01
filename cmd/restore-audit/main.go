package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"horizon-core/internal/db"
	"gorm.io/gorm"
)

func merkleRoot(payloads []string) string {
	if len(payloads) == 0 {
		return hex.EncodeToString(make([]byte, 32))
	}
	leaves := make([][32]byte, len(payloads))
	for i, p := range payloads {
		leaves[i] = sha256.Sum256([]byte(p))
	}
	for len(leaves) > 1 {
		var next [][32]byte
		for i := 0; i < len(leaves); i += 2 {
			var pair [64]byte
			copy(pair[:32], leaves[i][:])
			if i+1 < len(leaves) {
				copy(pair[32:], leaves[i+1][:])
			} else {
				copy(pair[32:], leaves[i][:])
			}
			next = append(next, sha256.Sum256(pair[:]))
		}
		leaves = next
	}
	return hex.EncodeToString(leaves[0][:])
}

func computeBlockHash(prevHash, root string, ts time.Time, txCount int) string {
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(root))
	h.Write([]byte(ts.Format(time.RFC3339Nano)))
	fmt.Fprintf(h, "|%d", txCount)
	return hex.EncodeToString(h.Sum(nil))
}

func must(tx *gorm.DB) {
	if tx.Error != nil {
		log.Fatalf("DB error: %v", tx.Error)
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func inspect() {
	var blocks []db.Block
	must(db.DB.Order("number asc").Find(&blocks))
	fmt.Printf("\n=== BLOCKS (%d) ===\n", len(blocks))
	for _, b := range blocks {
		fmt.Printf("  #%-3d hash=%s  prev=%s  tx=%-3d  type=%-10s  ts=%s\n",
			b.Number, trunc(b.BlockHash, 16), trunc(b.PreviousHash, 16),
			b.TxCount, b.MerkleType, b.Timestamp.Format("2006-01-02 15:04:05"))
	}

	var txs []db.Transaction
	must(db.DB.Order("block_num asc, created_at asc").Find(&txs))
	fmt.Printf("\n=== TRANSACTIONS (%d) ===\n", len(txs))
	for _, t := range txs {
		fmt.Printf("  %s  block=%-3d  lic=%-20s  user=%-15s  ts=%s\n",
			t.TxID, t.BlockNum, t.LicenseID, t.UserID,
			t.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}

func rebuild() {
	fmt.Println("\n[1] Wiping blocks and transactions (agents/licenses preserved)...")
	must(db.DB.Exec("DELETE FROM transactions"))
	must(db.DB.Exec("DELETE FROM blocks"))

	// ─── Genesis ───
	ts0 := time.Date(2026, 10, 16, 15, 0, 0, 0, time.UTC)
	genesisRoot := hex.EncodeToString(make([]byte, 32))
	genesisHash := computeBlockHash(
		"0000000000000000000000000000000000000000000000000000000000000000",
		genesisRoot, ts0, 0,
	)
	genesis := db.Block{
		Number:       0,
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
		MerkleRoot:   genesisRoot,
		MerkleType:   "genesis",
		TxCount:      0,
		Timestamp:    ts0,
		BlockHash:    genesisHash,
	}
	must(db.DB.Create(&genesis))
	fmt.Printf("  #0  GENESIS   hash=%s\n", genesisHash)

	// ─── License data from audit record ───
	type licDef struct {
		LicenseID string
		UserID    string
		ProductID string
		Volume    int64
		Duration  int
		Ts        time.Time
	}

	lics := []licDef{
		{"HORIZON-GOLD-001", "bank-markazi", "horizon-core", 1000000, 365, time.Date(2026, 10, 16, 15, 29, 22, 561545900, time.UTC)},
		{"HORIZON-GOLD-002", "bank-markazi", "horizon-core", 1000000, 365, time.Date(2026, 10, 16, 15, 30, 0, 0, time.UTC)},
		{"HORIZON-GOLD-004", "bank-markazi", "horizon-core", 1000000, 365, time.Date(2026, 10, 16, 15, 32, 0, 0, time.UTC)},
	}

	prevHash := genesisHash
	for i, lic := range lics {
		// Canonical JSON payload
		payload := fmt.Sprintf(
			`{"license_id":"%s","user_id":"%s","product_id":"%s","volume":%d,"duration":%d}`,
			lic.LicenseID, lic.UserID, lic.ProductID, lic.Volume, lic.Duration,
		)

		root := merkleRoot([]string{payload})
		num := uint64(i + 1)
		hash := computeBlockHash(prevHash, root, lic.Ts, 1)

		block := db.Block{
			Number:       num,
			PreviousHash: prevHash,
			MerkleRoot:   root,
			MerkleType:   "binary",
			TxCount:      1,
			Timestamp:    lic.Ts,
			BlockHash:    hash,
		}
		must(db.DB.Create(&block))

		shortID := lic.LicenseID[len("HORIZON-GOLD-"):]
		txID := fmt.Sprintf("tx_audit_b%d_%s", num, shortID)

		tx := db.Transaction{
			TxID:       txID,
			BlockNum:   num,
			LicenseID:  lic.LicenseID,
			UserID:     lic.UserID,
			ProductID:  lic.ProductID,
			Volume:     lic.Volume,
			Duration:   lic.Duration,
			Signature:  "audit_record_rebuilt_" + shortID,
			PublicKey:  "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214",
			CreatedAt:  lic.Ts,
		}
		must(db.DB.Create(&tx))

		fmt.Printf("  #%-3d BLOCK     hash=%s  prev=%s  lic=%s\n",
			num, trunc(hash, 16), trunc(prevHash, 16), lic.LicenseID)
		prevHash = hash
	}

	fmt.Println("\n✅ Chain rebuilt from audit record.")
	fmt.Printf("   Height: %d  |  Last hash: %s\n", len(lics), prevHash)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: restore-audit <inspect|rebuild>")
		os.Exit(1)
	}
	db.Init()
	switch os.Args[1] {
	case "inspect":
		inspect()
	case "rebuild":
		rebuild()
		inspect()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
