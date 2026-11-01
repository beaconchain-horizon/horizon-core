package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"horizon-core/internal/db"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OldBlock struct {
	Number       uint64    `gorm:"column:number"`
	PreviousHash string    `gorm:"column:previous_hash"`
	MerkleRoot   string    `gorm:"column:merkle_root"`
	MerkleType   string    `gorm:"column:merkle_type"`
	Timestamp    time.Time `gorm:"column:timestamp"`
	TxCount      uint64    `gorm:"column:tx_count"`
	BlockHash    string    `gorm:"column:block_hash"`
	AuthoritySig string    `gorm:"column:authority_sig"`
}

type OldTx struct {
	ID          uint64    `gorm:"column:id"`
	TxID        string    `gorm:"column:tx_id"`
	BlockNumber uint64    `gorm:"column:block_number"`
	TxIndex     int       `gorm:"column:tx_index"`
	TxType      string    `gorm:"column:tx_type"`
	Payload     string    `gorm:"column:payload"`
	Signature   string    `gorm:"column:signature"`
	PublicKey   string    `gorm:"column:public_key"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

type RegistryAgent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FilePath    string `json:"file_path"`
	Category    string `json:"category"`
}

func main() {
	oldPath := "data/horizon-chain.db.bak-20261014-193759"

	oldDB, err := gorm.Open(sqlite.Open(oldPath), &gorm.Config{})
	if err != nil { log.Fatalf("open old: %v", err) }
	var oldBlocks []OldBlock
	oldDB.Table("blocks").Order("number asc").Find(&oldBlocks)
	var oldTxs []OldTx
	oldDB.Table("transactions").Order("block_number asc, tx_index asc").Find(&oldTxs)
	fmt.Printf("   Read old: %d blocks, %d txs\n", len(oldBlocks), len(oldTxs))

	data, err := os.ReadFile("data/agents_registry.json")
	if err != nil { log.Fatalf("read agents: %v", err) }
	var rawAgents []RegistryAgent
	if err := json.Unmarshal(data, &rawAgents); err != nil { log.Fatalf("parse agents: %v", err) }
	fmt.Printf("   Read agents JSON: %d entries\n", len(rawAgents))

	for _, f := range []string{"horizon-core.db", "horizon-core.db-wal", "horizon-core.db-shm"} {
		os.Remove(f)
	}
	fmt.Println("   [init] db.Init()")
	db.Init()

	var preBlocks int64
	db.DB.Model(&db.Block{}).Count(&preBlocks)
	fmt.Printf("   After db.Init(): %d blocks in DB\n", preBlocks)

	// ─── Agents ───
	fmt.Println("   [insert] agents")
	agentCount, skipped := 0, 0
	skip := map[string]bool{"PULL_REQUEST_TEMPLATE": true, "README": true, "LICENSE": true}
	for _, a := range rawAgents {
		if a.ID == "" { skipped++; continue }
		if skip[strings.ToUpper(a.ID)] { skipped++; continue }
		if strings.Contains(a.FilePath, ".github") { skipped++; continue }
		if strings.Contains(strings.ToUpper(a.FilePath), "HISTORY") { skipped++; continue }
		
		role := a.Category
		if role == "" { role = "general" }
		agent := db.Agent{
			AgentID: a.ID, UserID: a.Name, Role: role,
			LicenseID: "LIC-" + strings.ToUpper(a.ID),
			Status: "Active", LastSeen: time.Now(),
		}
		if err := db.DB.Create(&agent).Error; err != nil {
			log.Printf("   ⚠️ agent %s: %v", a.ID, err)
		} else { agentCount++ }
	}
	fmt.Printf("   ✅ agents inserted: %d (skipped %d)\n", agentCount, skipped)

	// ─── Blocks ───
	fmt.Println("   [insert] blocks (OnConflict DoNothing)")
	blockCount, blockSkip := 0, 0
	for _, b := range oldBlocks {
		blk := db.Block{
			Number: b.Number, PreviousHash: b.PreviousHash,
			MerkleRoot: b.MerkleRoot, MerkleType: b.MerkleType,
			TxCount: b.TxCount, Timestamp: b.Timestamp,
			BlockHash: b.BlockHash, AuthoritySig: b.AuthoritySig,
		}
		res := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&blk)
		if res.Error != nil {
			log.Printf("   ⚠️ block %d: %v", b.Number, res.Error)
		} else if res.RowsAffected > 0 { blockCount++ } else { blockSkip++ }
	}
	fmt.Printf("   ✅ blocks: inserted=%d, skipped=%d (total=%d)\n", blockCount, blockSkip, len(oldBlocks))

	// ─── Transactions ───
	fmt.Println("   [insert] transactions")
	migrated, txSkip := 0, 0
	for _, t := range oldTxs {
		parts := strings.Split(t.Payload, "|")
		licenseID, productID, userID, hardwareID := "", "", "", ""
		var volume int64
		if len(parts) >= 1 { licenseID = parts[0] }
		if len(parts) >= 2 { productID = parts[1] }
		if len(parts) >= 3 { userID = parts[2] }
		if len(parts) >= 4 { volume, _ = strconv.ParseInt(parts[3], 10, 64) }
		if len(parts) >= 5 { hardwareID = parts[4] }

		txID := t.TxID
		var exists int64
		db.DB.Raw("SELECT COUNT(*) FROM transactions WHERE tx_id = ?", txID).Scan(&exists)
		if exists > 0 { txID = fmt.Sprintf("%s_%d", t.TxID, t.ID) }

		tx := db.Transaction{
			TxID: txID, BlockNum: t.BlockNumber,
			LicenseID: licenseID, UserID: userID, ProductID: productID, HardwareID: hardwareID,
			Volume: volume, Duration: 365,
			Signature: t.Signature, PublicKey: t.PublicKey, CreatedAt: t.CreatedAt,
		}
		res := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&tx)
		if res.Error != nil {
			log.Printf("   ⚠️ tx %s: %v", txID, res.Error)
		} else if res.RowsAffected > 0 { migrated++ } else { txSkip++ }
	}
	fmt.Printf("   ✅ txs: inserted=%d, skipped=%d (total=%d)\n", migrated, txSkip, len(oldTxs))

	var fb, ft, fa, fl int64
	db.DB.Model(&db.Block{}).Count(&fb)
	db.DB.Model(&db.Transaction{}).Count(&ft)
	db.DB.Model(&db.Agent{}).Count(&fa)
	db.DB.Model(&db.SystemLaw{}).Count(&fl)
	fmt.Printf("\n   ═══ FINAL ═══\n")
	fmt.Printf("     blocks:       %d\n", fb)
	fmt.Printf("     transactions: %d\n", ft)
	fmt.Printf("     agents:       %d\n", fa)
	fmt.Printf("     system_laws:  %d\n", fl)
}
