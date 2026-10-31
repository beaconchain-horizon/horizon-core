package db

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	// Using pure Go sqlite driver (glebarez) to avoid CGO dependency on Windows
	DB, err = gorm.Open(sqlite.Open("horizon-core.db?_pragma=journal_mode(WAL)"), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB Init Failed: %v", err)
	}

	err = DB.AutoMigrate(&Block{}, &Transaction{}, &Agent{}, &SystemLaw{})
	if err != nil {
		log.Fatalf("DB Migration Failed: %v", err)
	}

	// Insert immutable law
	DB.FirstOrCreate(&SystemLaw{
		LawID:       "LAW-001",
		Title:       "Zero-Storage Private Key",
		Status:      "MANDATORY",
		Description: "Private key and seed MUST NOT be stored anywhere in the database. Licenses are generated offline only.",
	})

	DB.FirstOrCreate(&SystemLaw{
		LawID:       "LAW-002",
		Title:       "Hardware-Bound License",
		Status:      "MANDATORY",
		Description: "A license is bound to a specific Hardware ID (e.g., Bank Sepah Branch 102). It cannot be used anywhere else.",
	})

	log.Println("✅ Database initialized with pure Go driver.")
}
