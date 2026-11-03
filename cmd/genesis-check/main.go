package main

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/sqlite"
)

func dump(db *sql.DB, label, q string) {
	fmt.Printf("\n=== %s ===\n", label)
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	fmt.Println("columns:", cols)
	n := 0
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		for i, c := range cols {
			fmt.Printf("  %s = %v\n", c, vals[i])
		}
		fmt.Println("---")
		n++
	}
	fmt.Println("rows:", n)
}

func main() {
	db, err := sql.Open("sqlite", "horizon-dev.db")
	if err != nil {
		fmt.Println("open err:", err)
		return
	}
	defer db.Close()

	dump(db, "GENESIS BLOCK (blocks.number=0)", "SELECT * FROM blocks WHERE number=0")
	dump(db, "GENESES DECLARATION", "SELECT * FROM geneses")
	dump(db, "MERKLE ROOTS @ GENESIS", "SELECT * FROM merkle_roots WHERE block_number=0")
	dump(db, "BLOCKS SUMMARY", "SELECT COUNT(*) AS cnt, MIN(number) AS min_n, MAX(number) AS max_n FROM blocks")
	dump(db, "GENESES COUNT", "SELECT COUNT(*) AS cnt FROM geneses")
}
