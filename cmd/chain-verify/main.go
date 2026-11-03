package main

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"horizon-core/internal/chain"
)

const genesisHash = "8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603"

func main() {
	dbPath := os.Getenv("SWITCH_DB")
	if dbPath == "" {
		dbPath = "./horizon-dev.db"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Println("open db:", err)
		os.Exit(1)
	}

	var blocks []chain.Block
	if err := db.Order("number ASC").Find(&blocks).Error; err != nil {
		fmt.Println("load blocks:", err)
		os.Exit(2)
	}
	if len(blocks) == 0 {
		fmt.Println("✗ no blocks")
		os.Exit(3)
	}

	genesis := blocks[0]
	if genesis.Number != 0 {
		fmt.Printf("✗ first block number=%d (want 0)\n", genesis.Number)
		os.Exit(4)
	}
	if genesis.BlockHash != genesisHash {
		fmt.Printf("✗ genesis hash mismatch\n  got  %s\n  want %s\n", genesis.BlockHash, genesisHash)
		os.Exit(5)
	}
	fmt.Printf("✓ genesis #0 ok: %s\n", genesis.BlockHash)

	prev := genesis
	for _, b := range blocks[1:] {
		if b.Number != prev.Number+1 {
			fmt.Printf("✗ gap: #%d after #%d\n", b.Number, prev.Number)
			os.Exit(6)
		}
		if b.PreviousHash != prev.BlockHash {
			fmt.Printf("✗ broken link at #%d\n  prev_hash=%s\n  want     =%s\n", b.Number, b.PreviousHash, prev.BlockHash)
			os.Exit(7)
		}
		sigStatus := "unsigned"
		if b.AuthoritySig != "" {
			sigStatus = fmt.Sprintf("signed(%d bytes)", len(b.AuthoritySig)/2)
		}
		fmt.Printf("✓ #%d ok  hash=%s.. prev=%s.. tx=%d %s\n",
			b.Number, b.BlockHash[:12], b.PreviousHash[:12], b.TxCount, sigStatus)
		prev = b
	}

	fmt.Printf("\n✓✓✓ CHAIN INTACT: %d blocks, head #%d\n", len(blocks), prev.Number)
}
