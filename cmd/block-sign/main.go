package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	s "github.com/decred/dcrd/dcrec/secp256k1/v4"
	e "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const expectedPub = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"

type Block struct {
	Number       int64
	BlockHash    string `gorm:"column:block_hash"`
	AuthoritySig string `gorm:"column:authority_sig"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: block-sign <block_number>")
		os.Exit(1)
	}
	n, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil || n <= 0 {
		fmt.Fprintln(os.Stderr, "refuse: block number must be > 0")
		os.Exit(1)
	}

	dbPath := os.Getenv("SWITCH_DB")
	if dbPath == "" {
		dbPath = "./horizon-dev.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		os.Exit(2)
	}

	var blk Block
	if err := db.Table("blocks").Where("number = ?", n).First(&blk).Error; err != nil {
		fmt.Fprintln(os.Stderr, "load block:", err)
		os.Exit(3)
	}
	if blk.AuthoritySig != "" {
		fmt.Fprintf(os.Stderr, "✗ block #%d already has authority_sig\n", blk.Number)
		os.Exit(4)
	}

	hashB, err := hex.DecodeString(blk.BlockHash)
	if err != nil || len(hashB) != 32 {
		fmt.Fprintf(os.Stderr, "✗ bad block_hash len=%d err=%v\n", len(hashB), err)
		os.Exit(5)
	}

	fmt.Fprintf(os.Stderr, "block #%d hash = %s\n", blk.Number, blk.BlockHash)

	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		fmt.Fprintln(os.Stderr, "✗ no key on stdin")
		os.Exit(6)
	}
	raw := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(sc.Text(), "0x"), "0X"))
	kb, err := hex.DecodeString(raw)
	if err != nil || len(kb) != 32 {
		fmt.Fprintln(os.Stderr, "✗ bad key")
		os.Exit(7)
	}

	priv := s.PrivKeyFromBytes(kb)
	pubHex := hex.EncodeToString(priv.PubKey().SerializeUncompressed())
	if pubHex != expectedPub {
		fmt.Fprintln(os.Stderr, "✗ pubkey mismatch — refusing to sign")
		os.Exit(8)
	}
	fmt.Fprintln(os.Stderr, "✓ pubkey matches locked authority")

	sig := e.Sign(priv, hashB)
	var compact [64]byte
	rVal := sig.R(); rVal.PutBytesUnchecked(compact[0:32])
	sVal := sig.S(); sVal.PutBytesUnchecked(compact[32:64])

	var rChk, sChk s.ModNScalar
	rChk.SetByteSlice(compact[0:32])
	sChk.SetByteSlice(compact[32:64])
	sigChk := e.NewSignature(&rChk, &sChk)
	if !sigChk.Verify(hashB, priv.PubKey()) {
		fmt.Fprintln(os.Stderr, "✗ local verify of compact sig failed")
		os.Exit(10)
	}
	sigHex := hex.EncodeToString(compact[:])
	fmt.Fprintln(os.Stderr, "✓ compact sig (64B R||S) verifies against block_hash + pubkey")

	res := db.Table("blocks").Where("number = ?", n).Update("authority_sig", sigHex)
	if res.Error != nil || res.RowsAffected != 1 {
		fmt.Fprintln(os.Stderr, "✗ update failed:", res.Error, "rows:", res.RowsAffected)
		os.Exit(11)
	}

	fmt.Printf("✓ block #%d signed: %s\n", n, sigHex)
}
