package main

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type LicenseRecord struct {
	ID string `gorm:"primaryKey"`
}

func (LicenseRecord) TableName() string { return "license_records" }

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: db-fix <db-path> <license-id>")
		os.Exit(2)
	}
	db, err := gorm.Open(sqlite.Open(os.Args[1]), &gorm.Config{})
	if err != nil { fmt.Println("open:", err); os.Exit(1) }
	var before []LicenseRecord
	db.Find(&before)
	fmt.Printf("before: %d records\n", len(before))
	for _, r := range before { fmt.Printf("  - %s\n", r.ID) }
	res := db.Where("id = ?", os.Args[2]).Delete(&LicenseRecord{})
	fmt.Printf("delete: rows=%d err=%v\n", res.RowsAffected, res.Error)
	var after []LicenseRecord
	db.Find(&after)
	fmt.Printf("after: %d records\n", len(after))
}
