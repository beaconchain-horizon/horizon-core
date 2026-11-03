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

	latest, err := chain.LatestBlock(db)
	if err != nil || latest == nil {
		fmt.Println("latest block:", err)
		os.Exit(2)
	}
	fmt.Printf("head: block #%d hash=%s\n", latest.Number, latest.BlockHash)

	if latest.Number != 0 {
		fmt.Println("✗ head is not genesis; already has blocks. aborting.")
		os.Exit(3)
	}
	if latest.BlockHash != genesisHash {
		fmt.Printf("✗ head hash != locked genesis\n  got  %s\n  want %s\n", latest.BlockHash, genesisHash)
		os.Exit(4)
	}

	tx := chain.PendingTx{
		TxID:      "block1-smoke-001",
		TxType:    "smoke",
		Payload:   `{"note":"first block after locked genesis","seq":1}`,
		Signature: "",
		PublicKey: "",
	}

	blk, err := chain.AppendBlock(db, []chain.PendingTx{tx})
	if err != nil {
		fmt.Println("append:", err)
		os.Exit(5)
	}

	if blk.PreviousHash != genesisHash {
		fmt.Printf("✗✗✗ link mismatch: got %s\n", blk.PreviousHash)
		os.Exit(6)
	}

	fmt.Printf("✓✓✓ block #%d appended\n  hash      = %s\n  prev_hash = %s\n  merkle    = %s\n",
		blk.Number, blk.BlockHash, blk.PreviousHash, blk.MerkleRoot)
}
