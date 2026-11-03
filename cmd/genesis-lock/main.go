package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	k256 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	k256ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	_ "github.com/glebarez/sqlite"
)

const (
	chainID    = "horizon-core-v1.4"
	genesisHx  = "8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603"
	pubkeyHx   = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
	compactSig = "afac005b7f7565fe69ca9d58dabdf0a3914fce0854ee05f214d788e9e478b02f7c82a05b35d1456b605a128e4074ed9d3c4e55932cc7b368516c0681d236face"
)

func verifyCompact(hashB, sigB []byte, pub *k256.PublicKey) bool {
	if len(sigB) != 64 { return false }
	var r, s k256.ModNScalar
	r.SetByteSlice(sigB[0:32])
	s.SetByteSlice(sigB[32:64])
	sig := k256ecdsa.NewSignature(&r, &s)
	return sig.Verify(hashB, pub)
}

func main() {
	pubB, _ := hex.DecodeString(pubkeyHx)
	pub, err := k256.ParsePubKey(pubB)
	if err != nil { fmt.Println("pub parse:", err); os.Exit(1) }
	hashB, _ := hex.DecodeString(genesisHx)
	sigB, _ := hex.DecodeString(compactSig)

	if !verifyCompact(hashB, sigB, pub) {
		fmt.Println("✗ compact sig does NOT verify — refusing to write DB")
		os.Exit(3)
	}
	fmt.Println("✓ compact sig verified against genesis hash + pubkey")

	// بکاپ
	src := "horizon-dev.db"
	backup := "horizon-dev.db.genesis-lock-" + time.Now().Format("20060102-150405") + ".bak"
	data, err := os.ReadFile(src)
	if err != nil { fmt.Println("read db:", err); os.Exit(4) }
	if err := os.WriteFile(backup, data, 0600); err != nil { fmt.Println("backup:", err); os.Exit(4) }
	fmt.Println("✓ backup:", backup)

	db, err := sql.Open("sqlite", src)
	if err != nil { fmt.Println("open:", err); os.Exit(5) }
	defer db.Close()

	var ts string
	db.QueryRow("SELECT timestamp FROM blocks WHERE number=0").Scan(&ts)
	fmt.Println("  genesis timestamp:", ts)

	tx, err := db.Begin()
	if err != nil { fmt.Println("begin tx:", err); os.Exit(6) }
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE blocks SET authority_sig=? WHERE number=0", compactSig); err != nil {
		fmt.Println("update blocks:", err); os.Exit(7)
	}
	fmt.Println("✓ blocks.authority_sig")

	if _, err := tx.Exec("UPDATE geneses SET signature=? WHERE id=1", compactSig); err != nil {
		fmt.Println("update geneses:", err); os.Exit(8)
	}
	fmt.Println("✓ geneses.signature")

	var cnt int
	tx.QueryRow("SELECT COUNT(*) FROM merkle_roots WHERE block_number=0").Scan(&cnt)
	if cnt == 0 {
		if _, err := tx.Exec("INSERT INTO merkle_roots (root_hash, block_number, tree_type, signature, timestamp) VALUES (?,0,'genesis',?,datetime('now'))", genesisHx, compactSig); err != nil {
			fmt.Println("insert merkle_roots:", err); os.Exit(9)
		}
		fmt.Println("✓ merkle_roots inserted")
	} else {
		if _, err := tx.Exec("UPDATE merkle_roots SET signature=?, root_hash=? WHERE block_number=0", compactSig, genesisHx); err != nil {
			fmt.Println("update merkle_roots:", err); os.Exit(9)
		}
		fmt.Println("✓ merkle_roots updated")
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("commit:", err); os.Exit(10)
	}

	fmt.Println("\n=== VERIFY ===")
	var bs, gs, ms string
	db.QueryRow("SELECT authority_sig FROM blocks WHERE number=0").Scan(&bs)
	db.QueryRow("SELECT signature FROM geneses WHERE id=1").Scan(&gs)
	db.QueryRow("SELECT signature FROM merkle_roots WHERE block_number=0").Scan(&ms)
	ok := bs == compactSig && gs == compactSig && ms == compactSig
	fmt.Println("  blocks:       ", bs == compactSig)
	fmt.Println("  geneses:      ", gs == compactSig)
	fmt.Println("  merkle_roots: ", ms == compactSig)
	if !ok { fmt.Println("✗ mismatch — restore from backup"); os.Exit(11) }

	// GENESIS.json
	gen := map[string]any{
		"chain_id":          chainID,
		"genesis_number":    0,
		"genesis_hash":      genesisHx,
		"merkle_root":       genesisHx,
		"merkle_type":       "genesis",
		"previous_hash":     "0000000000000000000000000000000000000000000000000000000000000000",
		"tx_count":          0,
		"timestamp":         ts,
		"authority_pub_key": pubkeyHx,
		"authority_sig":     compactSig,
		"locked_at":         time.Now().UTC().Format(time.RFC3339),
	}
	jb, _ := json.MarshalIndent(gen, "", "  ")
	if err := os.WriteFile("GENESIS.json", jb, 0644); err != nil {
		fmt.Println("write GENESIS.json:", err); os.Exit(12)
	}
	fmt.Println("\n✓ GENESIS.json نوشته شد")
	fmt.Println("✓✓✓ جنسیس برای همیشه قفل شد")
}
